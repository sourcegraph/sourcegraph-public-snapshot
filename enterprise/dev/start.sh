#!/usr/bin/env bash

set -euf -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"/..

DEV_PRIVATE_PATH=$PWD/../../dev-private

if [ ! -d "$DEV_PRIVATE_PATH" ]; then
  echo "Expected to find github.com/sourcegraph/dev-private checked out to $DEV_PRIVATE_PATH, but path wasn't a directory" 1>&2
  exit 1
fi

# Warn if dev-private needs to be updated.
if [ -f "$DEV_PRIVATE_PATH/enterprise/dev/critical-config.json" ]; then
  echo "Error: You need to update dev-private to a commit that incorporates https://github.com/sourcegraph/dev-private/pull/10."
  echo
  echo "Try running:"
  echo
  echo "    cd $DEV_PRIVATE_PATH && git pull"
  echo
  exit 1
fi

# shellcheck disable=SC1090
source "$DEV_PRIVATE_PATH/enterprise/dev/env"

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
