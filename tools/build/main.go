package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const appName = "chapteriser"

func main() {
	cmd := "build"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	var err error
	switch cmd {
	case "build":
		err = build()
	case "run":
		err = run()
	case "test":
		err = test()
	case "clean":
		err = clean()
	case "help":
		usage()
	default:
		usage()
		err = fmt.Errorf("Unknown command %q", cmd)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: go run ./tools/build [build|run|test|clean|help]\n")
}

func build() error {
	binPath := binaryPath()
	fmt.Printf("Building %s...\n", binPath)
	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		return err
	}
	if err := runCommand("go", "build", "-o", binPath, "."); err != nil {
		return err
	}
	fmt.Println("Build done.")
	return nil
}

func run() error {
	if err := build(); err != nil {
		return err
	}
	return runCommand(binaryPath(), "-h")
}

func test() error {
	fmt.Println("Testing...")
	return runCommand("go", "test", "./...")
}

func clean() error {
	fmt.Println("Cleaning...")
	if err := os.RemoveAll("bin"); err != nil {
		return err
	}
	fmt.Println("Removed bin.")
	return nil
}

func binaryPath() string {
	name := appName
	if targetGOOS() == "windows" {
		name += ".exe"
	}
	return filepath.Join("bin", name)
}

func targetGOOS() string {
	if goos := os.Getenv("GOOS"); goos != "" {
		return goos
	}
	return runtime.GOOS
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
