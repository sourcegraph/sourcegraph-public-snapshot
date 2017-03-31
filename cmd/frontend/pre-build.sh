#!/bin/sh
set -ex

cd ui
yarn install
yarn run build
cd ..
go generate ./cmd/frontend/internal/app/assets ./cmd/frontend/internal/app/templates
