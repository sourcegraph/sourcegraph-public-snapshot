#!/usr/bin/env bash

# This file is called by nix-shell when setting up the shell. It is
# responsible for setting up the development environment outside of what nix's
# package management.
#
# The main goal of this is to start stateful services which aren't managed by
# sourcegraph's developer tools. In particular this is our databases, which
# are used by both our tests and development server.

set -eu

cd "$(dirname "${BASH_SOURCE[0]}")"

# TODO build with nix
if [ ! -e ../../libsqlite3-pcre.so ]; then
  echo 'Building libsqlite3-pcre...'
  NIX_ENFORCE_PURITY=0 ../libsqlite3-pcre/build.sh
fi

. ./start-postgres.sh
