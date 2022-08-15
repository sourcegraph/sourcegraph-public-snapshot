#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../../../../..
set -eux

# TODO: Check $1 is set.

TEMP_DATA_DIR=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$TEMP_DATA_DIR"
}
trap cleanup EXIT

mkdir "${TEMP_DATA_DIR}/out"

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
  -v "${TEMP_DATA_DIR}/out:/var/lib/firecracker" \
  --cap-add=CAP_MKNOD \
  --entrypoint /bin/sh \
  docker \
  -- /script.sh

imageID="$(ls "${TEMP_DATA_DIR}"/out/image)"
mv "${TEMP_DATA_DIR}/out/image/${imageID}/image.ext4" "$1/image.ext4"
mv "${TEMP_DATA_DIR}/out/image/${imageID}/metadata.json" "$1/metadata.json"
