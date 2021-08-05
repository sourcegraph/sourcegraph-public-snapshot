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

Refer to our [installation guide](./index.md#installation) for more details on how to deploy Sourcegraph.

Migrating from another [deployment type](../index.md)? Refer to our [migration guides](../migrate-backup.md).

## Configure

Refer to our [Configuration guide](configure.md).

## Upgrade

- See the [Updating Sourcegraph docs](update.md) on how to upgrade.<br/>
- See the [Updating a Kubernetes Sourcegraph instance docs](../../updates/kubernetes.md) for details on changes in each version to determine if manual migration steps are necessary.

## Troubleshoot

See the [Troubleshooting docs](troubleshoot.md).

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
