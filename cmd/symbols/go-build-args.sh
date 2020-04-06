#!/usr/bin/env bash

# This script exports the build arguments required to run go-build.sh

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

. ./dev/libsqlite3-pcre/go-build-args.sh
