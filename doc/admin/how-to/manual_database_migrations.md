# How to run `migrator` operations

> NOTE: It is encouraged for users to always use a recent release of the `migrator`, even against older Sourcegraph instances. This is especially true for commands such as `downgrade`, `drift`, `run-out-of-band-migrations`, and `upgrade`, which all work against Sourcegraph versions as old as 3.20.
>
> The `up` command is a notable exception. See the command documentation for additional details.

The `migrator` is a service that runs as an initial step of the startup process for [Kubernetes](../deploy/kubernetes/update.md#database-migrations) and [Docker-compose](../deploy/docker-compose/index.md#database-migrations) instance deployments. This service is also designed to be invokable directly by a site administrator to perform common tasks dealing with database state.

The [commands](#commands) section below details the legal commands with which the `migrator` service can be invoked. The [environments](#environments) section below details how to supply those commands to a `migrator` instance that has access to your Sourcegraph database.

## Commands

The `migrator` service exposes the following commands:

### upgrade

The `upgrade` command performs database schema migrations and out-of-band migrations to rewrite existing instance data in-place into the shaped expected by a given target Sourcegraph version. This command is used by site-administrators to perform [multi-version upgrades](../updates/index.md#multi-version-upgrades).

```
upgrade \
    -from=<current version> -to=<target version> \
    [-dry-run=false] \
    [-disable-animation=false] \
    [-skip-version-check=false] [-skip-drift-check=false] \
    [-unprivileged-only=false] [-noop-privileged=false] [-privileged-hash=<hash>]
```

**Required arguments**:

- `-from`: The current Sourcegraph release version (_without the patch_; e.g., `v3.36`)
- `-to`: The target Sorucegraph release version (_without the patch_; e.g., `v4.0`)

**Optional arguments**:

- `-dry-run`: Print the steps of the upgrade but do not execute them.
- `-disable-animation`: Print plain log messages instead of an animated progress bar.
- `-skip-version-check`: Skip comparing the current instance version against `-from`.
- `-skip-drift-check`: Skip comparing the database schema shape against the schema defined by `-from`.
- `-unprivileged-only` and `-noop-privileged`: Controls behavior of schema migrations the presence of [privileged definitions](./privileged_migrations.md).

**Notes**:

- Successive invocations of this command will re-attempt the last failed or attempted (but incomplete) migration. This command run as if the `-ignore-single-{dirty,pending}-log` flags supplied by the commands `up`, `upto`, and `downto` were enabled.
- This command checks that the schema of the database is in the correct state for the current version, if schema drift is detected it must be resolved before completing the upgrade. [Learn more here.](./schema-drift.md).
- Successive invocations of this command may *cause* database drift when partial progress is made. When making a subsequent upgrade attempt, invoke this command with `-skip-drift-check` ignore the failing startup check.

### drift

The `drift` command describes the current (live) database schema and compares it against the expected schema at the given version. The output of this command will include all relevant schema differences that could affect application correctness and performance. When schema drift is detected, a diff of the expected and actual Postgres object definitions will be shown, along with instructions on how to manually resolve the disparity. [Learn more here.](./schema-drift.md)

```
drift \
    -db=<schema> \
    [-version=<version>] \
    [-file=<path to description file>]
```

**Required arguments**:

- `-db`: The target schema to inspect. *Ex: frontend, codeintel, codeinsights*

**Mutually exclusive arguments**:

Exactly one of `-version` and `-file` must be supplied.

- `-version`: The instance's current Sourcegraph release version _including a patch_ (e.g., `v3.42.1`).
- `-file`: The filepath to a local schema description file. This is useful for airgapped instances that do not have access to the public Sourcegraph GitHub repository or the public GCS bucket where old revisions have been backfilled.

### downgrade

The `downgrade` command performs database schema migrations and (reverse-applied) out-of-band migrations to rewrite existing instance data in-place into the shaped expected by a given target Sourcegraph version.

```
downgrade \
    -from=<current version> -to=<target version> \
    [-dry-run=false] \
    [-disable-animation=false] \
    [-skip-version-check=false] [-skip-drift-check=false] \
    [-unprivileged-only=false] [-noop-privileged=false] [-privileged-hash=<hash>]
```

**Required arguments**:

- `-from`: The current Sourcegraph release version (_without the patch_; e.g., `v3.36`)
- `-to`: The target Sorucegraph release version (_without the patch_; e.g., `v4.0`)

**Optional arguments**:

- `-dry-run`: Print the steps of the upgrade but do not execute them.
- `-disable-animation`: Print plain log messages instead of an animated progress bar.
- `-skip-version-check`: Skip comparing the current instance version against `-from`.
- `-skip-drift-check`: Skip comparing the database schema shape against the schema defined by `-from`.
- `-unprivileged-only` and `-noop-privileged`: Controls behavior of schema migrations the presence of [privileged definitions](./privileged_migrations.md).

**Notes**:

- Successive invocations of this command will re-attempt the last failed or attempted (but incomplete) migration. This command run as if the `-ignore-single-{dirty,pending}-log` flags supplied by the commands `up`, `upto`, and `downto` were enabled.
- Successive invocations of this command may *cause* database drift when partial progress is made. When making a subsequent downgrade attempt, invoke this command with `-skip-drift-check` ignore the failing startup check.

### add-log

The `add-log` command adds an entry to the `migration_logs` table after a site administrator has explicitly applied the contents of a migration definition. This command may be performed by a site-administrator as part of [repairing a dirty database](./dirty_database.md#3-add-a-migration-log-entry).

```
add-log \
    -db=<schema> \
    -version=<version> \
    [-up=true]
```

**Required arguments**:

- `-db`: The target schema to modify.
- `-version`: The migration identifier noted on the log entry.

**Optional arguments**:

- `-up`: The migration direction noted on the log entry.

### validate

The `validate` command validates the current state of the database (both schema and data migration progress). This command is used on Sourcegraph instance startup of database-dependent services to ensure that the migrator has been run to the expected version.

```
validate \
    [-db=all] \
    [-skip-out-of-band-migrations=false]
```

**Optional arguments**:

- `-db`: The target schema(s) to validate. Comma-separated values are allowed.
- `skip-out-of-band-migrations`: Skip validation of out-of-band migrations. Validate the schema only.

**Notes**:

- If `DISABLE_CODE_INSIGHTS` is not set and the `codeinsights-db` is not available, then this command will fail with the default value for the `-db` flag. To resolve, supply `-db=frontend,codeintel` instead.

### up

The `up` command (the default behavior of the `migrator` service) applies all migrations defined at the time the `migrator` was built. This command is used on Sourcegraph instance startup to ensure the database schema is up to date prior to starting other services that depend on the database.

> WARNING: The target migration leaves of this command are defined at `migrator` **compile time** and does not accept a version argument. This is the only command where the Sourcegraph instance version and `migrator` version are expected to match.

Users should generally prefer the command [`upto`](#upto), which accepts more explicit bounds and does not depend on the migrator compilation version.

```
up \
    [-db=all] \
    [-skip-upgrade-validation=false] \
    [-skip-oobmigration-validation=false]
    [-ignore-single-dirty-log=false] [-ignore-single-pending-log=false] \
    [-unprivileged-only=false] [-noop-privileged=false] [-privileged-hash=<hash>]
```

**Optional arguments**:

- `-db`: The target schema(s) to modify. Comma-separated values are allowed.
- `-skip-upgrade-validation`: Skip asserting that the [standard upgrade policy](../updates/index.md#standard-upgrades) is being followed. 
- `-skip-oobmigration-validation`: Skip reading the progress of out-of-band migrations to assert completion of newly deprecated migrations.
- `-ignore-single-dirty-log` and `-ignore-single-pending-log`: Re-attempt to apply the **next** migration that was marked as errored or as incomplete (respectively). See [how to troubleshoot a dirty database](dirty_database.md#0-attempt-re-application).
- `-unprivileged-only` and `-noop-privileged`: Controls behavior of schema migrations the presence of [privileged definitions](./privileged_migrations.md).

**Notes**:

- If `DISABLE_CODE_INSIGHTS` is not set and the `codeinsights-db` is not available, then this command will fail with the default value for the `-db` flag. To resolve, supply `-db=frontend,codeintel` instead.

### upto

The `upto` command ensures a given migration has been applied, and may apply dependency migrations.

```
upto \
    -db=<schema> \
    -target=<target>,<target>,... \
    [-ignore-single-dirty-log=false] [-ignore-single-pending-log=false] \
    [-unprivileged-only=false] [-noop-privileged=false] [-privileged-hash=<hash>]
```

**Required arguments**:

- `-db`: The target schema to modify.
- `-target`: The migration identifier(s) to target (these and unapplied descendants will be applied). Comma-separated values are accepted.

**Optional arguments**:

- `-ignore-single-dirty-log` and `-ignore-single-pending-log`: Re-attempt to apply the **next** migration that was marked as errored or as incomplete (respectively). See [how to troubleshoot a dirty database](dirty_database.md#0-attempt-re-application).
- `-unprivileged-only` and `-noop-privileged`: Controls behavior of schema migrations the presence of [privileged definitions](./privileged_migrations.md).

### downto

The `downto` command revert any applied migrations that are children of the given targets—this effectively "resets" the schema to the target version.

```
downto \
    -db=<schema> \
    -target=<target>,<target>,... \
    [-ignore-single-dirty-log=false] [-ignore-single-pending-log=false] \
    [-unprivileged-only=false] [-noop-privileged=false] [-privileged-hash=<hash>]
```
**Required arguments**:

- `-db`: The target schema to modify.
- `-target`: The migration identifier(s) to target (proper ancestors will be reverted). Comma-separated values are accepted.

**Optional arguments**:

- `-ignore-single-dirty-log` and `-ignore-single-pending-log`: Re-attempt to apply the **next** migration that was marked as errored or as incomplete (respectively). See [how to troubleshoot a dirty database](dirty_database.md#0-attempt-re-application).
- `-unprivileged-only` and `-noop-privileged`: Controls behavior of schema migrations the presence of [privileged definitions](./privileged_migrations.md).

### run-out-of-band-migrations

The `run-out-of-band-migrations` command runs out-of-band migrations within the `migrator`. This command may be performed by a site-administrator as part of [repairing an unfinished migration](./unfinished_migration.md).

```
run-out-of-band-migrations \
    [-id <id>]+ \
    [-apply-reverse=false] \
    [-disable-animation=false]
```

**Required arguments**:

- `-id`: The identifier(s) of the migrations to apply. Multiple flags can be supplied. If no flag is supplied, all migrations are applied.
- `-apply-reverse`: Run migrations in the reverse direction.

**Optional arguments**:

- `-disable-animation`: Print plain log messages instead of an animated progress bar.

### describe

The `describe` command outputs a dump of your database schema.

```
describe \
    -db=<schema> \
    -format=<json|psql> \
    [-out=stdout] \
    [-force=false] \
    [-no-color=false]
```

**Required arguments**:

- `-db`: The target schema to describe.
- `-format`: The format in which the description is output.

**Optional arguments**:

- `-out`: The target output file. If not supplied, the output is printed to stdout.
- `-force`: Overwrite the file.
- `-no-color`: Do not print ANSI color sequences.

## Environments

To run a `migrator` command, follow the guide for your Sourcegraph distribution type:

- [Kubernetes](#kubernetes)
- [Docker / Docker-compose](#docker--docker-compose)
- [Local development](#local-development)

### Kubernetes

Run the following commands in the root of your `deploy-sourcegraph` fork.

First, modify the `migrator` manifest to update two fields: the `spec.template.spec.containers[0].args` field, which selects the target operation, and the `spec.template.spec.containers[0].image` field, which controls the version of the migrator binary (and, consequently, the set of embedded migration definitions).

The following example uses `yq`, but these values may also be updated manually by opening the `configure/migrator/migrator.Job.yaml` file in an editor of your choice and editing the `image` and `args` key value pairs.

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

The above `yq` commands alter the `configure/migrator/migrator.Job.yaml` file. For example producing the following key pairs:
```yaml
image: "index.docker.io/sourcegraph/migrator:4.1.3@sha256:0dc6543f0a755e46d962ba572d501559915716fa55beb3aa644a52f081fcd57e"
```
```yaml
args: ["upgrade", "--from=3.40.2", "--to=4.1.2"]
```

Next, apply the job and wait for it to complete.

> NOTE: These values will work for a standard deployment of Sourcegraph with all three databases running in-cluster. If you've customized your deployment (e.g., using an external database service), you will have to modify the environment variables in `configure/migrator/migrator.Job.yaml` accordingly.

```bash
kubectl delete -f configure/migrator/migrator.Job.yaml --ignore-not-found=true

# Apply the manifest and wait for the operation to complete before continuing
# Note: -1s timeout will wait "forever"
kubectl apply -f configure/migrator/migrator.Job.yaml
kubectl wait -f configure/migrator/migrator.Job.yaml --for=condition=complete --timeout=-1s

# Optionally: check migrator logs for progress
kubectl logs job.batch/migrator -f
```

You should see something like the following printed to the terminal:

```text
job.batch "migrator" deleted
job.batch/migrator created
job.batch/migrator condition met
```

The log output of the `migrator` should include `INFO`-level logs and successfully terminate with `migrator exited with code 0`. If you see an error message or any of the databases have been flagged as "dirty", please follow ["How to troubleshoot a dirty database"](../../../admin/how-to/dirty_database.md). A dirty database will not affect your ability to use Sourcegraph however it will need to be resolved to upgrade further. If you are unable to resolve the issues, contact support at <mailto:support@sourcegraph.com> for further assistance. Otherwise, you are now safe to upgrade Sourcegraph.

### Docker / Docker compose

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
