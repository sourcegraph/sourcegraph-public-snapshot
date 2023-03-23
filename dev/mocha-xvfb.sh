#!/usr/bin/env bash

 # Xvfb "$DISPLAY" -screen 0 1280x1024x24 &
 # x11vnc -display "$DISPLAY" -forever -rfbport 5900 >/x11vnc.log 2>&1 &
 # ffmpeg -y -f x11grab -video_size 1280x1024 -i "$DISPLAY" -pix_fmt yuv420p qatest.mp4 >ffmpeg.log 2>&1 &

BUILDKITE=${BUILDKITE:-"0"}

if [[ $BUILDKITE -eq "0" ]]; then
  # only start Xvfb if we're not in CI
  echo "NOT CI - STARTING Xvfb"
  /usr/X11/bin/Xvfb ":99" -screen 0 1280x1024x24 &
else
  echo "IN CI - NOT STARTING Xvfb"
fi

cmd=$1
shift

echo "Running: '$cmd'"
"$cmd" "$@" || exit 1
