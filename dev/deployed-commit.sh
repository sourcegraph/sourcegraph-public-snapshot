#!/bin/bash

# This script determines the commit of sourcegraph/sourcegraph that is currently
# running on sourcegraph.com. The returned version has one of the following
# formats:
#
#   - v{semver}
#   - {buildkite build num}_{date}_{abbreviated commit}
#
# In either case, the following cut will give us something parseable by rev-parse.

set -o pipefail
git rev-parse "$(curl -sf https://sourcegraph.com/__version | cut -d'_' -f3)"
