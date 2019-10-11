#!/bin/bash

set -eo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../migrations"

hash migrate 2>/dev/null || {
    if [[ $(uname) == "Darwin" ]]; then
        brew install golang-migrate
    else
        echo "You need to install the 'migrate' tool: https://github.com/golang-migrate/migrate/"
        exit 1
    fi
}

if [ -z "$1" ]; then
    echo "USAGE: $0 <migration>"
    exit 1
fi

# First, apply migrations up to the version we want to squash
migrate -database "postgres://${PGHOST}:${PGPORT}/${PGDATABASE}" -path . goto $1

# Dump the database into a temporary file. Exclude the schema_migrations
# table which is created before migrations are run. This causes the squashed
# migrations to fail due to conflict (and there's no flag to emit IF NOT EXISTS).
pg_dump -s --no-owner --exclude-table schema_migrations -f tmp.sql

# Now clean up all of the old migration files. `ls` will return files in
# alphabetical order, so we can delete all files from the migration directory
# until we hit our squashed migration.

for file in $(ls *.sql); do
    rm $file

    # There should be two files prefixed with this schema version. The down
    # version comes first, then the up version. Make sure we only break the
    # loop once we remove both files.

    if [[ "$file" == "$1"* && "$file" == *'up.sql' ]]; then
        break
    fi
done

# Move the new migrations into place
mv tmp.sql "./$1_squash_migrations.up.sql"
echo 'SELECT 1' > "./$1_squash_migrations.down.sql"
