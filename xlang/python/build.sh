#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=us.gcr.io/sourcegraph-dev/xlang-python
export TAG=${TAG-latest}

set -x

if [ ! -d "python-langserver" ]; then
    git clone https://github.com/sourcegraph/python-langserver python-langserver
else
    cd python-langserver && git pull && cd ..
fi

docker build -t $IMAGE:$TAG .
gcloud docker -- push $IMAGE:$TAG
