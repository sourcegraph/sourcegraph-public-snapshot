#!/usr/bin/env bash
#
# This script tests the backward-compatibility of the current DB schema at revision $HEAD with the
# DB unit tests at revision $OLD. $HEAD should be set to the currently checked out revision.
#
# * It first checks the precondition that the schema of the DB has been updated to match the latest
#   migration existing at $HEAD.
# * It then checks out the $OLD revision and runs the db package unit tests, explicitly skipping the
#   migration test to avoid downgrading the schema.
# * It checks that the DB schema version still corresponds to the latest schema version.
# * It then checks out $HEAD again and exits with the exit code of the db unit tests it ran in the
#   old revision.

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

if [ -z "$HEAD" ] || [ -n "$(git diff "$HEAD"..HEAD)" ]; then
  # shellcheck disable=SC2016
  echo 'Must set $HEAD to currently checked out branch.'
  set -x
  git diff "$HEAD"..HEAD
  git log -n1 "$HEAD"
  git log -n1 HEAD
  exit 1
fi
if [ -z "$OLD" ]; then
  # shellcheck disable=SC2016
  echo 'Must set $OLD to old commit.'
  exit 1
fi

if [ -n "$(git status --porcelain)" ]; then
  git status
  echo 'Work tree is dirty, aborting.'
  exit 1
fi

echo "Running backcompat test between $HEAD (HEAD) and $OLD"

function getLatestMigrationVersion() {
  find ./migrations -type f -name '[0-9]*.up.sql' | cut -d'_' -f 1 | cut -d'/' -f 3 | sort -n | tail -n1
}

LATEST_SCHEMA=$(getLatestMigrationVersion)
CURRENT_DB_SCHEMA=$(psql -t -d sourcegraph-test-db -c 'select version from schema_migrations' | xargs echo -n)

if [ "$LATEST_SCHEMA" != "$CURRENT_DB_SCHEMA" ]; then
  echo "Latest migration schema version ($LATEST_SCHEMA) does not match schema in test DB ($CURRENT_DB_SCHEMA)."
  # shellcheck disable=SC2016
  echo '    You can run `go test -count=1 -v ./cmd/frontend/db/  -run=TestMigrations` to update the test DB schema.'
  exit 1
fi

function runTest() {
  (
    set -ex
    git checkout "$OLD"
    asdf install # in case the go version has changed in between these two commits
    set +ex

    NOW_LATEST_SCHEMA=$(getLatestMigrationVersion)
    cat <<-EOF
	Running DB tests against old commit: $(git rev-parse HEAD)
	    Latest migration version as of this commit:	${NOW_LATEST_SCHEMA}
	    Latest migration version overall:		${LATEST_SCHEMA}
	    DB schema version:				${CURRENT_DB_SCHEMA}
	EOF

    # All DB tests are assumed to import the internal/db/dbtesting package, so use that to
    # find which packages' tests need to be run.
    # shellcheck disable=SC2016
    mapfile -t PACKAGES_WITH_DB_TEST < <(go list -f '{{$printed := false}}{{range .TestImports}}{{if and (not $printed) (eq . "github.com/sourcegraph/sourcegraph/internal/db/dbtesting")}}{{$.ImportPath}}{{$printed = true}}{{end}}{{end}}' ./...)

    set -ex
    # Test without cache, because schema change does not
    # necessarily mean Go source has changed.
    TEST_SKIP_DROP_DB_BEFORE_TESTS=true SKIP_MIGRATION_TEST=true go test -count=1 -v "${PACKAGES_WITH_DB_TEST[@]}"
    set +ex

    NOW_DB_SCHEMA=$(psql -t -d sourcegraph-test-db -c 'select version from schema_migrations' | xargs echo -n)
    if [ "$LATEST_SCHEMA" != "$NOW_DB_SCHEMA" ]; then
      echo ""
      echo "FAIL: DB schema ${NOW_DB_SCHEMA} no longer matches latest schema version ${LATEST_SCHEMA} after running tests."
      echo ""
      exit 1
    else
      echo "DB schema ${NOW_DB_SCHEMA} still matches latest schema version ${LATEST_SCHEMA} after running tests."
    fi
  )
}
runTest
EXIT_CODE="$?"

set -x
git checkout "$HEAD"
asdf install # in case the go version has changed in between these two commits
set +x
echo "Restored HEAD commit: $HEAD: $(git rev-parse HEAD)"

exit "$EXIT_CODE"
