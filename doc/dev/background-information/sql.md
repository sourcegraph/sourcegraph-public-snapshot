# High-performance SQL

This document contains recommendations on how to write efficient SQL for certain classes of tasks.

## Batch operations

### Insertions

If a large number of rows are being inserted into the same table, use of a [batch inserter](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:%5Einternal/database/batch/batch%5C.go+content:%27func+NewInserter%28%27&patternType=literal) instance should be preferred over issuing an `INSERT` statement for each row. This inserter also emits `INSERT` statements, but batches values to insert together, so that the minimum number of round trips to the database are made. The batch inserter instance is also aware of the maximum size of payloads that can be sent to PostgreSQL in one query and will adjust the flush rate accordingly to prevent large queries from being rejected.

The package provides many convenience functions, but basic usage is as follows. An inserter is created with a table name and a list of column names for which values will be supplied. Then, the `Insert` method is called for each row to be inserted. It is expected that the number of values supplied to each call to `Insert` matches the number of columns supplied at construction of the inserter. On each call to `Insert`, if the current batch is full, it will be prepared and sent to the database, leaving an empty batch for future operations. A final call to `Flush` will ensure that any remaining batched rows are sent to the database.

```go
inserter := batch.NewInserter(ctx, db, "table", "col1", "col2", "col3" /* , ... */)

for /* ... */ {
    if err := inserter.Insert(ctx, val1, val2, val3 /* , ... */); err != nil {
        return err
    }
}

if err := inserter.Flush(ctx); err != nil {
    return err
}
```

It is recommended to pass a database handle which is already wrapped in a transaction both for increased performance as well as atomicity in case one batch of rows fails to be inserted.

Sample uses:

- [data_write.go](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@b795806a03468f565702cd8f1990a7fbc969722a/-/blob/enterprise/internal/codeintel/stores/lsifstore/data_write.go#L290:16): Code intelligence uses a batch inserter to write processed code intelligence index data to the codeintel-db. This use instantiates a number of inserters that read values to insert from a shared channel, all operating in parallel.
- [changeset_jobs.go](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@b795806a03468f565702cd8f1990a7fbc969722a/-/blob/enterprise/internal/batches/store/changeset_jobs.go#L99:15): Batch changes uses a batch inserter to insert a large number of changeset jobs atomically. Before using a batch inserter, this method would fail when a large number of rows were inserted. PostgreSQL accepts a maximum of `(2^15) - 1` parameters per query, which can be spent surprisingly quickly when inserting a large number of columns.

### Insertion with common values

During batch insertions, it is common for each of the rows being inserted to be logically related. For example, they may all share the same value for a portion of their primary key, a foreign key, or an insertion timestamp. In such cases, it is wasteful to send the same value explicitly in every row. This waste may be negligible for most insertions, but will be large when inserting a large number of rows or inserting a large number of identical columns.

There is an alternate strategy to supplying values explicitly for each column:

First, create a temporary table containing only the columns that will contain distinct values. This table should be used as the target of the batch inserter. Because this table has fewer columns, the batch inserter will be able to insert a greater number of rows per query and reduce the number of total round trips to the database.

```sql
tempTableQuery := `
  CREATE TEMPORARY TABLE temp_table (
      col3 text NOT NULL, 
      col4 text NOT NULL
  ) ON COMMIT DROP
`
```
```go
if err := db.Exec(ctx, sqlf.Sprintf(tempTableQuery)); err != nil {
    return err
}
```

Here, we defined the temporary table with the clause `ON COMMIT DROP`, which will ensure that the temporary table is only visible to the current transaction and is dropped on the next commit or rollback of the transaction.

Next, create and use a batch inserter instance just as descried in the previous section, but target the newly created temporary table. Only the columns that are defined on the temporary table need to be supplied when calling the `Insert` method.

```go
inserter := batch.NewInserter(ctx, db, "temp_table", "col3", "col4")

for /* ... */ {
    if err := inserter.Insert(ctx, val3, val4); err != nil {
        return err
    }
}

if err := inserter.Flush(ctx); err != nil {
    return err
}
```

Finally, issue a final `INSERT` statement that uses the entire contents of the temporary table as well as the values common to all rows (supplied exactly once) as the insertion source.

```sql
insertQuery := `
    INSERT INTO table
    SELECT
        %s as col1, %s as col2,  -- insert canned col1, col2
        source.col3, source.col4 -- insert remaining columns from the temporary table
    FROM temp_table source
`
```
```go
if err := db.Exec(ctx, sqlf.Sprintf(insertQuery, val1, val2)); err != nil {
    return err
}
```

### Updates

Not all batch operations can be accomplished with insertions alone. We have several cases where denormalized (or cached) data is persisted to PostgreSQL en masse. The obvious implementation is to delete any conflicting data prior to a full insertion of the data set. This works, but has an obvious drawback: if the data set changes incrementally between recalculation, then many identical rows may be deleted then re-inserted. This has the effect of creating null operations that requires two writes per row, which is obviously inefficient. However, the real danger comes from increased table and index bloat.

When PostgreSQL deletes or updates a row, it simply marks it as invisible to all operations occurring after the deletion time. A background process, the autovacuum daemon, will periodically scan heap pages and indexes to remove references to data that is unreachable to any active or future operation. A table or index's _bloat_ is the proportion of data within that object that is no longer reachable. Objects with a high bloat factor are inefficient in space (both live and dead tuples must be stored), and queries over such objects are inefficient in time (searching over a larger data structure takes longer, and more heap pages must be retrieved to check visibility information).

Bloat factor grows when the rate of creation of unreachable rows outpaces the autovacuum daemon's removal of unreachable rows. This exact situation happened at a large enterprise customer and caused their database disk to fill completely. This caused major instability in their instance due to database restarts. One remediation is to tune the autovaccum daemon (at the cost of taking CPU and memory resources away from the query path).

A better solution is to reduce the number of dead rows created during these batch updates. We can use a technique similar to the section above, but instead of issuing a single `INSERT` statement, we can issue an `INSERT` statement for all _new_ rows, a `DELETE` statement for all _missing_ rows, and an `UPDATE` statement for all _changed_ rows. This is likely to affect a much smaller number of total rows when the new and old data set have a large overlap (i.e., the majority of rows are unchanged between updates).

In the following, we assume that `col1`, `col2`, and `col3` cover the identity of the row, and `col4` is a simple data point (which can be updated without changing the identity the row). For this to be optimally efficient, `col4` should not be present in any index so that a [HOT update](https://www.cybertec-postgresql.com/en/hot-updates-in-postgresql-for-better-performance/) is possible.

#### Insert new rows

```sql
insertQuery := `
    INSERT INTO table
    SELECT %s as col1, %s as col2, source.col3, source.col4
    FROM temp_table source
    WHERE NOT EXISTS (
        -- Skip insertion of any rows that already exist in the table
        SELECT 1 FROM table t WHERE t.col1 = %s AND t.col2 = %s AND t.col3 = source.col3
    )
`
```
```go
if err := db.Exec(ctx, sqlf.Sprintf(insertQuery, val1, val2, val1, val2)); err != nil {
    return err
}
```

#### Delete missing rows

```sql
deleteQuery := `
    DELETE FROM table
    WHERE col1 = %s AND col2 = %s AND NOT EXISTS (
        -- NOTE: col1 = val1 and col2 = val2 for all rows in temp_table,
        -- so matching rows can be determined by the value of col3 alone.
        SELECT 1 FROM temp_table t WHERE t.col3 = source.col3
    )
`
```
```go
if err := db.Exec(ctx, sqlf.Sprintf(deleteQuery, val1, val2)); err != nil {
    return err
}
```

#### Update changed rows

```sql
updateQuery := `
    UPDATE table t SET col4 = source.col4
    FROM temp_table source
    -- Update rows with matching identity but distinct col4 values
    WHERE t.col1 = %s AND t.col2 = %s AND t.col3 = source.col3 AND t.col4 != source.col4
`
```
```go
if err := db.Exec(ctx, sqlf.Sprintf(updateQuery, val1, val2)); err != nil {
    return err
}
```

Sample uses:

- [commits.go](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@b795806a03468f565702cd8f1990a7fbc969722a/-/blob/enterprise/internal/codeintel/stores/dbstore/commits.go#L292:16): Code intelligence uses this update-in-place technique over a set of tables to store an indexed and compressed view of the commit graph of a repository. This operation runs frequently (after each precise code intelligence index upload, and after updates from the code host) and generally changes only in a local way between updates.

## Materialized cache

For tables that get very large, it may be too expensive to calculate common aggregates over large portions of the table online using a basic `COUNT(...)` function. Consider the following table `conditions`, which stores data sampled every second from a number of temperature sensors over time.

```sql
CREATE TABLE conditions (
    time timestamptz NOT NULL,
    location_id int NOT NULL,
    temperature_celsius double precision NOT NULL,
    PRIMARY KEY (time, location_id)
);

CREATE INDEX conditions_location_id_time ON conditions(location_id, time);
```

Suppose a frequent operation is to select unique locations present in this table. To make this a more realistic example, we can also suppose that the table is partitioned by time and the hot table is still massive. The most obvious query to attain this data uses `DISTINCT` or `GROUP BY`.

```sql
SELECT DISTINCT location_id FROM conditions;
```

In a randomly generated table with 5 million rows and 50k unique locations, the query plan looks as follows.

```
                                                         QUERY PLAN
----------------------------------------------------------------------------------------------------------------------------
 HashAggregate  (cost=94347.90..94841.93 rows=49403 width=4) (actual time=1155.067..1171.370 rows=50000 loops=1)
   Group Key: location_id
   Batches: 5  Memory Usage: 4145kB  Disk Usage: 3368kB
   ->  Seq Scan on conditions  (cost=0.00..81847.92 rows=4999992 width=4) (actual time=0.025..392.017 rows=5000000 loops=1)
 Planning Time: 0.055 ms
 Execution Time: 1176.541 ms
(6 rows)

Time: 1176.797 ms (00:01.177)
```

Aggregate queries will usually either sort an intermediate result set, or build a hash table while the underlying table is scanned. Queries that need to scan a large amount of the table will frequently devolve into a sequential scan over that table. We cannot change the former, but we can reduce the effect of the latter by reducing the size of the table that gets scanned.

In the following solution, we create an aggregate table that stores pre-calculated aggregate data from the source table that is kept updated with triggers. Note that this solution can be expanded to calculate more complex aggregates than show here. For example, data can be pre-calculated to take conditions into account. This will require more bookkeeping at around the triggers, but the complexity may be worth the trade for efficiency.

#### Create table

For this example, we create a `condition_locations` table that materializes the result of the query `SELECT DISTINCT location_id FROM conditions`.

```sql
CREATE TABLE condition_locations (
    location_id int NOT NULL,
    PRIMARY KEY (location_id)
);
```

Note that we are explicitly choosing not to use a [materialized view](https://www.postgresql.org/docs/12/rules-materializedviews.html) here as they are refreshed by running the source query and storing the results. The refresh cost is too high for our purposes, and we need to maintain the materialized view incrementally.

#### Triggers

Triggers should be added to the source table to keep the counts in the aggregate table up to date on all relevant operations. Some tables may have different write semantics such as append/write-only tables or tables whose records are only ever soft deleted. Such tables may not need to apply a trigger on every type of write operation if it cannot change value of the aggregates we are trying to keep up to date.

In our running example, the insertion of a record may introduce a new location. Therefore, we need to update the aggregate value on insertions. In the following, we create a trigger that invokes the `update_condition_locations_insert` for each *statement* that inserts into the `conditions` table. Note that this will be called _once_ for an insert statement that happens to write multiple rows to the table. This function inserts the distinct location from the new batch of rows into the aggregate table, no-oping the insert if the value already exists.

```sql
CREATE OR REPLACE FUNCTION update_condition_locations_insert() RETURNS trigger AS $$ BEGIN
    INSERT INTO condition_locations SELECT DISTINCT location_id FROM newtab ON CONFLICT (location_id) DO NOTHING;
    RETURN NULL;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER condition_locations_insert
AFTER INSERT ON conditions REFERENCING NEW TABLE AS newtab
FOR EACH STATEMENT EXECUTE PROCEDURE update_condition_locations_insert();
```

Similarly, the deletion of a record may remove the last reference of a location. Therefore, we need to update the aggregate value on deletions as well. In the following, we create a trigger that invokes the `update_condition_locations_delete` function on each statement that deletes from the `conditions` table. This function determines if any of the locations removed by the statement are still referenced by the remaining rows. If there are no remaining references, the associated location is deleted from the aggregate table.

```sql
CREATE OR REPLACE FUNCTION update_condition_locations_delete() RETURNS trigger AS $$ BEGIN
    DELETE FROM condition_locations cl
    WHERE
        cl.location_id IN (SELECT location_id FROM oldtab) AND
        NOT EXISTS (SELECT 1 FROM conditions c WHERE c.location_id = cl.location_id);
    RETURN NULL;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER condition_locations_delete
AFTER DELETE ON conditions REFERENCING OLD TABLE AS oldtab
FOR EACH STATEMENT EXECUTE PROCEDURE update_condition_locations_delete();
```

Sample uses:

- [update_lsif_data_documents_schema_versions_insert](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+CREATE+FUNCTION+update_lsif_data_documents_schema_versions_insert&patternType=literal): Code intelligence stores the minimum and maximum row versions for each LSIF document grouped by dump identifier. This is used to efficiently determine sets of records that need to be migrated from an old version to a new version.
- [lsif_data_documentation_pages_insert](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%27CREATE+OR+REPLACE+FUNCTION+lsif_data_documentation_pages_insert%27&patternType=regexp): API Docs updates three aggregate counts including the number of rows, the number of distinct dump identifiers, and the number of distinct indexed dump identifiers (a conditional value).

### Result

Looping back to our original query, we can achieve the same result with much lower latency by querying the aggregate table instead of the source table.

```sql
SELECT location_id FROM condition_locations;
```

The query plan shows that the query is orders of magnitude more efficient, even though the shape of the query plan is the same, including the same aggregate hash table build and underlying sequential scan. The differing factor is the size of the underlying data and the benefits of having a smaller working set (data locality, hot buffers, etc).

```
                                                         QUERY PLAN
-----------------------------------------------------------------------------------------------------------------------------
 HashAggregate  (cost=847.00..1347.00 rows=50000 width=4) (actual time=11.717..17.368 rows=50000 loops=1)
   Group Key: location_id
   Batches: 5  Memory Usage: 4145kB  Disk Usage: 200kB
   ->  Seq Scan on condition_locations  (cost=0.00..722.00 rows=50000 width=4) (actual time=0.006..3.415 rows=50000 loops=1)
 Planning Time: 0.045 ms
 Execution Time: 20.475 ms
(6 rows)

Time: 20.766 ms
```
