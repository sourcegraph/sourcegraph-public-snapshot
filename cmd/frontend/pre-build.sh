#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

# Build the enterprise webapp typescript code.
yarn --frozen-lockfile
yarn add "@sourcegraph/webapp@${OSS_WEBAPP_VERSION:-latest}"
NODE_ENV=production yarn run build

go generate github.com/sourcegraph/enterprise/cmd/frontend/internal/assets
go generate github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/templates
