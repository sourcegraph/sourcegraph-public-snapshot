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

1. Start the migrations (run these commands from the root of your `deploy-sourcegraph` fork):

    > NOTE: This script makes the assumption that the environment has all three databases enabled. If the configuration flag `DISABLE_CODE_INSIGHTS` is set and the `codeinsights-db` is unavailable, the `migrator` container will fail. Please see the [Migrating Without Code Insights](#migrating-without-code-insights) section below for more info.
    
    ```bash
    # Update the "image" value of the migrator container in the manifest
    export SOURCEGRAPH_VERSION="the version you're upgrading to"
    yq eval -i \
        '.spec.template.spec.containers[0].image = "index.docker.io/sourcegraph/migrator:" + strenv(SOURCEGRAPH_VERSION)' \
        configure/migrator/migrator.Job.yaml

    # If you do not have yq, you can update the image tag manually to:
    #   "index.docker.io/sourcegraph/migrator:$SOURCEGRAPH_VERSION"

    # Apply and wait for migrations to complete before continuing
    kubectl delete -f configure/migrator/migrator.Job.yaml --ignore-not-found=true
    kubectl apply -f configure/migrator/migrator.Job.yaml
    # -1s timeout will wait "forever"
    kubectl wait -f configure/migrator/migrator.job.yaml --for=condition=complete --timeout=-1s
    ```

    You should see something like the following printed to the terminal:

    ```text
    job.batch "migrator" deleted
    job.batch/migrator created
    job.batch/migrator condition met
    ```

    The log output of the `migrator` container should look similar to:

    ```
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Checked current version" schema=frontend appliedVersions="[1528395834 1528395835 ... 15283959051 1528395951]" pendingVersions=[] failedVersions=[]
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Checked current version" schema=codeintel appliedVersions="[1000000015 1000000016 ... 1000000024 1000000025]" pendingVersions=[1000000031] failedVersions=[]
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Checked current version" schema=codeinsights appliedVersions="[1000000010 1000000011 ... 1000000015 1000000016]" pendingVersions=[1000000031] failedVersions=[]

    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Desugaring `upgrade` to `targeted up` operation" schema=frontend leafVersions=[1528395972]
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Checked current version" schema=frontend appliedVersions="[1528395834 1528395835 ... 15283959051 1528395951]" pendingVersions=[] failedVersions=[]
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migrations" schema=frontend up=true count=21
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migration" schema=frontend migrationID=1528395952 up=true
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migration" schema=frontend migrationID=1528395953 up=true
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Checked progress of index creation" tableName=repo indexName=repo_hashed_name_idx exists=false isValid=false in-progress=false
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Creating index concurrently" schema=frontend migrationID=1528395954 tableName=repo indexName=repo_hashed_name_idx
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Checked current version" schema=frontend appliedVersions="[1528395834 1528395835 ... 15283959051 1528395951]" pendingVersions=[] failedVersions=[]
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migrations" schema=frontend up=true count=18
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migration" schema=frontend migrationID=1528395955 up=true
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migration" schema=frontend migrationID=1528395956 up=true
    ...
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migration" schema=frontend migrationID=1528395971 up=true
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migration" schema=frontend migrationID=1528395972 up=true
    
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Desugaring `upgrade` to `targeted up` operation" schema=codeintel leafVersions=[1000000030]
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Checked current version" schema=codeintel appliedVersions="[1000000015 1000000016 ... 1000000024 1000000025]" pendingVersions=[1000000031] failedVersions=[]
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migrations" schema=codeintel up=true count=5
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migration" schema=codeintel migrationID=1000000026 up=true
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migration" schema=codeintel migrationID=1000000027 up=true
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migration" schema=codeintel migrationID=1000000028 up=true
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migration" schema=codeintel migrationID=1000000029 up=true
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migration" schema=codeintel migrationID=1000000030 up=true

    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Desugaring `upgrade` to `targeted up` operation" schema=codeinsights leafVersions=[1000000018]
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Checked current version" schema=codeinsights appliedVersions="[1000000010 1000000011 ... 1000000015 1000000016]" pendingVersions=[1000000031] failedVersions=[]
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migrations" schema=codeinsights up=true count=2
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migration" schema=codeinsights migrationID=1000000017 up=true
    migrator | t=2022-02-08T03:14:35+0000 lvl=info msg="Applying migration" schema=codeinsights migrationID=1000000018 up=true

    migrator exited with code 0
    ```

If you see an error message or any of the databases have been flagged as "dirty", please follow ["How to troubleshoot a dirty database"](../../../admin/how-to/dirty_database.md). A dirty database will not affect your ability to use Sourcegraph however it will need to be resolved to upgrade further. If you are unable to resolve the issues, contact support at <mailto:support@sourcegraph.com> for further assistance and provide the output of the three `psql` commands. Otherwise, you are now safe to upgrade Sourcegraph.

### Migrating Without Code Insights

If the `DISABLE_CODE_INSIGHTS=true` feature flag is set in Sourcegraph and the `codeinsights-db` is unavailable to the `migrator` container, the migration process will fail. To work around this, the `configure/migrator/migrator.Job.yaml` file will need to be updated. Please make the following changes to your fork of `deploy-sourcegraph`'s `migrator.Job.yaml` file.

1. Modify the `migrator` manifest to update the `spec.template.spec.containers[0].args` field to `["up", "-db", "frontend,codeintel"]`

You should now be able to apply the job and continue the migration and upgrade process as normal.

### Troubleshooting

See the [troubleshooting page](troubleshoot.md).
