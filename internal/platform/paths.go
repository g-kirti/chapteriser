package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// FindTool prefers a bundled executable, then falls back to PATH. An explicit
// path is useful for system package-manager installations.
func FindTool(name, configured string) (string, error) {
	if configured != "" {
		return validateFile(configured, name)
	}

	if dir, err := executableDir(); err == nil {
		for _, candidate := range toolNames(name) {
			path := filepath.Join(dir, candidate)
			if isFile(path) {
				return path, nil
			}
		}
	}

	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s was not found beside chapteriser or on PATH; install FFmpeg or pass -%s", name, name)
	}
	return path, nil
}

func FindVoskLibrary(configured string) (string, error) {
	if configured != "" {
		return validateFile(configured, "Vosk library")
	}
	if env := os.Getenv("VOSK_LIBRARY_PATH"); env != "" {
		return validateFile(env, "VOSK_LIBRARY_PATH")
	}
	if dir, err := executableDir(); err == nil {
		for _, candidate := range voskLibraryNames() {
			path := filepath.Join(dir, candidate)
			if isFile(path) {
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("Vosk library was not found beside chapteriser; reinstall the release bundle or pass -vosk-lib")
}

// ResolveModelPath supports both development from the repository root and an
// installed bundle launched from any working directory.
func ResolveModelPath(path string) string {
	if filepath.IsAbs(path) || isDir(path) {
		return path
	}
	if dir, err := executableDir(); err == nil {
		candidate := filepath.Join(dir, path)
		if isDir(candidate) {
			return candidate
		}
	}
	return path
}

func executableDir() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(path), nil
}

func toolNames(name string) []string {
	if runtime.GOOS == "windows" {
		return []string{name + ".exe", name}
	}
	return []string{name}
}

func voskLibraryNames() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"libvosk.dll", "vosk.dll"}
	case "darwin":
		return []string{"libvosk.dylib"}
	default:
		return []string{"libvosk.so"}
	}
}

func validateFile(path, label string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", label, err)
	}
	if !isFile(abs) {
		return "", fmt.Errorf("%s does not exist or is not a file: %s", label, abs)
	}
	return abs, nil
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
