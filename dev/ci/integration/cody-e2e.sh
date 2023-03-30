#!/usr/bin/env bash

# shellcheck disable=SC1091
source /root/.profile
set -ex

mkdir -p /var/run/dbus
echo "--- starting dbus session"
export DISPLAY=":99"
export DBUS_SESSION_BUS_ADDRESS=$(dbus-daemon --system --address "unix:path=/run/dbus/system_bus_socket" --print-address --fork)

echo "--- start cody e2e"
pnpm install --frozen-lockfile --fetch-timeout 60000
pnpm --filter cody-ai run test:integration
