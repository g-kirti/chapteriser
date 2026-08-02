package media

import (
	"bytes"
	"fmt"
	"g-kirti/chapteriser/internal/metadata"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type SplitOptions struct {
	InputPath              string
	MetadataPath           string
	OutputDir              string
	AttachedPicStreamIndex int
	IncludeCover           bool
	FFmpegPath             string
}

var simpleNumberWords = map[string]int{
	"zero":      0,
	"one":       1,
	"two":       2,
	"three":     3,
	"four":      4,
	"five":      5,
	"six":       6,
	"seven":     7,
	"eight":     8,
	"nine":      9,
	"ten":       10,
	"eleven":    11,
	"twelve":    12,
	"thirteen":  13,
	"fourteen":  14,
	"fifteen":   15,
	"sixteen":   16,
	"seventeen": 17,
	"eighteen":  18,
	"nineteen":  19,
	"twenty":    20,
	"thirty":    30,
	"forty":     40,
	"fifty":     50,
	"sixty":     60,
	"seventy":   70,
	"eighty":    80,
	"ninety":    90,
}

func SplitAudioFile(opts SplitOptions) error {

	// create output directory
	err := os.MkdirAll(opts.OutputDir, os.ModePerm)
	if err != nil {
		return err
	}

	chapters, err := metadata.ParseFile(opts.MetadataPath)
	if err != nil {
		return err
	}

	outputExt := filepath.Ext(opts.InputPath)
	for _, ch := range chapters {
		normalizedTitle := normalizeChapterTitle(ch.Title)
		outputName := filepath.Join(opts.OutputDir, normalizedTitle+outputExt)
		log.Printf("[run] Processing: %s (%.2fs - %.2fs)\n", normalizedTitle, ch.Start, ch.End)

		args := []string{
			"-y",
			"-i", opts.InputPath,
			"-ss", fmt.Sprintf("%f", ch.Start),
			"-to", fmt.Sprintf("%f", ch.End),
			"-map", "0:a",
			"-c:a", "copy",
		}

		if opts.IncludeCover && opts.AttachedPicStreamIndex >= 0 {
			args = append(args,
				"-map", fmt.Sprintf("0:%d", opts.AttachedPicStreamIndex),
				"-c:v", "copy",
				"-disposition:v:0", "attached_pic",
			)
		}
		args = append(args, outputName)

		cmd := exec.Command(opts.FFmpegPath, args...)

		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		if err := cmd.Run(); err != nil {
			return fmt.Errorf(
				"Failed to split audio file %q: %w; ffmpeg output: %s",
				normalizedTitle,
				err,
				strings.TrimSpace(out.String()),
			)
		}
	}
	return nil
}

func normalizeChapterTitle(title string) string {
	parts := strings.Fields(title)
	if len(parts) < 2 {
		return title
	}

	prefix := strings.ToLower(parts[0])
	if prefix != "chapter" && prefix != "part" && prefix != "volume" {
		return title
	}

	n, ok := parseNumberWords(parts[1:])
	if !ok {
		return title
	}

	return fmt.Sprintf("%s %d", parts[0], n)
}

func parseNumberWords(words []string) (int, bool) {
	total := 0
	current := 0

	for _, word := range words {
		w := strings.ToLower(word)
		switch w {
		case "and":
			continue
		case "hundred":
			if current == 0 {
				current = 1
			}
			current *= 100
		case "thousand":
			if current == 0 {
				current = 1
			}
			total += current * 1000
			current = 0
		default:
			value, ok := simpleNumberWords[w]
			if !ok {
				return 0, false
			}
			current += value
		}
	}

	return total + current, true
}
