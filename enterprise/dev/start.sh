#!/usr/bin/env bash

set -euf -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/..

DEV_PRIVATE_PATH=$PWD/../../dev-private

if [ ! -d "$DEV_PRIVATE_PATH" ]; then
    echo "Expected to find github.com/sourcegraph/dev-private checked out to $DEV_PRIVATE_PATH, but path wasn't a directory" 1>&2
    exit 1
fi

echo "Installing enterprise web dependencies..."
yarn --check-files

source "$DEV_PRIVATE_PATH/enterprise/dev/env"

# set to true if unset so set -u won't break us
: ${SOURCEGRAPH_COMBINE_CONFIG:=false}

SOURCEGRAPH_CONFIG_FILE=$DEV_PRIVATE_PATH/enterprise/dev/config.json GOMOD_ROOT=$PWD PROCFILE=$PWD/dev/Procfile ENTERPRISE_COMMANDS="frontend" ../dev/launch.sh
