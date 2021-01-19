## codeinsights-db Docker Image

We republish the TimescaleDB (open source) Docker image under sourcegraph/codeinsights-db to ensure it uses our standard naming and versioning scheme. This is done in `docker-images/codeinsights-db/`.

## Getting a psql prompt (dev server)

```sh
docker exec -it codeinsights-db psql -U postgres
```

## Migrations

Since TimescaleDB is just Postgres (with an extension), we use the same SQL migration framework we use for our other Postgres databases. `migrations/codeinsights` contains the migrations for the Code Insights database, they are executed when the frontend starts up (as is the same with e.g. codeintel DB migrations.)

### Add a new migration

To add a new migration, use:

```
./dev/db/add_migration.sh codeinsights MIGRATION_NAME
```

See [migrations/README.md](migrations/README.md) for more information

# Random stuff

## Upsert repo names

```
WITH e AS(
    INSERT INTO repo_names(name)
    VALUES ('github.com/gorilla/mux-original')
    ON CONFLICT DO NOTHING
    RETURNING id
)
SELECT * FROM e
UNION
    SELECT id FROM repo_names WHERE name='github.com/gorilla/mux-original';

WITH e AS(
    INSERT INTO repo_names(name)
    VALUES ('github.com/gorilla/mux-renamed')
    ON CONFLICT DO NOTHING
    RETURNING id
)
SELECT * FROM e
UNION
    SELECT id FROM repo_names WHERE name='github.com/gorilla/mux-renamed';
```

## Upsert event metadata

Upsert metadata, getting back ID:

```
WITH e AS(
    INSERT INTO metadata(metadata)
    VALUES ('{"hello": "world", "languages": ["Go", "Python", "Java"]}')
    ON CONFLICT DO NOTHING
    RETURNING id
)
SELECT * FROM e
UNION
    SELECT id FROM metadata WHERE metadata='{"hello": "world", "languages": ["Go", "Python", "Java"]}';
```

## Inserting gauge events

```
INSERT INTO gauge_events(
    time,
    value,
    metadata_id,
    repo_id,
    repo_name_id,
    original_repo_name_id
) VALUES(
    now(),
    0.5,
    (SELECT id FROM metadata WHERE metadata = '{"hello": "world", "languages": ["Go", "Python", "Java"]}'),
    2,
    (SELECT id FROM repo_names WHERE name = 'github.com/gorilla/mux-renamed'),
    (SELECT id FROM repo_names WHERE name = 'github.com/gorilla/mux-original')
);
```

## Query data

### All data

```
SELECT * FROM gauge_events ORDER BY time DESC LIMIT 100;
```

### Filter by repo name, returning metadata (may be more optimally queried separately)

```
SELECT *
FROM gauge_events
JOIN metadata ON metadata.id = metadata_id
WHERE repo_name_id IN (
    SELECT id FROM repo_names WHERE name ~ '.*-renamed'
)
ORDER BY time
DESC LIMIT 100;
```

### Filter by metadata containing `{"hello": "world"}`

```
SELECT *
FROM gauge_events
JOIN metadata ON metadata.id = metadata_id
WHERE metadata @> '{"hello": "world"}'
ORDER BY time
DESC LIMIT 100;
```

### Filter by metadata containing Go languages

```
SELECT *
FROM gauge_events
JOIN metadata ON metadata.id = metadata_id
WHERE metadata @> '{"languages": ["Go"]}'
ORDER BY time
DESC LIMIT 100;
```
