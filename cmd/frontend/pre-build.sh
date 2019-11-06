#!/bin/bash
set -euxo pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../..

# Start postgres (for the dev/generate.sh scripts)
gosu postgres /usr/lib/postgresql/9.6/bin/pg_ctl initdb
## Allow pgsql to listen to all IPs
## See https://stackoverflow.com/a/52381997 for more information
## Try changing lock location to /tmp since Kaniko doesn't allow writes to '/var/run'
## See https://github.com/GoogleContainerTools/kaniko/issues/506 and https://forums.postgresql.fr/viewtopic.php?id=3984
gosu postgres /usr/lib/postgresql/9.6/bin/pg_ctl -o "-c listen_addresses='*' -c unix_socket_directories='/tmp'"  -w start
export PGHOST='/tmp'

# Build the webapp typescript code.
echo "--- yarn"
yarn --frozen-lockfile --network-timeout 60000

pushd web
echo "--- yarn run build"
NODE_ENV=production DISABLE_TYPECHECKING=true yarn run build
popd

echo "--- go generate"
go generate ./cmd/frontend/internal/app/assets ./cmd/frontend/internal/app/templates
