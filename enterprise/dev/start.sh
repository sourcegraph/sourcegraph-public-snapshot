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

export CRITICAL_CONFIG_FILE=$DEV_PRIVATE_PATH/enterprise/dev/critical-config.json
export SITE_CONFIG_FILE=$DEV_PRIVATE_PATH/enterprise/dev/site-config.json
export EXTSVC_CONFIG_FILE=$DEV_PRIVATE_PATH/enterprise/dev/external-services-config.json
export GLOBAL_SETTINGS_FILE=$PWD/../dev/global-settings.json
export SITE_CONFIG_ALLOW_EDITS=true
export GLOBAL_SETTINGS_ALLOW_EDITS=true
export EXTSVC_CONFIG_ALLOW_EDITS=true

export WATCH_ADDITIONAL_GO_DIRS="$PWD/cmd $PWD/dev $PWD/internal"
export ENTERPRISE_COMMANDS="frontend repo-updater"
export ENTERPRISE=1
../dev/start.sh
