package main

import (
	"flag"
	"g-kirti/chapteriser/internal/app"
	"log"
	"os"
)

func main() {
	cfg := app.DefaultConfig()

	flag.StringVar(&cfg.InputPath, "i", "", "input audio file (any ffmpeg-readable format)")
	flag.StringVar(&cfg.OutputPath, "o", "", "output .m4b file path")
	flag.StringVar(&cfg.MetadataInput, "metadata-input", "", "path to an existing ffmetadata.txt file")
	flag.StringVar(&cfg.ModelPath, "model-path", cfg.ModelPath, "path to Vosk model directory")
	flag.IntVar(&cfg.Workers, "workers", cfg.Workers, "number of chunks to transcribe concurrently (0 = auto)")
	flag.IntVar(&cfg.ChunkMinutes, "chunk-minutes", cfg.ChunkMinutes, "chunk size in minutes")
	flag.StringVar(&cfg.Bitrate, "bitrate", cfg.Bitrate, "AAC bitrate for output .m4b")
	flag.BoolVar(&cfg.SkipMux, "skip-mux", false, "skip final m4b creation (only write ffmetadata)")
	flag.StringVar(&cfg.KeepTemp, "keep-temp", "", "keep temp dir; optional value sets base dir (e.g. '.')")
	flag.BoolVar(&cfg.SplitAudio, "split-audio", false, "split audio into respective chapters")

	flag.Parse()

	if err := app.Run(cfg); err != nil {
		log.Printf("[error] %v", err)
		os.Exit(1)
	}
}
