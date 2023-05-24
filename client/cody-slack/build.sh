#!/usr/bin/env bash

# Navigate to the directory two levels above the one where the current script is located.
cd "$(dirname "${BASH_SOURCE[0]}")"/../..
set -ex

# Create a temporary directory for output and ensure it's cleaned up on exit
OUTPUT=$(mktemp -d -t cody_slack_dockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

# Define targets and directories for the build
declare -A targets=(
  ["//client/cody-slack:bundle"]="$OUTPUT/dist"
  ["//client/cody-slack:package_json_prod"]="$OUTPUT/package"
)

# Build and copy the targets
for target in "${!targets[@]}"; do
  ./dev/ci/bazel.sh build "$target"
  mkdir -p "${targets[$target]}"

  mapfile -t files < <(./dev/ci/bazel.sh cquery "$target" --output=files)
  for file in "${files[@]}"
  do
    cp "$file" "${targets[$target]}"
  done
done

# Build the docker image
docker build -f client/cody-slack/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION

exit $?
