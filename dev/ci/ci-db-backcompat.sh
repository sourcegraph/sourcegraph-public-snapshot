#!/usr/bin/env bash
#
# This is wrapper that runs the DB schema backcompat test (db-backcompat.sh) in the CI environment.
#
# It finds the newest commit that is currently running in sourcegraph.com and to run the DB unit
# tests against the *present* schema. This ensures that sourcegraph.com can write to the new
# database schema if upgraded directly to this commit.

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

BRANCH=$(git branch --show-current)
if [[ "$BRANCH" =~ ^[0-9.]+$ ]]; then
  echo 'Skipping db-backcompat test for versioned branch'
  exit 0
fi

HEAD=$(git symbolic-ref --short HEAD || git rev-parse HEAD)
if [ -z "$HEAD" ]; then
  # shellcheck disable=SC2016
  echo 'Could not set $HEAD to current revision'
  exit 1
fi

CURRENTLY_DEPLOYED=$(./dev/deployed-commit.sh)

cat <<EOF
Running ci-db-backcompat.sh with the following parameters:
  HEAD:                            ${HEAD}
  git rev-parse HEAD:              $(git rev-parse HEAD)
  git rev-parse --abbrev-ref HEAD: $(git rev-parse --abbrev-ref HEAD)
  current deployed commit:         ${CURRENTLY_DEPLOYED}
EOF

# Recreate the test DB and run TestMigrations once to ensure that the schema version is the latest.
set -ex
asdf install # in case the go version has changed in between these two commits
go test -count=1 -v ./internal/db/ -run=TestMigrations
HEAD="$HEAD" OLD="${CURRENTLY_DEPLOYED}" ./dev/ci/db-backcompat.sh
set +ex

echo "SUCCESS"
