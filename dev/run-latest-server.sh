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
IMAGE=sourcegraph/server:insiders
docker pull $IMAGE

echo "starting server..."
gcloud docker -- run "$@" \
 --publish 7080:7080 --rm \
 --network lsp \
 --name sourcegraph \
 --volume $DATA/config:/etc/sourcegraph \
 --volume $DATA/data:/var/opt/sourcegraph \
 -v /var/run/docker.sock:/var/run/docker.sock \
 $IMAGE
