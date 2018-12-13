#!/usr/bin/env bash

# This script ensures pkg/conf is NOT imported by the management console
# indirectly. This would inherently mean the management console is tied to the
# frontend working in some way, which would be really bad.

echo "--- mgmt console conf import"

set -euf -o pipefail

template='{{with $pkg := .}}{{ range $pkg.Deps }}{{ printf "%s imports %s\n" $pkg.ImportPath .}}{{end}}{{end}}'

if go list ./../../cmd/management-console/... \
        | xargs go list -f "$template" \
        | grep -E "github.com/sourcegraph/sourcegraph/pkg/conf$"; then
    echo "Error: the management console is not allowed to import pkg/conf"
    exit 1
fi
