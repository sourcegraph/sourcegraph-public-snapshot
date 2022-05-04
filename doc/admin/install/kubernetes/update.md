# Updating Sourcegraph with Kubernetes

> WARNING: This guide applies exclusively to a Kubernetes deployment **without** Helm. If using Helm, please go to [Updating Sourcegraph in the Helm guide](helm.md#upgrading-sourcegraph). 
> If you have not deployed Sourcegraph yet, it is higly recommended to use Helm as it simplifies the configuration and greatly simplifies the upgrade process. See our [Helm guide](helm.md) for more information.

A new version of Sourcegraph is released every month (with patch releases in between, released as needed). Check the [Sourcegraph blog](https://about.sourcegraph.com/blog) for release announcements.

> WARNING: Please check the [Kubernetes update notes](../../updates/kubernetes.md) before upgrading to any particular version of Sourcegraph to check if any manual migrations are necessary.

## Steps

**These steps assume that you have created a `release` branch following the [instructions in the configuration guide](configure.md)**.

1. Merge the new version of Sourcegraph into your release branch.

   ```bash
   cd $DEPLOY_SOURCEGRAPH_FORK
   # get updates
   git fetch upstream
   # to merge the upstream release tag into your release branch.
   git checkout release
   # Choose which version you want to deploy from https://github.com/sourcegraph/deploy-sourcegraph/releases
   git merge $NEW_VERSION
   ```

2. Deploy the updated version of Sourcegraph to your Kubernetes cluster:

   ```bash
   ./kubectl-apply-all.sh
   ```

3. Monitor the status of the deployment.

   ```bash
   kubectl get pods -o wide --watch
   ```

## Rollback

You can rollback by resetting your `release` branch to the old state and proceeding with step 2 above.

_If an update includes a database migration, rollback will require some manual DB
modifications. We plan to eliminate these in the near future, but for now,
email <mailto:support@sourcegraph.com> if you have concerns before updating to a new release._

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

## Database Migrations

> NOTE: The `migrator` service is only available in versions `3.37` and later.

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
