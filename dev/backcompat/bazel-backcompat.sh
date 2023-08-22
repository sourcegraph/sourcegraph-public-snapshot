#!/usr/bin/env bash
set -eu

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

bazelrcs=(--bazelrc=.bazelrc)
current_commit=$(git rev-parse HEAD)
tag="5.1.0"

function restore_current_commit() {
  git checkout --force "${current_commit}"
}

EXIT_CODE=0
git diff --quiet --exit-code || EXIT_CODE=$? # do not fail on non-zero exit
if [[ $EXIT_CODE -ne 0 ]]; then
  echo "🚨 WARNING: Backcompat tests does destructive operations on the repository. Please make sure your changes are commited"
fi

if [[ ${CI:-} == "true" ]]; then
  bazelrcs=(--bazelrc=.bazelrc --bazelrc=.aspect/bazelrc/ci.bazelrc --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc)
else
  if [[ $EXIT_CODE -ne 0 ]]; then
    echo "The following files have changes:"
    git diff --name-only --exit-code
  else
    trap restore_current_commit EXIT
  fi
fi

echo "--- :git::rewind: checkout v${tag}"
git checkout --force "v${tag}"

echo "--- :git: checkout migrations and scripts at ${current_commit}"
git checkout --force "${current_commit}" -- migrations/ dev/backcompat/patch_flakes.sh dev/backcompat/flakes.json

echo "--- :snowflake: patch flake for tag ${tag}"
./dev/backcompat/patch_flakes.sh ${tag}

echo "--- :bazel: bazel test"
bazel "${bazelrcs[@]}" \
  test --test_tag_filters=go -- \
  //cmd/... \
  //lib/... \
  //internal/... \
  //enterprise/cmd/... \
  //enterprise/internal/...\
  -//cmd/migrator/... \
  -//enterprise/cmd/migrator/...
