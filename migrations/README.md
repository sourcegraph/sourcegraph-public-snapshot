This directory contains database migrations for the frontend Postgres DB.

## Usage

Migrations are handled by the [migrate](https://github.com/golang-migrate/migrate/tree/master/cli#installation) tool. Migrations get applied automatically at application startup. The CLI tool can also be used to manually test migrations.

### Add a new migration

Run the following:

```
./dev/add_migration.sh
```

There will be up/down `.sql` migration files created in this directory. Add SQL statements to these
files that will perform the desired migration. **NOTE**: the migration runner wraps each migration
script in a transaction block; do not add explicit transaction blocks to the migration script as
this has caused issues in the past.

```sql
# Enter statements here
```

After adding SQL statements to those files, embed them into the Go code and update the schema doc:

```
./dev/generate.sh
```

or, to only run the DB generate scripts (subset of the command above):

```
go generate ./cmd/frontend/internal/db/migrations/
go generate ./cmd/frontend/internal/db/
```

### Migrating up/down

Migrations happen automatically on server start-up after running the generate scripts. They can also
be run manually using the migrate CLI. Run `./dev/migrate.sh` for more info.

You can run `./dev/migrate.sh down 1` to rollback the previous migration. If a migration fails and
you need to revert to a previous state `./dev/migrate.sh force` may be helpful. Alternatively use
the `dropdb` and `createdb` commands to wipe your local DB and start from a clean state.

### Running down migrations for customer rollbacks

If a customer needs to roll back across a DB migration, they will need to do the following steps:

* Roll back Sourcegraph to the previous version. On startup, the frontend pods will log a migration
  error and the `dirty` column in the `schema_migrations` table will be set to true.
* Determine the two commits that correspond to the previous and new versions of Sourcegraph. Check
  out each commit and run `ls -1` in the `migrations` directory. The order of the migrations is the
  same as the alphabetical order of the migration scripts, so take the diff between the two `ls -1`s to determine which migrations should be run.
* Apply the down migration scripts in **reverse chronological order**. Wrap each down migration in a
  transaction block. If there are any errors, stop and resolve the issue before proceeding with the
  next down migration.
* After all down migrations have been applied, run `update schema_migrations set version=$VERSION, dirty=false;`, where `$VERSION` is the numerical prefix of the migration script corresponding to
  the first migration you _didn't_ just apply. In other words, it is the numerical prefix of the
  last migration script as of the rolled-back-to commit.
* Kill all frontend pods. On restart, they should spin up successfully.
