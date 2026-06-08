package app

import (
	"context"
	"fmt"
	"g-kirti/chapteriser/internal/domain"
	"g-kirti/chapteriser/internal/media"
	"g-kirti/chapteriser/internal/metadata"
	"g-kirti/chapteriser/internal/transcribe"
	"g-kirti/chapteriser/internal/workspace"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type chunkJob struct {
	index    int
	startSec int
	durSec   int
}

type chunkResult struct {
	index    int
	chapters []domain.Chapter
	err      error
}

func chooseWorkerCount(requested int, chunkCount int) int {
	if chunkCount < 1 {
		return 1
	}
	if requested > 0 {
		if requested > chunkCount {
			return chunkCount
		}
		return requested
	}

	// when workers requested is 0 (auto)
	workers := runtime.GOMAXPROCS(0)
	if workers > 2 {
		workers--
	}
	if workers < 1 {
		workers = 1
	}
	if workers > chunkCount {
		workers = chunkCount
	}
	return workers
}

func Run(cfg Config) error {
	// show time elapsed
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		log.Printf("[summary] Time taken to complete: %v\n", elapsed)
	}()

	// interrupt signal context
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// get resolved config
	resolved, err := cfg.Resolve()
	if err != nil {
		return err
	}

	// probe metadata
	chunkSeconds := resolved.ChunkMinutes * 60
	probe, err := media.ProbeInput(resolved.InputAbs)
	if err != nil {
		return err
	}
	if !probe.HasAudio {
		return fmt.Errorf("Input file has no audio stream: %s", resolved.InputAbs)
	}

	totalSeconds := probe.DurationSeconds
	if totalSeconds < 1 {
		return fmt.Errorf("Input duration is 0 seconds")
	}
	chunkCount := totalSeconds / chunkSeconds
	if totalSeconds%chunkSeconds != 0 {
		chunkCount++
	}
	if chunkCount < 1 {
		return fmt.Errorf("No audio chunks were created")
	}

	log.Println("[probe] Metadata checked")

	// skip detection if user provides metadata
	if resolved.SplitAudio && resolved.MetadataInputAbs != "" {
		log.Printf("[mux] Splitting audio file using provided metadata: %s", resolved.MetadataInputAbs)
		if err := media.SplitAudioFile(media.SplitOptions{
			InputPath:              resolved.InputAbs,
			MetadataPath:           resolved.MetadataInputAbs,
			OutputDir:              resolved.OutputAbs,
			AttachedPicStreamIndex: probe.AttachedPicStreamIndex,
			IncludeCover:           probe.HasAttachedPic,
		}); err != nil {
			return err
		}
		log.Printf("[mux] Done. Split audio written to %s", resolved.OutputAbs)
		return nil
	} else if resolved.MetadataInputAbs != "" {
		log.Printf("[mux] Creating chaptered m4b at %s (bitrate - %s)", resolved.OutputAbs, resolved.Bitrate)
		if err := media.CreateM4B(media.M4BOptions{
			InputPath:              resolved.InputAbs,
			MetadataPath:           resolved.MetadataInputAbs,
			OutputPath:             resolved.OutputAbs,
			Bitrate:                resolved.Bitrate,
			AttachedPicStreamIndex: probe.AttachedPicStreamIndex,
			IncludeCover:           probe.HasAttachedPic,
		}); err != nil {
			return err
		}
		log.Printf("[mux] Done. Muxxed audio written to %s", resolved.OutputAbs)
		return nil
	}

	// initiate temporary workspace to store metadata and chunks
	ws, err := workspace.New(resolved.TempDirAbs, "chapteriser-*", resolved.KeepTemp)
	if err != nil {
		return fmt.Errorf("Failed to create temp dir: %w", err)
	} else {
		log.Printf("[run] Creating temp dir: %s", ws.Dir)
	}
	if resolved.KeepTemp {
		log.Printf("[run] Keeping temp dir: %s", ws.Dir)
	}
	defer ws.Cleanup()

	// create temporary files
	metadataPath := ws.Path("ffmetadata.txt")

	// initialise Vosk model
	model, err := transcribe.NewSharedModel(resolved.ModelPath)
	if err != nil {
		return err
	} else {
		log.Printf("[run] Vosk model at: %s created", resolved.ModelPath)
	}
	defer model.Free()

	// ready engine
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	jobs := make(chan chunkJob, chunkCount)
	results := make(chan chunkResult, chunkCount)

	for i := range chunkCount {
		startSec := i * chunkSeconds
		durSec := chunkSeconds
		remaining := totalSeconds - startSec
		if remaining < durSec {
			durSec = remaining
		}
		jobs <- chunkJob{
			index:    i,
			startSec: startSec,
			durSec:   durSec,
		}
	}
	close(jobs)

	workerCount := chooseWorkerCount(resolved.Workers, chunkCount)
	log.Printf("[transcribe] Using %d worker(s) for %d chunk(s) at %d minute chunk size", workerCount, chunkCount, resolved.ChunkMinutes)

	// start engine
	var wg sync.WaitGroup
	var finalResultMu sync.Mutex
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}

				chapterList, err := transcribe.TranscribeChunkRange(ctx, model, resolved.InputAbs, job.startSec, job.durSec, &finalResultMu)
				if err != nil {
					results <- chunkResult{
						index: job.index,
						err:   fmt.Errorf("Failed to transcribe chunk %d (start=%ds,duration=%ds): %w", job.index, job.startSec, job.durSec, err),
					}
					return
				}

				offset := float64(job.index * chunkSeconds)
				for i := range chapterList {
					chapterList[i].Start += offset
					chapterList[i].End += offset
				}

				results <- chunkResult{
					index:    job.index,
					chapters: chapterList,
				}
			}
		}()
	}

	// close results channel after all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// re-order whilst reading from results channel
	ordered := make([][]domain.Chapter, chunkCount)
	var firstErr error
	for result := range results {
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
				cancel()
			}
			continue
		}
		ordered[result.index] = result.chapters
	}
	if firstErr != nil {
		return firstErr
	}

	// ~~~ finish ~~~ //
	var final []domain.Chapter
	for i := range ordered {
		final = append(final, ordered[i]...)
	}

	if err := metadata.WriteChaptersToFile(final, metadataPath, probe.Tags); err != nil {
		return fmt.Errorf("Failed to write chapter metadata: %w", err)
	} else {
		log.Println("[metadata] Finished writing ffmetadata.txt")
	}

	// split audio into chapters and output to directory
	if resolved.SplitAudio {
		log.Printf("[mux] Splitting audio file")
		if err := media.SplitAudioFile(media.SplitOptions{
			InputPath:              resolved.InputAbs,
			MetadataPath:           metadataPath,
			OutputDir:              resolved.OutputAbs,
			AttachedPicStreamIndex: probe.AttachedPicStreamIndex,
			IncludeCover:           probe.HasAttachedPic,
		}); err != nil {
			return err
		}
		log.Printf("[mux] Done. Split audio written to %s", resolved.OutputAbs)
		return nil
	}

	// skip final conversion to .m4b if -skip-mux is set
	if resolved.SkipMux && resolved.KeepTemp {
		log.Printf("[mux] Skipped m4b creation")
		log.Printf("[metadata] Metadata in %s", ws.Dir)
		return nil
	} else if resolved.SkipMux {
		log.Printf("[mux] Skipped m4b creation")
		return nil
	}

	// create M4B
	log.Printf("[mux] Creating chaptered m4b at %s (bitrate - %s)", resolved.OutputAbs, resolved.Bitrate)
	if err := media.CreateM4B(media.M4BOptions{
		InputPath:              resolved.InputAbs,
		MetadataPath:           metadataPath,
		OutputPath:             resolved.OutputAbs,
		Bitrate:                resolved.Bitrate,
		AttachedPicStreamIndex: probe.AttachedPicStreamIndex,
		IncludeCover:           probe.HasAttachedPic,
	}); err != nil {
		return err
	}

	if resolved.KeepTemp {
		log.Printf("[metadata] Metadata in %s", ws.Dir)
	}

	log.Printf("[mux] Done. Output written to %s", resolved.OutputAbs)
	return nil
}
