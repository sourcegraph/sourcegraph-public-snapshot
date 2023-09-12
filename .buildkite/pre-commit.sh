#!/usr/bin/env bash

# Script that calls pre-commit in CI

## Temp stuff
mkdir -p .bin && curl -L --retry 3 --retry-max-time 120 https://github.com/pre-commit/pre-commit/releases/download/v3.3.2/pre-commit-3.3.2.pyz --output .bin/pre-commit-3.3.2.pyz --silent

python3 .bin/pre-commit-3.3.2.pyz run --from-ref $(git merge-base ${BUILDKITE_PULL_REQUEST_BASE_BRANCH} HEAD) --to-ref HEAD --show-diff-on-failure
