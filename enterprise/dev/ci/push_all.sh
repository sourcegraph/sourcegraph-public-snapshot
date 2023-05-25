#!/bin/bash

set -eu

registries=(
  "index.docker.io/sourcegraph"
  "us.gcr.io/sourcegraph-dev"
)

date_fragment="$(date +%y-%m-%d)"

tags=(
  "${BUILDKITE_COMMIT:0:12}"
  "${BUILDKITE_COMMIT:0:12}_${date_fragment}"
  "${BUILDKITE_COMMIT:0:12}_${BUILDKITE_BUILD_NUMBER}"
  "${BUILDKITE_COMMIT:0:12}_${date_fragment}_${BUILDKITE_BUILD_NUMBER}"
)

if [ "$BUILDKITE_BRANCH" == "main" ]; then
  tags+=("insiders")
fi

if [[ "$BUILDKITE_TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  tags+=("${BUILDKITE_TAG:1}")
fi

echo "--- :docker: tags"
for tag in "${tags[@]}"; do
  for registry in "${registries[@]}"; do
    echo -e "\t ${registry}/\$IMAGE:${tag}"
  done
done
echo "--- "

tags_args=""
for t in "${tags[@]}"; do
  tags_args="$tags_args --tag $t"
done

function tag_and_push_image() {
  repository="$1"
  target="$2"
  echo "--- :bazel::docker: Pushing $repository"

  repositories_args=""
  for registry in "${registries[@]}"; do
    repositories_args="$repositories_args --repository ${registry}/${repository}"
  done

  bazel \
    --bazelrc=.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh \
    run
    -- "$tags_args" "$repositories_args"

  bazel run "$target" --platforms @zig_sdk//platform:linux_amd64 --extra_toolchains @zig_sdk//toolchain:linux_amd64_gnu.2.31 --workspace_status_command=./dev/bazel_stamp_vars.sh --stamp -- $tags_args $repositories_args
  echo "--- "
}

images=$(bazel query 'kind("oci_push rule", //...)')
for target in ${images[@]}; do
  [[ "$target" =~ ([A-Za-z0-9_-]+): ]]
  name="${BASH_REMATCH[1]}"
  tag_and_push_image "$name" "$target"
done
