#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
set -ex

OUTPUT=$(mktemp -d -t cody_slack_dockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

# TODO(valery): implement Bazel support
# if [[ "${DOCKER_BAZEL:-false}" == "true" ]]; then
#   ./dev/ci/bazel.sh build /client/cody-slack/
#   out=$(./dev/ci/bazel.sh cquery //client/cody-slack --output=files)
#   cp "$out" "$OUTPUT"

#   docker build -f client/cody-slack/Dockerfile -t "$IMAGE" "$OUTPUT" \
#     --progress=plain \
#     --build-arg COMMIT_SHA \
#     --build-arg DATE \
#     --build-arg VERSION
#   exit $?
# fi

echo "--- pnpm build"
pkg="github.com/sourcegraph/sourcegraph/client/cody-slack"
pnpm install
pnpm run build

echo "--- docker build $IMAGE"
docker build -f client/cody-slack/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
