This directory contains database migrations for the frontend Postgres DB.

## Usage 

Migrations are handled by the [migrate](https://github.com/mattes/migrate/tree/master/cli#installation) tool. Migrations get applied automatically at application startup. The CLI tool can also be used to manually test migrations.

### Add a new migration

Run the following, with the first argument being a descriptive name for the migration:

```
./dev/add_migration.sh "migration name"
```

There will be up/down `.sql` migration files created in this directory. Add SQL statements to those files, then embed them into the Go code:

```
go generate ./pkg/localstore/migrations/
```
