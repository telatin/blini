# Compiles blini executables for all platforms.

set -e

EXE=blini
VERSION=v0.3.1
OUTDIR=../release
FLAGS="-ldflags=-s -X main.version=$VERSION"

rm -fr $OUTDIR
mkdir $OUTDIR

# Linux
go build "$FLAGS" -o $OUTDIR ./$EXE
zip -j $OUTDIR/${EXE}_linux.zip $OUTDIR/$EXE
rm $OUTDIR/$EXE

# Mac
GOOS=darwin go build "$FLAGS" -o $OUTDIR ./$EXE
zip -j $OUTDIR/${EXE}_mac.zip $OUTDIR/$EXE
rm $OUTDIR/$EXE

# Windows
GOOS=windows go build "$FLAGS" -o $OUTDIR ./$EXE
zip -j $OUTDIR/${EXE}_win.zip $OUTDIR/$EXE.exe
rm $OUTDIR/$EXE.exe

ls -lh $OUTDIR
