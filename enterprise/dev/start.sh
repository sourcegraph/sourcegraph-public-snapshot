#!/usr/bin/env bash

set -euf -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/..

DEV_PRIVATE_PATH=$PWD/../../dev-private

if [ ! -d "$DEV_PRIVATE_PATH" ]; then
    echo "Expected to find github.com/sourcegraph/dev-private checked out to $DEV_PRIVATE_PATH, but path wasn't a directory" 1>&2
    exit 1
fi

echo "Installing enterprise web dependencies..."
[ -n "${OFFLINE-}" ] || yarn --check-files

source "$DEV_PRIVATE_PATH/enterprise/dev/env"

# set to true if unset so set -u won't break us
: ${SOURCEGRAPH_COMBINE_CONFIG:=false}

export DEV_OVERRIDE_CRITICAL_CONFIG=$DEV_PRIVATE_PATH/enterprise/dev/critical-config.json
export DEV_OVERRIDE_SITE_CONFIG=$DEV_PRIVATE_PATH/enterprise/dev/site-config.json
export DEV_OVERRIDE_EXTSVC_CONFIG=$DEV_PRIVATE_PATH/enterprise/dev/external-services-config.json
export GOMOD_ROOT=$PWD
export PROCFILE=$PWD/dev/Procfile
export ENTERPRISE_COMMANDS="frontend management-console"
../dev/launch.sh
