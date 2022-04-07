#!/usr/bin/env bash

set -eu

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

if test -f ".gitmodules"; then
  echo "ERROR: Found .gitmodules in root directory."
  echo "Using git submodules is not allowed; they take a long time to clone in CI."
  echo "Moreover, Buildkite doesn't allow configuring submodule usage on a per-job basis."
  echo ""
  echo "For more context, see:"
  echo "- https://github.com/buildkite/agent/issues/1053#issuecomment-989784531"
  echo "- https://github.com/sourcegraph/sourcegraph/issues/33384"
  echo ""
  echo "If you'd like to change this, please discuss in #dev-experience first."
  exit 1
fi
