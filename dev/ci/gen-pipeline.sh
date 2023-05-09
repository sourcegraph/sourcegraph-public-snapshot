#!/usr/bin/env bash

set -e

echo "--- :books: Annotating build with Glossary"
buildkite-agent annotate --style info <./enterprise/dev/ci/glossary.md

echo "--- :bazel: Build pipeline generator"
bazel \
  --bazelrc=.bazelrc \
  --bazelrc=.aspect/bazelrc/ci.bazelrc \
  --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
  build \
  //enterprise/dev/ci:ci

pipeline_gen="$(bazel cquery //enterprise/dev/ci:ci --output files)"

echo "--- :writing_hand: Generate pipeline"
$pipeline_gen | tee generated-pipeline.yml

echo ""
echo "--- :arrow_up: Upload pipeline"
buildkite-agent pipeline upload generated-pipeline.yml
