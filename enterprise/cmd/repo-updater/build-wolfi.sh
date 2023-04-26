#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"/../../..

./cmd/repo-updater/build-wolfi.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/repo-updater
