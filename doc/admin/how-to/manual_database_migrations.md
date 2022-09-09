# How to run `migrator` operations

> NOTE: It is encouraged for users to always use a recent release of the `migrator`, even against older Sourcegraph instances. This is especially true for commands such as `downgrade`, `drift`, `run-out-of-band-migrations`, and `upgrade`, which all work against Sourcegraph versions as old as 3.20.
>
> The `up` command is a notable exception. See the command documentation for additional details.

The `migrator` is a service that runs as an initial step of the startup process for [Kubernetes](../deploy/kubernetes/update.md#database-migrations) and [Docker-compose](../deploy/docker-compose/index.md#database-migrations) instance deployments. This service is also designed to be invokable directly by a site administrator to perform common tasks dealing with database state.

The [commands](#commands) section below details the legal commands with which the `migrator` service can be invoked. The [environments](#environments) section below details how to supply those commands to a `migrator` instance that has access to your Sourcegraph database.

## Commands

The `migrator` service exposes the following commands:

### add-log

Usage: **`add-log -db=<schema> -version=<version> [-up=true]`**

The `add-log` command adds an entry to the migration log after a site administrator has explicitly applied the contents of a migration file. The `-db` flag specifies the target schema to modify. The `-version` flag specifies the migration version. The `-up` flag specifies the migration direction.

This command may be performed by a site-administrator as part of [repairing a dirty database](./dirty_database.md#3-add-a-migration-log-entry).

### downgrade

Usage: **`downgrade --from=<current version> --to=<target version> [--skip-version-check=false] [--dry-run=false]`**

The `downgrade` command performs database schema migrations and (reverse-applied) out-of-band migrations to transform existing data into the shaped expected by an older Sourcegraph instance version. The `--from` and `--to` flags both accept Sourcegraph release versions _without the patch_ (e.g., `v3.42`) and dictate the bounds of the migration.

If the flag `--skip-version-check` is supplied, then the `migrator` will not assert that the previously running instance version matches the given `--from` value.

If the flag `--dry-run` is supplied, then the downgrade plan will be printed but not executed.

The flags `--unprivileged-only` and `--noop-privileged` are also defined on this command to control the behavior of the `migrator` in the presence of [privileged migrations](./privileged_migrations.md).

Note that the flags `--ignore-single-dirty-log` and `ignore-single-pending-log` available on the commands `up`, `upto`, and `downto` are essentially on-by-default for this command. Successive invocations of `downgrade` and `upgrade` will always re-attempt the last failed or attempted-but-unfinished migration.

### downto

Usage: **`downto -db=<schema> -target=<target>,<target>,...`**

The `downto` command revert any applied migrations that are children of the given targets - this effectively "resets" the schema to the target version. The `-db` flag signifies the target schema to modify. The `-target` flag signifies a set of targets whose proper ancestors should be reverted. Comma-separated values are accepted.

The flags `--ignore-single-dirty-log` and `--ignore-single-pending-log` can be supplied to re-attempt to apply the **next** migration that was marked as errored or as incomplete (respectively). For usage, see the guide on [How to troubleshoot a dirty database](dirty_database.md#0-attempt-re-application).

The flags `--unprivileged-only` and `--noop-privileged` are also defined on this command to control the behavior of the `migrator` in the presence of [privileged migrations](./privileged_migrations.md).

### drift

Usage: **`drift -db=<schema> [--version=<version>] [--file=<path to description file>]`**

The `drift` command describes the current (live) database schema and compares it against the expected schema at the given version. The output of this command will include all relevant schema differences that could affect application correctness and performance. When schema drift is detected, a diff of the expected and actual Postgres object definitions will be shown, along with instructions on how to manually resolve the disparity.

The `--version` flag is expected to be a Sourcegraph release version _including a patch_ (e.g., `v3.42.1`). If the flag `--file` is supplied, then the given filepath is read as the schema description file. This is useful for airgapped instances that do not have access to the public Sourcegraph GitHub repository or the public GCS bucket where old revisions have been backfilled. Exactly one of `--version` and `--file` must be supplied.

### run-out-of-band-migrations

Usage: **`run-out-of-band-migrations [-id <id> [, -id <id> ...]] [--apply-reverse=false]`**

The `run-out-of-band-migrations` command runs out-of-band migrations within the `migrator`. If no `-id` is supplied, then all registered out-of-band migrations will be invoked. If the flag `--apply-reverse` is supplied, then the invoked migrations will be migrated in the down direction.

This command may be performed by a site-administrator as part of [repairing an unfinished migration](./unfinished_migration.md).

### up

Usage: **`up [-db=all] [--skip-upgrade-validation=false] [--skip-oobmigration-validation=false]`**

> WARNING: The target migration leaves of this command are defined at `migrator` **compile time** and does not accept a version argument. This is the only command where the Sourcegraph instance version and `migrator` version are expected to match.

The `up` command (the default behavior of the `migrator` service) applies all migrations defined at the time the `migrator` was built. The `-db` flag signifies the target schema(s) to modify. Comma-separated values are accepted. Supply `all` (the default) to migrate all schemas.

This command is used on Sourcegraph instance startup to ensure the database schema is up to date prior to starting other services that depend on the database. Users should prefer the command [`upto`](#upto), which accepts more explicit bounds and does not depend on the migrator compilation version.

Note that if `-db=all`, the configuration flag `DISABLE_CODE_INSIGHTS` is not enabled, and the `codeinsights-db` is unavailable, the operation will fail. In this case, supply an explicit `-db` flag (e.g., `-db=frontend,codeintel`).

If the flag `--skip-upgrade-validation` is supplied, then the current version of the schema will not be read to assert the [standard upgrade policy](../updates/index.md#standard-upgrades) is being followed. If the flag `--skip-oobmigration-validation` is supplied, then the progress of out-fo-band migrationsw ill not be read to assert completion of newly deprecated migrations.

The flags `--ignore-single-dirty-log` and `--ignore-single-pending-log` can be supplied to re-attempt to apply the **next** migration that was marked as errored or as incomplete (respectively). For usage, see the guide on [How to troubleshoot a dirty database](dirty_database.md#0-attempt-re-application).

The flags `--unprivileged-only` and `--noop-privileged` are also defined on this command to control the behavior of the `migrator` in the presence of [privileged migrations](./privileged_migrations.md).

### upgrade

Usage: **`upgrade --from=<current version> --to=<target version> [--skip-version-check=false] [--dry-run=false]`**

The `upgrade` command performs database schema migrations and out-of-band migrations to transform existing data into the shaped expected by a given Sourcegraph instance version. The `--from` and `--to` flags both accept Sourcegraph release versions _without the patch_ (e.g., `v3.42`) and dictate the bounds of the migration.

This command is used by site-administrators to perform [multi-version upgrades](../updates/index.md#multi-version-upgrades).

If the flag `--skip-version-check` is supplied, then the `migrator` will not assert that the previously running instance version matches the given `--from` value.

If the flag `--dry-run` is supplied, then the upgrade plan will be printed but not executed.

The flags `--unprivileged-only` and `--noop-privileged` are also defined on this command to control the behavior of the `migrator` in the presence of [privileged migrations](./privileged_migrations.md).

Note that the flags `--ignore-single-dirty-log` and `ignore-single-pending-log` available on the commands `up`, `upto`, and `downto` are essentially on-by-default for this command. Successive invocations of `upgrade` and `downgrade` will always re-attempt the last failed or attempted-but-unfinished migration.

### upto

Usage: **`upto -db=<schema> -target=<target>,<target>,...`**

The `upto` command ensures a given migration has been applied, and may apply dependency migrations. The `-db` flag signifies the target schema to modify. The `-target` flag signifies the migration to apply. Comma-separated values are accepted.

If the flag `--skip-upgrade-validation` is supplied, then the current version of the schema will not be read to assert the [standard upgrade policy](../updates/index.md#standard-upgrades) is being followed. If the flag `--skip-oobmigration-validation` is supplied, then the progress of out-fo-band migrationsw ill not be read to assert completion of newly deprecated migrations.

The flags `--ignore-single-dirty-log` and `--ignore-single-pending-log` can be supplied to re-attempt to apply the **next** migration that was marked as errored or as incomplete (respectively). For usage, see the guide on [How to troubleshoot a dirty database](dirty_database.md#0-attempt-re-application).

The flags `--unprivileged-only` and `--noop-privileged` are also defined on this command to control the behavior of the `migrator` in the presence of [privileged migrations](./privileged_migrations.md).

### validate

Usage: **`validate [-db=all] [--skip-out-of-band-migrations=false]`**

The `validate` command validates the current state of the database (both schema as well as data related to out-of-band migrations). The `-db` flag signifies the target schema(s) to validate. Comma-separated values are accepted. Supply `all` (the default) to validate all schemas.

This command is used on Sourcegraph instance startup of database-dependent services to ensure that the migrator has been run to the expected version.

Note that if `-db=all`, the configuration flag `DISABLE_CODE_INSIGHTS` is not enabled, and the `codeinsights-db` is unavailable, the operation will fail. In this case, supply an explicit `-db` flag (e.g., `-db=frontend,codeintel`).

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

To run the migrator locally, simply run `go run ./cmd/migrator` or `go run ./enterprise/cmd/migrator`.

Many of the commands detailed above are also available via `sg`. Replace `migrator` with `sg migration ...`. There are a few command registered on the `migrator` but not on `sg` (e.g., `upgrade` and `downgrade`), as local environments are a bit of a different beast than environments performing upgrades only along tagged releases.
