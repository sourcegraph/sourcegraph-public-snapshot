# Sizes (before)

| table_schema | table_name              | row_estimate  | total  | index  | toast  | table   |
| ------------ | ----------------------- | ------------- | ------ | ------ | ------ | ------- |
| public       | lsif_data_definitions   | 3.077555e+08  | 230 GB | 68 GB  | 186 MB | 162 GB  |
| public       | lsif_data_documents     | 1.1777075e+07 | 145 GB | 793 MB | 141 GB | 3159 MB |
| public       | lsif_data_references    | 4.4469523e+08 | 765 GB | 83 GB  | 432 GB | 250 GB  |
| public       | lsif_data_result_chunks | 4.878871e+06  | 64 GB  | 105 MB | 64 GB  | 284 MB  |

# Original query

```sql
WITH candidates AS (
	SELECT dump_id
	FROM lsif_data_definitions_schema_versions
	WHERE
		min_schema_version <= 1 AND
		max_schema_version >= 1
	ORDER BY dump_id
)
SELECT dump_id, scheme, identifier, data
FROM lsif_data_definitions
WHERE
	dump_id IN (SELECT dump_id FROM candidates) AND
	schema_version = 1
ORDER BY dump_id
LIMIT 10;
```

# Alternate query

```sql
SELECT dump_id, scheme, identifier, data
FROM lsif_data_definitions d
WHERE
    EXISTS (
        SELECT 1
        FROM lsif_data_definitions_schema_versions sv
        WHERE
            sv.dump_id = d.dump_id AND
            sv.min_schema_version <= 1 AND
            sv.max_schema_version >= 1
    ) AND
	schema_version = 1
ORDER BY dump_id
LIMIT 10;
```

# Possible indexes to help query

## New index on dump_id, schema_version

```sql
CREATE INDEX lsif_data_definitions_dump_id_schema_version ON lsif_data_definitions (dump_id, schema_version);
```

Time: 673473.870 ms (11:13.474)

| table_schema | table_name            | row_estimate  | total  | index | toast  | table  |
| ------------ | --------------------- | ------------- | ------ | ----- | ------ | ------ |
| public       | lsif_data_definitions | 3.077555e+08  | 236 GB | 74 GB | 186 MB | 162 GB |
| public       | lsif_data_references  | 4.2517418e+08 | 774 GB | 92 GB | 432 GB | 250 GB |

Original query:

Limit (cost=2.85..34.47 rows=10 width=218) (actual time=523.298..523.309 rows=10 loops=1)
-> Merge Semi Join (cost=2.85..4866515.53 rows=1538777 width=218) (actual time=523.297..523.306 rows=10 loops=1)
Merge Cond: (lsif_data_definitions.dump_id = lsif_data_definitions_schema_versions.dump_id)
-> Index Scan using lsif_data_definitions_dump_id_schema_version on lsif_data_definitions (cost=0.57..4845913.12 rows=1538777 width=218) (actual time=521.830..521.835 rows=10 loops=1)
Index Cond: (schema_version = 1)
-> Index Scan using lsif_data_definitions_schema_versions_pkey on lsif_data_definitions_schema_versions (cost=0.29..1014.33 rows=29210 width=4) (actual time=0.018..1.345 rows=1145 loops=1)
Filter: ((min_schema_version <= 1) AND (max_schema_version >= 1))
Rows Removed by Filter: 3695
Planning Time: 0.385 ms
Execution Time: 523.344 ms

Alternate query:

Limit (cost=0.86..43.75 rows=10 width=218) (actual time=5.877..5.884 rows=10 loops=1)
-> Nested Loop (cost=0.86..5872946.80 rows=1369479 width=218) (actual time=5.876..5.881 rows=10 loops=1)
-> Index Scan using lsif_data_definitions_schema_versions_pkey on lsif_data_definitions_schema_versions sv (cost=0.29..17383.52 rows=29210 width=4) (actual time=0.005..1.609 rows=1145 loops=1)
Filter: ((min_schema_version <= 1) AND (max_schema_version >= 1))
Rows Removed by Filter: 3695
-> Index Scan using lsif_data_definitions_dump_id_schema_version on lsif_data_definitions d (cost=0.57..198.46 rows=200 width=218) (actual time=0.003..0.003 rows=0 loops=1145)
Index Cond: ((dump_id = sv.dump_id) AND (schema_version = 1))
Planning Time: 0.358 ms
Execution Time: 5.917 ms

## New index on dump_id including schema_version

```sql
CREATE INDEX lsif_data_definitions_dump_id_with_schema_version ON lsif_data_definitions (dump_id) INCLUDE (schema_version);
```

Time: 568091.245 ms (09:28.091)

| table_schema | table_name            | row_estimate | total  | index | toast  | table  |
| ------------ | --------------------- | ------------ | ------ | ----- | ------ | ------ |
| public       | lsif_data_definitions | 3.077555e+08 | 236 GB | 74 GB | 186 MB | 162 GB |

DOES NOT HELP EITHER QUERY (both use primary key and filter out rows without include)

## Replace primary key

```sql
CREATE UNIQUE INDEX lsif_data_definitions_pkey_alt ON lsif_data_definitions (dump_id, scheme, identifier) INCLUDE (schema_version);
ALTER TABLE lsif_data_definitions DROP CONSTRAINT lsif_data_definitions_pkey;
ALTER TABLE lsif_data_definitions ADD CONSTRAINT lsif_data_definitions_pkey PRIMARY KEY USING INDEX lsif_data_definitions_pkey_alt;
```

Time: 2621394.080 ms (43:41.394)

| table_schema | table_name            | row_estimate | total  | index | toast  | table  |
| ------------ | --------------------- | ------------ | ------ | ----- | ------ | ------ |
| public       | lsif_data_definitions | 3.077555e+08 | 187 GB | 24 GB | 186 MB | 162 GB |
