#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to enterprise/
set -euxo pipefail

../dev/generate.sh

go list ./... | xargs go generate -x
