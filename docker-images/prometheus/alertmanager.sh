#!/bin/sh
set -e

# Bazel migration moves the binary
if [ -x /usr/bin/alertmanager ]; then
  # shellcheck disable=SC2086
  exec /usr/bin/alertmanager --storage.path=/alertmanager --data.retention=168h $ALERTMANAGER_ADDITIONAL_FLAGS "$@"
else
  # shellcheck disable=SC2086
  exec /bin/alertmanager --storage.path=/alertmanager --data.retention=168h $ALERTMANAGER_ADDITIONAL_FLAGS "$@"
fi
