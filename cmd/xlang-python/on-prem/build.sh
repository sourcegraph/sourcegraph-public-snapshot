#!/bin/bash

die () {
    echo >&2 "$@"
    exit 1
}

[ "$#" -eq 1 ] || die "Please provide a version with which to tag this image"

set -e
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=us.gcr.io/sourcegraph-dev/xlang-python
export TAG=$1

set -x

if [ ! -d "python-langserver" ]; then
    git clone https://github.com/sourcegraph/python-langserver python-langserver
else
    cd python-langserver && git checkout master && git pull origin master && cd ..
fi

docker build -t $IMAGE:$TAG .
gcloud docker -- push $IMAGE:$TAG
