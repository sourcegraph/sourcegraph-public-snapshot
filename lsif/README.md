# LSIF code intelligence

[LSIF](https://code.visualstudio.com/blogs/2019/02/19/lsif) is a file format that stores code intelligence information such as hover docstrings, definitions, and references.

Sourcegraph receives LSIF data via [src-cli](https://github.com/sourcegraph/src-cli), usually from a build running within CI, then uses that information to provide fast and precise code intelligence when viewing files.

The HTTP [server](src/server.ts) runs behind Sourcegraph (for auth) and services requests for relevant LSP queries. LSIF dump uploads received by the server are stored in a temporary file to be asynchronously processed by the [worker](src/worker.ts). The server and worker communicate via [node-resque](https://github.com/taskrabbit/node-resque), a work broker powered by Redis.

## Usage documentation

See [the usage documentation on Sourcegraph.com](https://docs.sourcegraph.com/user/code_intelligence/lsif).

## Database Configuration

The LSIF processes store most of its data in SQLite repositories on a shared disk that are written once by a worker on LSIF dump upload, and read many times by the APIs to answer LSIF/LSP queries. Cross-repository and commit graph data is stored in Postgres, as this database requires many concurrent writers (which is an unsafe operation for SQLite in a networked application). The LSIF processes retrieve PostgreSQL connection configuration from the frontend process on startup.

We rely on the Sourcegraph frontend to apply our DB migrations. These live in the `/migrations` folder. This means:

- The server and worker wait for the frontend to apply the migration version it cares about before starting.
- We (and more importantly, site admins) only have to care about a single set of DB schema migrations. This is the primary property we benefit from by doing this.

## Migrations

To add a new migration for the tables used by the LSIF processes, create a new migration in the frontend according to the instructions in [the migration documentation](../migrations/README.md). Then, update the value of `MINIMUM_MIGRATION_VERSION` in [connection.ts](./src/connection.ts) to be the timestamp from the generated filename.

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

## Service documentation

See [the docs/ directory](./docs).
