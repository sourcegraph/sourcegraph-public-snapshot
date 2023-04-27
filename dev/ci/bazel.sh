#!/bin/bash

set -x

if [[ "${CI:-false}" == "true" ]]; then
  bazel \
    --bazelrc=.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
    "$@" \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64
else
  bazel "$@"
fi
