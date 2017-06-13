#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")

./gofmt.sh
./template-lint.sh
./template-inlines.sh
./go-generate.sh
./go-lint.sh
./todo-security.sh
