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
    echo "USAGE: $0 <tag>"
    echo ""
    echo "This tool will squash all migrations up to and including the last migration defined"
    echo "in the given tag branch. The input to this tool should be three minor releases before"
    echo "the current release. For example, if we're currently on 3.10, the input should be the"
    echo "tag 'v3.7.0' as we need to maintain compatibility with versions 3.8 and 3.9."
    echo ""
    exit 1
fi

# Find the last migration defined in the given tag
VERSION=$(git ls-tree -r --name-only "$1" ./ \
    | cut -d'_' -f1 \
    | grep -v [^0-9] \
    | sort \
    | tail -n1)

if [ -z "${VERSION}" ]; then
    echo "failed to retrieve migration version"
    exit 1
fi

# First, apply migrations up to the version we want to squash
migrate -database "postgres://${PGHOST}:${PGPORT}/${PGDATABASE}?sslmode=disable" -path . goto "${VERSION}"

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

    if [[ "$file" == "${VERSION}"* && "$file" == *'up.sql' ]]; then
        break
    fi
done

# Wrap squashed migration in transaction
printf "BEGIN;\n" > "./${VERSION}_squashed_migrations.up.sql"
cat tmp_squashed.sql >> "./${VERSION}_squashed_migrations.up.sql"
printf "\nCOMMIT;\n" >> "./${VERSION}_squashed_migrations.up.sql"
rm tmp_squashed.sql

# Create down migration. This needs to drop everything, so we just drop the
# schema and recreate it. This happens to also drop the schema_migrations
# table, which blows up the migrate tool if we don't put it back.

cat > "./${VERSION}_squashed_migrations.down.sql" << EOL
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;

CREATE TABLE IF NOT EXISTS schema_migrations (
    version bigint NOT NULL PRIMARY KEY,
    dirty boolean NOT NULL
);
EOL

echo ""
echo "squashed migrations written to ${VERSION}_squashed_migrations.{up,down}.sql"

# Regenerate bindata
go generate
