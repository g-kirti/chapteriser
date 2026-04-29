package metadata

import (
	"g-kirti/chapteriser/internal/domain"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteChaptersToFile_WritesTagsAndCapitalizedTitles(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "ffmetadata.txt")

	chapters := []domain.Chapter{
		{Title: "chapter one", Start: 1.0, End: 1.2},
		{Title: "part two", Start: 12.0, End: 12.2},
	}
	tags := map[string]string{
		"title":  "my book",
		"artist": "frank herbert",
	}

	if err := WriteChaptersToFile(chapters, outFile, tags); err != nil {
		t.Fatalf("WriteChaptersToFile failed: %v", err)
	}

	b, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	got := string(b)

	if !strings.Contains(got, ";FFMETADATA1\n") {
		t.Fatalf("missing ffmetadata header")
	}
	if !strings.Contains(got, "artist=frank herbert\n") {
		t.Fatalf("missing artist tag: %s", got)
	}
	if !strings.Contains(got, "title=my book\n") {
		t.Fatalf("missing title tag: %s", got)
	}
	if !strings.Contains(got, "title=Chapter One\n") {
		t.Fatalf("expected capitalized chapter title: %s", got)
	}
	if !strings.Contains(got, "title=Part Two\n") {
		t.Fatalf("expected capitalized chapter title: %s", got)
	}
}
