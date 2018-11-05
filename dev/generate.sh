#!/bin/bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

# A hard-coded list of top level go paths is passed directly to `go list`
# because `go list` doesn't support excluding directories. node_modules/ is
# problematic because on of the transitive devDependencies of the browser
# extension includes some Go code that is incompatible with Sourcegraph. The
# offending path is:
#
# node_modules/snyk-go-plugin/gosrc/resolver
echo cmd dev enterprise pkg schema xlang | xargs -n 1 -I name go list ./name/... | xargs go generate
