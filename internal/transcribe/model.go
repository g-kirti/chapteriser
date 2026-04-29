package transcribe

import (
	"bytes"
	"context"
	"fmt"
	"g-kirti/chapteriser/internal/domain"
	"io"
	"os/exec"
	"strconv"
	"sync"

	vosk "github.com/alphacep/vosk-api/go"
)

func NewSharedModel(modelPath string) (*vosk.VoskModel, error) {
	vosk.SetLogLevel(-1)
	model, err := vosk.NewModel(modelPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load model from %q: %w", modelPath, err)
	}
	return model, nil
}

func TranscribeChunkRange(ctx context.Context, model *vosk.VoskModel, inputPath string, startSeconds int, durationSeconds int, finalResultMu *sync.Mutex) ([]domain.Chapter, error) {
	sampleRate := 16000.0
	rec, err := vosk.NewRecognizer(model, sampleRate)
	if err != nil {
		return nil, fmt.Errorf("Failed to load recogniser: %w", err)
	}
	defer rec.Free()

	rec.SetWords(1)

	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-ss", strconv.Itoa(startSeconds),
		"-t", strconv.Itoa(durationSeconds),
		"-i", inputPath,
		"-vn",
		"-ac", "1",
		"-ar", "16000",
		"-f", "s16le",
		"-acodec", "pcm_s16le",
		"pipe:1",
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	var ffmpegErr bytes.Buffer
	cmd.Stderr = &ffmpegErr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Failed to open ffmpeg stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("Failed to start ffmpeg chunk pipe: %w", err)
	}

	buf := make([]byte, 64*1024)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		n, err := stdout.Read(buf)
		if n > 0 {
			rec.AcceptWaveform(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Read error: %w", err)
		}
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("ffmpeg chunk pipe failed: %w; ffmpeg output: %s", err, ffmpegErr.String())
	}

	var jsonString string

	// lock before getting final result from Vosk recogniser
	if finalResultMu != nil {
		finalResultMu.Lock()
		defer finalResultMu.Unlock()
		jsonString = rec.FinalResult()
	} else {
		jsonString = rec.FinalResult()
	}

	return FindHeadings(jsonString), nil
}
