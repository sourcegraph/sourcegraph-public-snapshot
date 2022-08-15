#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -ex

export GO111MODULE=on

pushd ./dev/sg
go build -o ../../tmp-sg
popd

function finish {
  rm -f ./tmp-sg
}
trap finish EXIT

# Squash migrations and create new SQL file; leave database as-is so we can re-describe in different formats
./tmp-sg migration squash-all -skip-teardown -db frontend -f migrations/frontend/squashed.sql
./tmp-sg migration squash-all -skip-teardown -db codeintel -f migrations/codeintel/squashed.sql
./tmp-sg migration squash-all -skip-teardown -db codeinsights -f migrations/codeinsights/squashed.sql

# Primarily for nix users, where postgres is listening on a unix socket
set +e -x
test -d "$PGHOST"
IS_UNIX_POSTGRES=$?
set -ex

if [ $IS_UNIX_POSTGRES -eq 0 ]; then
  if [ -z "$PGDATASOURCE" ]; then
    echo "\$PGDATASOURCE expected to be set when \$PGHOST points to the filesystem."
    exit 1
  fi
  PGDATASOURCE_BASE=$PGDATASOURCE
  export PGDATASOURCE="${PGDATASOURCE_BASE}&dbname=sg-squasher-frontend"
  export CODEINTEL_PGDATASOURCE="${PGDATASOURCE_BASE}&dbname=sg-squasher-codeintel"
  export CODEINSIGHTS_PGDATASOURCE="${PGDATASOURCE_BASE}&dbname=sg-squasher-codeinsights"
else
  export PGDATASOURCE="postgres://${PGUSER}:${PGPASSWORD}@${PGHOST}:${PGPORT}/sg-squasher-frontend"
  export CODEINTEL_PGDATASOURCE="postgres://${PGUSER}:${PGPASSWORD}@${PGHOST}:${PGPORT}/sg-squasher-codeintel"
  export CODEINSIGHTS_PGDATASOURCE="postgres://${PGUSER}:${PGPASSWORD}@${PGHOST}:${PGPORT}/sg-squasher-codeinsights"
fi

# Output-psql formatted schema description
./tmp-sg migration describe -db frontend --format=psql -force -out internal/database/schema.md
./tmp-sg migration describe -db codeintel --format=psql -force -out internal/database/schema.codeintel.md
./tmp-sg migration describe -db codeinsights --format=psql -force -out internal/database/schema.codeinsights.md

# Output json-formatted schema description
./tmp-sg migration describe -db frontend --format=json -force -out internal/database/schema.json
./tmp-sg migration describe -db codeintel --format=json -force -out internal/database/schema.codeintel.json
./tmp-sg migration describe -db codeinsights --format=json -force -out internal/database/schema.codeinsights.json
