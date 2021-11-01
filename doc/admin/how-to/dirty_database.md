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

When migrations run, the `schema_migrations` table is updated to show the state of migrations. The `dirty` column indicates whether a migration is in-process, and the `version` column indicates the version of the migration the database is on or converting to. On startup, the frontend will not start if the `schema_migrations` `dirty` column is set to true.

**Check schema version, by querying the database version table:** `SELECT * FROM schema_migrations;` **If it's dirty, then note the version number for use in step 2.**

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

Sourcegraphs migration files take for form of `sql` files following the snake case naming schema `<version>_<description>.<up or down>.sql` and can be found [here](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/migrations) in subdirectories for the specific database. _Note frontend is the pgsql database_.

1. **Find the up migration starting with the version number identified in [step 1](#1-identify-incomplete-migration):** [https://github.com/sourcegraph/sourcegraph/tree/main/migrations](https://github.com/sourcegraph/sourcegraph/tree/main/migrations)

2. **Run the code from the identified migration _up_ file explicitly using the `psql` CLI:**
   * It’s possible that one or more commands from the migration ran successfully already. In these cases you may need to run the sql transaction in pieces. For example if a migration file creates multiple indexes and one index already exist you'll need to manually run this transaction skipping that line or adding `IF NOT EXISTS` to the transaction.
   * If you’re running into unique index creation errors because of duplicate values please let us know at support@sourcegraph.com or via your enterprise support channel. 

### 3. Verify database is clean and declare `dirty=false`

1. **Manually clear the dirty flag on the `schema_migrations` table**, example `psql` query:
```
UPDATE schema_migrations SET version=1528395918, dirty=false;
```
**Do not mark the migration table as clean if you have not verified that the migration was successfully completed.**

Checking on the status of the migration requires looking at the migration’s commands. In its migration file. Many migrations do nothing but create tables and/or indexes or alter them.

You can get a description of a table and its associated indexes quickly using the `\d` command (note lack of semicolon). Using this information, you can determine whether a table exists, what columns it contains, and what indexes on it exist. Use this inforamtion to determine if commands in a migration ran successfully before setting `dirty=false`.

1. **Start Sourcegraph again and the remaining migrations should succeed, otherwise repeat this procedure again from [_1. Identify incomplete migration_](#1-identify-incomplete-migration)**

## Additional Information

### `CREATE_INDEX_CONCURRENTLY`
Some migrations utilize the `CREATE INDEX CONCURRENTLY` migration option which runs the query as a background process (learn more [here](https://www.postgresql.org/docs/9.1/sql-createindex.html#SQL-CREATEINDEX-CONCURRENTLY)). If one of these migrations fails to complete, you will see the index when you describe the table, the migration will see it there, but it will be an unusable. You will then need to `REINDEX CONURRENTLY` or drop and recreate the index.

You can discover if such a damaged index exists by running the following query:

```
SELECT
    current_database() AS datname,
    pc.relname AS relname,
    1 AS count
FROM pg_class pc
JOIN pg_index pi ON pi.indexrelid = pc.oid
WHERE
    NOT indisvalid AND
    NOT EXISTS (SELECT 1 FROM pg_stat_progress_create_index pci WHERE pci.index_relid = pi.indexrelid)
```
Additionally grafana will alert you of an index is this state. The grafana alert can be found unders its database charts. Ex: `Site Admin > Monitoring > Postgres > Invalid Indexes (unusable by query planner)`

## Further resources

* [Sourcegraph - Upgrading Sourcegraph to a new version](https://docs.sourcegraph.com/admin/updates)
* [Migrations README.md](https://github.com/sourcegraph/sourcegraph/blob/main/migrations/README.md) (Note some of the info contained here pertains to running Sourcegraphs development environment and should not be used on production instances)
