#!/usr/bin/env bash

set -e
pushd "$(dirname "${BASH_SOURCE[0]}")/../.." >/dev/null

gsutil cp ui/assets/webpack.manifest.json gs://sourcegraph-assets/
