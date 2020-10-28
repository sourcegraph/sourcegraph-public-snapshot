#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.." || exit
set -x

# shellcheck disable=SC1091
source /root/.profile
Xvfb "$DISPLAY" -screen 0 1280x1024x24 &
x11vnc -display "$DISPLAY" -forever -rfbport 5900 >/x11vnc.log 2>&1 &

curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src

asdf install
yarn
yarn generate

ffmpeg -y -f x11grab -video_size 1280x1024 -i "$DISPLAY" -pix_fmt yuv420p qatest.mp4 >ffmpeg.log 2>&1 &

IMAGE=sourcegraph/server:insiders ./dev/run-server-image.sh -d --name sourcegraph-server

sleep 15

pushd test/qa-test || exit
go run main.go
popd || exit

source /root/.profile
pushd client/web || exit
yarn run test:regression:core
popd || exit
PID=$(pgrep ffmpeg)
kill "$PID"
