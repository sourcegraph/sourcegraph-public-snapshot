#!/usr/bin/env bash

set -eu

function preview_tags() {
  IFS=' ' read -r -a registries <<<"$1"
  tags="$2"

  for registry in "${registries[@]}"; do
    echo -e "\t ${registry}/\$IMAGE with tags: ${tags}"
  done
}

function create_push_command() {
  IFS=' ' read -r -a registries <<<"$1"
  repository="$2"
  target="$3"
  tags_args="$4"

  repositories_args=""
  for registry in "${registries[@]}"; do
    repositories_args="$repositories_args --repository ${registry}/${repository}"
  done

  cmd="bazel \
    --bazelrc=.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
    run \
    $target \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh"

  echo "$cmd -- $tags_args $repositories_args"
}

dev_registries=(
  "us.gcr.io/sourcegraph-dev"
)
prod_registries=(
  "index.docker.io/sourcegraph"
)

date_fragment="$(date +%Y-%m-%d)"

qa_prefix="bazel"

dev_tags=(
  "${BUILDKITE_COMMIT:0:12}"
  "${BUILDKITE_COMMIT:0:12}_${date_fragment}"
  "${BUILDKITE_COMMIT:0:12}_${BUILDKITE_BUILD_NUMBER}"
)
prod_tags=(
  "${PUSH_VERSION}"
)

# Only push to production if PUSH_PRODUCTION is set by caller
BK_PUSH_PROD=$PUSH_PRODUCTION

push_prod=false

if [ "$BUILDKITE_BRANCH" == "main" ] && [ -n "$BK_PUSH_PROD" ]; then
  dev_tags+=("insiders")
  prod_tags+=("insiders")
  push_prod=true
fi

if [[ "$BUILDKITE_TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] && [ -n "$BK_PUSH_PROD" ]; then
  dev_tags+=("${BUILDKITE_TAG:1}")
  prod_tags+=("${BUILDKITE_TAG:1}")
  push_prod=true
fi

echo "--- done"

registries=()
tags_args=""
if $push_prod; then
  registries=("${prod_registries[@]}")

  for t in "${prod_tags[@]}"; do
    tags_args="$tags_args --tag ${qa_prefix}-${t}"
  done
else
  registries=("${dev_registries[@]}")

  for t in "${dev_tags[@]}"; do
    tags_args="$tags_args --tag ${qa_prefix}-${t}"
  done
fi

preview_tags "${registries[*]}" "$tags_args"

images=$(bazel query 'kind("oci_push rule", //...)')

job_file=$(mktemp)
# shellcheck disable=SC2064
trap "rm -rf $job_file" EXIT

# shellcheck disable=SC2068
for target in ${images[@]}; do
  [[ "$target" =~ ([A-Za-z0-9_-]+): ]]
  name="${BASH_REMATCH[1]}"
  # Append push command
  create_push_command "${registries[*]}" "$name" "$target" "$tags_args" >>"$job_file"
done

echo "-- jobfile"
cat "$job_file"
echo "--- done"

echo "--- :bazel::docker: Pushing images..."
log_file=$(mktemp)
# shellcheck disable=SC2064
trap "rm -rf $log_file" EXIT
parallel --jobs=16 --line-buffer --joblog "$log_file" -v <"$job_file"

# Pretty print the output from gnu parallel
while read -r line; do
  # Skip the first line (header)
  if [[ "$line" != Seq* ]]; then
    cmd="$(echo "$line" | cut -f9)"
    [[ "$cmd" =~ (\/\/[^ ]+) ]]
    target="${BASH_REMATCH[1]}"
    exitcode="$(echo "$line" | cut -f7)"
    duration="$(echo "$line" | cut -f4 | tr -d "[:blank:]")"
    if [ "$exitcode" == "0" ]; then
      echo "--- :docker::arrow_heading_up: $target ${duration}s :white_check_mark:"
    else
      echo "--- :docker::arrow_heading_up: $target ${duration}s: failed with $exitcode) :red_circle:"
    fi
  fi
done <"$log_file"

echo "--- :bazel::docker: detailed summary"
cat "$log_file"
echo "--- done"
