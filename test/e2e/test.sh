#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.." || exit
set -x

# shellcheck disable=SC1091
source /root/.profile
bash test/setup-deps.sh
bash test/setup-display.sh

# ==========================

asdf install
yarn
yarn generate
pushd enterprise || exit
./cmd/server/pre-build.sh
./cmd/server/build.sh
popd || exit
./dev/ci/e2e.sh
docker image rm -f "${IMAGE}"
