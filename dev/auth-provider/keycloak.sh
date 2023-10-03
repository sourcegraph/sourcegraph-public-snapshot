#!/usr/bin/env bash

# This script is a wrapper around Keycloak's Docker image.

set -e

if [ -z "$USE_KEYCLOAK" ]; then
  echo Not using Keycloak. Keycloak authentication providers will not work.
  exit 0
fi

unset CDPATH
cd "$(dirname "${BASH_SOURCE[0]}")" # cd to dev/auth-provider dir

source scripts/common.sh

# Kill Keycloak container on exit.
function finish() {
  docker rm -f "$CONTAINER"
}
trap finish EXIT

# Pull Keycloak image separately from running it, so that our wait-until-ready step below doesn't
# fail if the one-time pull takes longer.
docker inspect "$IMAGE" >/dev/null 2>&1 || docker pull "$IMAGE"

# Run Keycloak server.
docker inspect "$CONTAINER" >/dev/null 2>&1 && docker rm -f "$CONTAINER"
docker run \
  --name="$CONTAINER" \
  --interactive \
  --detach --rm \
  --publish 3220:8080 \
  -e DB_VENDOR=H2 \
  -e KEYCLOAK_USER="$KEYCLOAK_USER" \
  -e KEYCLOAK_PASSWORD="$KEYCLOAK_PASSWORD" \
  -e KEYCLOAK_LOGLEVEL="${KEYCLOAK_LOGLEVEL-INFO}" \
  "$IMAGE" \
  >/dev/null

# Wait for Keycloak server to be ready.
echo Waiting for Keycloak server to be ready...
for i in $(seq 1 20); do
  curl --retry 60 --retry-delay 2 --retry-max-time 120 -sSL "$KEYCLOAK" >/dev/null 2>&1 && break
  sleep 2
done
curl -sSL "$KEYCLOAK" >/dev/null
echo Configuring Keycloak...

# Set up users, clients, etc.
RESET=1 scripts/configure-keycloak.sh

echo Keycloak is ready: "$KEYCLOAK"
exec docker attach --no-stdin "$CONTAINER" | (
  trap '' 2
  while read -r i; do echo "$i"; done
)
