#!/bin/bash

echo "--- docsite check (lint Markdown files in doc/)"

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

export GOBIN="$PWD/.bin"
export PATH=$GOBIN:$PATH
export GO111MODULE=on

go install github.com/sourcegraph/docsite/cmd/docsite

# Check broken links, etc., in Markdown files in doc/.

echo
echo

docsite check || {
    echo
    echo Errors found in Markdown documentation files. Fix the errors in doc/ and try again.
    echo
    exit 1
}
