# Configuration

The LSIF processes store most of their data in SQLite repositories on a shared disk that are written once by a dump processor on LSIF dump upload, and read many times by the APIs to answer LSIF/LSP queries. Cross-repository and commit graph data is stored in Postgres, as this database requires many concurrent writers (which is an unsafe operation for SQLite in a networked application). The LSIF processes retrieve PostgreSQL connection configuration from the frontend process on startup.

We rely on the Sourcegraph frontend to apply our DB migrations. These live in the `/migrations` folder. This means:

- The server and dump processor wait for the frontend to apply the migration version it cares about before starting.
- We (and more importantly, site admins) only have to care about a single set of DB schema migrations. This is the primary property we benefit from by doing this.

## Migrations

To add a new migration for the tables used by the LSIF processes, create a new migration in the frontend according to the instructions in [the migration documentation](../../migrations/README.md). Then, update the value of `MINIMUM_MIGRATION_VERSION` in [postgres.ts](../src/shared/database/postgres.ts) to be the timestamp from the generated filename.
