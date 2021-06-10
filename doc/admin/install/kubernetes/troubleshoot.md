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


### Error: Error from server (Forbidden): error when creating "base/frontend/sourcegraph-frontend.Role.yaml": roles.rbac.authorization.k8s.io "sourcegraph-frontend" is forbidden: attempt to grant extra privileges.

- The account you are using to apply the Kubernetes configuration doesn't have sufficient permissions to create roles.
- GCP: `kubectl create clusterrolebinding cluster-admin-binding --clusterrole cluster-admin --user $YOUR_EMAIL`


### "kubectl get pv" shows no Persistent Volumes, and/or "kubectl get events" shows a `Failed to provision volume with StorageClass "sourcegraph"` error.

Check that a storage class named "sourcegraph" exists via:

```bash
kubectl get storageclass
```

If one does exist, run `kubectl get storageclass sourcegraph -o=yaml` and verify that the zone indicated in the output matches the zone of your cluster.

- Google Cloud Platform users may need to [request an increase in storage quota](https://cloud.google.com/compute/quotas).


### Error: error retrieving RESTMappings to prune: invalid resource networking.k8s.io/v1, Kind=Ingress, Namespaced=true: no matches for kind "Ingress" in version "networking.k8s.io/v1".

- Make sure the client version of your kubectl matches the one used by the server. Run `kubectl version` to check.
- See the ["Configure network access"](configure.md#security-configure-network-access)
- Check for duplicate `sourcegraph-frontend` using `kubectl get ingresses -A`
  - Delete duplicate using `kubectl delete ingress sourcegraph-frontend -n default`


### Error: error when creating "base/cadvisor/cadvisor.ClusterRoleBinding.yaml": subjects[0].namespace: Required value

Add `namespace: default` to the [base/cadvisor/cadvisor.ClusterRoleBinding.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/cadvisor/cadvisor.ClusterRoleBinding.yaml) file under `subjects`.


### Many pods are stuck in Pending status.

One thing to check for is insufficient resources. To obtain a dump of the logs: `kubectl cluster-info dump > dump.txt`

```error
  "Reason": "FailedScheduling",
  "Message": "0/3 nodes are available: 1 Insufficient memory, 3 Insufficient cpu.",
```

The message above shows that your cluster is under provisioned (i.e. has too few nodes, or not enough CPU and memory).
If you're using Google Cloud Platform, note that the default node type is `n1-standard-1`, a machine
with only one CPU, and that some components request a 2-CPU node. When creating a cluster, use
`--machine-type=n1-standard-16`.


### ImagePullBackOff / 429 Too Many Requests Errors.

This indicates the instance is getting rate-limited by Docker Hub([link](https://www.docker.com/increase-rate-limits)), where our images are stored, as unauthenticated users are limited to 100 image pulls within a 6 hour period. Possible solutions included:
- Create a Docker Hub account with a higher rate limit > configure an `ImagePullSecrets` K8S object with your Docker Hub service that contains your docker credentials ([link to tutorial](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/)) > add these credentials to the default service account that's running the same namespace that Sourcegraph is running in ([link to tutorial](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#add-imagepullsecrets-to-a-service-account))
- Wait until the rate limits are reset
- [**OPTIONAL**] Upgrade your account to a Docker Pro or Team subscription ([See Docker Hub for more information](https://www.docker.com/increase-rate-limits))


### Prometheus Pod is constantly down when using the namespace overlays.

This is most likely due to cadvisor picking up other metrics from the cluster.
You can confirm this theory by checking your [prometheus.ConfigMap.yaml](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph@3.27/-/blob/base/prometheus/prometheus.ConfigMap.yaml#L248-250) file, where the `source_labels: [container_label_io_kubernetes_pod_namespace]` fields under `metric_relabel_configs` should be commented out and the `regex` field must be updated with your namespace.


### I don't see any metrics on my Grafana Dashboard.

This means Sourcegraph is having issues connecting to the Kubernetes API. For instance, using the non-privileged overlay is most likely going to prevent Sourcegraph from picking up metrics from the Kubernetes API. One of the potential solutions is to give Prometheus and cAdvisor root access by adding the ClusterRoleBinding.yaml files for both services from the base layer to the non-privileged overlay.


### Which metrics are using the most resources?

You can port-forward the instance with `kubectl port-forward pod prometheus-$$ 9090:9090`, then go to [http://localhost:9090/](http://localhost:9090/) and run `topk(10, count by (__name__)({__name__=~".+"}))` to check the values.


### You can't access Sourcegraph.

See the [Troubleshooting ingress-nginx docs](https://kubernetes.github.io/ingress-nginx/troubleshooting/).

If you followed our instructions, the namespace of the ingress-controller is `ingress-nginx`.

### Healthcheck failing with Strconv.Atoi: parsing "{$portName}": invalid syntax error

This can occur when a port does not have a name but the that name is used within the Readiness or Liveness probe.
Ensure that the port name is consistent with upstream.

```
ports:
  - containerPort: 3188
    name: minio
...
livenessProbe:
  httpGet:
    path: /minio/health/live
    port: minio   #this port name MUST exist in the same spec
```

Any other issues? Contact us at [@sourcegraph](https://twitter.com/sourcegraph)
or <mailto:support@sourcegraph.com>, or file issues on
our [public issue tracker](https://github.com/sourcegraph/issues/issues).
