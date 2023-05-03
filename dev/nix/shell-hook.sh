#!/usr/bin/env bash

# This file is called by nix-shell when setting up the shell. It is
# responsible for setting up the development environment outside of what nix's
# package management.
#
# The main goal of this is to start stateful services which aren't managed by
# sourcegraph's developer tools. In particular this is our databases, which
# are used by both our tests and development server.

if [ -f /etc/NIXOS ]; then
  cat <<EOF > .bazelrc-nix
build --host_platform=@rules_nixpkgs_core//platforms:host
build --extra_toolchains=@nixpkgs_nodejs_toolchain//:nodejs_nix,@nixpkgs_rust_toolchain//:rust_nix
EOF
fi

pushd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null || exit

. ./start-postgres.sh
. ./start-redis.sh

# We disable postgres_exporter since it expects postgres to be running on TCP.
export SRC_DEV_EXCEPT="${SRC_DEV_EXCEPT:-postgres_exporter}"

popd >/dev/null || exit
