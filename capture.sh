#!/bin/bash

set -eou pipefail

cleanup() {
    if [[ -n "${PID-}" ]]; then
        kill $PID 2>/dev/null || true
    fi
    if [[ -n "${CHILD_PID-}" ]]; then
        kill "$CHILD_PID" 2>/dev/null || true
    fi
}
# Trap SIGINT and SIGTERM to call the cleanup function
trap cleanup SIGINT SIGTERM ERR EXIT

# Launch the program
go run . &
# Capture it's PID
PID=$!
# Wait for it to finish launching
sleep 1
# Get the child process it spawns, the actual OpenGL window
CHILD_PID=$(pgrep -P $PID)

OUT_DIR="$PWD/screenshots/"
mkdir -p "$OUT_DIR"
OUT_FILENAME=$1

[ -f "$OUT_DIR/$OUT_FILENAME.png" ] && rm -f "$OUT_DIR/$OUT_FILENAME.png"

# flameshot was the only tool that would take a screenshot of a screen on Wayland
flameshot screen --path "$OUT_DIR/$OUT_FILENAME.png"

# No tool can screenshot a region, so crop the region out of the screensize shot
CROP_REGION="800x600+880+420"
convert "$OUT_DIR/$OUT_FILENAME.png" -crop $CROP_REGION "$OUT_DIR/$OUT_FILENAME.png"

# Close the processes.
kill "$CHILD_PID"
kill $PID
