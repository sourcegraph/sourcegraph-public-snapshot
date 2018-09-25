#!/bin/bash
#
### HACK(beyang): remove this when we move build scripts to enterprise. This exists solely as a stopgap to fix builds.

set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

./cmd/frontend/ci-pre-build.sh
