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
    echo ""
    echo "To use this tool, first find the migration version you want to squash. All"
    echo "migrations prior to this version will be squashed into that version."
    echo ""
    echo "For example, if the latest release of Sourcegraph is v3.9.x then since we must"
    echo "maintain backwards compatability for upgrading from 3.7.x and 3.8.x (two minor"
    echo "versions), you should find the last migration that v3.6.x performed and specify"
    echo "that as the version to squash to."
    exit 1
fi

# First, apply migrations up to the version we want to squash
migrate -database "postgres://${PGHOST}:${PGPORT}/${PGDATABASE}" -path . goto $1

# Dump the database into a temporary file that we need to post-process
pg_dump -s --no-owner --no-comments --clean --if-exists -f tmp_squashed.sql

# Remove settings header from pg_dump output
sed -i '' -e 's/^SET .*$//g' tmp_squashed.sql
sed -i '' -e 's/^SELECT pg_catalog.set_config.*$//g' tmp_squashed.sql

# Do not drop extensions if they already exists. This causes some
# weird problems with the back-compat tests as the extensions are
# not dropped in the correct order to honor dependencies.
sed -i '' -e 's/^DROP EXTENSION .*$//g' tmp_squashed.sql

# Remove references to public schema
sed -i '' -e 's/public\.//g' tmp_squashed.sql
sed -i '' -e 's/ WITH SCHEMA public//g' tmp_squashed.sql

# Remove comments, multiple blank lines
sed -i '' -e 's/^--.*$//g' tmp_squashed.sql
sed -i '' -e '/^$/N;/^\n$/D' tmp_squashed.sql

# Now clean up all of the old migration files. `ls` will return files in
# alphabetical order, so we can delete all files from the migration directory
# until we hit our squashed migration.

for file in $(ls *.sql); do
    rm $file
    echo "squashed migration $file"

    # There should be two files prefixed with this schema version. The down
    # version comes first, then the up version. Make sure we only break the
    # loop once we remove both files.

    if [[ "$file" == "$1"* && "$file" == *'up.sql' ]]; then
        break
    fi
done

# Wrap squashed migration in transaction
printf "BEGIN;\n" > "./$1_squashed_migrations.up.sql"
cat tmp_squashed.sql >> "./$1_squashed_migrations.up.sql"
printf "\nCOMMIT;\n" >> "./$1_squashed_migrations.up.sql"
rm tmp_squashed.sql

# Create down migration. This needs to drop everything, so we just drop the
# schema and recreate it. This happens to also drop the schema_migrations
# table, which blows up the migrate tool if we don't put it back.

cat > "./$1_squashed_migrations.down.sql" << EOL
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;

CREATE TABLE IF NOT EXISTS schema_migrations (
    version bigint NOT NULL PRIMARY KEY,
    dirty boolean NOT NULL
);
EOL

echo ""
echo "squashed migrations written to $1_squashed_migrations.{up,down}.sql"

# Regenerate bindata
go generate
