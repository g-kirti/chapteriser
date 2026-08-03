# chapteriser

`chapteriser` detects spoken chapter headings in long audio files, writes FFmpeg chapter metadata, and can create a chaptered `.m4b` or separate audio files.

## Install

Release bundles include the application, Vosk speech-recognition engine, English model, FFmpeg, and FFprobe. They support 64-bit Windows and 64-bit Linux.

### Windows

1. Download `chapteriser_<version>_windows_amd64.zip` from the [latest release](https://github.com/g-kirti/chapteriser/releases/latest) and extract it.
2. In PowerShell, run the extracted `install.ps1`:

   ```powershell
   Set-ExecutionPolicy -Scope Process Bypass
   .\install.ps1
   ```

3. Open a new terminal and run `chapteriser -version`.

The installer places the bundle in `%LOCALAPPDATA%\Programs\chapteriser` and adds it to your user `PATH`. Run `uninstall.ps1` from that directory to remove it.

### Linux

1. Download `chapteriser_<version>_linux_amd64.tar.gz` from the [latest release](https://github.com/g-kirti/chapteriser/releases/latest).
2. Extract it and run the included installer:

   ```sh
   tar -xzf chapteriser_<version>_linux_amd64.tar.gz
   cd chapteriser_<version>_linux_amd64
   ./install.sh
   ```

3. Ensure `~/.local/bin` is on `PATH`, open a new shell, then run `chapteriser -version`.

The installer stores the bundle in `~/.local/lib/chapteriser` and creates `~/.local/bin/chapteriser`. Run `~/.local/lib/chapteriser/uninstall.sh` to remove it.

## Usage

Create a chaptered `.m4b`:
```sh
chapteriser -i book.mp3 -o title
```

Keep temporary files in the current directory and only generate metadata:
```sh
chapteriser -i book.mp3 -keep-temp . -skip-mux
```

Split the input into detected chapter files:
```sh
chapteriser -i book.mp3 -split-audio
```

Split using manually edited FFmpeg metadata rather than automatic detection:
```sh
chapteriser -i book.mp3 -split-audio -metadata-input ffmetadata.txt
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
   chapteriser -i book.mp3 -keep-temp . -skip-mux
   ```

2. Edit the generated `ffmetadata.txt` file.

3. Re-run chapteriser using the edited metadata and skip automatic detection:

   ```bash
   chapteriser -i book.mp3 -metadata-input ./ffmetadata.txt
   ```

You can also combine `-metadata-input` with `-split-audio` to split audio using your manually edited chapter metadata.

## Limitations

- Detection works best when headings are spoken clearly
- Dialogue can produce false-positive headings
- Only English for now
