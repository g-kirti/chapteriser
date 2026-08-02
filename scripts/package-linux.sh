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
bundle="chapteriser_${version}_linux_amd64"
stage="$output_dir/$bundle"

for required in "$vosk_dir/libvosk.so" "$ffmpeg_dir/ffmpeg" "$ffmpeg_dir/ffprobe"; do
	if [ ! -f "$required" ]; then
		echo "Missing required release dependency: $required" >&2
		exit 1
	fi
done

rm -rf "$stage"
mkdir -p "$stage"
(
	cd "$root"
	CGO_ENABLED=1 go build -trimpath -ldflags "-s -w -X main.version=$version" -o "$stage/chapteriser" ./cmd/chapteriser
)
cp "$ffmpeg_dir/ffmpeg" "$ffmpeg_dir/ffprobe" "$stage/"
cp "$vosk_dir"/libvosk.so* "$stage/"
cp -R "$root/model" "$stage/model"
cp "$root/README.md" "$root/packaging/linux/install.sh" "$root/packaging/linux/uninstall.sh" "$stage/"
chmod 755 "$stage/chapteriser" "$stage/ffmpeg" "$stage/ffprobe" "$stage/install.sh" "$stage/uninstall.sh"

mkdir -p "$output_dir"
tar -C "$output_dir" -czf "$output_dir/$bundle.tar.gz" "$bundle"
printf 'Created %s\n' "$output_dir/$bundle.tar.gz"
