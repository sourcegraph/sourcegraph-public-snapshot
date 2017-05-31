#!/bin/bash

set -e

mydir="$(dirname "${BASH_SOURCE[0]}")"
cd $mydir
go run up.go -f "../local-installer/sourcegraph/docker-compose.yml" -o . $@

echo '';
echo 'Hit [ENTER] to run `docker-compose up`';
echo "  > If this doesn't work, try re-running one or more of the host commands after \`docker-compose up\`.";
echo "  > If this still doesn't work, consult the native run commands or config (Procfile, dev/server.sh) to see if something changed.";
read;

docker-compose up;
