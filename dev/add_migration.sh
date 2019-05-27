#!/bin/bash

cd $(dirname "${BASH_SOURCE[0]}")/../migrations
set -e

if [ -z "$1" ]; then
    echo "USAGE: $0 <name>"
    exit 1
fi

migrate create -ext sql -dir . -digits 10 -seq "$1"

files=$(ls -1 | grep '^[0-9]'.*\.sql | sort -n | tail -n2)

for f in $files; do
    cat > $f <<EOF
BEGIN;

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
