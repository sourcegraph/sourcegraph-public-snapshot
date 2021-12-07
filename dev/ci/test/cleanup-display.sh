#!/usr/bin/env bash

procs=(ffmpeg Xvfb)

for p in "${procs[@]}"; do pgrep "$p" | xargs kill; done
