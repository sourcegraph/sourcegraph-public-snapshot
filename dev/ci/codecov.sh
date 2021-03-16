#!/usr/bin/env bash

# Wrapper script for downloading codecov helper. Also includes buildkite log
# output.

echo "--- codecov"

exec bash <(curl -s https://codecov.io/bash) "$@"
