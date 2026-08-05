package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Config struct {
	Language        string
	InputPath       string
	OutputPath      string
	MetadataInput   string
	ModelPath       string
	VoskLibraryPath string
	FFmpegPath      string
	FFprobePath     string
	Workers         int
	ChunkMinutes    int
	Bitrate         string
	SkipMux         bool
	KeepTemp        string
	SplitAudio      bool
}

type ResolvedConfig struct {
	Language         string
	InputAbs         string
	OutputAbs        string
	MetadataInputAbs string
	ModelPath        string
	VoskLibraryPath  string
	FFmpegPath       string
	FFprobePath      string
	TempDirAbs       string
	Workers          int
	ChunkMinutes     int
	Bitrate          string
	SkipMux          bool
	KeepTemp         bool
	SplitAudio       bool
}

func DefaultConfig() Config {
	return Config{
		Language:     "en",
		ModelPath:    "model/vosk-model-small-en-us-0.15",
		Workers:      4,
		ChunkMinutes: 30,
		Bitrate:      "96k",
	}
}

func defaultOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	return base + ".chaptered.m4b"
}

func defaultSplitOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	return strings.TrimSuffix(inputPath, ext)
}

func normaliseLanguageCode(userInput string) (string, error) {
	cleanInput := strings.TrimSpace(strings.ToLower(userInput))

	switch cleanInput {
	case "en", "english", "eng":
		return "en", nil
	case "es", "spanish", "spa", "español", "espanol":
		return "es", nil
	case "fr", "french", "fre", "fra", "français", "francais":
		return "fr", nil
	default:
		return "", fmt.Errorf("%q is unsupported.", userInput)
	}
}

var bitratePattern = regexp.MustCompile(`(?i)^[1-9][0-9]*(k|m)$`)

func (c Config) Resolve() (ResolvedConfig, error) {
	if c.InputPath == "" {
		return ResolvedConfig{}, fmt.Errorf("Missing input file: use -i <audio-file>")
	}
	if c.Workers < 0 {
		return ResolvedConfig{}, fmt.Errorf("Flag '-workers' must be >= 0")
	}
	if c.ChunkMinutes < 1 {
		return ResolvedConfig{}, fmt.Errorf("Flag '-chunk-minutes' must be >= 1")
	}
	if c.Bitrate == "" {
		return ResolvedConfig{}, fmt.Errorf("Flag '-bitrate' cannot be empty")
	}
	if !bitratePattern.MatchString(c.Bitrate) {
		return ResolvedConfig{}, fmt.Errorf("Flag '-bitrate' must include a unit like '96k' or '1m'")
	}
	if strings.TrimSpace(c.MetadataInput) == "" && c.ModelPath == "" {
		return ResolvedConfig{}, fmt.Errorf("Flag '-model-path' cannot be empty")
	}

	keepTemp := false
	tempOverride := strings.TrimSpace(c.KeepTemp)
	if tempOverride != "" {
		switch strings.ToLower(tempOverride) {
		case "true", "1":
			keepTemp = true
			tempOverride = ""
		case "false", "0":
			keepTemp = false
			tempOverride = ""
		default:
			keepTemp = true
		}
	}

	langCode, err := normaliseLanguageCode(c.Language)
	if err != nil {
		return ResolvedConfig{}, fmt.Errorf("Failed to resolve language: %w", err)
	}

	outputPath := c.OutputPath
	if outputPath == "" {
		if c.SplitAudio {
			outputPath = defaultSplitOutputPath(c.InputPath)
		} else {
			outputPath = defaultOutputPath(c.InputPath)
		}
	}
	if !c.SplitAudio && !strings.EqualFold(filepath.Ext(outputPath), ".m4b") {
		outputPath += ".m4b"
	}

	inputAbs, err := filepath.Abs(c.InputPath)
	if err != nil {
		return ResolvedConfig{}, fmt.Errorf("Failed to resolve input path: %w", err)
	}
	outputAbs, err := filepath.Abs(outputPath)
	if err != nil {
		return ResolvedConfig{}, fmt.Errorf("Failed to resolve output path: %w", err)
	}
	metadataInputAbs := ""
	if strings.TrimSpace(c.MetadataInput) != "" {
		metadataInputAbs, err = filepath.Abs(c.MetadataInput)
		if err != nil {
			return ResolvedConfig{}, fmt.Errorf("Failed to resolve metadata-input path: %w", err)
		}
	}

	tempDirPath := tempOverride
	if tempDirPath == "" {
		tempDirPath = os.TempDir()
	}
	tempDirAbs, err := filepath.Abs(tempDirPath)
	if err != nil {
		return ResolvedConfig{}, fmt.Errorf("Failed to resolve temp-dir path: %w", err)
	}

	return ResolvedConfig{
		InputAbs:         inputAbs,
		OutputAbs:        outputAbs,
		Language:         langCode,
		MetadataInputAbs: metadataInputAbs,
		ModelPath:        c.ModelPath,
		VoskLibraryPath:  c.VoskLibraryPath,
		FFmpegPath:       c.FFmpegPath,
		FFprobePath:      c.FFprobePath,
		TempDirAbs:       tempDirAbs,
		Workers:          c.Workers,
		ChunkMinutes:     c.ChunkMinutes,
		Bitrate:          c.Bitrate,
		SkipMux:          c.SkipMux,
		SplitAudio:       c.SplitAudio,
		KeepTemp:         keepTemp,
	}, nil
}
