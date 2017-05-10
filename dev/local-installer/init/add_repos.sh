#!/bin/bash
set -x
set -e
echo "this script imports non-GitHub repositories into Sourcegraph"

if [ $# -ne 1 ]; then
    echo $0: usage: ./setup.sh repo-list-path
    exit 1
fi

REPO_LIST=$1

for repo in $(cat $REPO_LIST); do
    set +e
    for i in {1..3}; do # retry max 3 times (might fail due to docker-compose initialization race)
        curl -XPOST "${FRONTEND_URL}/.api/repos-ensure" -d "[\"${repo}\"]"
        if [ $? -eq 0 ]; then
            break;
        fi
        sleep 1;
    done;
    set -e
done;
