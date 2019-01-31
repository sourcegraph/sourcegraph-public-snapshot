#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

# Build the webapp typescript code.
yarn || yarn --frozen-lockfile --network-timeout 60000

go generate ./cmd/frontend/internal/app/assets ./cmd/frontend/internal/app/templates ./cmd/frontend/docsite
