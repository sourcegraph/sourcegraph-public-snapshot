#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../../..
set -ex

./cmd/repo-updater/build.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/repo-updater
