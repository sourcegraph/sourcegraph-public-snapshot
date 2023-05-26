#!/bin/bash

set -eu

registries=(
  # "index.docker.io/sourcegraph"
  "us.gcr.io/sourcegraph-dev"
)

date_fragment="$(date +%y-%m-%d)"

qa_prefix="bazel"

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
  tags_args="$tags_args --tag ${qa_prefix}-${t}"
done

function create_push_command() {
  repository="$1"
  target="$2"
  # echo "--- :bazel::docker: Pushing $repository"

  repositories_args=""
  for registry in "${registries[@]}"; do
    repositories_args="$repositories_args --repository ${registry}/${repository}"
  done

  cmd="bazel \
    --bazelrc=.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
    run \
    "$target" \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh"

  echo "$cmd -- $tags_args $repositories_args"
}

images=$(bazel query 'kind("oci_push rule", //...)')
commands=()

job_file=$(mktemp)
# shellcheck disable=SC2064
trap "rm -rf $job_file" EXIT

for target in ${images[@]}; do
  [[ "$target" =~ ([A-Za-z0-9_-]+): ]]
  name="${BASH_REMATCH[1]}"
  echo $(create_push_command "$name" "$target") >> "$job_file"
done

echo "-- jobfile"
cat "$job_file"
echo "--- "

echo "--- :bazel::docker: Pushing images..."
log_file=$(mktemp)
# shellcheck disable=SC2064
trap "rm -rf $log_file" EXIT
parallel --jobs=8 --line-buffer --joblog "$log_file" -v < "$job_file"
cat "$log_file"
echo "--- "
