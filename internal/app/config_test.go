package app

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ModelPath != "model/vosk-model-small-en-us-0.15" {
		t.Fatalf("unexpected model path default: %q", cfg.ModelPath)
	}
	if cfg.Workers != 4 {
		t.Fatalf("unexpected workers default: %d", cfg.Workers)
	}
	if cfg.ChunkMinutes != 30 {
		t.Fatalf("unexpected chunk-minutes default: %d", cfg.ChunkMinutes)
	}
	if cfg.Bitrate != "96k" {
		t.Fatalf("unexpected bitrate default: %q", cfg.Bitrate)
	}
}

func TestResolve_AddsDefaultOutputAndM4BExtension(t *testing.T) {
	cfg := Config{
		InputPath:    "book.mp3",
		ModelPath:    "model",
		Workers:      0,
		ChunkMinutes: 30,
		Bitrate:      "96k",
	}

	resolved, err := cfg.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if !filepath.IsAbs(resolved.InputAbs) {
		t.Fatalf("expected absolute input path, got %q", resolved.InputAbs)
	}
	if !filepath.IsAbs(resolved.OutputAbs) {
		t.Fatalf("expected absolute output path, got %q", resolved.OutputAbs)
	}
	if !strings.HasSuffix(strings.ToLower(resolved.OutputAbs), ".m4b") {
		t.Fatalf("expected .m4b output, got %q", resolved.OutputAbs)
	}
}

func TestResolve_SplitAudioDoesNotForceM4BExtension(t *testing.T) {
	cfg := Config{
		InputPath:    "book.mp3",
		OutputPath:   "chapters",
		ModelPath:    "model",
		Workers:      0,
		ChunkMinutes: 30,
		Bitrate:      "96k",
		SplitAudio:   true,
	}

	resolved, err := cfg.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if filepath.Base(resolved.OutputAbs) != "chapters" {
		t.Fatalf("expected output path to remain directory-like, got %q", resolved.OutputAbs)
	}
}

func TestResolve_ValidationErrors(t *testing.T) {
	tests := []Config{
		{},
		{InputPath: "in.mp3", ModelPath: "model", Workers: -1, ChunkMinutes: 30},
		{InputPath: "in.mp3", ModelPath: "model", Workers: 0, ChunkMinutes: 0},
		{InputPath: "in.mp3", ModelPath: "", Workers: 0, ChunkMinutes: 30},
	}

	for i, cfg := range tests {
		if _, err := cfg.Resolve(); err == nil {
			t.Fatalf("case %d: expected error, got nil", i)
		}
	}
}

func TestResolve_KeepTempOverridesTempDir(t *testing.T) {
	cfg := Config{
		InputPath:    "book.mp3",
		ModelPath:    "model",
		Workers:      0,
		ChunkMinutes: 30,
		Bitrate:      "96k",
		KeepTemp:     ".",
	}

	resolved, err := cfg.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if !resolved.KeepTemp {
		t.Fatalf("expected KeepTemp true")
	}
	if resolved.TempDirAbs == "" {
		t.Fatalf("expected temp dir to be set")
	}
}

func TestResolve_SkipMuxPropagation(t *testing.T) {
	cfg := Config{
		InputPath:    "book.mp3",
		ModelPath:    "model",
		Workers:      0,
		ChunkMinutes: 30,
		Bitrate:      "96k",
		SkipMux:      true,
	}

	resolved, err := cfg.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if !resolved.SkipMux {
		t.Fatalf("expected SkipMux true")
	}
}

func TestResolve_MetadataInputPath(t *testing.T) {
	cfg := Config{
		InputPath:     "book.mp3",
		Workers:       0,
		ChunkMinutes:  30,
		Bitrate:       "96k",
		MetadataInput: "ffmetadata.txt",
	}

	resolved, err := cfg.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if resolved.MetadataInputAbs == "" {
		t.Fatalf("expected metadata input path to be resolved")
	}
}
