#!/usr/bin/env bash

# This file is required, because docker images are always assumed to live in the root of the cmd directory.
# Since we build various versions of executor images, we want to forward this for now and might reconsider
# making this an option in our CI framework at some point.

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"/docker-image

./build.sh "$@"

#cd "$(dirname "${BASH_SOURCE[0]}")"/kubernetes
#
#./build.sh "$@"
