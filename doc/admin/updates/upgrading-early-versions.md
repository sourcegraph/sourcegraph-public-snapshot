# Upgrading Early Versions

- The migrator service was introduced in v3.37.0
- Migrator upgrade commands were introduced in v4.0.0

Why is your instance soooooooo old

Please stop distribution is hard enough managing srcerers 

docker compose warning

1. **If upgrading from 3.26 or before to 3.27 or later**, the `pgsql` and `codeintel-db` databases must be upgraded from Postgres 11 to Postgres 12. If this step is not performed, then the following upgrade procedure will fail fast (and leave all existing data untouched).
  - If using an external database, follow the [upgrading external PostgreSQL instances](../../postgres.md#upgrading-external-postgresql-instances) guide.
  - Otherwise, perform the following steps from the [upgrading internal Postgres instances](../../postgres.md#upgrading-internal-postgresql-instances) guide:
      1. It's assumed that your fork of `deploy-sourcegraph-docker` is up to date with your instance's current version. Pull the upstream changes for `v3.27.0` and resolve any git merge conflicts. We need to temporarily boot the containers defined at this specific version to rewrite existing data to the new Postgres 12 format.
      1. Run `docker-compose up pgsql` to launch new Postgres 12 containers and rewrite the old Postgres 11 data. This may take a while, but streaming container logs should show progress.
      1. Wait until the database container is accepting connections. Once ready, run the command `docker exec pgsql -- psql -U sg -c 'REINDEX database sg;'` to repair indexes that were silently invalidated by the previous data rewrite step. **If you skip this step**, then some data may become inaccessible under normal operation, the following steps are not guaranteed to work, and **data loss will occur**.
      1. Follow the same steps for the `codeintel-db`:
          - Run `docker-compose up codeintel-db` to launch Postgres 12.
          - Run `docker exec codeintel-db -- pgsql -U sg -c 'REINDEX database sg;'` to reindex the database.


kubernetes warning

**If upgrading from 3.26 or before to 3.27 or later**, the `pgsql` and `codeintel-db` databases must be upgraded from Postgres 11 to Postgres 12. If this step is not performed, then the following upgrade procedure will fail fast (and leave all existing data untouched).
   - If using an external database, follow the [upgrading external PostgreSQL instances](../../postgres.md#upgrading-external-postgresql-instances) guide.
   - Otherwise, perform the following steps from the [upgrading internal Postgres instances](../../postgres.md#upgrading-internal-postgresql-instances) guide:
      1. It's assumed that your fork of `deploy-sourcegraph` is up to date with your instance's current version. Pull the upstream changes for `v3.27.0` and resolve any git merge conflicts. We need to temporarily boot the containers defined at this specific version to rewrite existing data to the new Postgres 12 format.
      2. Run `kubectl apply -l deploy=sourcegraph -f base/pgsql` to launch a new Postgres 12 container and rewrite the old Postgres 11 data. This may take a while, but streaming container logs should show progress. **NOTE**: The Postgres migration requires enough capacity in its attached volume to accommodate an additional copy of the data currently on disk. Resize the volume now if necessary—the container will fail to start if there is not enough free disk space.
      3. Wait until the database container is accepting connections. Once ready, run the command `kubectl exec pgsql -- psql -U sg -c 'REINDEX database sg;'` issue a reindex command to Postgres to repair indexes that were silently invalidated by the previous data rewrite step. **If you skip this step**, then some data may become inaccessible under normal operation, the following steps are not guaranteed to work, and **data loss will occur**.
      4. Follow the same steps for the `codeintel-db`:
          - Run `kubectl apply -l deploy=sourcegraph -f base/codeintel-db` to launch Postgres 12.
          - Run `kubectl exec codeintel-db -- psql -U sg -c 'REINDEX database sg;'` to issue a reindex command to Postgres.
      5. Leave these versions of the databases running while the subsequent migration steps are performed. If `codeinsights-db` is a container new to your instance, now is a good time to start it as well.
