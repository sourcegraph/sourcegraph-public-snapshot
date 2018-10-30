#!/usr/bin/env bash

set -e
cd $(dirname "${BASH_SOURCE[0]}")/../..

if grep -q 'replace github.com/sourcegraph/sourcegraph => ../sourcegraph' go.mod; then
    echo "go.mod contains a replace directive for github.com/sourcegraph/sourcegraph"
    echo "This replace directive is only for local development and must not be committed"
    exit 1
fi
