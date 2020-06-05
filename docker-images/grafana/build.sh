#!/usr/bin/env bash

set -ex

cd "$(dirname "${BASH_SOURCE[0]}")"

# Copy over everything needed to build the monitoring-generator and grafana-wrapper.
# Since grafana-wrapper depends on internal/conf, which has a myriad of dependencies
# (many of which are internal and hence not go-get-able), we just copy the (almost)
# entire repository back into the build context. We use rsync to do this since it
# supports excluding files better.
rm -rf ./sourcegraph
rsync -r --exclude={'.*','docker-images','node_modules','browser','web','ui','doc','packages','migrations'} ../../ sourcegraph

BUILD_CACHE="--no-cache"
if [[ "$CACHE" == "true" ]]; then
  BUILD_CACHE=""
fi

# shellcheck disable=SC2086
docker build ${BUILD_CACHE} -t "${IMAGE:-sourcegraph/grafana}" . \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION

# Clean up for convenience
rm -rf ./sourcegraph
