#!/usr/bin/env bash

set -euxo pipefail

# This migration script gives the "sourcegraph" user ownership over all the folders
# in Sourcegraph's external volumes so that the container itself can be run
# as a non-"root" user ("sourcegraph").
#
# This script expects to run as root, and it can be deleted in the future once all of our users
# have ran through this migration process.

if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root."
    exit 1
fi

chown -R sourcegraph:sourcegraph /etc/sourcegraph
chown -R sourcegraph:sourcegraph /var/opt/sourcegraph

exec runuser -u sourcegraph "$@"
