# How to apply privileged migrations

Postgres database migrations can be _privileged_ or _unprivileged_. The vast majority of migrations are unprivileged, and should require relatively low capabilities within the connected database. Some migrations are privileged and contain queries that require additional capabilities within the database (for example, the installation of Postgres extensions).

When [running the migrator](manual_database_migrations.md) on your Postgres instance, supply the flag optional command flag `-unprivileged-only` to ensure that only unprivileged migrations will be applied. If the migration runner encounters an unapplied privileged migration, it will stop with an error message similar to the following:

```
error: failed to run migration for schema "frontend": refusing to apply a privileged migration: schema "frontend" requires database migration 1645717519 to be applied by a database user with elevated permissions
The migration runner is currently being run with -unprivileged-only. The indicated migration is marked as privileged and cannot be applied by this invocation of the migration runner. Before re-invoking the migration runner, follow the instructions on https://docs.sourcegraph.com/admin/how-to/privileged_migrations. Please contact support@sourcegraph.com for further assistance.
```

To allow the migrator to make additional progress, the site-administrator must run the privileged query manually. This process is the same as starting from **Step 2** in the guide [_How to resolve a dirty database_](dirty_database.md#2-run-the-sql-queries-to-finish-incomplete-migrations). Once the privileged migration has been applied a migration log entry has been created, the migrator can be re-invoked to complete the remaining unprivileged migrations.
