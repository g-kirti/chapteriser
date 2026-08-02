package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindToolUsesConfiguredFile(t *testing.T) {
	dir := t.TempDir()
	tool := filepath.Join(dir, "ffmpeg")
	if err := os.WriteFile(tool, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := FindTool("ffmpeg", tool)
	if err != nil {
		t.Fatalf("FindTool returned error: %v", err)
	}
	if got != tool {
		t.Fatalf("FindTool = %q, want %q", got, tool)
	}
}

func TestFindVoskLibraryUsesEnvironmentOverride(t *testing.T) {
	dir := t.TempDir()
	library := filepath.Join(dir, "libvosk.so")
	if err := os.WriteFile(library, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("VOSK_LIBRARY_PATH", library)

	got, err := FindVoskLibrary("")
	if err != nil {
		t.Fatalf("FindVoskLibrary returned error: %v", err)
	}
	if got != library {
		t.Fatalf("FindVoskLibrary = %q, want %q", got, library)
	}
}
