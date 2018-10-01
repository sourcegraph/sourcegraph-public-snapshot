#!/bin/bash

set -e
unset CDPATH

cd "$(dirname "${BASH_SOURCE[0]}")/../.." # cd to repo root dir

source ./dev/phabricator/shared.sh

# if [ "$(docker ps -aq -f name=$db_container)" ];
# then
    # docker stop $db_container
    # docker rm $db_container
# fi

docker ps -aq -f network=$network | xargs docker stop | xargs docker rm

# if [ "$(docker ps -aq -f name=$app_container)" ];
# then
    # docker stop $app_container
    # docker rm $app_container
# fi
