#!/usr/bin/env bash

set -e

echo "--- gofmt"

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

# Check if all code is gofmt'd

# IGNORE:
#   vendored: Vendored code doesn't matter for gofmt
#   syntax-highlighter/crates: has embedded go test files

DIFF=$(
  find . \( \
    -path ./vendor \
    -o -path ./vendored \
    -o -path ./docker-images/syntax-highlighter/crates/sg-syntax/languages/tree-sitter-go \
    \) -prune -o -name '*.go' -exec gofmt -s -w -d {} +
)
if [ -z "$DIFF" ]; then
  echo "Success: gofmt check passed."
  exit 0
else
  echo "ERROR: gofmt check failed:"
  echo -e "\`\`\`term\n$DIFF\n\`\`\`" >./annotations/gofmt
  echo "^^^ +++"
  exit 1
fi
