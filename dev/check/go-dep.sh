#!/bin/bash
set -e
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

dep=dep-$(go env GOOS)-$(go env GOARCH)

curl -L -o /tmp/$dep https://github.com/golang/dep/releases/download/v0.4.1/$dep
chmod +x /tmp/$dep

working_copy_hash=$((git diff; git status) | (md5sum || md5) 2> /dev/null)

/tmp/$dep ensure

new_working_copy_hash=$((git diff; git status) | (md5sum || md5) 2> /dev/null)

if [[ ${working_copy_hash} = ${new_working_copy_hash} ]]; then
    echo "SUCCESS: dep ensure did not change the working copy"
else
    echo "FAIL: dep ensure changed the working copy"
    git diff
    git status
    exit 2
fi
