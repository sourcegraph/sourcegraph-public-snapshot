# Troubleshooting

If Sourcegraph does not start up or shows unexpected behavior, there are a variety of ways you can determine the root
cause of the failure. 

## The most helpful commands:

List all pods in your cluster and the corresponding health status of each pod:

```bash
kubectl get pods -o=wide
```

Tail the logs for the specified pod:

```bash
kubectl logs -f $POD_NAME
```

If Sourcegraph is unavailable and the `sourcegraph-frontend-*` pod(s) are not in status `Running`, then view their logs with `kubectl logs -f sourcegraph-frontend-$POD_ID` (filling in `$POD_ID` from the `kubectl get pods` output). Inspect both the log messages printed at startup (at the beginning of the log output) and recent log messages.

## Less frequently used commands:

Display detailed information about the status of a single pod:

```bash
kubectl describe $POD_NAME
```

List all Persistent Volume Claims (PVCs) and their statuses:

```bash
kubectl get pvc
```

List all Persistent Volumes (PVs) that have been provisioned. In a healthy cluster, there should
  be a one-to-one mapping between PVs and PVCs:

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

## Common errors

> `Error from server (Forbidden): error when creating "base/frontend/sourcegraph-frontend.Role.yaml": roles.rbac.authorization.k8s.io "sourcegraph-frontend" is forbidden: attempt to grant extra privileges`

- The account you are using to apply the Kubernetes configuration doesn't have sufficient permissions to create roles.
- GCP: `kubectl create clusterrolebinding cluster-admin-binding --clusterrole cluster-admin --user $YOUR_EMAIL`

> `kubectl get pv` shows no Persistent Volumes, and/or `kubectl get events` shows a `Failed to provision volume with StorageClass "default"` error.

Check that a storage class named "default" exists via:

```bash
kubectl get storageclass
```

If one does exist, run `kubectl get storageclass default -o=yaml` and verify that the zone indicated in the output matches the zone of your cluster.

- Google Cloud Platform users may need to [request an increase in storage quota](https://cloud.google.com/compute/quotas).

> `error retrieving RESTMappings to prune: invalid resource networking.k8s.io/v1, Kind=Ingress, Namespaced=true: no matches for kind "Ingress" in version "networking.k8s.io/v1"`
- Make sure the client version of your kubectl matches the one used by the server. Run `kubectl version` to check.
- See the ["Configure network access"](configure.md#security-configure-network-access)
- Check for duplicate `sourcegraph-frontend` using `kubectl get ingresses -A`
  - Delete duplicate using `kubectl delete ingress sourcegraph-frontend -n default`

> `error when creating "base/cadvisor/cadvisor.ClusterRoleBinding.yaml": subjects[0].namespace: Required value`
Add `namespace: default` to the [base/cadvisor/cadvisor.ClusterRoleBinding.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/cadvisor/cadvisor.ClusterRoleBinding.yaml) file under `subjects`.

> `error when creating "base/prometheus/prometheus.ClusterRoleBinding.yaml": subjects[0].namespace: Required value`
Add `namespace: default` to the [base/prometheus/prometheus.ClusterRoleBinding.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/prometheus/prometheus.ClusterRoleBinding.yaml) file under `subjects`.

> Many pods are stuck in Pending status. 

One thing to check for is insufficient resources. To obtain a dump of the logs: `kubectl cluster-info dump > dump.txt`

```error
  "Reason": "FailedScheduling",
  "Message": "0/3 nodes are available: 1 Insufficient memory, 3 Insufficient cpu.",
```

The message above shows that your cluster is under provisioned (i.e. has too few nodes, or not enough CPU and memory).
If you're using Google Cloud Platform, note that the default node type is `n1-standard-1`, a machine
with only one CPU, and that some components request a 2-CPU node. When creating a cluster, use
`--machine-type=n1-standard-16`.

> You can't access Sourcegraph.

See the [Troubleshooting ingress-nginx docs](https://kubernetes.github.io/ingress-nginx/troubleshooting/). 
If you followed our instructions, the namespace of the ingress-controller is `ingress-nginx`.



Any other issues? Contact us at [@srcgraph](https://twitter.com/srcgraph)
or <mailto:support@sourcegraph.com>, or file issues on
our [public issue tracker](https://github.com/sourcegraph/issues/issues).
