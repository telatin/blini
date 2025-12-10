#!/usr/bin/env bash
# Compiles blini executables for all platforms.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
EXE=blini
SOURCE="$SCRIPT_DIR/$EXE/blini.go"
# Exrtact version from "blini/blini.go" where there is a line like 'var version = "0.4.1"'
VERSION=$(sed -n 's/.*version[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' "$SOURCE" | head -n 1)

if [[ -z "$VERSION" ]]; then
	echo "Could not determine version from $SOURCE" >&2
	exit 1
fi

OUTDIR="$SCRIPT_DIR/../release"
LDFLAGS="-s -X main.version=$VERSION"

rm -fr "$OUTDIR"
mkdir -p "$OUTDIR"

echo "Building blini $VERSION"

# Linux
go build -ldflags "$LDFLAGS" -o "$OUTDIR/$EXE" "$SCRIPT_DIR/$EXE"
zip -j "$OUTDIR/${EXE}_${VERSION}_linux.zip" "$OUTDIR/$EXE"
rm "$OUTDIR/$EXE"

# Mac
GOOS=darwin go build -ldflags "$LDFLAGS" -o "$OUTDIR/$EXE" "$SCRIPT_DIR/$EXE"
zip -j "$OUTDIR/${EXE}_${VERSION}_mac.zip" "$OUTDIR/$EXE"
rm "$OUTDIR/$EXE"

# Windows
GOOS=windows go build -ldflags "$LDFLAGS" -o "$OUTDIR/$EXE.exe" "$SCRIPT_DIR/$EXE"
zip -j "$OUTDIR/${EXE}_${VERSION}_win.zip" "$OUTDIR/$EXE.exe"
rm "$OUTDIR/$EXE.exe"

ls -lh "$OUTDIR"
