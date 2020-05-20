#!/usr/bin/env bash

set -e
cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to enterprise/

../dev/generate.sh

go list ./... | xargs go generate -x
