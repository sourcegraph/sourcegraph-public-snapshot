# Postgres Migrations

The children of this directory contain migrations for each Postgres database instance:

- `frontend` is the main database (things should go here unless there is a good reason)
- `codeintel` is a database containing only processed LSIF data (which can become extremely large)
- `codeinsights` is a TimescaleDB database, containing only Code Insights time series data

The migration path for each database instance is the same and is described below. Each of the database instances described here are deployed separately, but are designed to be _overlayable_ to reduce friction during development. That is, we assume that the names in each database do not overlap so that the same connection parameters can be used for both database instances.

## Migrating up and down

Up migrations will happen automatically in development on service startup. In production environments, they are run by the [`migrator`](../cmd/migrator) instance. You can run migrations manually during development via `sg`:

- `sg migration up` runs all migrations to the latest version
- `sg migration up -db=frontend -target=<version>` runs up migrations (relative to the current database version) on the frontend database until it hits the target version
- `sg migration undo -db=codeintel` runs one _down_ migration (relative to the current database version) on the codeintel database

## Adding a migration

**IMPORTANT:** All migrations must be backwards-compatible, meaning that _existing_ code must be able to operate successfully against the _new_ (post-migration) database schema. Consult [_Writing database migrations_](https://docs.sourcegraph.com/dev/background-information/sql/migrations.md) in our developer documentation for additional context.

To create a new migration file, run the following command.

```
$ sg migration add -db=<db_name> <my_migration_name>
Migration files created
 Up query file: ~/migrations/codeintel/1644260831/up.sql
 Down query file: ~/migrations/codeintel/1644260831/down.sql
 Metadata file: ~/migrations/codeintel/1644260831/metadata.yaml
```

This will create an _up_ and _down_ pair of migration files (whose path is printed by the following command). Add SQL statements to these files that will perform the desired migration. After adding SQL statements to those files, update the schema doc via `go generate ./internal/database/` (or regenerate everything via `./dev/generate.sh`).

To pass CI, you'll additionally need to:

- Ensure that your new migrations run against the current Go unit tests
- Ensure that your new migrations can be run up, then down, then up again (idempotency test)
- Ensure that your new migrations do not break the Go unit tests published with the previous release (backwards-compatibility test)

## Customer rollbacks

Running _down_ migrations in a rollback **should NOT** be necessary if all migrations are backward-compatible. In case the customer must run a down migration, they will need perform do the following steps.

- Roll back Sourcegraph to the previous version. On startup, the frontend pods will log a migration warning stating that the schema has been migrated to a newer version. This warning should **NOT** indicate that the database is dirty.

- Determine if a database is dirty by running the following commands.

  **frontend database**:

  ```
  kubectl exec $(kubectl get pod -l app=pgsql -o jsonpath='{.items[0].metadata.name}') -- psql -U sg -c 'SELECT * FROM schema_migrations'
  ```

  **codeintel database**:

  ```
  kubectl exec $(kubectl get pod -l app=pgsql-codeintel -o jsonpath='{.items[0].metadata.name}') -- psql -U sg -c 'SELECT * FROM codeintel_schema_migrations'
  ```

  **codeinsights database**:

  ```
  kubectl exec $(kubectl get pod -l app=codeinsights-db -o jsonpath='{.items[0].metadata.name}') -- psql -U sg -c 'SELECT * FROM codeinsights_schema_migrations'
  ```

  For each dirty database, follow the steps in the _Dirty schema_ section below.

- For each database `<db_name>` with the schema version table `<schema_version_table_name>`, do the following:
  - Determine the two commits that correspond to the previous and new versions of Sourcegraph. Check out each commit and run `ls -1` in the `migrations/<db_name>` directory. The order of the migrations is the same as the alphabetical order of the migration scripts, so take the diff between the two list outputs to determine which migrations should be run.
  - Apply the down migration scripts in **reverse chronological order**. Wrap each down migration in a transaction block. If there are any errors, stop and resolve the issue before proceeding with the next down migration.
  - After all down migrations have been applied, run
    ```
    update <schema_version_table_name> set version=$VERSION;
    ```
    where `$VERSION` is the numerical prefix of the migration script corresponding to the first migration you _didn't_ just apply. In other words, it is the numerical prefix of the last migration script as of the rolled-back-to commit.
  - Restart frontend pods. On restart, they should spin up successfully.

### Reverting a migration

If a reverted PR contains a DB migration, it may still have been applied to Sourcegraph.com, k8s.sgdev.org, etc. due to their rollout schedules. In some cases, it may also have been part of a Sourcegraph release. To fix this, you should create a PR to revert the migrations of that commit. The `sg migration revert <commit>` command automates all the necessary changes the migration definitions.
