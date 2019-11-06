#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../..

set -euxo pipefail

./cmd/frontend/pre-build.sh
