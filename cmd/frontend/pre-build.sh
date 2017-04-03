#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

cd ui
yarn install
yarn run build
cd ..
go generate ./cmd/frontend/internal/app/assets ./cmd/frontend/internal/app/templates
