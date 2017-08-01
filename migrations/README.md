This directory contains database migrations for the frontend Postgres DB.

## Usage 

Migrations are handled by the [migrate](https://github.com/mattes/migrate/tree/master/cli#installation) tool. Install the tool to apply and run migrations.

```
$ go get -u -d github.com/mattes/migrate/cli github.com/lib/pq
$ go build -tags 'postgres' -o /usr/local/bin/migrate github.com/mattes/migrate/cli
```

### Update database to latest version 

Run the following from the project root to apply all the latest migrations:
```
./dev/migrate.sh
```

### Add a new migration

Run the following, with the first argument being a descriptive name for the migration:

```
./dev/add_migration.sh "migration name"
```

There will be up/down `.sql` migration files created in this directory.

*NOTE:* migrations need to be idempotent with respect to the final schema for now. i.e. prefer to use `CREATE TABLE IF NOT EXISTS` statements over just `CREATE TABLE`. This is so migrations can apply cleanly over a new database that has already created the initial set of tables. We will run into issues if a migration moves or alters tables so this is an issue that will be addressed. See https://github.com/sourcegraph/sourcegraph/pull/6475#issuecomment-319164301
