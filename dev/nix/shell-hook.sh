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
build --action_env=PATH=$BAZEL_ACTION_PATH
EOF
  if [ ! -f /lib64/ld-linux-x86-64.so.2 ]; then
    echo "---------------------------------------------------------------------------------------------------------------------------------------"
    echo "⚠️ Bazel provisions binaries that expect the program interpreter (ld.so) at /lib64/ld-linux-x86-64.so.2, but it wasn't found.       ⚠️"
    echo "⚠️ You may need to enable nix-ld[1] or symlink /lib64/ld-linux-x86-64.so.2 to a glibc's loader[2] if you intend to use bazel.       ⚠️"
    echo "⚠️ [1] https://nixos.org/manual/nixos/unstable/options.html#opt-programs.nix-ld.enable                                              ⚠️"
    echo "⚠️ [2] https://sourcegraph.com/github.com/Strum355/nix-dotfiles@7951c755/-/blob/hosts/noah-nixos-desktop/configuration.nix?L272-276 ⚠️"
    echo "---------------------------------------------------------------------------------------------------------------------------------------"
  fi
fi

pushd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null || exit

. ./start-postgres.sh
. ./start-redis.sh

# We disable postgres_exporter since it expects postgres to be running on TCP.
export SRC_DEV_EXCEPT="${SRC_DEV_EXCEPT:-postgres_exporter}"

popd >/dev/null || exit
