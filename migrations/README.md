This directory contains database migrations for the frontend Postgres DB.

## Usage

Migrations are handled by the [migrate](https://github.com/mattes/migrate/tree/master/cli#installation) tool. Migrations get applied automatically at application startup. The CLI tool can also be used to manually test migrations.

### Add a new migration

Run the following, with the first argument being a descriptive name for the migration (e.g. "add users table"):

```
./dev/add_migration.sh "migration name"
```

There will be up/down `.sql` migration files created in this directory. Add SQL statements to these files that will perform the desired migration. If using mutiple statements, wrap them in a transaction to prevent partially-executed migration state in the event of a failure.

```sql
BEGIN;

# Enter statements here

COMMIT;
```

After adding SQL statements to those files, embed them into the Go code and update the schema doc:

```
./dev/generate.sh
```

or, to only run the DB generate scripts (subset of the command above):

```
go generate ./pkg/localstore/migrations/
go generate ./pkg/localstore/
```

### Migrating up/down

Migrations happen automatically on server start-up after running the generate scripts. They can also be run manually using the migrate CLI. Run `./dev/migrate.sh` for more info.

You can run `./dev/migrate.sh down 1` to rollback the previous migration. If a migration fails and you need to revert to a previous state `./dev/migrate.sh force` may be helpful. Alternatively use the `dropdb` and `createdb` commands to wipe your local DB and start from a clean state.
