#!/bin/bash

# This runs the latest version of server that has passed CI on docker-images/server

DATA=/tmp/sourcegraph

echo -n "Do you want to delete $DATA and start clean? [Y/n] "
read clean 

if [ "$clean" != "n" ] && [ "$clean" != "N" ]; then
    echo "deleting $DATA"
    rm -rf $DATA
fi

echo "deleting old docker images..."
docker rmi us.gcr.io/sourcegraph-dev/server

echo "creating lsp network bridge..."
docker network create --driver bridge lsp

echo "starting server..."
gcloud docker -- run \
 --publish 7080:7080 --rm \
 --network lsp \
 --name sourcegraph \
 --volume $DATA/config:/etc/sourcegraph \
 --volume $DATA/data:/var/opt/sourcegraph \
 us.gcr.io/sourcegraph-dev/server:latest
