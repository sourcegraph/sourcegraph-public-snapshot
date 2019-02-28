#!/bin/bash

cd $(dirname "${BASH_SOURCE[0]}")/../migrations
set -e

# The name is intentionally empty ('') so that it forces a merge conflict if two branches attempt to
# create a migration at the same sequence number (because they will both add a file with the same
# name, like `migrations/1528277032_.up.sql`).

migrate create -ext sql -dir . -digits 10 -seq ''

files=$(ls -1 | grep '^[0-9]'.*\.sql | sort -n | tail -n2)

for f in $files; do
    cat > $f <<EOF
BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

COMMIT;
EOF
   
    echo "Created migrations/$f"
done
