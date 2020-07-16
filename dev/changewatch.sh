#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

# Wrapper for watchman. To debug which changes it detect set the environment
# variable WATCHMAN_DEBUG=t

if [ ! -x "$(command -v watchman)" ]; then
  echo "Please install watchman"
  echo
  echo "  brew install watchman"
  exit 1
fi

set -e
pushd dev/watchmanwrapper
go build
popd

exec dev/watchmanwrapper/watchmanwrapper dev/handle-change.sh <<-EOT
["subscribe", ".", "gochangewatch", {
  "expression": ["allof",
    ["not", ["anyof",
      ["match", ".*"],
      ["suffix", "_test.go"]]],
    ["anyof",
      ["suffix", "go"],
      ["dirname", "cmd/symbols"],
      ["dirname", "schema"],
      ["dirname", "docker-images/grafana/jsonnet"],
      ["dirname", "monitoring"],
      ["name", "cmd/frontend/graphqlbackend/schema.graphql", "wholename"]]],
  "fields": ["name"]
}]
EOT
