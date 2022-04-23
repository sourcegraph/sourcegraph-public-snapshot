#!/usr/bin/env bash

# This script builds the symbols go binary.
# Requires a single argument which is the path to the target bindir.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

OUTPUT="${1:?no output path provided}"

cmd/symbols/build.sh

docker cp "$(docker create --rm "$IMAGE")":/usr/local/bin/symbols "$OUTPUT/symbols"
