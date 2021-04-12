#!/usr/bin/env bash
set -euo pipefail

sed -i "" "s/search-blitz:.\{1,2\}\..\{1,2\}\..\{1,2\}/search-blitz:$1/" ../deploy-sourcegraph-dot-com/configure/search-blitz/search-blitz.StatefulSet.yaml
