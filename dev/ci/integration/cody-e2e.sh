#!/usr/bin/env bash

# shellcheck disable=SC1091
source /root/.profile
set -ex

# ==========================
echo "--- starting dbus session"
export DBUS_SESSION_BUS_ADDRESS=$(dbus-daemon --session --print-address --fork)
export DISPLAY=":99"

echo "--- start cody e2e"
pnpm install --frozen-lockfile --fetch-timeout 60000
pnpm --filter cody-ai run test:integration
