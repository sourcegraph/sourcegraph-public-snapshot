#!/usr/bin/env bash

set -e

bazelrc=(--bazelrc=.bazelrc --bazelrc=.aspect/bazelrc/ci.bazelrc --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc)

echo "--- :books: Annotating build with Glossary"
buildkite-agent annotate --style info <./dev/ci/glossary.md

echo "--- :bazel: Build pipeline generator"
bazel "${bazelrc[@]}" \
  build \
  //dev/ci:ci

pipeline_gen="$(bazel "${bazelrc[@]}" cquery //dev/ci:ci --output files)"

echo "--- :writing_hand: Generate pipeline"
$pipeline_gen | tee generated-pipeline.yml

echo ""
echo "--- :arrow_up: Upload pipeline"
buildkite-agent pipeline upload generated-pipeline.yml
