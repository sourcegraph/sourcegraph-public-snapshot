#!/usr/bin/env bash

set -e

docker exec -it "$(docker ps -aq -f name=phabricator$)" sh -c "cd /opt/bitnami/phabricator && bin/phd restart"
