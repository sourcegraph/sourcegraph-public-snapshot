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
PRIVATE_REGISTRY="us.gcr.io"

docker pull us.gcr.io/sourcegraph-dev/gitserver:insiders || true
docker pull $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4cli || true
docker pull $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4-fusion || true
docker pull $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4-coursier || true

docker build \
  --target p4cli \
  --cache-from $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4cli \
  -t $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4cli \
  -f cmd/gitserver/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION

docker build \
  --target p4-fusion \
  --cache-from $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4cli \
  --cache-from $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4-fusion \
  -t $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4-fusion \
  -f cmd/gitserver/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION

docker build \
  --target coursier \
  --cache-from $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4cli \
  --cache-from $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4-fusion \
  --cache-from $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:coursier \
  -t $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:coursier \
  -f cmd/gitserver/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION

docker push $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4cli
docker push $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4-fusion
docker push $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4-coursier

docker build \
  --cache-from $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4cli \
  --cache-from $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:p4-fusion \
  --cache-from $PRIVATE_REGISTRY/sourcegraph-dev/gitserver:coursier \
  --cache-from us.gcr.io/sourcegraph-dev/server:insiders \
  -f cmd/gitserver/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
