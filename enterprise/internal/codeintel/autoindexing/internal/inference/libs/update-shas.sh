#!/bin/bash

set -u

DOCKER_USER=${DOCKER_USER:?"No DOCKER_USER is set."}
DOCKER_PASS=${DOCKER_PASS:?"No DOCKER_PASS is set."}

for indexer in lsif-clang lsif-go lsif-rust scip-java scip-python scip-typescript scip-ruby; do
  tag="latest"
  if [[ "${indexer}" = "scip-python" ]] || [[ "${indexer}" = "scip-typescript" || "${indexer}" = "scip-ruby" ]]; then
    tag="autoindex"
  fi

  sha=$(docker manifest inspect sourcegraph/${indexer}:${tag} -v | jq -s .[0].Descriptor.digest)

  sed -i.bak \
    "s|\("'"'"sourcegraph/${indexer}"'"'":\).*|\1${sha},|g" \
    indexes.go

  echo "Updated tag for ${indexer}"
  rm indexes.go.bak
done

go fmt indexes.go
