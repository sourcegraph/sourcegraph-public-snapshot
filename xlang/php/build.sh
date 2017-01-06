#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=us.gcr.io/sourcegraph-dev/xlang-php
export TAG=${TAG-latest}

composer install --prefer-dist --no-interaction --no-progress --no-plugins
composer run-script parse-stubs

docker build -t $IMAGE:$TAG .
gcloud docker -- push $IMAGE:$TAG

# Also push latest if on CI
[ -z "$CI" ] || (docker tag $IMAGE:$TAG $IMAGE:latest && gcloud docker -- push $IMAGE:latest)
