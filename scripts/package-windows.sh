#!/usr/bin/env sh
set -eu

if [ "$#" -ne 4 ]; then
	echo "Usage: $0 VERSION VOSK_DIR FFMPEG_DIR OUTPUT_DIR" >&2
	exit 2
fi

version=$1
vosk_dir=$2
ffmpeg_dir=$3
output_dir=$4
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
bundle="chapteriser_${version}_windows_amd64"
stage="$output_dir/$bundle"
cc=${CC:-x86_64-w64-mingw32-gcc}

for command in "$cc" zip; do
	if ! command -v "$command" >/dev/null 2>&1; then
		echo "Required command is not available: $command" >&2
		exit 1
	fi
done

for required in "$vosk_dir/libvosk.dll" "$ffmpeg_dir/ffmpeg.exe" "$ffmpeg_dir/ffprobe.exe"; do
	if [ ! -f "$required" ]; then
		echo "Missing required release dependency: $required" >&2
		exit 1
	fi
done

rm -rf "$stage"
mkdir -p "$stage"
(
	cd "$root"
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC="$cc" \
		go build -trimpath -ldflags "-s -w -X main.version=$version" -o "$stage/chapteriser.exe" ./cmd/chapteriser
)

cp "$ffmpeg_dir/ffmpeg.exe" "$ffmpeg_dir/ffprobe.exe" "$stage/"
cp "$vosk_dir"/*.dll "$stage/"
cp -R "$root/model" "$stage/model"
cp "$root/README.md" "$root/packaging/windows/install.ps1" "$root/packaging/windows/uninstall.ps1" "$stage/"

mkdir -p "$output_dir"
(
	cd "$output_dir"
	zip -qr "$bundle.zip" "$bundle"
)
printf 'Created %s\n' "$output_dir/$bundle.zip"
