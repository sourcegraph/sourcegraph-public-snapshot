#!/usr/bin/env bash

# This script ensures pkg/dbconn is only imported by services allowed to
# directly speak with the database.

set -euf -o pipefail

allowed='^github.com/sourcegraph/sourcegraph/cmd/frontend|github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend|github.com/sourcegraph/sourcegraph/cmd/management-console'
template='{{with $pkg := .}}{{ range $pkg.Deps }}{{ printf "%s imports %s\n" $pkg.ImportPath .}}{{end}}{{end}}'

if go list ./../../cmd/... ../../enterprise/cmd/... \
        | grep -Ev "$allowed" \
        | xargs go list -f "$template" \
        | grep "github.com/sourcegraph/sourcegraph/pkg/dbconn"; then
    echo "Error: the above service(s) are not allowed to import pkg/dbconn"
    exit 1
fi
