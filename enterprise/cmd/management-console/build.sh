#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../../..
set -ex

./cmd/management-console/build.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/management-console
