#!/usr/bin/env bash
#
# This is wrapper that runs the DB schema backcompat test (db-backcompat.sh) in the CI environment.
#
# It finds the last migration by listing the migration SQL files (alphabetical order corresponds to
# chronological order), then finds the commit in which those SQL files were added. It then uses the
# commit immediately before that commit to run the DB unit tests against the *present* schema.

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

HEAD=$(git symbolic-ref --short HEAD || git rev-parse HEAD)
if [ -z "$HEAD" ]; then
  # shellcheck disable=SC2016
  echo 'Could not set $HEAD to current revision'
  exit 1
fi

cat <<EOF
Running ci-db-backcompat.sh with the following parameters:
  HEAD:					$HEAD
  git rev-parse HEAD:			$(git rev-parse HEAD)
  git rev-parse --abbrev-ref HEAD: 	$(git rev-parse --abbrev-ref HEAD)
EOF

LAST_MIGRATION=$(find ./migrations -type f -name '[0-9]*.up.sql' | cut -d'_' -f 1 | cut -d'/' -f 3 | sort -n | tail -n1)
COMMIT_OF_LAST_MIGRATION=$(git log --pretty=format:"%H" "./migrations/${LAST_MIGRATION}"* | tail -n1)
COMMIT_BEFORE_LAST_MIGRATION=$(git log -n1 --pretty=format:"%H" "${COMMIT_OF_LAST_MIGRATION}"^)

echo "Last migration was	${LAST_MIGRATION},	added in     	${COMMIT_OF_LAST_MIGRATION}."
echo "Testing current schema	${LAST_MIGRATION},	with tests at	${COMMIT_BEFORE_LAST_MIGRATION}."
echo ""
git log -n2 --stat "${COMMIT_OF_LAST_MIGRATION}" | sed 's/^/  /'
echo ""

# Recreate the test DB and run TestMigrations once to ensure that the schema version is the latest.
set -ex
asdf install # in case the go version has changed in between these two commits
go test -count=1 -v ./cmd/frontend/db/ -run=TestMigrations
HEAD="$HEAD" OLD="${COMMIT_BEFORE_LAST_MIGRATION}" ./dev/ci/db-backcompat.sh
set +ex

echo "SUCCESS"
