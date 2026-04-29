package metadata

import (
	"bufio"
	"g-kirti/chapteriser/internal/domain"
	"os"
	"strconv"
	"strings"
)

// something of a state-based parser
func ParseFile(path string) ([]domain.Chapter, error) {
	metadata, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer metadata.Close()

	var chapters []domain.Chapter
	var current *domain.Chapter

	scanner := bufio.NewScanner(metadata)
	for scanner.Scan() {
		line := scanner.Text()

		// identify chapters with sentinel [CHAPTER]
		if line == "[CHAPTER]" {
			if current != nil {
				chapters = append(chapters, *current)
			}
			current = &domain.Chapter{}
			continue
		}

		// skip header lines
		if current == nil {
			continue
		}

		// collect chapter info
		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 {
			continue
		}

		key, val := parts[0], parts[1]
		switch key {
		case "START":
			ms, _ := strconv.ParseFloat(val, 64)
			current.Start = ms / 1000.0
		case "END":
			ms, _ := strconv.ParseFloat(val, 64)
			current.End = ms / 1000.0
		case "title":
			current.Title = val
		}
	}
	// append last chapter
	if current != nil {
		chapters = append(chapters, *current)
	}

	return chapters, nil
}
