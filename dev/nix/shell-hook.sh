#!/usr/bin/env bash

# This file is called by nix-shell when setting up the shell. It is
# responsible for setting up the development environment outside of what nix's
# package management.
#
# The main goal of this is to start stateful services which aren't managed by
# sourcegraph's developer tools. In particular this is our databases, which
# are used by both our tests and development server.

pushd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null || exit

# TODO build with nix
NIX_ENFORCE_PURITY=0 ../libsqlite3-pcre/build.sh

. ./start-postgres.sh
. ./start-redis.sh

# We disable postgres_exporter since it expects postgres to be running on TCP.
export SRC_DEV_EXCEPT="${SRC_DEV_EXCEPT:-postgres_exporter}"

popd >/dev/null || exit
