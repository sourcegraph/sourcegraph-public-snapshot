#!/bin/bash

set -e

mydir="$(dirname "${BASH_SOURCE[0]}")"
cd $mydir
go run up.go -f "../local-installer/sourcegraph/docker-compose.yml" $@ > docker-compose.yml

echo 'Hit [ENTER] to run `docker-compose up`';
read;

docker-compose up;
