#!/bin/bash

set -eou pipefail

# Size of window
WIDTH=800
HEIGHT=600
# Location of window
X=880
Y=420

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
print_usage() {
    echo "Usage: $0 [-g] [-h]"
    echo "  -g    Record a GIF instead of taking a screenshot."
    echo "  -h    Display this help message."
}

# Parse arguments
RECORD_GIF=false
while getopts "gh" opt; do
    case ${opt} in
        g )
            RECORD_GIF=true
            ;;
        h )
            print_usage
            exit 0
            ;;
        \? )
            print_usage
            exit 1
            ;;
    esac
done
shift $((OPTIND -1))

# Launch the program
go run . &
# Capture it's PID
PID=$!
# Wait for it to finish launching
sleep 1
# Get the child process it spawns, the actual OpenGL window
CHILD_PID=$(pgrep -P $PID)

OUT_DIR="$PWD/screenshots"
mkdir -p "$OUT_DIR"
OUT_FILENAME=${1:-$(git branch --show-current)}

if $RECORD_GIF; then
    [ -f "$OUT_DIR/$OUT_FILENAME.gif" ] && rm -f "$OUT_DIR/$OUT_FILENAME.gif"

    echo "Not supported!" >&2
    exit 1
else
    [ -f "$OUT_DIR/$OUT_FILENAME.png" ] && rm -f "$OUT_DIR/$OUT_FILENAME.png"

    # flameshot was the only tool that would take a screenshot of a screen on Wayland
    flameshot screen --path "$OUT_DIR/$OUT_FILENAME.png"

    # No tool can screenshot a region, so crop the region out of the screensize shot
    CROP_REGION="${WIDTH}x${HEIGHT}+$X+$Y"
    magick "$OUT_DIR/$OUT_FILENAME.png" -crop $CROP_REGION "$OUT_DIR/$OUT_FILENAME.png"
fi

# Close the processes.
kill "$CHILD_PID"
kill $PID
