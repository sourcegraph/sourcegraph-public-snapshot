# Troubleshoot Sourcegraph with Kubernetes

If [Sourcegraph with Kubernetes](./index.md) does not start up or shows unexpected behavior, there are a variety of ways you can determine the root cause of the failure.

Also refer to our [operations guide](./operations.md) for useful commands and operations.

Still need additional help? Contact us at [@sourcegraph](https://twitter.com/sourcegraph)
or <mailto:support@sourcegraph.com>, or file issues on
our [public issue tracker](https://github.com/sourcegraph/issues/issues).

## Common errors

#### Error: Error from server (Forbidden): error when creating "base/frontend/sourcegraph-frontend.Role.yaml": roles.rbac.authorization.k8s.io "sourcegraph-frontend" is forbidden: attempt to grant extra privileges.

- The account you are using to apply the Kubernetes configuration doesn't have sufficient permissions to create roles.
- GCP: `kubectl create clusterrolebinding cluster-admin-binding --clusterrole cluster-admin --user $YOUR_EMAIL`


#### "kubectl get pv" shows no Persistent Volumes, and/or "kubectl get events" shows a `Failed to provision volume with StorageClass "sourcegraph"` error.

Check that a storage class named "sourcegraph" exists via:

```bash
kubectl get storageclass
```

If one does exist, run `kubectl get storageclass sourcegraph -o=yaml` and verify that the zone indicated in the output matches the zone of your cluster.

- Google Cloud Platform users may need to [request an increase in storage quota](https://cloud.google.com/compute/quotas).


#### Error: error retrieving RESTMappings to prune: invalid resource networking.k8s.io/v1, Kind=Ingress, Namespaced=true: no matches for kind "Ingress" in version "networking.k8s.io/v1".

- Make sure the client version of your kubectl matches the one used by the server. Run `kubectl version` to check.
- See the ["Configure network access"](configure.md#security-configure-network-access)
- Check for duplicate `sourcegraph-frontend` using `kubectl get ingresses -A`
  - Delete duplicate using `kubectl delete ingress sourcegraph-frontend -n default`


#### Error: error when creating "base/cadvisor/cadvisor.ClusterRoleBinding.yaml": subjects[0].namespace: Required value

Add `namespace: default` to the [base/cadvisor/cadvisor.ClusterRoleBinding.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/cadvisor/cadvisor.ClusterRoleBinding.yaml) file under `subjects`.


#### Many pods are stuck in Pending status.

One thing to check for is insufficient resources. To obtain a dump of the logs: `kubectl cluster-info dump > dump.txt`

```error
  "Reason": "FailedScheduling",
  "Message": "0/3 nodes are available: 1 Insufficient memory, 3 Insufficient cpu.",
```

The message above shows that your cluster is under provisioned (i.e. has too few nodes, or not enough CPU and memory).
If you're using Google Cloud Platform, note that the default node type is `n1-standard-1`, a machine
with only one CPU, and that some components request a 2-CPU node. When creating a cluster, use
`--machine-type=n1-standard-16`.


#### ImagePullBackOff / 429 Too Many Requests Errors.

This indicates the instance is getting rate-limited by Docker Hub([link](https://www.docker.com/increase-rate-limits)), where our images are stored, as unauthenticated users are limited to 100 image pulls within a 6 hour period. Possible solutions included:
- Create a Docker Hub account with a higher rate limit > configure an `ImagePullSecrets` K8S object with your Docker Hub service that contains your docker credentials ([link to tutorial](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/)) > add these credentials to the default service account that's running the same namespace that Sourcegraph is running in ([link to tutorial](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#add-imagepullsecrets-to-a-service-account))
- Wait until the rate limits are reset
- [**OPTIONAL**] Upgrade your account to a Docker Pro or Team subscription ([See Docker Hub for more information](https://www.docker.com/increase-rate-limits))


#### Irrelevant cAdvisor metrics are causing strange alerts and performance issues.

This is most likely due to cAdvisor picking up other metrics from the cluster.
A workaround is available: [Filtering cAdvisor metrics](./configure.md#filtering-cadvisor-metrics).

#### I don't see any metrics on my Grafana Dashboard.

This means Sourcegraph is having issues connecting to the Kubernetes API. For instance, using the non-privileged overlay is most likely going to prevent Sourcegraph from picking up metrics from the Kubernetes API. One of the potential solutions is to give Prometheus and cAdvisor root access by adding the ClusterRoleBinding.yaml files for both services from the base layer to the non-privileged overlay.


#### Which metrics are using the most resources?

You can port-forward the instance with `kubectl port-forward pod prometheus-$$ 9090:9090`, then go to [http://localhost:9090/](http://localhost:9090/) and run `topk(10, count by (__name__)({__name__=~".+"}))` to check the values.


#### You can't access Sourcegraph.

See the [Troubleshooting ingress-nginx docs](https://kubernetes.github.io/ingress-nginx/troubleshooting/).

If you followed our instructions, the namespace of the ingress-controller is `ingress-nginx`.

#### Healthcheck failing with Strconv.Atoi: parsing "{$portName}": invalid syntax error

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

## Service mesh

Known issues when using a service mesh (e.g. Istio, Linkerd, etc.)

#### Error message: `Git command [git rev-parse HEAD] failed (stderr: ""): strconv.Atoi: parsing "": invalid syntax`

<img class="screenshot w-100" src="https://user-images.githubusercontent.com/68532117/178506378-3d047bc5-d672-487a-920f-8f228ae5cb27.png"/>

This error occurs because Envoy, the proxy used by Istio, [drops proxied trailers for the requests made over HTTP/1.1 protocol by default](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/protocol.proto#config-core-v3-http1protocoloptions). To resolve this issue, enable trailers in your instance following the examples provided for [Kubernetes](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/overlays) and [Kubernetes with Helm](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples).

### Symbols sidebar and hovers are not working

<img class="screenshot w-100" src="https://user-images.githubusercontent.com/68532117/212372086-4c53e817-be3d-46b6-9cc1-fc34e695d30c.png"/>

This issue occurs when the "symbols" component attempts to connect to other services using the built-in service discovery system provided by Sourcegraph, instead of using the service discovery system provided by Kubernetes or a service mesh (such as Istio).

In general, service meshes like Istio use a feature called mutual Transport Layer Security (mTLS) to secure communication between services. mTLS relies on services communicating with each other using DNS names, rather than IP addresses. These DNS names are used to identify the specific services or pods that the communication is intended for.

In this case, when the Envoy sidecar (a component used to manage communication between services) intercepts a request from the "frontend" component to the "symbols" component, it is unable to locate the correct upstream service. This is because Sourcegraph's built-in service discovery makes requests to specific pods via IP address, rather than using more standard Kubernetes service discovery. This allows Sourcegraph to optimize search resolution using cached data for certain services, but violates assumptions made by envoy.

For example, envoy expects, the communication flow would be: frontend -> http://symbol:3184/ -> envoy -> [look up upstream service using DNS name] -> envoy -> symbols

With our built-in service discovery, the communication flow would be: frontend -> http://symbol_pod_ip:3184 -> envoy -> [symbol_pod_ip not found since it's an IP address not a DNS name] -> envoy -> symbols

To resolve this issue, the solution is to redeploy the frontend after specifying the service address for symbols by setting the SYMBOLS_URL environment variable in frontend.

```yaml
SYMBOLS_URL=http:symbols:3184
```

Please make sure the old frontend pods are removed.

### Squirrel.LocalCodeIntel http status 502

<img class="screenshot w-100" src="https://user-images.githubusercontent.com/68532117/212374098-dc2dfe69-4d26-4f5e-a78b-37a53c19ef22.png"/>
The issue described is related to the Code Intel hover feature, where it may get stuck in a loading state or return a 502 error with the message `Squirrel.LocalCodeIntel http status 502`. This is caused by the same issue as above ("Symbols sidebar and hovers are not working"). See that section for the solution.
