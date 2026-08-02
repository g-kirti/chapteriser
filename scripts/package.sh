#!/usr/bin/env bash

version=$1
build="$PROJECT_ROOT/dist"

if [ ! "$version" ]; then
	echo "Missing version number" >&2
	exit 1
fi

declare -A env_vars
env_vars[linux]="$VOSK_LINUX_X86_64"
env_vars[windows]="$VOSK_WIN64"

declare -A ffmpeg_dirs
ffmpeg_dirs[linux]="$FFMPEG_LINUX"
ffmpeg_dirs[windows]="$FFMPEG_WINDOWS"

for os in "${!env_vars[@]}"; do
  lib="${env_vars[$os]}"
  ffmpeg="${ffmpeg_dirs[$os]}"
  
  "$PROJECT_ROOT/scripts/package-${os}.sh" "$version" "$lib" "$ffmpeg" "$build"
done
