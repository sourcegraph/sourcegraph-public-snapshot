#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir
set -euxo pipefail

go list ./... | xargs go generate -x
