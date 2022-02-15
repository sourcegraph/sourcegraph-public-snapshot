#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd "$(dirname "${BASH_SOURCE[0]}")"/../..
set -ex

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

cp -a ./cmd/gitserver/p4-fusion-install-alpine.sh "$OUTPUT"

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

pkg="github.com/sourcegraph/sourcegraph/cmd/gitserver"
go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION  -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" -buildmode exe -tags dist -o "$OUTPUT/$(basename $pkg)" "$pkg"

# Can't use it because it's not yet https
# PRIVATE_REGISTRY="private-docker-registry:5000"
# PRIVATE_REGISTRY="us.gcr.io"

docker pull index.docker.io/sourcegraph/gitserver:insiders || true
docker pull index.docker.io/sourcegraph/gitserver:p4cli || true
docker pull index.docker.io/sourcegraph/gitserver:p4-fusion || true
docker pull index.docker.io/sourcegraph/gitserver:coursier || true

docker build \
  --target p4cli \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  --cache-from index.docker.io/sourcegraph/gitserver:p4cli \
  -t index.docker.io/sourcegraph/gitserver:p4cli \
  -f cmd/gitserver/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION

docker build \
  --target p4-fusion \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  --cache-from index.docker.io/sourcegraph/gitserver:p4cli \
  --cache-from index.docker.io/sourcegraph/gitserver:p4-fusion \
  -t index.docker.io/sourcegraph/gitserver:p4-fusion \
  -f cmd/gitserver/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION

docker build \
  --target coursier \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  --cache-from index.docker.io/sourcegraph/gitserver:p4cli \
  --cache-from index.docker.io/sourcegraph/gitserver:p4-fusion \
  --cache-from index.docker.io/sourcegraph/gitserver:coursier \
  -t index.docker.io/sourcegraph/gitserver:coursier \
  -f cmd/gitserver/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION

docker push index.docker.io/sourcegraph/gitserver:p4cli
docker push index.docker.io/sourcegraph/gitserver:p4-fusion
docker push index.docker.io/sourcegraph/gitserver:coursier

docker build \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  --cache-from index.docker.io/sourcegraph/gitserver:p4cli \
  --cache-from index.docker.io/sourcegraph/gitserver:fusion \
  --cache-from index.docker.io/sourcegraph/gitserver:coursier \
  --cache-from us.gcr.io/sourcegraph-dev/server:insiders \
  -f cmd/gitserver/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
