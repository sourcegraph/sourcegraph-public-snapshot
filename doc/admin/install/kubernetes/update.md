# Updating Sourcegraph with Kubernetes

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

> NOTE: This feature is only available in versions `3.36` and later

By default, database migrations will be performed during application startup by the `frontend` application. These migrations _must_ succeed before Sourcegraph will become available. If the databases are large, these migrations may take a long time.

In some situations, administrators may wish to migrate their databases before upgrading the rest of the system to reduce downtime. Sourcegraph guarantees database backward compatibility to the most recent minor point release so the database can safely be upgraded before the application code.

To execute the database migrations independently, run the following commands in your fork of `deploy-sourcegraph`.

> NOTE: These values will work for a standard deployment of Sourcegraph with all three databases running in-cluster. If you've customized your deployment (e.g., using an external database service), you will have to modify the environment variables in `configure/migrator/migrator.Job.yaml` accordingly.


1. Check the current migration versions of all three databases:
    ```
    # This will output the current migration version for the frontend db
    kubectl exec $(kubectl get pod -l app=pgsql -o jsonpath="{.items[0].metadata.name}") -c pgsql -- psql -U sg -c "SELECT * FROM schema_migrations;"

    version   | dirty
    ------------+-------
    1528395964 | f
    (1 row)

    # This will output the current migration version for the codeintel db
    kubectl exec $(kubectl get pod -l app=codeintel-db -o jsonpath="{.items[0].metadata.name}") -c pgsql -- psql -U sg -c "SELECT * FROM codeintel_schema_migrations;"

    version   | dirty
    ------------+-------
    1000000030 | f
    (1 row)

    # This will output the current migration version for the codeinsights db
    kubectl exec $(kubectl get pod -l app=codeinsights-db -o jsonpath="{.items[0].metadata.name}") -- psql -U postgres -c "SELECT * FROM codeinsights_schema_migrations;"

    version   | dirty
    ------------+-------
    1000000024 | f
    (1 row)
    ```

1. Start the migrations:
    ```
    # Update the "image" value of the migrator container in the manifest
    export SOURCEGRAPH_VERSION="the version you're upgrading to"
    yq eval -i '.spec.template.spec.containers[0].image = strenv(SOURCEGRAPH_VERSION)' configure/migrator/migrator.Job.yaml

    # Apply and wait for migrations to complete before continuing
    kubectl delete -f configure/migrator/migrator.Job.yaml --ignore-not-found=true
    kubectl apply -f configure/migrator/migrator.Job.yaml
    # -1s timeout will wait "forever"
    kubectl wait -f configure/migrator/migrator.job.yaml --for=condition=complete --timeout=-1s
    ```

    You should see something like the following printed to the terminal:

    > job.batch "migrator" deleted
    > job.batch/migrator created
    > job.batch/migrator condition met

    The log output of the `migrator` container should look similar to:
    > t=2022-01-14T23:47:47+0000 lvl=info msg="Checked current version" schema=frontend version=1528395964 dirty=false
    > t=2022-01-14T23:47:47+0000 lvl=info msg="Checked current version" schema=codeintel version=1000000030 dirty=false
    > t=2022-01-14T23:47:47+0000 lvl=info msg="Checked current version" schema=codeinsights version=1000000024 dirty=false
    > t=2022-01-14T23:47:47+0000 lvl=info msg="Checked current version" schema=codeinsights version=1000000024 dirty=false
    > t=2022-01-14T23:47:47+0000 lvl=info msg="Upgrading schema" schema=codeinsights

1. Repeat the three `psql` commands from the first step to verify the migration versions and that none of the databases are flagged as `dirty`. The versions reported should match the last output version from the migrator container.

If you see an error message or any of the databases have been flagged as dirty, contact support at support@sourcegraph.com for further assistance and provide the output of the three `psql` commands. Your database may not have been migrated correctly. Otherwise, you are now safe to upgrade Sourcegraph.

### Troubleshooting

See the [troubleshooting page](troubleshoot.md).
