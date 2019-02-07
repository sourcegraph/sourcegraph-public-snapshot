# Postgres

## 🔢 Version requirements

Sourcegraph uses Postgres as its main internal database. We support any version **starting from 9.6**.

## ⤴️ Upgrades

This section describes the possible Postgres upgrade procedures for the different Sourcegraph deployment types.

⚠️ Upgrades **require downtime**. Ensure the existing Sourcegraph deployment is stopped before proceeding and that you communicated to your users about this.

----

### 🌈 Automatic upgrades in `sourcegraph/server` deployments

Existing Postgres data will be automatically migrated when a release of the `sourcegraph/server` Docker image ships with a new version of Postgres. For the upgrade to proceed, the Docker socket **must be mounted** the first time the new Docker image is ran. This is needed to run the [Postgres upgrade containers](https://github.com/tianon/docker-postgres-upgrade) in the
Docker host. When the upgrade is done, the container can be restarted without mounting the Docker socket.

**Ensure** the previous `sourcegraph/server` image is completely stopped before running:

```bash
# Add "--env=SRC_LOG_LEVEL=dbug" below for verbose logging.
docker run -p 7080:7080 -p 2633:2633 --rm \
  -v ~/.sourcegraph/config:/etc/sourcegraph \
  -v ~/.sourcegraph/data:/var/opt/sourcegraph \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  sourcegraph/server:3.0.1
```

When using the `sourcegraph/server` image in other environments (e.g. Kubernetes), please refer to official documentation on how to mount the Docker socket for the upgrade procedure.

----

### 🐾 Manual upgrades in `sourcegraph/server` deployments

These instructions can be followed when manual Postgres upgrades are preferred. **Ensure** the previous `sourcegraph/server` image is completely stopped before proceeding.

Assuming Postgres data must be upgraded from `9.6` to `11` and your Sourcegraph directory is at `$HOME/.sourcegraph`, here is how it would be done:


```bash
#!/usr/bin/env bash

set -xeuo pipefail

export OLD=${OLD:-"9.6"}
export NEW=${NEW:-"11"}
export SRC_DIR=${SRC_DIR:-"$HOME/.sourcegraph"}

docker run \
  -w /tmp/upgrade \
  -v "$SRC_DIR/data/postgres-$NEW-upgrade:/tmp/upgrade" \
  -v "$SRC_DIR/data/postgresql:/var/lib/postgresql/$OLD/data" \
  -v "$SRC_DIR/data/postgresql-$NEW:/var/lib/postgresql/$NEW/data" \
  "tianon/postgres-upgrade:$OLD-to-$NEW"

mv "$SRC_DIR/data/"{postgresql,postgresql-$OLD}
mv "$SRC_DIR/data/"{postgresql-$NEW,postgresql}

curl -fsSL -o "$SRC_DIR/data/postgres-$NEW-upgrade/optimize.sh" https://raw.githubusercontent.com/sourcegraph/sourcegraph/master/cmd/server/rootfs/postgres-optimize.sh

docker run \
  --entrypoint "/bin/bash" \
  -w /tmp/upgrade \
  -v "$SRC_DIR/data/postgres-$NEW-upgrade:/tmp/upgrade" \
  -v "$SRC_DIR/data/postgresql:/var/lib/postgresql/data" \
  "postgres:$NEW" \
  -c 'chown -R postgres $PGDATA . && gosu postgres bash ./optimize.sh $PGDATA'
```

----

### 🌈 Automatic upgrades in `github.com/sourcegraph/deploy-sourcegraph` deployments

The automatic upgrade process runs at startup time in the `sourcegraph/postgresql-11` image and it requires certain environment variables to be set. If you have previously customized `PGUSER`, `PGDATABASE` or `PGDATA` then you are required to specify the corresponding `PG*OLD` and `PG*NEW` environment variables. Below are the defaults and documentation on what each variable is used for:

- `POSTGRES_PASSWORD=''`: Password of `PGUSERNEW` if it is newly created (i.e when `PGUSERNEW` didn't exist in the old database).
- `PGUSEROLD=sg`: A user that exists in the old database that can be used to authenticate intermediate upgrade operations.
- `PGUSERNEW=sg`: A user that must exist in the new database after the upgrade is done (i.e. it'll be created if it didn't exist already).
- `PGDATABASEOLD=sg`: A database that exists in the old database that can be used to authenticate intermediate upgrade operations. (e.g `psql -d`)
- `PGDATABASENEW=sg`: A database that must exist in the new database after the upgrade is done (i.e. it'll be created if it didn't exist already).
- `PGDATAOLD=/data/pgdata`: The data directory containing the files of the old Postgres database to be upgraded.
- `PGDATANEW=/data/pgdata-11`: The data directory containing the upgraded Postgres data files, used by the new version of Postgres.

Additionally the upgrade process assumes it can write to the parent directory of `PGDATAOLD`.

----


### ℹ️ Upgrades of external Postgres clusters or instances

When running an external Postgres cluster (or instance) please refer to the documentation of your provider on how to perform upgrade procedures.

- [AWS RDS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_UpgradeDBInstance.PostgreSQL.html)
- [AWS Aurora](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/USER_UpgradeDBInstance.Upgrading.html)
- [GCP CloudSQL](https://cloud.google.com/sql/docs/postgres/db-versions)
- [Azure DB](https://docs.microsoft.com/en-us/azure/postgresql/concepts-supported-versions#managing-updates-and-upgrades)
- [Heroku](https://devcenter.heroku.com/articles/upgrading-heroku-postgres-databases)
- [EnterpriseDB](https://www.enterprisedb.com/docs/en/9.6/pg/upgrading.html)
- [Citus](http://docs.citusdata.com/en/v8.1/admin_guide/upgrading_citus.html)
- [Aiven Postgres](https://help.aiven.io/postgresql/operations/how-to-perform-a-postgresql-in-place-major-version-upgrade)
- [Your own Postgres](https://www.postgresql.org/docs/11/pgupgrade.html)
