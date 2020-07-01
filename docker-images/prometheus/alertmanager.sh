#!/bin/sh

# shellcheck disable=SC2086
exec /bin/prometheus --storage.path=/alertmanager $ALERTMANAGER_ADDITIONAL_FLAGS "$@"
