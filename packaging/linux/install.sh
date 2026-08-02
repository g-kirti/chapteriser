#!/usr/bin/env sh
set -eu

source_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
install_dir=${CHAPTERISER_HOME:-"$HOME/.local/lib/chapteriser"}
bin_dir="$HOME/.local/bin"

for required in chapteriser ffmpeg ffprobe libvosk.so; do
	if [ ! -f "$source_dir/$required" ]; then
		printf 'This installer must be run from an extracted chapteriser Linux release bundle. Missing: %s\n' "$source_dir/$required" >&2
		exit 1
	fi
done

mkdir -p "$install_dir" "$bin_dir"
cp -R "$source_dir"/. "$install_dir"/
chmod 755 "$install_dir/chapteriser" "$install_dir/ffmpeg" "$install_dir/ffprobe"
ln -sfn "$install_dir/chapteriser" "$bin_dir/chapteriser"

printf 'Installed chapteriser to %s\n' "$install_dir"
case ":${PATH}:" in
*":$bin_dir:"*) printf 'Run chapteriser -version to verify the installation.\n' ;;
*) printf 'Add %s to PATH, then open a new shell.\n' "$bin_dir" ;;
esac
