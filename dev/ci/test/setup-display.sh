#!/usr/bin/env bash

Xvfb "$DISPLAY" -screen 0 1280x1024x24 &
x11vnc -display "$DISPLAY" -forever -rfbport 5900 >/x11vnc.log 2>&1 &
ffmpeg -y -f x11grab -video_size 1280x1024 -i "$DISPLAY" -pix_fmt yuv420p qatest.mp4 >ffmpeg.log 2>&1 &
