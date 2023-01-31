# Updating Sourcegraph with Kubernetes

> WARNING: This guide applies exclusively to a Kubernetes deployment **without** Helm. If using Helm, please go to [Updating Sourcegraph in the Helm guide](helm.md#upgrading-sourcegraph).
> If you have not deployed Sourcegraph yet, it is higly recommended to use Helm as it simplifies the configuration and greatly simplifies the upgrade process. See our [Helm guide](helm.md) for more information.

A new version of Sourcegraph is released every month (with patch releases in between, released as needed). Check the [Sourcegraph blog](https://about.sourcegraph.com/blog) for release announcements.

## Upgrades

### Standard upgrades

A [standard upgrade](../../updates/index.md#standard-upgrades) occurs between two minor versions of Sourcegraph. If you are looking to jump forward several versions, you must perform a [multi-version upgrade](#multi-version-upgrades) instead.

**Before upgrading:**

- Read our [update policy](../../updates/index.md#update-policy) to learn about Sourcegraph updates.
- Find the relevant entry for your update in the [update notes for Sourcegraph with Kubernetes](../../updates/kubernetes.md).

**The following steps assume that you have created a `release` branch following the [instructions in the configuration guide](configure.md)**.

First, merge the new version of Sourcegraph into your release branch.

```bash
cd $DEPLOY_SOURCEGRAPH_FORK
# get updates
git fetch upstream
# to merge the upstream release tag into your release branch.
git checkout release
# Choose which version you want to deploy from https://github.com/sourcegraph/deploy-sourcegraph/releases
git merge $NEW_VERSION
```

Then, deploy the updated version of Sourcegraph to your Kubernetes cluster:

```bash
./kubectl-apply-all.sh
```

> NOTE: By default, this script applies our base manifests using [`kubectl apply`](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply) with a variety of arguments specific to the [reference repository](./index.md#reference-repository)'s layout.
> If you have specific commands that should be run whenever you apply your manifests, you should modify this script as needed. For example, if you use [overlays to make changes to the manifests](./configure.md#overlays), you should modify this script to apply your generated cluster instead.

Monitor the status of the deployment to determine its success.

```bash
kubectl get pods -o wide --watch
```

### Multi-version upgrades

A [multi-version upgrade](../../updates/index.md#multi-version-upgrades) is a downtime-incurring upgrade from version 3.20 or later to any future version. Multi-version upgrades will run both schema and data migrations to ensure the data available from the instance remains available post-upgrade.

> NOTE: It is highly recommended to **take an up-to-date snapshot of your databases** prior to starting a multi-version upgrade. The upgrade process aggressively mutates the shape and contents of your database, and undiscovered errors in the migration process or unexpected environmental differences may cause an unusable instance or data loss.
>
> We recommend performing the entire upgrade procedure on an idle clone of the production instance and switch traffic over on success, if possible. This may be low-effort for installations with a canary environment or a blue/green deployment strategy.
>
> **If you do not feel confident running this process solo**, contact customer support team to help guide you thorough the process.

**Before performing a multi-version upgrade**:

- Read our [update policy](../../updates/index.md#update-policy) to learn about Sourcegraph updates.
- Find the entries that apply to the version range you're passing through in the [update notes for Sourcegraph with Kubernetes](../../updates/kubernetes.md#multi-version-upgrade-procedure).

To perform a multi-version upgrade on a Sourcegraph instance running on Kubernetes:

1. Spin down any pods that access the database. This must be done for the following deployments and stateful sets listed below. This can be performed directly via a series of `kubectl` commands (given below), or by setting `replicas: 0` in each deployment/stateful set's definitions and re-applying the configuration.
  - Deployments (e.g., `kubectl scale deployment <name> --replicas=0`)
      - precise-code-intel-worker
      - repo-updater
      - searcher
      - sourcegraph-frontend
      - sourcegraph-frontend-internal
      - symbols
      - worker
  - Stateful sets (e.g., `kubectl scale sts <name> --replicas=0`):
      - gitserver
      - indexed-search
1. **If upgrading from 3.26 or before to 3.27 or later**, the `pgsql` and `codeintel-db` databases must be upgraded from Postgres 11 to Postgres 12. If this step is not performed, then the following upgrade procedure will fail fast (and leave all existing data untouched).
  - If using an external database, follow the [upgrading external PostgreSQL instances](../../postgres.md#upgrading-external-postgresql-instances) guide.
  - Otherwise, perform the following steps from the [upgrading internal Postgres instances](../../postgres.md#upgrading-internal-postgresql-instances) guide:
      1. It's assumed that your fork of `deploy-sourcegraph` is up to date with your instance's current version. Pull the upstream changes for `v3.27.0` and resolve any git merge conflicts. We need to temporarily boot the containers defined at this specific version to rewrite existing data to the new Postgres 12 format.
      1. Run `kubectl apply -l deploy=sourcegraph -f base/pgsql` to launch a new Postgres 12 container and rewrite the old Postgres 11 data. This may take a while, but streaming container logs should show progress. **NOTE**: The Postgres migration requires enough capacity in its attached volume to accommodate an additional copy of the data currently on disk. Resize the volume now if necessary—the container will fail to start if there is not enough free disk space.
      1. Wait until the database container is accepting connections. Once ready, run the command `kubectl exec pgsql -- psql -U sg -c 'REINDEX database sg;'` issue a reindex command to Postgres to repair indexes that were silently invalidated by the previous data rewrite step. **If you skip this step**, then some data may become inaccessible under normal operation, the following steps are not guaranteed to work, and **data loss will occur**.
      1. Follow the same steps for the `codeintel-db`:
          - Run `kubectl apply -l deploy=sourcegraph -f base/codeintel-db` to launch Postgres 12.
          - Run `kubectl exec codeintel-db -- psql -U sg -c 'REINDEX database sg;'` to issue a reindex command to Postgres.
      1. Leave these versions of the databases running while the subsequent migration steps are performed. If `codeinsights-db` is a container new to your instance, now is a good time to start it as well.
1. Pull the upstream changes for the target instance version and resolve any git merge conflicts. The [standard upgrade procedure](#standard-upgrades) describes this step in more detail.
1. Follow the instructions on [how to run the migrator job in Kubernetes](../../how-to/manual_database_migrations.md#kubernetes) to perform the upgrade migration. For specific documentation on the `upgrade` command, see the [command documentation](../../how-to/manual_database_migrations.md#upgrade). The following specific steps are an easy way to run the upgrade command:
  1. Edit the file `configure/migrator/migrator.Job.yaml` and set the value of the `args` key to `["upgrade", "--from=<old version>", "--to=<new version>"]`. It is recommended to also add the `--dry-run` flag on a trial invocation to detect if there are any issues with database connection, schema drift, or mismatched versions that need to be addressed. If your instance has in-use code intelligence data it's recommended to also temporarily increase the CPU and memory resources allocated to this job. A symptom of underprovisioning this job will result in an `OOMKilled`-status container.
  1. Run `kubectl delete -f configure/migrator/migrator.Job.yaml` to ensure no previous job invocations will conflict with our current invocation.
  1. Start the migrator job via `kubectl apply -f configure/migrator/migrator.Job.yaml`.
  1. Run `kubectl wait -f configure/migrator/migrator.Job.yaml --for=condition=complete --timeout=-1s` to wait for the job to complete. Run `kubectl logs job.batch/migrator -f` stream the migrator's stdout logs for progress.
1. The remaining infrastructure can now be updated. The [standard upgrade procedure](#standard-upgrades) describes this step in more detail.
  - Ensure that the replica counts adjusted in the previous steps are turned back up.
  - Run `./kubectl-apply-all.sh` to deploy the new pods to the Kubernetes cluster.
  - Monitor the status of the deployment via `kubectl get pods -o wide --watch`.

## Rollback

You can rollback by resetting your `release` branch to the old state and proceeding re-running the following:

```
./kubectl-apply-all.sh
```

If you are rolling back more than a single version, then you must also [rollback your database](../../how-to/rollback_database.md), as database migrations (which may have run at some point during the upgrade) are guaranteed to be compatible with one previous minor version.

## Improving update reliability and latency with node selectors

Some of the services that comprise Sourcegraph require more resources than others, especially if the
default CPU or memory allocations have been overridden. During an update when many services restart,
you may observe that the more resource-hungry pods (e.g., `gitserver`, `indexed-search`) fail to
restart, because no single node has enough available CPU or memory to accommodate them. This may be
especially true if the cluster is heterogeneous (i.e., not all nodes have the same amount of
CPU/memory).

If this happens, do the following:

- Use `kubectl drain $NODE` to drain a node of existing pods, so it has enough allocation for the larger
  service.
- Run `watch kubectl get pods -o wide` and wait until the node has been drained. Run `kubectl get pods` to check that all pods except for the resource-hungry one(s) have been assigned to a node.
- Run `kubectl uncordon $NODE` to enable the larger pod(s) to be scheduled on the drained node.

Note that the need to run the above steps can be prevented altogether with [node
selectors](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector), which
tell Kubernetes to assign certain pods to specific nodes. See the [docs on enabling node
selectors](scale.md#node-selector) for Sourcegraph on Kubernetes.

## High-availability updates

Sourcegraph is designed to be a high-availability (HA) service, but upgrades by default require a 10m downtime
window. If you need zero-downtime upgrades, please contact us. Services employ health checks to test the health
of newly updated components before switching live traffic over to them by default. HA-enabling features include
the following:

- Replication: nearly all of the critical services within Sourcegraph are replicated. If a single instance of a
  service fails, that instance is restarted and removed from operation until it comes online again.
- Updates are applied in a rolling fashion to each service such that a subset of instances are updated first while
  traffic continues to flow to the old instances. Once the health check determines the set of new instances is
  healthy, traffic is directed to the new set and the old set is terminated. By default, some database operations
  may fail during this time as migrations occur so a scheduled 10m downtime window is required.
- Each service includes a health check that detects whether the service is in a healthy state. This check is specific to
  the service. These are used to check the health of new instances after an update and during regular operation to
  determine if an instance goes down.
- Database migrations are handled automatically on update when they are necessary.

## Database migrations

By default, database migrations will be performed during application startup by a `migrator` init container running prior to the `frontend` deployment. These migrations **must** succeed before Sourcegraph will become available. If the databases are large, these migrations may take a long time.

In some situations, administrators may wish to migrate their databases before upgrading the rest of the system to reduce downtime. Sourcegraph guarantees database backward compatibility to the most recent minor point release so the database can safely be upgraded before the application code.

To execute the database migrations independently, follow the [Kubernetes instructions on how to manually run database migrations](../../how-to/manual_database_migrations.md#kubernetes). Running the `up` (default) command on the `migrator` of the *version you are upgrading to* will apply all migrations required by the next version of Sourcegraph.

### Failing migrations

Migrations may fail due to transient or application errors. When this happens, the database will be marked by the migrator as _dirty_. A dirty database requires manual intervention to ensure the schema is in the expected state before continuing with migrations or application startup.

In order to retrieve the error message printed by the migrator on startup, you'll need to use the `kubectl logs <frontend pod> -c migrator` to specify the init container, not the main application container. Using a bare `kubectl logs` command will result in the following error:

```
Error from server (BadRequest): container "frontend" in pod "sourcegraph-frontend-69f4b68d75-w98lx" is waiting to start: PodInitializing
```

Once a failing migration error message can be found, follow the guide on [how to troubleshoot a dirty database](../../how-to/dirty_database.md).

## Troubleshooting

See the [troubleshooting page](troubleshoot.md).
