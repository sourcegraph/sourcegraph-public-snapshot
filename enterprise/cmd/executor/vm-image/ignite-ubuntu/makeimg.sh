#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../../../../..
set -ex

if [[ -z $1 ]]; then
  echo "Must set out dir as first arg."
  exit 1
fi

set -u

TEMP_DATA_DIR=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$TEMP_DATA_DIR"
}
trap cleanup EXIT

mkdir "${TEMP_DATA_DIR}/out"

# On macOS we use docker to create the image file.
if [[ "$OSTYPE" == "darwin"* ]]; then
  cat <<EOF >"${TEMP_DATA_DIR}/script.sh"
set -exu

apk add curl e2fsprogs e2fsprogs-extra
curl -fL https://github.com/weaveworks/ignite/releases/download/v0.10.0/ignite-amd64 --output /usr/bin/ignite
chmod +x /usr/bin/ignite
ignite image import --runtime docker sourcegraph/ignite-ubuntu:insiders
EOF

  docker run \
    --platform linux/amd64 \
    --rm \
    --privileged \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v "${TEMP_DATA_DIR}/script.sh:/script.sh:ro" \
    -v "$1:/var/lib/firecracker" \
    --cap-add=CAP_MKNOD \
    --entrypoint /bin/sh \
    docker \
    -- /script.sh
else
  mkdir -p /var/lib/firecracker
  mv /var/lib/firecracker /var/lib/firecracker.bak
  curl -fL https://github.com/weaveworks/ignite/releases/download/v0.10.0/ignite-amd64 --output "${TEMP_DATA_DIR}/ignite"
  chmod +x "${TEMP_DATA_DIR}/ignite"
  "${TEMP_DATA_DIR}/ignite" image import --runtime docker sourcegraph/ignite-ubuntu:insiders
  mv "/var/lib/firecracker" "$1"
  mv /var/lib/firecracker.bak /var/lib/firecracker
fi
