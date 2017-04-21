#!/bin/bash

./add_repos.sh <(./find_local.sh)

if [ ! -z "$ENSURE_REPOS_REMOTE" ]; then
    echo "Adding remote repositories..."
    ./add_repos.sh <(echo $ENSURE_REPOS_REMOTE)
    echo "Done adding remote repositories..."
fi

# Wait around indefinitely
while true; do sleep 86400; done;
