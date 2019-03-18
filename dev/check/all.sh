#!/usr/bin/env bash

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
./bash-syntax.sh

# TODO(sqs): Reenable this check when about.sourcegraph.com is reliable. Most failures come from its
# downtime, not from broken URLs.
#
# ./broken-urls.bash
