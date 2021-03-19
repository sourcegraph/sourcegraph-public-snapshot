#!/usr/bin/env bash

# This script attempts to help when rebasing a branch with one or more
# migrations by renumbering a database migration to a non-conflicting number.

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../migrations"

if [ $# -lt 2 ]; then
  echo "USAGE: $0 <db_name> <migration_to_rebase...>"
  echo
  echo "NOTE: <migration_to_rebase> can be either an up or down migration file."
  exit 1
fi

if [ ! -d "$1" ]; then
  echo "Unknown database '$1'"
  exit 1
fi
database="$1"
pushd "$database" >/dev/null || exit 1

cat <<EOF
SOURCERER WARNING

Assuming your migration is the latest one, it's a lot easier to migrate down
before rebasing. If you're in the middle of a rebase, we can't migrate down now
because the migrate command will get confused by the duplicate migrations we're
about to fix.

If you haven't already migrated down, you may want to consider aborting your
rebase, migrating down, then re-running this script:

    git rebase --abort
    ./dev/db/migrate.sh down $(($# - 1))
    git rebase origin/main          # or whatever you're rebasing on
    $0 $@

EOF

read -p 'Do you want to continue? [y/N] ' -n 1 -r
echo
case "$REPLY" in
  Y | y) ;;
  *)
    echo 'No problemo! This script will still be here when you need it.'
    exit 0
    ;;
esac

# OK. Let's figure out what we're renaming.
shift 1
migrations=()
for migration; do
  migration="$(echo "$migration" | sed -E -e 's/^.*\///' -e 's/\.sql$//' -e 's/\.(down|up)$//')"

  for suffix in .down.sql .up.sql; do
    if [ ! -f "$migration$suffix" ]; then
      echo "ERROR: cannot find migration file $migration"
      exit 1
    fi
  done

  migrations+=("$migration")
done

# Find the highest version, relying on ls's default sorting behaviour.
#
# Disable shellcheck: it doesn't like the ls | grep construction, and honestly
# nor do I, but the alternatives (using a glob with ./, or find | sort) involve
# more string munging due to including preceding path elements in the output.
# (This is one of the exceptions noted in the relevant shellcheck page.)
#
# shellcheck disable=SC2010
version="$(ls -1 | grep '\.sql$' | tail -1 | cut -d _ -f 1)"

# Now we'll go through and rename the files.
for migration in "${migrations[@]}"; do
  name="$(echo "$migration" | cut -d _ -f 2-)"
  version=$((version + 1))
  echo "Renumbering $migration to $version..."
  git mv "${migration}.down.sql" "${version}_${name}.down.sql"
  git mv "${migration}.up.sql" "${version}_${name}.up.sql"
done

cat <<EOF
Done!

Don't forget to regenerate schema before continuing your
rebase:

    ./dev/db/migrate.sh $database up
    go generate ./internal/database

Then git add everything and continue the rebase.
EOF
