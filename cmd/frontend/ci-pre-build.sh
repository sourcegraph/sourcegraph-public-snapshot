#!/bin/bash
#
### HACK(beyang): remove this when we move build scripts to enterprise. This exists solely as a stopgap to fix builds.

set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

cp ../sourcegraph/cmd/frontend/internal/app/templates/data_vfsdata.go ./vendor/github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/templates/
cp ../sourcegraph/cmd/frontend/internal/app/assets/assets_vfsdata.go ./vendor/github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets/
