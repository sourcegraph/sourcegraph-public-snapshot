#!/bin/bash
set -x
set -e
echo "this script imports non-GitHub repositories into Sourcegraph"

if [ $# -ne 1 ]; then
    echo $0: usage: ./setup.sh repo-list-path
    exit 1
fi

REPO_LIST=$1

counter=10000
for repo in $(cat $REPO_LIST); do
    if [[ $repo == github* ]];
    then
	echo "Skipping $repo, because it's from github."
	continue
    fi
    set +e
    docker-compose exec postgres psql -U sg -c "INSERT INTO repo VALUES (${counter},'${repo}','','${repo}','','','','','','master','',false,false,false,false,false,now(),now(),now(),now(),null,false,'',0,'');";
    set -e
    counter=$((counter+1))
done;

