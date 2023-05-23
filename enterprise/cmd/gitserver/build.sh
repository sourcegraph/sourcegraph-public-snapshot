#!/usr/bin/env bash

# the build process for the OSS gitserver is identical to the build process for the Enterprise gitserver
# pull some shenanigans up front so that we don't have to sprinkle "enterprise" all throughout the enterprise version

exedir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

path="cmd/gitserver"

if [[ ${exedir} = */enterprise/cmd/gitserver ]]; then
  # We want to build multiple go binaries, so we use a custom build step on CI.
  cd "${exedir}"/../../.. || exit 1
  path="enterprise/${path}"
else
  # We want to build multiple go binaries, so we use a custom build step on CI.
  cd "${exedir}"/../.. || exit 1
fi

### OSS and Enterprise builds should be identical after this point

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)

cleanup() {
  rm -rf "$OUTPUT"
}

trap cleanup EXIT

for f in p4-fusion-install-alpine.sh p4-fusion-wrapper-detect-kill.sh process-stats-watcher.sh; do
  cp -a "./${path}/${f}" "${OUTPUT}"
done

if [[ "${DOCKER_BAZEL:-false}" == "true" ]]; then
  ./dev/ci/bazel.sh build //${path}
  out=$(./dev/ci/bazel.sh cquery //${path} --output=files)
  cp "$out" "$OUTPUT"

  docker build -f ${path}/Dockerfile -t "$IMAGE" "$OUTPUT" \
    --progress=plain \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
  exit $?
fi

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

pkg="github.com/sourcegraph/sourcegraph/${path}"
go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION  -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" -buildmode exe -tags dist -o "$OUTPUT/$(basename $pkg)" "$pkg"

docker build -f ${path}/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
