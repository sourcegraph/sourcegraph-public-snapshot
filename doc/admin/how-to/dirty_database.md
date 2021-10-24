# How to troubleshoot a dirty database

This document will take you through how to resolve a 'dirty database' error. During an upgrade, the database has to be migrated. If the upgrade was interrupted during the migration, this can result in a 'dirty database' error.

The error will look something like this:

```
ERROR: Failed to migrate the DB. Please contact support@sourcegraph.com for further assistance: Dirty database version 1528395797. Fix and force version.
```

## Prerequisites

* This document assumes that you are installing Sourcegraph or were attempting an upgrade when an error occurred
* **NOTE: If you encountered this error during an upgrade, ensure you followed the [proper step upgrade process documented here.](https://docs.sourcegraph.com/admin/updates) If you skipped a minor version during an upgrade, you will need to revert back to the last minor version your instance was on before following the steps in this document.**

## Steps to resolve

1. Check schema version. If it's dirty, then note the version number by using this command:

`SELECT * FROM schema_migrations;`

2. Find the up migration with that version in https://github.com/sourcegraph/sourcegraph/tree/main/migrations/frontend

3. Run the code there explicitly
4. Manually clear the dirty flag on the `schema_migrations` table
5. Start up again and the remaining migrations should succeed, otherwise repeat

* When migrations run, the `schema_migrations` table is updated to show the state of migrations.
  * The `dirty` column indicates whether a migration is in-process, and
  * The `version` column indicates the version of the migration the database is on or converting to.
  * On startup, frontend will abort if the `dirty` column is set to true. (The table has only one row.)
* If frontend fails at startup with a complaint about a dirty migration, a migration was started but not recorded as completing.
  * It’s possible that one or more commands from the migration ran successfully.
* Do not mark the migration table as clean if you have not verified that the migration was successfully completed.
* To check the state of the migration table:
  * `SELECT * FROM schema_migrations;`
  * `version   | dirty`
  * `------------+-------`
  * `1528395539 | t`
  * `(1 row)`

* This indicates that migration 1528395539 was running, but has not yet completed.
* Check on the actual state of the migration directly; *if* it has completed, you can manually clear the dirty bit:
  * `UPDATE schema_migrations SET dirty = 'f' WHERE version = 1528395539;`
* Checking on the status of the migration requires looking at the migration’s commands.
  * The source for each migration is in `sourcegraph/sourcegraph/migrations`, in a file named something like `1528395539_.up.sql`
    * The number indicates a migration serial number
    * The text (usually empty in recent migrations) after the serial number describes the purpose of the migration
    * There should be a corresponding `.down.sql` file to reverse the migration.
* Many migrations do nothing but create tables and/or indexes or alter them.
* You can get a description of a table and its associated indexes quickly using the `\d` command (note lack of semicolon):
  * `sg=# \d global_dep`
  * Using this information, you can determine whether a table exists, what columns it contains, and what indexes on it exist.
  * This allows you to determine whether a given command ran successfully.

## Further resources

* [Sourcegraph - Upgrading Sourcegraph to a new version](https://docs.sourcegraph.com/admin/updates)
