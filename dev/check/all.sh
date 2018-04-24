#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")

go version
go env
./gofmt.sh
./template-inlines.sh
./go-generate.sh
./go-lint.sh
./todo-security.sh
./no-localhost-guard.sh
./broken-urls.bash
