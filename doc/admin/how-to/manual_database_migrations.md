# How to run `migrator` operations

> NOTE: The `migrator` service is only available in versions `3.37` and later.

The `migrator` is a service that runs as an initial step of the upgrade process for [Kubernetes](../deploy/kubernetes/update.md#database-migrations) and [Docker-compose](../deploy/docker-compose/index.md#database-migrations) instance deployments. This service is also designed to be invokable directly by a site administrator to perform common tasks dealing with database state.

The [commands](#commands) section below details the legal commands with which the `migrator` service can be invoked. The [environments](#environments) section below details how to supply those commands to a `migrator` instance that has access to your Sourcegraph database.

## Commands

The `migrator` service exposes the following commands:

### up

Usage: **`up [-db=all]`**

The `up` command (the default behavior of the `migrator` service) applies all migrations. The `-db` flag signifies the target schema(s) to modify. Comma-separated values are accepted. Supply `all` (the default) to migrate all schemas.

> NOTE: These default behavior applies to all three databases. If the configuration flag `DISABLE_CODE_INSIGHTS` is set and the `codeinsights-db` is unavailable, the operation will fail. To work around this, explicitly supply database(s) via the `-db` flag (e.g., `-db=frontend,codeintel`).

### upto

Usage: **`upto -db=<schema> -target=<target>,<target>,...`**

The `upto` command ensures a given migration has been applied, and may apply dependency migrations. The `-db` flag signifies the target schema to modify. The `-target` flag signifies the migration to apply. Comma-separated values are accepted.

### downto

Usage: **`downto -db=<schema> -target=<target>,<target>,...`**

The `downto` command revert any applied migrations that are children of the given targets - this effectively "resets" the schema to the target version. The `-db` flag signifies the target schema to modify. The `-target` flag signifies a set of targets whose proper ancestors should be reverted. Comma-separated values are accepted.

### validate

Usage: **`validate [-db=all]`**

The `validate` command validates the current state of the database. The `-db` flag signifies the target schema(s) to validate. Comma-separated values are accepted. Supply `all` (the default) to validate all schemas.

> NOTE: These default behavior applies to all three databases. If the configuration flag `DISABLE_CODE_INSIGHTS` is set and the `codeinsights-db` is unavailable, the operation will fail. To work around this, explicitly supply database(s) via the `-db` flag (e.g., `-db=frontend,codeintel`).

### add-log

Usage: **`add-log -db=<schema> -version=<version> [-up=true]`**

The `add-log` command adds an entry to the migration log after a site administrator has explicitly applied the contents of a migration file. The `-db` flag specifies the target schema to modify. The `-version` flag specifies the migration version. The `-up` flag specifies the migration direction.

## Environments

To run a `migrator` command, follow the guide for your Sourcegraph distribution type:

- [Kubernetes](#kubernetes) 
- [Docker-compose](#docker-compose)
- [Local development](#local-development)

### Kubernetes

Run the following commands in the root of your `deploy-sourcegraph` fork.

First, modify the `migrator` manifest to update two fields: the `spec.template.spec.containers[0].args` field, which selects the target operation, and the `spec.template.spec.containers[0].image` field, which controls the version of the migrator binary (and, consequently, the set of embedded migration definitions).

The following example uses `yq`, but these values can also be updated manually in thee `configure/migrator/migrator.Job.yaml` file.

```bash
export MIGRATOR_SOURCEGRAPH_VERSION="..."

# Update the "image" value of the migrator container in the manifest
yq eval -i \
  '.spec.template.spec.containers[0].image = "index.docker.io/sourcegraph/migrator:" + strenv(MIGRATOR_SOURCEGRAPH_VERSION)' \
  configure/migrator/migrator.Job.yaml

# Optional (defaults to `["up", "-db", "all"]`)
# Update the "args" value of the migrator container in the manifest
yq eval -i \
  '.spec.template.spec.containers[0].args = ["add", "quoted", "arguments"]' \
  configure/migrator/migrator.Job.yaml
```

Next, apply the job and wait for it to complete.

> NOTE: These values will work for a standard deployment of Sourcegraph with all three databases running in-cluster. If you've customized your deployment (e.g., using an external database service), you will have to modify the environment variables in `configure/migrator/migrator.Job.yaml` accordingly.

```bash
kubectl delete -f configure/migrator/migrator.Job.yaml --ignore-not-found=true

# Apply the manifest and wait for the operation to complete before continuing
# Note: -1s timeout will wait "forever"
kubectl apply -f configure/migrator/migrator.Job.yaml
kubectl wait -f configure/migrator/migrator.Job.yaml --for=condition=complete --timeout=-1s
```

You should see something like the following printed to the terminal:

```text
job.batch "migrator" deleted
job.batch/migrator created
job.batch/migrator condition met
```

The log output of the `migrator` should include `INFO`-level logs and successfully terminate with `migrator exited with code 0`. If you see an error message or any of the databases have been flagged as "dirty", please follow ["How to troubleshoot a dirty database"](../../../admin/how-to/dirty_database.md). A dirty database will not affect your ability to use Sourcegraph however it will need to be resolved to upgrade further. If you are unable to resolve the issues, contact support at <mailto:support@sourcegraph.com> for further assistance. Otherwise, you are now safe to upgrade Sourcegraph.

### Docker compose

Run the following commands on your Docker host.

> NOTE: These values will work for a standard docker-compose deployment of Sourcegraph. If you've customized your deployment (e.g., using an external database service), you will have to modify the environment variables accordingly.

```bash
export MIGRATOR_SOURCEGRAPH_VERSION="..."

docker run \
  --rm \
  --name migrator_$MIGRATOR_SOURCEGRAPH_VERSION \
  -e PGHOST='pgsql' \
  -e PGPORT='5432' \
  -e PGUSER='sg' \
  -e PGPASSWORD='sg' \
  -e PGDATABASE='sg' \
  -e PGSSLMODE='disable' \
  -e CODEINTEL_PGHOST='codeintel-db' \
  -e CODEINTEL_PGPORT='5432' \
  -e CODEINTEL_PGUSER='sg' \
  -e CODEINTEL_PGPASSWORD='sg' \
  -e CODEINTEL_PGDATABASE='sg' \
  -e CODEINTEL_PGSSLMODE='disable' \
  -e CODEINSIGHTS_PGHOST='codeinsights-db' \
  -e CODEINSIGHTS_PGPORT='5432' \
  -e CODEINSIGHTS_PGUSER='postgres' \
  -e CODEINSIGHTS_PGPASSWORD='password' \
  -e CODEINSIGHTS_PGDATABASE='postgres' \
  -e CODEINSIGHTS_PGSSLMODE='disable' \
  --network=docker-compose_sourcegraph \
  sourcegraph/migrator:$MIGRATOR_SOURCEGRAPH_VERSION \
  # Optional (defaults to `["up", "-db", "all"]`)
  "add" "quoted" "arguments"
```

Observe the output of the `migrator` container via:

```bash
docker logs migrator_$SOURCEGRAPH_VERSION
```

The log output of the `migrator` should include `INFO`-level logs and successfully terminate with `migrator exited with code 0`. If you see an error message or any of the databases have been flagged as "dirty", please follow ["How to troubleshoot a dirty database"](../../../admin/how-to/dirty_database.md). A dirty database will not affect your ability to use Sourcegraph however it will need to be resolved to upgrade further. If you are unable to resolve the issues, contact support at <mailto:support@sourcegraph.com> for further assistance. Otherwise, you are now safe to upgrade Sourcegraph.

### Local development

The command detailed document are also available via `sg`. Replace `migrator` with `sg migration ...`.
