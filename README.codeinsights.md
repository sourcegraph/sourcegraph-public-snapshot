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

## Insert data

```
INSERT INTO histogram_events(time,value,metadata,repo_id) VALUES(now(), 0.5, '{"hello": "world"}', 2);
```

## Query data

```
SELECT * FROM histogram_events ORDER BY time DESC LIMIT 100;
```
