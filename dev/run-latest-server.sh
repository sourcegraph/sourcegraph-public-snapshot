#!/bin/bash

# This runs the latest version of server that has passed CI on docker-images/server

DATA=/tmp/sourcegraph

echo -n "Do you want to delete $DATA and start clean? [Y/n] "
read clean

if [ "$clean" != "n" ] && [ "$clean" != "N" ]; then
    echo "deleting $DATA"
    rm -rf $DATA
fi

echo "creating lsp network bridge..."
docker network create --driver bridge lsp

echo "pulling new docker image..."
IMAGE=us.gcr.io/sourcegraph-dev/server:latest
docker pull us.gcr.io/sourcegraph-dev/server:latest

echo "starting server..."
gcloud docker -- run "$@" \
 --publish 7080:7080 --rm \
 --network lsp \
 --name sourcegraph \
 --volume $DATA/config:/etc/sourcegraph \
 --volume $DATA/data:/var/opt/sourcegraph \
 -v /var/run/docker.sock:/var/run/docker.sock \
 $IMAGE
