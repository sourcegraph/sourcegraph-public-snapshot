#!/usr/bin/env bash

# This script ensures pkg/db/dbconn is only imported by services allowed to
# directly speak with the database.

echo "--- go dbconn import"

set -euf -o pipefail

allowed='^sourcegraph.com/cmd/frontend|sourcegraph.com/enterprise/cmd/frontend|sourcegraph.com/cmd/management-console|sourcegraph.com/enterprise/cmd/management-console'
template='{{with $pkg := .}}{{ range $pkg.Deps }}{{ printf "%s imports %s\n" $pkg.ImportPath .}}{{end}}{{end}}'

if go list ./../../cmd/... ../../enterprise/cmd/... \
        | grep -Ev "$allowed" \
        | xargs go list -f "$template" \
        | grep "sourcegraph.com/pkg/db/dbconn"; then
    echo "Error: the above service(s) are not allowed to import pkg/db/dbconn"
    exit 1
fi
