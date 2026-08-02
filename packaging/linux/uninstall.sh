#!/usr/bin/env sh
set -eu

install_dir=${CHAPTERISER_HOME:-"$HOME/.local/lib/chapteriser"}
rm -f "$HOME/.local/bin/chapteriser"
rm -rf "$install_dir"
printf 'chapteriser has been removed.\n'
