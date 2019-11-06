#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -euxo pipefail

docker build -f cmd/symbols/libsqlite3-pcre/Dockerfile -t "$IMAGE" cmd/symbols/libsqlite3-pcre
