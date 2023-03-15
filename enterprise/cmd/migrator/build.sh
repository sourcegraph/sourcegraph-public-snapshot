#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"/../../..

./cmd/migrator/build.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/migrator
