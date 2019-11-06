#!/usr/bin/env bash
# "Experimental" - warm up the cache by attempting to build sourcegraph once

set -exo pipefail

DOWNLOAD_DIR=`mktemp -d -t sourcegraph_builder_XXXXXXX`
cleanup() {
    rm -rf "$DOWNLOAD_DIR"
}
trap cleanup EXIT

# Set the default python version to 2.7
# (for sourcegraph-frontend's node-sass depdendency)
ln -s /usr/bin/python2.7 /usr/bin/python
trap "rm /usr/bin/python" EXIT

SOURCEGRAPH_COMMIT=${SOURCEGRAPH_COMMIT:-e93d3379f68441062b54ec219f44046c280a0b18}

curl -L https://github.com/sourcegraph/sourcegraph/archive/${SOURCEGRAPH_COMMIT}.tar.gz | tar xz -C $DOWNLOAD_DIR --strip-components 1
cd $DOWNLOAD_DIR

./enterprise/cmd/server/pre-build.sh
