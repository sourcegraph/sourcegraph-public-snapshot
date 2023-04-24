#!/bin/bash

set -euf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

tmpdir=$(mktemp -d -t wolfi-bin.XXXXXXXX)
function cleanup() {
  echo "Removing $tmpdir"
  rm -rf "$tmpdir"
}
trap cleanup EXIT

# TODO: Install these binaries as part of the buildkite base image
(
  cd "$tmpdir"
  mkdir bin

  # Install apko from Sourcegraph cache
  # Source: https://github.com/chainguard-dev/apko/releases/download/v0.6.0/apko_0.6.0_linux_amd64.tar.gz
  wget https://storage.googleapis.com/package-repository/ci-binaries/apko_0.6.0_linux_amd64.tar.gz
  tar zxf apko_0.6.0_linux_amd64.tar.gz
  mv apko_0.6.0_linux_amd64/apko bin/apko

  # Install apk from Sourcegraph cache
  # Source: https://gitlab.alpinelinux.org/api/v4/projects/5/packages/generic//v2.12.11/x86_64/apk.static
  wget https://storage.googleapis.com/package-repository/ci-binaries/apk-v2.12.11.tar.gz
  tar zxf apk-v2.12.11.tar.gz
  chmod +x apk
  mv apk bin/apk
)

export PATH="$tmpdir/bin:$PATH"

if [ $# -eq 0 ]; then
  echo "No arguments supplied - provide the base image name to build"
  exit 0
fi

name=${1%/}
if [ ! -f "wolfi-images/${name}.yaml" ]; then
  echo "File '$name.yaml' does not exist"
  exit 1
fi

tag=${2-latest}

cd "wolfi-images/"

# Build base image with apko
echo " * Building base image '$name' with apko..."
image_name="sourcegraph-wolfi/${name}-base"
tarball="sourcegraph-wolfi-${name}-base.tar"
apko build --debug "${name}.yaml" \
  "$image_name:latest" \
  "$tarball" ||
  (echo "*** Build failed ***" && exit 1)

# Tag image and upload to GCP Artifact Registry
docker load <"$tarball"

docker tag "$image_name" "us.gcr.io/sourcegraph-dev/wolfi-${name}-base:$tag"
docker push "us.gcr.io/sourcegraph-dev/wolfi-${name}-base:$tag"
# Temporary convenience during initial development, as this doesn't scale to multiple branches!
docker tag "$image_name" "us.gcr.io/sourcegraph-dev/wolfi-${name}-base:latest"
docker push "us.gcr.io/sourcegraph-dev/wolfi-${name}-base:latest"
