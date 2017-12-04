#!/bin/bash
set -e
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

export GOBIN="$PWD/vendor/.bin"
export PATH=$GOBIN:$PATH

go install sourcegraph.com/sourcegraph/sourcegraph/vendor/golang.org/x/tools/cmd/stringer

# Runs generate.sh and ensures no files changed. This relies on the go
# generation that ran are idempotent.

working_copy_hash=$((git diff; git status) | (md5sum || md5) 2> /dev/null)

./dev/generate.sh

new_working_copy_hash=$((git diff; git status) | (md5sum || md5) 2> /dev/null)

if [[ ${working_copy_hash} = ${new_working_copy_hash} ]]; then
    echo "SUCCESS: go generate did not change the working copy"
else
    echo "FAIL: go generate changed the working copy"
    git diff
    git status
    exit 2
fi
