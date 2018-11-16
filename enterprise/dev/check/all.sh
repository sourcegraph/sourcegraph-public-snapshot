#!/usr/bin/env bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")

go version
go env
./go-generate.sh

