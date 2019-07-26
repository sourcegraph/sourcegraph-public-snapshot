#!/usr/bin/env bash

# This script ensures OSS sourcegraph does not import enterprise sourcegraph.

echo "--- go enterprise import"

set -euxf -o pipefail

prefix=sourcegraph.com/enterprise
template='{{with $pkg := .}}{{ range $pkg.Imports }}{{ printf "%s imports %s\n" $pkg.ImportPath .}}{{end}}{{end}}'

if go list ./../../... \
        | grep -v "^$prefix" \
        | xargs go list -f "$template" \
        | grep "$prefix"; then
    echo "Error: OSS is not allowed to import enterprise"
    exit 1
fi
