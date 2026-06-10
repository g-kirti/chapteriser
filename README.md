# chapteriser

`chapteriser` is a Go tool that uses goroutines to scan monolithic audiobooks, detect chapters and key sections, and generate `ffmetadata.txt`, with options to mux into a chaptered `.m4b` or split by chapter.

## How it works

1. Splits audio into chunks
2. Runs speech recognition on each chunk
3. Detects phrases like "Chapter 5", "Introduction", etc.
4. Generates `ffmetadata.txt`
5. Muxes input into an `.m4b` or splits it by chapter

## Requirements

- Go
- GCC C compiler
- `ffmpeg` and `ffprobe`
- System-installed Vosk library

## Build

Build with:
```bash
go run ./tools/build build
```

This creates `./bin/chapteriser`.

If you have `make` installed, this also works:
```bash
make build
```

## Usage

Create a chaptered `.m4b` (default):
```bash
./bin/chapteriser -i book.mp3 -o title
```

Keep temp files in current directory and skip muxxing:
```bash
./bin/chapteriser -i book.mp3 -keep-temp . -skip-mux
```

Split by detected chapters:
```bash
./bin/chapteriser -i book.mp3 -split-audio
```

Split by user provided metadata (skips detection):
```bash
./bin/chapteriser -i book.mp3 -split-audio -metadata-input /path/to/ffmetadata.txt
```

### Useful flags

- `-model-path` Path to Vosk model (`vosk-model-small-en-us-0.15` is included in this repo)
- `-workers` Number of concurrent workers (`0` = auto, **default**: `4`)
- `-chunk-minutes` Audio chunk duration in minutes (**default**: `30`)
- `-bitrate` Output bitrate for `.m4b` (**default**: `96k`)

You can tweak these to suit your desired memory and CPU usage.

## Manual chapter corrections

Because chapter detection is based on speech recognition and keyword matching, you may wish to manually correct chapters.

If you want to review or edit the generated chapters, you could:–

1. Generate metadata and keep temporary files:

   ```bash
   ./bin/chapteriser -i book.mp3 -keep-temp . -skip-mux
   ```

2. Edit the generated `ffmetadata.txt` file.

3. Re-run chapteriser using the edited metadata and skip automatic detection:

   ```bash
   ./bin/chapteriser -i book.mp3 -metadata-input ./ffmetadata.txt
   ```

You can also combine `-metadata-input` with `-split-audio` to split audio using your manually edited chapter metadata.

## Limitations

- Works best when chapter titles are spoken clearly
- May include false positive headings when keywords are spoken in dialogue
- Only English for now
