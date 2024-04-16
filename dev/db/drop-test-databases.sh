#!/usr/bin/env bash

set -eo pipefail

DB_NAMES=$(psql -Xc "copy (select datname from pg_database where datname like 'sourcegraph-test-%') to stdout")
for dbname in $DB_NAMES; do
  dropdb "$dbname"
  echo "dropped $dbname"
done
