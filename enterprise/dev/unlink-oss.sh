#!/usr/bin/env bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")/..

yarn unlink @sourcegraph/webapp
yarn --check-files

go mod edit -dropreplace=github.com/sourcegraph/sourcegraph
