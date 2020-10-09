#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -exo pipefail

# shellcheck disable=SC1091
source /root/.profile
Xvfb "$DISPLAY" -screen 0 1280x1024x24 &
x11vnc -display "$DISPLAY" -forever -rfbport 5900 >/x11vnc.log 2>&1 &

asdf install
yarn
yarn generate

ffmpeg -y -f x11grab -video_size 1280x1024 -i "$DISPLAY" -pix_fmt yuv420p e2e.mp4 >ffmpeg.log 2>&1 &

IMAGE=sourcegraph/server:3.20.1 ./dev/run-server-image.sh

pushd client/web
sleep 10
yarn run test:regression
popd
