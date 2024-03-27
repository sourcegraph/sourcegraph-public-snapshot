#!/usr/bin/env bash

# This file is called by nix-shell when setting up the shell. It is
# responsible for setting up the development environment outside of what nix's
# package management.
#
# The main goal of this is to start stateful services which aren't managed by
# sourcegraph's developer tools. In particular this is our databases, which
# are used by both our tests and development server.

pushd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null || exit

. ./start-postgres.sh
. ./start-redis.sh

# We disable postgres_exporter since it expects postgres to be running on TCP.
export SRC_DEV_EXCEPT="${SRC_DEV_EXCEPT:-postgres_exporter}"

popd >/dev/null || exit

# Empty out .bazelrc-nix
echo -n > .bazelrc-nix

# Use a hermetic and non-host CC compiler on NixOS.
if [ -f /etc/NIXOS ]; then
  cat <<EOF > .bazelrc-nix
build --extra_toolchains=@zig_sdk//toolchain:linux_amd64_gnu.2.34
EOF
fi

# without this, zig cc is forced to rebuild on every sandboxed GoLink action, which
# adds ~1m of time to GoLink actions. The reason it's on _every_ GoLink action is
# because sandboxes are ephemeral and don't persist non-mounted paths between actions.
mkdir -p /tmp/zig-cache

# We run this check afterwards so we can read the values exported by the
# start-*.sh scripts. We need to smuggle in these envvars for tests on both
# linux and darwin.
cat <<EOF >> .bazelrc-nix
build --action_env=PATH=$BAZEL_ACTION_PATH
build --action_env=REDIS_ENDPOINT
build --action_env=PGHOST
build --action_env=PGDATA
build --action_env=PGDATABASE
build --action_env=PGDATASOURCE
build --action_env=PGUSER
build --sandbox_add_mount_pair=/tmp/zig-cache
build --sandbox_writable_path=/tmp/zig-cache
build --noincompatible_sandbox_hermetic_tmp
EOF
