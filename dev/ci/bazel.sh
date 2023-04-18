#!/bin/bash

if [[ "${CI:-false}" == "true" ]]; then
  bazel \
    --bazelrc=.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
    "$@"
else
  bazel "$@"
fi
