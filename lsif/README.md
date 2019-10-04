# LSIF code intelligence

[LSIF](https://code.visualstudio.com/blogs/2019/02/19/lsif) is a file format that stores code intelligence information such as hover docstrings, definitions, and references.

Sourcegraph receives and stores LSIF files uploaded using [upload.sh](upload.sh) (usually used in CI, similar to [Codecov's Bash Uploader](https://docs.codecov.io/docs/about-the-codecov-bash-uploader)), then uses that information to provide fast and precise code intelligence when viewing files.

The HTTP [server](src/server.ts) runs behind Sourcegraph (for auth) and services requests for relevant LSP queries. LSIF dump uploads received by the server are stored in a temporary file to be asynchronously processed by the [worker](src/worker.ts). The server and worker communicate via [node-resque](https://github.com/taskrabbit/node-resque), a work broker powered by Redis.

## Database Configuration

The LSIF processes store most of its data in SQLite repositories on a shared disk that are written once by a worker on LSIF dump upload, and read many times by the APIs to answer LSIF/LSP queries. Cross-repository and cross-commit data is stored in Postgres, as this database requires many concurrent writer (which is an unsafe operation for SQLite in a networked application). The LSIF processes read PostgreSQL connection configuration from the [`PG*` environment variables](http://www.postgresql.org/docs/current/static/libpq-envars.html). These value should be configured to point to the same Postgres database as the rest of Sourcegraph. The LSIF processes will connect to the database named `{PGDATABASE}_lsif`, where `{PGDATABASE}` is the primary Sourcegraph database. The processes also require access to the primary Sourcegraph database to check the status of migration on application startup. Although it currently resides on the same Postgres instance (pod, container, or physical machine depending on the deployment context), separation of this data allows the LSIF database to be easily moved to a completely separate Postgres instance if we find that LSIF access and write patterns put too much stress on a single node.

## Migrations

To add a new migration for the LSIF database, simply create a migration to be read by the frontend with the following content. This will cause the frontend to execute the SQL text in the context of the LSIF database instead of the primary one (to which it is connected).

```
SELECT remote_exec('_lsif', '
    BEGIN;
    -- migration to apply in LSIF database
    COMMIT;
');
```

You will also need to update the `preview` in [conection.ts](./src/connection.ts) to be the timestamp from the generated filename.

## API

### `/upload`

Receives an LSIF dump encoded as JSON lines.

URL query parameters:

- `repository`: the name of the repository (e.g. `github.com/sourcegraph/codeintellify`)
- `commit`: the 40 character hash of the commit

The request body must be HTML form data with a single file (e.g. `curl -F "data=@file.lsif" ...`).

### `/request`

Performs a `hover`, a `definitions`, or a `references` request for the given repository@commit and returns the result. Fails if there is no LSIF data for the given repository@commit.

The request body must be a JSON object with these properties:

- `repository`: the name of the repository (e.g. `github.com/sourcegraph/codeintellify`)
- `commit`: the 40 character hash of the commit
- `method`: `hover`, `definitions`, or `references`
- `path`: the file path in the repository.
- `position`: the zero-based `{ line, character }` in the file at which the request is being made
