# Upgrading Early Versions

Why is your instance soooooooo old

Please stop distribution is hard enough managing srcerers 

1. **If upgrading from 3.26 or before to 3.27 or later**, the `pgsql` and `codeintel-db` databases must be upgraded from Postgres 11 to Postgres 12. If this step is not performed, then the following upgrade procedure will fail fast (and leave all existing data untouched).
  - If using an external database, follow the [upgrading external PostgreSQL instances](../../postgres.md#upgrading-external-postgresql-instances) guide.
  - Otherwise, perform the following steps from the [upgrading internal Postgres instances](../../postgres.md#upgrading-internal-postgresql-instances) guide:
      1. It's assumed that your fork of `deploy-sourcegraph-docker` is up to date with your instance's current version. Pull the upstream changes for `v3.27.0` and resolve any git merge conflicts. We need to temporarily boot the containers defined at this specific version to rewrite existing data to the new Postgres 12 format.
      1. Run `docker-compose up pgsql` to launch new Postgres 12 containers and rewrite the old Postgres 11 data. This may take a while, but streaming container logs should show progress.
      1. Wait until the database container is accepting connections. Once ready, run the command `docker exec pgsql -- psql -U sg -c 'REINDEX database sg;'` to repair indexes that were silently invalidated by the previous data rewrite step. **If you skip this step**, then some data may become inaccessible under normal operation, the following steps are not guaranteed to work, and **data loss will occur**.
      1. Follow the same steps for the `codeintel-db`:
          - Run `docker-compose up codeintel-db` to launch Postgres 12.
          - Run `docker exec codeintel-db -- pgsql -U sg -c 'REINDEX database sg;'` to reindex the database.
