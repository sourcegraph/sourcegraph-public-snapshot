#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."
set -euo pipefail

revision="${1:?no revision provided}"
outdir="${2:?no output directory provided}"

if [[ -f "${outdir}/${revision}-internal_database_schema.json" ]]; then
  echo "exists!"
  exit 0
fi

dropdbs() {
  dropdb --if-exists sg-squasher-frontend 2>/dev/null
  dropdb --if-exists sg-squasher-codeintel 2>/dev/null
  dropdb --if-exists sg-squasher-codeinsights 2>/dev/null
  docker stop squasher >/dev/null 2>&1 || true
}

cleanup() {
  rm -rf temp-squash
  git checkout -- migrations
  git clean -qfd migrations
  dropdbs
}
trap cleanup EXIT

export PGDATASOURCE="postgres://${PGUSER}:${PGPASSWORD}@${PGHOST}:${PGPORT}/sg-squasher-frontend"
export CODEINTEL_PGDATASOURCE="postgres://${PGUSER}:${PGPASSWORD}@${PGHOST}:${PGPORT}/sg-squasher-codeintel"
export CODEINSIGHTS_PGDATASOURCE="postgres://${PGUSER}:${PGPASSWORD}@${PGHOST}:${PGPORT}/sg-squasher-codeinsights"

codeinsights_container_args=""
if (("${revision:3:2}" < 37)); then
  # If minor version < 37, launch and target TimescaleDB container
  codeinsights_container_args="-in-timescaledb-container"
  export CODEINSIGHTS_PGDATASOURCE="postgres://postgres:password@${PGHOST}:5433/codeinsights"
fi

go build ./dev/sg # Currently requires migration at compile time; doesn't read from disk
echo "Rewriting migration definitions as they were at ${revision}..."
./sg migration rewrite -db frontend -rev "${revision}" >/dev/null
./sg migration rewrite -db codeintel -rev "${revision}" >/dev/null
./sg migration rewrite -db codeinsights -rev "${revision}" >/dev/null

go build ./dev/sg # Currently requires migration at compile time; doesn't read from disk
echo "Squashing migration definitions as they were at ${revision}..."
dropdbs
./sg migration squash-all -skip-data -skip-teardown -db frontend -f temp-squash
./sg migration squash-all -skip-data -skip-teardown -db codeintel -f temp-squash
if [[ "${codeinsights_container_args}" == "" ]]; then
  ./sg migration squash-all -skip-data -skip-teardown -db codeinsights -f temp-squash
else
  ./sg migration squash-all -skip-data -skip-teardown -db codeinsights -f temp-squash "${codeinsights_container_args}"
fi

echo "Describing migration definitions as they were at ${revision}..."
./sg migration describe -db frontend --format=json -force -out "${outdir}/${revision}-internal_database_schema.json"
./sg migration describe -db codeintel --format=json -force -out "${outdir}/${revision}-internal_database_schema.codeintel.json"
./sg migration describe -db codeinsights --format=json -force -out "${outdir}/${revision}-internal_database_schema.codeinsights.json"
