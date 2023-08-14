#!/usr/bin/env bash

if [[ "$CI" != "1" ]]; then
  pnpm prettier --write --list-different --ignore-unknown --config prettier.config.js
else
  BAZEL_BINDIR=. bazel run //:prettier --run_under="cd $PWD &&" -- --list-different --ignore-unknown --config prettier.config.js
fi
