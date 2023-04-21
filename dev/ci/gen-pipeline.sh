#!/usr/bin/env bash

set -e

echo "--- generate pipeline"
bazel \
  --bazelrc=.bazelrc \
  --bazelrc=.aspect/bazelrc/ci.bazelrc \
  --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
  run \
  --run_under="cd $PWD && " \
  //enterprise/dev/ci:ci -- | tee generated-pipeline.yml

echo ""
echo "--- upload pipeline"
buildkite-agent pipeline upload generated-pipeline.yml
