#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -x

# shellcheck disable=SC1091
source /root/.profile
Xvfb "$DISPLAY" -screen 0 1280x1024x24 &
x11vnc -display "$DISPLAY" -forever -rfbport 5900 >/x11vnc.log 2>&1 &

asdf install
yarn
yarn generate

ffmpeg -y -f x11grab -video_size 1280x1024 -i "$DISPLAY" -pix_fmt yuv420p e2e.mp4 >ffmpeg.log 2>&1 &

IMAGE=sourcegraph/server:insiders ./dev/run-server-image.sh

sleep 15

pushd test/regression
go run main.go
popd

source /root/.profile
pushd client/web
yarn run test:regression:core
yarn run test:regression:codeintel
popd
PID=$(pgrep ffmpeg)
kill "$PID"
