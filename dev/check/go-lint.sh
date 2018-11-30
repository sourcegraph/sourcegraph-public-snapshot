#!/bin/bash
set -e
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

export GOBIN="$PWD/.bin"
export PATH=$GOBIN:$PATH
export GO111MODULE=on

pkgs=${@:-./...}

go install honnef.co/go/tools/cmd/staticcheck

set -x

go install -buildmode=archive ${pkgs}

go vet ${pkgs}

staticcheck -ignore '*:ST1005' ${pkgs}
