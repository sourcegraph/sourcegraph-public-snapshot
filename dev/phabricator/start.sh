#!/usr/bin/env bash

set -e
unset CDPATH

oldest_supported="2017.09-r1"

version="$1"
if [ -z "$1" ]; then
  echo "Usage: dev/phabricator/start.sh <tag>"
  echo "where <tag> is a release version from https://hub.docker.com/r/bitnami/phabricator/tags/"
  echo "NOTE: the oldest release we support is ${oldest_supported}"
  echo
  echo "Continuing with oldest version supported..."
  echo

  version="${oldest_supported}"
fi

cd "$(dirname "${BASH_SOURCE[0]}")/../.." # cd to repo root dir

# shellcheck source=./shared.sh
source ./dev/phabricator/shared.sh

# Ensure old instances are cleaned up
if [ "$(docker ps -aq -f network=$network)" ]; then
  ./dev/phabricator/stop.sh

  docker volume rm $db_volume
  docker volume rm $app_volume
fi

# Create network if not exists
! (docker network ls | grep -q $network) && docker network create $network

# Create db volume if not exists
! (docker volume ls | grep -q $db_volume) && docker volume create --name $db_volume

# Create application data volume if not exists
! (docker volume ls | grep -q $app_volume) && docker volume create --name $app_volume

# Start the db
docker run -d --name $db_container -e ALLOW_EMPTY_PASSWORD=yes \
  --net $network \
  --volume $db_volume:/bitnami \
  bitnami/mariadb:latest

# Start the application
docker run -d --name $app_container -p 80:80 -p 443:443 \
  --net $network \
  --volume $app_volume:/bitnami \
  -e PHABRICATOR_USERNAME=admin \
  -e PHABRICATOR_PASSWORD=sourcegraph \
  -e PHABRICATOR_EMAIL=phabricator@sourcegraph.com \
  -e MARIADB_HOST=$db_container \
  bitnami/phabricator:"${version}"

echo
echo "Phabricator ${version} is now running at http://127.0.0.1"
echo "Login credentials: admin/sourcegraph"
echo "Exec into container: docker exec -it $(docker ps -aq -f name=phabricator$) /bin/bash"
echo "Phabricator root directory: /opt/bitnami/phabricator"
echo "To restart Phabricator, run: dev/phabricator/restart.sh"
echo "To stop Phabricator, run: dev/phabricator/stop.sh"
echo "Find more version to test at https://hub.docker.com/r/bitnami/phabricator/tags/"
