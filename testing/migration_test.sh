#!/usr/bin/env bash

set -eu

LOAD_MIGRATOR_IMAGE="$1"
TEST_VERSION="$2"

function cleanup {
  echo "Cleaning our mess"
  docker kill wg_pgsql >/dev/null 2>&1 || true
  docker kill wg_codeintel-db >/dev/null 2>&1 || true
  docker kill wg_codeinsights-db >/dev/null 2>&1 || true
  docker kill wg_migrator >/dev/null 2>&1 || true
  docker network rm wg_test >/dev/null 2>&1 || true
}

trap cleanup EXIT

function run_migrator {
  docker run \
    --rm \
    --name wg_migrator \
    --platform linux/amd64 \
    -e PGHOST='wg_pgsql' \
    -e PGPORT='5432' \
    -e PGUSER='sg' \
    -e PGPASSWORD='sg' \
    -e PGDATABASE='sg' \
    -e PGSSLMODE='disable' \
    -e CODEINTEL_PGHOST='wg_codeintel-db' \
    -e CODEINTEL_PGPORT='5432' \
    -e CODEINTEL_PGUSER='sg' \
    -e CODEINTEL_PGPASSWORD='sg' \
    -e CODEINTEL_PGDATABASE='sg' \
    -e CODEINTEL_PGSSLMODE='disable' \
    -e CODEINSIGHTS_PGHOST='wg_codeinsights-db' \
    -e CODEINSIGHTS_PGPORT='5432' \
    -e CODEINSIGHTS_PGUSER='sg' \
    -e CODEINSIGHTS_PGPASSWORD='sg' \
    -e CODEINSIGHTS_PGDATABASE='postgres' \
    -e CODEINSIGHTS_PGSSLMODE='disable' \
    --network=wg_test \
    "$1" \
    up -db all
}

# Create a docker network for the purpose of our tests
echo "--- 🐋 creating network"
docker network create wg_test >/dev/null

# Create the test databases
echo "--- 🐋 creating frontend db"
docker run \
  --rm \
  --detach \
  --platform linux/amd64 \
  --name wg_pgsql \
  --network=wg_test \
  sourcegraph/postgres-12-alpine:5.1.0 >/dev/null

echo "--- 🐋 creating codeintel db"
docker run \
  --rm \
  --detach \
  --platform linux/amd64 \
  --name wg_codeintel-db \
  --network=wg_test \
  sourcegraph/codeintel-db:5.1.0 >/dev/null

echo "--- 🐋 creating codeinsights db"
docker run \
  --rm \
  --detach \
  --platform linux/amd64 \
  --name wg_codeinsights-db \
  --network=wg_test \
  sourcegraph/codeinsights-db:5.1.0 >/dev/null

# Migrate those databases to a given version
echo "--- 💽 Migration databases up to v5.1.0"
run_migrator sourcegraph/migrator:"$TEST_VERSION"

# Load the docker image for migrator
echo "--- 🐋 Loading locally built migrator"
"$1" >/dev/null

echo "--- 💽 Migration databases up to latest"
run_migrator migrator:candidate
