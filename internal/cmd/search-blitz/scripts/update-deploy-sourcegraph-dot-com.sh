#!/usr/bin/env bash
set -euo pipefail
pushd "$(dirname "${BASH_SOURCE[0]}")/../../../.." >/dev/null

sed -i "" "s/search-blitz:.\{1,2\}\..\{1,2\}\..\{1,2\}/search-blitz:$1/" ../deploy-sourcegraph-dot-com/configure/search-blitz/search-blitz.StatefulSet.yaml

popd >/dev/null
