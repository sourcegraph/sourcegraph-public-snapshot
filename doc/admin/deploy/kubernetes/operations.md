# Operations guides for Sourcegraph with Kubernetes

Operations guides specific to managing [Sourcegraph with Kubernetes](./index.md) installations.

Trying to deploy Sourcegraph with Kubernetes? Refer to our [installation guide](./index.md#installation).

## Featured guides

<div class="getting-started">
  <a href="configure" class="btn btn-primary" alt="Configure">
   <span>Configure</span>
   </br>
   Configure your Sourcegraph deployment with our deployment reference.
  </a>

  <a href="#upgrade" class="btn" alt="Upgrade">
   <span>Upgrade</span>
   </br>
   Upgrade your deployment to the latest Sourcegraph release.
  </a>

  <a href="troubleshoot" class="btn" alt="Backup and restore">
   <span>Troubleshoot</span>
   </br>
   Troubleshoot common issues with your Sourcegraph instance.
  </a>
</div>

## Deploy

Refer to our [installation guide](./index.md#installation) for details on how to deploy Sourcegraph.

Migrating from another [deployment type](../index.md)? Refer to our [migration guides](../migrate-backup.md).

### Applying manifests

In general, Sourcegraph with Kubernetes is deployed by applying the [Kubernetes](./index.md#kubernetes) manifests in our [deploy-sourcegraph reference repository](./index.md#reference-repository) - see our [configuration guide](./configure.md) for more details.

We provide a `kubectl-apply-all.sh` script that you can use to do this, usually by running the following from the root directory of the [deploy-sourcegraph reference repository](./index.md#reference-repository):

```sh
./kubectl-apply-all.sh
```

> NOTE: By default, this script applies our base manifests using [`kubectl apply`](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply) with a variety of arguments specific to the [reference repository](./index.md#reference-repository)'s layout.
> If you have specific commands that should be run whenever you apply your manifests, you should modify this script as needed. For example, if you use [overlays to make changes to the manifests](./configure.md#overlays), you should modify this script to apply your generated cluster instead.

Once you have applied your changes:

- *Watch* - verify your deployment has started:

  ```bash
  kubectl get pods -A -o wide --watch
  ```

- *Port-foward* - verify Sourcegraph is running by temporarily making the frontend port accessible:

  ```sh
  kubectl port-forward svc/sourcegraph-frontend 3080:30080
  ```

- *Log in* - browse to your Sourcegraph deployment, login, and verify the instance is working as expected.

## Configure

We strongly recommend referring to our [Configuration guide](configure.md) to learn about how to configure your Sourcegraph with Kubernetes instance.

## Upgrade

- See the [Updating Sourcegraph docs](update.md) on how to upgrade.<br/>
- See the [Updating a Kubernetes Sourcegraph instance docs](../../updates/kubernetes.md) for details on changes in each version to determine if manual migration steps are necessary.

## List pods in cluster

List all pods in your cluster and the corresponding health status of each pod:

```bash
kubectl get pods -o=wide
```

## Tail logs for specific pod

Tail the logs for the specified pod:

```bash
kubectl logs -f $POD_NAME
```

If Sourcegraph is unavailable and the `sourcegraph-frontend-*` pod(s) are not in status `Running`, then view their logs with `kubectl logs -f sourcegraph-frontend-$POD_ID` (filling in `$POD_ID` from the `kubectl get pods` output). Inspect both the log messages printed at startup (at the beginning of the log output) and recent log messages.

## Retrieving resource information

Display detailed information about the status of a single pod:

```bash
kubectl describe $POD_NAME
```

List all Persistent Volume Claims (PVCs) and their statuses:

```bash
kubectl get pvc
```

List all Persistent Volumes (PVs) that have been provisioned.
In a healthy cluster, there should be a one-to-one mapping between PVs and PVCs:

```bash
kubectl get pv
```

List all events in the cluster's history:

```bash
kubectl get events
```

Delete failing pod so it gets recreated, possibly on a different node:

```bash
kubectl delete pod $POD_NAME
```

Remove all pods from a node and mark it as unschedulable to prevent new pods from arriving

```bash
kubectl drain --force --ignore-daemonsets --delete-local-data $NODE
```

Restarting Sourcegraph Instance:

```bash
kubectl rollout restart deployment sourcegraph-frontend
```

## Access the database

Get the id of one `pgsql` Pod:

```bash
kubectl get pods -l app=pgsql
NAME                     READY     STATUS    RESTARTS   AGE
pgsql-76a4bfcd64-rt4cn   2/2       Running   0          19m
```

Make sure you are operating under the correct namespace (i.e. add `-n prod` if your pod is under the `prod` namespace).

Open a PostgreSQL interactive terminal:

```bash
kubectl exec -it pgsql-76a4bfcd64-rt4cn -- psql -U sg
```

Run your SQL query:

```sql
SELECT * FROM users;
```

> NOTE: To execute an SQL query against the database without first creating an interactive session (as below), append `--command "SELECT * FROM users;"` to the docker container exec command.

## Backup and restore

The following instructions are specific to backing up and restoring the sourcegraph databases in a Kubernetes deployment. These do not apply to other deployment types.

> WARNING: **Only core data will be backed up**.
>
> These instructions will only back up core data including user accounts, configuration, repository-metadata, etc. Other data will be regenerated automatically:
>
> - Repositories will be re-cloned
> - Search indexes will be rebuilt from scratch
>
> The above may take a while if you have a lot of repositories. In the meantime, searches may be slow or return incomplete results. This process rarely takes longer than 6 hours and is usually **much** faster.

> NOTE: In some places you will see `$NAMESPACE` used. Add `-n $NAMESPACE` to commands if you are not using the default namespace
> More kubectl configuration options can be found here: [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)

### Back up Sourcegraph databases

These instructions will back up the primary `sourcegraph` database and the [codeintel](../../../code_navigation/index.md) database.

A. Verify deployment running

```bash
kubectl get pods -A
```

B. Stop all connections to the database by removing the frontend deployment

```bash
kubectl scale --replicas=0 deployment/sourcegraph-frontend
# or
kubectl delete deployment sourcegraph-frontend
```

C. Check for corrupt database indexes.  If amcheck returns errors, please reach out to [support@sourcegraph.com](mailto:support@sourcegraph.com)

```sql
create extension amcheck;

select bt_index_parent_check(c.oid, true), c.relname, c.relpages
from pg_index i
join pg_opclass op ON i.indclass[0] = op.oid
join pg_am am ON op.opcmethod = am.oid
join pg_class c ON i.indexrelid = c.oid
join pg_namespace n ON c.relnamespace = n.oid
where am.amname = 'btree'
-- Don't check temp tables, which may be from another session:
and c.relpersistence != 't'
-- Function may throw an error when this is omitted:
and i.indisready AND i.indisvalid;
```

D. Generate the database dumps

```bash
kubectl exec -it $pgsql_POD_NAME -- bash -c 'pg_dump -C --username sg sg' > sourcegraph_db.out
kubectl exec -it $codeintel-db_POD_NAME -- bash -c 'pg_dump -C --username sg sg' > codeintel_db.out
```

Ensure the `sourcegraph_db.out` and `codeintel_db.out` files are moved to a safe and secure location.

### Restore Sourcegraph databases

#### Restoring Sourcegraph databases into a new environment

The following instructions apply only if you are restoring your databases into a new deployment of Sourcegraph ie: a new virtual machine

If you are restoring a previously running environment, see the instructions for [restoring a previously running deployment](#restoring-sourcegraph-databases-into-an-existing-environment)

A. Copy the database dump files (eg. `sourcegraph_db.out` and `codeintel_db.out`) into the root of the `deploy-sourcegraph` directory

B. Start the database services by running the following command from the root of the [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) directory

```bash
kubectl rollout restart deployment pgsql
kubectl rollout restart deployment codeintel-db
```

C. Copy the database files into the pods by running the following command from the root of the [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) directory

```bash
kubectl cp sourcegraph_db.out $NAMESPACE/$pgsql_POD_NAME:/tmp/sourcegraph_db.out
kubectl cp codeintel_db.out $NAMESPACE/$codeintel-db_POD_NAME:/tmp/codeintel_db.out
```

D. Restore the databases

```bash
kubectl exec -it $pgsql_POD_NAME -- bash -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/sourcegraph_db.out sg'
kubectl exec -it $codeintel-db_POD_NAME -- bash -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/condeintel_db.out sg'
```

E. Check for corrupt database indexes.  If amcheck returns errors, please reach out to [support@sourcegraph.com](mailto:support@sourcegraph.com)

```sql
create extension amcheck;

select bt_index_parent_check(c.oid, true), c.relname, c.relpages
from pg_index i
join pg_opclass op ON i.indclass[0] = op.oid
join pg_am am ON op.opcmethod = am.oid
join pg_class c ON i.indexrelid = c.oid
join pg_namespace n ON c.relnamespace = n.oid
where am.amname = 'btree'
-- Don't check temp tables, which may be from another session:
and c.relpersistence != 't'
-- Function may throw an error when this is omitted:
and i.indisready AND i.indisvalid;
```

F. Start the remaining Sourcegraph services by following the steps in [applying manifests](#applying-manifests).

#### Restoring Sourcegraph databases into an existing environment

A. Stop the existing deployment by removing the frontend deployment

```bash
kubectl scale --replicas=0 deployment/sourcegraph-frontend
# or
kubectl delete deployment sourcegraph-frontend
```

B. Remove any existing volumes for the databases in the existing deployment

```bash
kubectl delete pvc pgsql
kubectl delete pvc codeintel-db
kubectl delete pv $pgsql_PV_NAME --force
kubectl delete pv $codeintel-db_PV_NAME --force
```

C. Copy the database dump files (eg. `sourcegraph_db.out` and `codeintel_db.out`) into the root of the `deploy-sourcegraph` directory

D. Start the database services only

```bash
kubectl rollout restart deployment pgsql
kubectl rollout restart deployment codeintel-db
```

E. Copy the database files into the pods by running the following command from the root of the [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) directory

```bash
kubectl cp sourcegraph_db.out $NAMESPACE/$pgsql_POD_NAME:/tmp/sourcegraph_db.out
kubectl cp codeintel_db.out $NAMESPACE/$codeintel-db_POD_NAME:/tmp/codeintel_db.out
```

F. Restore the databases

```bash
kubectl exec -it $pgsql_POD_NAME -- bash -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/sourcegraph_db.out sg'
kubectl exec -it $codeintel-db_POD_NAME -- bash -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/condeintel_db.out sg'
```

G. Check for corrupt database indexes.  If amcheck returns errors, please reach out to [support@sourcegraph.com](mailto:support@sourcegraph.com)

```sql
create extension amcheck;

select bt_index_parent_check(c.oid, true), c.relname, c.relpages
from pg_index i
join pg_opclass op ON i.indclass[0] = op.oid
join pg_am am ON op.opcmethod = am.oid
join pg_class c ON i.indexrelid = c.oid
join pg_namespace n ON c.relnamespace = n.oid
where am.amname = 'btree'
-- Don't check temp tables, which may be from another session:
and c.relpersistence != 't'
-- Function may throw an error when this is omitted:
and i.indisready AND i.indisvalid;
```

H. Start the remaining Sourcegraph services by following the steps in [applying manifests](#applying-manifests).

## Troubleshoot

See the [Troubleshooting docs](troubleshoot.md).
