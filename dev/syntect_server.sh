#!/usr/bin/env bash

set -euf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

if [[ ${USE_SYNTECT_SERVER_FROM_PATH-} == t* ]]; then
  # NB: this is NOT the common path - below is.
  export QUIET='true'
  export ROCKET_SECRET_KEY="SeerutKeyIsI7releuantAndknvsuZPluaseIgnorYA="
  export ROCKET_ENV="production"
  export ROCKET_LIMITS='{json=10485760}'
  export ROCKET_PORT=9238
  if [[ "${INSECURE_DEV:-}" == '1' ]]; then
    export ROCKET_ADDRESS='127.0.0.1'
  fi
  exec syntect_server
fi

addr=()
if [[ "${INSECURE_DEV:-}" == '1' ]]; then
  addr+=("-e" "ROCKET_ADDRESS=0.0.0.0")
fi

docker inspect syntect_server >/dev/null 2>&1 && docker rm -f syntect_server
exec docker run --name=syntect_server --rm -p9238:9238 -e WORKERS=1 "${addr[@]}" sourcegraph/syntect_server:9089f98@sha256:83ff65809e6647b466bd400de4c438a32feeabe8e791b12e15c67c84529ad2de
