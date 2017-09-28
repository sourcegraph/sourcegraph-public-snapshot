#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=${IMAGE-us.gcr.io/sourcegraph-dev/xlang-python}

set -x

if [ ! -d "python-langserver" ]; then
    git clone https://github.com/sourcegraph/python-langserver python-langserver
else
    cd python-langserver && git checkout master && git pull origin master && cd ..
fi

docker build -t $IMAGE .
