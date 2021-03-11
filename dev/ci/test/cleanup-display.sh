#!/usr/bin/env bash

PID=$(pgrep ffmpeg)
kill "$PID"
