#!/usr/bin/env bash

set -e

echo "--- :bazel: build pipeline generator"
bazel \
  --bazelrc=.bazelrc \
  --bazelrc=.aspect/bazelrc/ci.bazelrc \
  --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
  build \
  //enterprise/dev/ci:ci

pipeline_gen="$(bazel cquery //enterprise/dev/ci:ci --output files)"

echo "--- :writing_hand: generate pipeline"
$pipeline_gen | tee generated-pipeline.yml

echo ""
echo "--- :arrow_up: upload pipeline"
buildkite-agent pipeline upload generated-pipeline.yml
