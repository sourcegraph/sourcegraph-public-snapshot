# How to troubleshoot a dirty database

This document will take you through how to resolve a 'dirty database' error. During an upgrade, the `pgsql`, `codeintel-db`, and `codeinsights-db` databases must be migrated. If the upgrade was interrupted during the migration, this can result in a 'dirty database' error. 

The error will look something like this:

```
ERROR: Failed to migrate the DB. Please contact support@sourcegraph.com for further assistance: Dirty database version 1528395797. Fix and force version.
```

Migration interruption usually results from a migration query that has run too long and is interupted by a timeout, or an SQL error. Resolving this error requires discovering which migration file failed to run, and manually attempting to run that migration. 

## Prerequisites

* This document assumes that you are installing Sourcegraph or were attempting an upgrade when an error occurred. 
* **NOTE: If you encountered this error during an upgrade, ensure you followed the [proper step upgrade process documented here.](https://docs.sourcegraph.com/admin/updates) If you skipped a minor version during an upgrade, you will need to revert back to the last minor version your instance was on before following the steps in this document.**

The following procedure requires that you are able to execute commands from inside the database container. Learn more about shelling into [kubernetes](https://docs.sourcegraph.com/admin/install/kubernetes/operations#access-the-database), [docker-compose](https://docs.sourcegraph.com/admin/install/docker-compose/operations#access-the-database), and [Sourcegraph single-container](https://docs.sourcegraph.com/admin/install/docker/operations#access-the-database) instances at these links. 

## Steps to resolve


### 1. Identify incomplete migration

Check schema version, by querying the database version table: `SELECT * FROM schema_migrations;` If it's dirty, then note the version number. 

Example:
```
SELECT * FROM schema_migrations;
version | dirty
------------+-------
1528395539 | t
(1 row)
```
This indicates that migration `1528395539` was running, but has not yet completed. 

_Note: for codeintel the schema version table is called `codeintel_schema_migrations` and for codeinsights its called `codeinsights_schema_migrations`_

### 2. Run the sql queries to finish incomplete migrations
Sourcegraphs migration files take for form of `sql` files following the snake case naming schema `<version>_<description>.<up or down>.sql` and can be found [here](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/migrations).

1. Find the up migration with that version in [https://github.com/sourcegraph/sourcegraph/tree/main/migrations/frontend](https://github.com/sourcegraph/sourcegraph/tree/main/migrations/frontend)

2. Run the code there explicitly

### 3. Verify database is clean and declare dirty=false
3. Manually clear the dirty flag on the `schema_migrations` table, example `psql` query:
   *  `UPDATE schema_migrations SET version=1528395918, dirty=false;` 
4. Start up again and the remaining migrations should succeed, otherwise repeat

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
