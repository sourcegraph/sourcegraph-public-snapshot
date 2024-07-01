#!/usr/bin/env bash

set -u

if [ -z "${DOCKER_USER:-}" ]; then
  echo "warning: DOCKER_USER is not set; may hit Docker rate limit"
fi

if [ -z "${DOCKER_PASS:-}" ]; then
  if [ -n "${DOCKER_USER:-}" ]; then
    echo "error: DOCKER_USER set but DOCKER_PASS was not set"
    exit 1
  fi
fi

SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"

# No scip-clang as that doesn't have a Docker image
for indexer in scip-go scip-rust scip-java scip-python scip-typescript scip-ruby scip-dotnet; do
  tag="latest"
  if [[ "${indexer}" = "scip-python" ]] || [[ "${indexer}" = "scip-typescript" || "${indexer}" = "scip-ruby" ]]; then
    tag="autoindex"
  fi

  sha=$(docker buildx imagetools inspect sourcegraph/${indexer}:${tag} --raw | sha256sum | awk '{print "\"" "sha256:" $1 "\""}')

  sed -i.bak \
    "s|\("'"'"sourcegraph/${indexer}"'"'":\).*|\1${sha},|g" \
    "$SCRIPT_DIR/indexes.go"

  echo "Updated tag for ${indexer}"
  rm "$SCRIPT_DIR/indexes.go.bak"
done

go fmt "$SCRIPT_DIR/indexes.go"

echo "Updating SHAs in test snapshots"
go test "$SCRIPT_DIR/../..." -update
