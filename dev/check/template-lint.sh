#!/bin/bash
set -e
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

export GOBIN="$PWD/vendor/.bin"
export PATH=$GOBIN:$PATH

go install sourcegraph.com/sourcegraph/sourcegraph/vendor/sourcegraph.com/sourcegraph/go-template-lint

go-template-lint -f cmd/frontend/internal/app/tmpl/tmpl_funcs.go -t cmd/frontend/internal/app/tmpl/tmpl.go -td cmd/frontend/internal/app/templates
