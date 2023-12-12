#!/usr/bin/env bash

SCRIPT_ROOT="$(dirname "${BASH_SOURCE[0]}")"
cd "${SCRIPT_ROOT}"

set -euo pipefail

SRC_DEVELOPMENT=true SRC_LOG_LEVEL=dbug go run .
