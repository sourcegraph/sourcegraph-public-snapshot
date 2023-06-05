#!/bin/bash

set -u

DOCKER_USER=${DOCKER_USER:?"No DOCKER_USER is set."}
DOCKER_PASS=${DOCKER_PASS:?"No DOCKER_PASS is set."}

for indexer in lsif-clang scip-go lsif-rust scip-rust scip-java scip-python scip-typescript scip-ruby; do
  tag="latest"
  if [[ "${indexer}" = "scip-python" ]] || [[ "${indexer}" = "scip-typescript" || "${indexer}" = "scip-ruby" ]]; then
    tag="autoindex"
  fi

  sha=$(docker buildx imagetools inspect sourcegraph/${indexer}:${tag} --raw | sha256sum | awk '{print "\"" "sha256:" $1 "\""}')

  sed -i.bak \
    "s|\("'"'"sourcegraph/${indexer}"'"'":\).*|\1${sha},|g" \
    indexes.go

  echo "Updated tag for ${indexer}"
  rm indexes.go.bak
done

go fmt indexes.go
