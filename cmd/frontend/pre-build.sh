#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

# Build the enterprise webapp typescript code.
yarn --frozen-lockfile
NODE_ENV=production yarn run build

go generate github.com/sourcegraph/enterprise/cmd/frontend/internal/assets
