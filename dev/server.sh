#!/bin/bash

# Runs a Sourcegraph server locally for development. This should still
# be run via `make serve-dev` in the parent directory so that the
# credential env vars get set.

set -euf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

GOBIN="$PWD"/vendor/.bin go get sourcegraph.com/sourcegraph/sourcegraph/vendor/sourcegraph.com/sqs/rego sourcegraph.com/sourcegraph/sourcegraph/vendor/github.com/mattn/goreman

export AUTH0_CLIENT_ID=onW9hT0c7biVUqqNNuggQtMLvxUWHWRC
export AUTH0_CLIENT_SECRET=cpse5jYzcduFkQY79eDYXSwI6xVUO0bIvc4BP6WpojdSiEEG6MwGrt8hj_uX3p5a
export AUTH0_DOMAIN=sourcegraph-dev.auth0.com
export AUTH0_MANAGEMENT_API_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJSYW1KekRwRmN6SFZZNTBpcmFSb0JMdTNRVmFHTE1VRiIsInNjb3BlcyI6eyJ1c2VycyI6eyJhY3Rpb25zIjpbInJlYWQiLCJ1cGRhdGUiXX0sInVzZXJfaWRwX3Rva2VucyI6eyJhY3Rpb25zIjpbInJlYWQiXX0sInVzZXJzX2FwcF9tZXRhZGF0YSI6eyJhY3Rpb25zIjpbInVwZGF0ZSJdfX0sImlhdCI6MTQ3NzA5NDQxOSwianRpIjoiMTA3YzYyMTZjNWZjYzVjNGNkYjYzZTgxNjRjYjg3ODgifQ.ANOcIGeFPH7X_ppl-AXcv2m0zI7hWwqDlRwJ6h_rMdI
export GITHUB_CLIENT_ID=6f2a43bd8877ff5fd1d5
export GITHUB_CLIENT_SECRET=c5ff37d80e3736924cbbdf2922a50cac31963e43
export LIGHTSTEP_PROJECT=sourcegraph-dev
export LIGHTSTEP_ACCESS_TOKEN=d60b0b2477a7ccb05d7783917f648816

export WEBPACK_DEV_SERVER_URL=${WEBPACK_DEV_SERVER_URL:-http://localhost:8080}
export WEBPACK_DEV_SERVER_ADDR=${WEBPACK_DEV_SERVER_ADDR:-127.0.0.1:8080}
export GITHUB_BASE_URL=http://127.0.0.1:3180
export SRC_REPOS_DIR=$HOME/.sourcegraph/repos
export DEBUG=true
export SRC_APP_DISABLE_SUPPORT_SERVICES=true
export SRC_GIT_SERVERS=127.0.0.1:3178
export LSP_PROXY=127.0.0.1:4388
export LSP_PROXY_BG=127.0.0.1:4388
export ZAP_SERVER="ws://$HOME/.sourcegraph/zap"

mkdir -p .bin
env GOBIN=$PWD/.bin go install -v sourcegraph.com/sourcegraph/sourcegraph/cmd/{gitserver,indexer,github-proxy,zap,xlang-go,lsp-proxy}

. dev/langservers.lib.bash
detect_dev_langservers

type ulimit > /dev/null && ulimit -n 10000
exec "$PWD"/vendor/.bin/goreman -f dev/Procfile start
