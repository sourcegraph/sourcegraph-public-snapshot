#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."
set -euxo pipefail

docker build -f cmd/symbols/internal/pkg/ctags/Dockerfile -t "$IMAGE" cmd/symbols \
    --build-arg CTAGS_VERSION
