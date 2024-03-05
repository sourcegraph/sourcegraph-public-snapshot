#!/usr/bin/env bash

if [[ "${CI:-false}" == "true" ]]; then
  aspectRC="/tmp/aspect-generated.bazelrc"
  rosetta bazelrc > "$aspectRC"
  bazelrc=(--bazelrc="$aspectRC")

  if [[ "$1"  == "build" || "$1" == "test" || "$1" == "run" ]]; then
    # shellcheck disable=SC2145
    echo "--- :bazel: bazel $@"
  fi
  bazel "${bazelrc[@]}" \
    "$@" \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64
else
  bazel "$@"
fi
