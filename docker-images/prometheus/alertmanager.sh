#!/bin/sh
set -e

# shellcheck disable=SC2086
exec /bin/alertmanager --storage.path=/alertmanager $ALERTMANAGER_ADDITIONAL_FLAGS "$@"
