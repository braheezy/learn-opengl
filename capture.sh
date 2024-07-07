#!/bin/bash

set -eou pipefail

# Launch the program
go run main.go &
# Capture it's PID
PID=$!
# Wait for it to finish launching
sleep 1
# Get the child process it spawns, the actual OpenGL window
CHILD_PID=$(pgrep -P $PID)

OUT_DIR="$PWD/screenshots/"
mkdir -p "$OUT_DIR"
OUT_FILE=$(git branch --show-current)

[ -f "$OUT_DIR/$OUT_FILE.png" ] && rm -f "$OUT_DIR/$OUT_FILE.png"

# flameshot was the only tool that would take a screenshot of a screen on Wayland
flameshot screen --path "$OUT_DIR/$OUT_FILE.png"

# No tool can screenshot a region, so crop the region out of the screensize shot
CROP_REGION="800x600+880+420"
convert "$OUT_DIR/$OUT_FILE.png" -crop $CROP_REGION "$OUT_DIR/$OUT_FILE.png"

# Close the processes.
kill "$CHILD_PID"
kill $PID
