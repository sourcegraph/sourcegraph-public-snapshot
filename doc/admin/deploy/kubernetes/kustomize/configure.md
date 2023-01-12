# Configure Sourcegraph with Kustomize

This guide will show you how to customize your deployment by creating a [Kustomize overlay for Sourcegraph](#overlay). 

Once you have an overlay ready, you can then use it to deploy Sourcegraph to your cluster with a simple kubectl command:

```bash
$ kubectl apply -k --prune -l deploy=sourcegraph -f $PATH_TO_OVERLAY
```

> NOTE: If you are deploying Sourcegraph with Helm, please refer to the [configuration guide for Helm](helm.md#configuration).

#### Create an overlay for Sourcegraph

**IMPORTANT NOTE**: Please create your own sets of overlays within the 'overlays' directory and refrain from making changes to the other directories to prevent potential conflicts during future updates and ensure a seamless upgrade process.


1. Ensure you have met [all the prerequisites](../kustomize/index.md#prerequisites) with your private reference repository available locally
2. Create a duplicate of the `new/overlays/template` folder within the `new/overlays` directory
3. Rename the duplicate to the name of your cluster environment (ex. `prod`, `staging`, `aws`, etc) --this would be the name of your overlay: $OVERLAY_NAME
4. Start configuring your overlay accordingly using the instructions below

---

## RBAC

RBAC must be enabled in your cluster for Sourcegraph to do the following:

- Communicates with the Kubernetes API for service discovery
- Runs janitor DaemonSets that clean up temporary cache data

### Non-privileged

By default, RBAC is not required to deploy the Sourcegraph main stacks; however, RBACs must be enabled for the monitoring stacks to be fully functional.

### Privileged

If you have RBAC enabled on your cluster and wish to deploy a Sourcegraph instance with privileged access, you can include the privileged component in your overlay.

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/privileged
```

## Monitoring Stacks

Add the `monitoring component` to deploy the Sourcegraph Monitoring Stacks:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/monitoring
```

> IMPORTANT: RBAC must be enabled for the monitoring stacks to be working properly.

### Non-privileged

To deploy the non-privileged version of the Sourcegraph Monitoring Stacks:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/monitoring/non-privileged
```

All the RBAC-required services are removed from the monitoring stacks. The UI for grafana might still be accessible, but most of the dashboards are expected to be empty.

## Remove DaemonSets

If your cluster does not allow DaemonSets, we recommend removing the Sourcegraph Monitoring Stacks from your deployment, or add the `remove/daemonset component` to remove all DaemonSets from the base resources:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/remove/daemonset
```

## Namespace

Update the [namespace](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/namespace/) field in your `kustomization.yaml` file will apply/override the namespace on all resources included in your overlay:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: default
components:
- ../../components/$COMPONENT_NAME
```

## Resources

We recommend setting up your Sourcegraph deployment usisng resouces configured for your [instance size](../../instance-size.md), and then add the `sizes component` for your instance size to your Overlay:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/sizes/xs
```

## Storage class

Sourcegraph by default requires a storage class for all persisent volumes claims. By default this storage class is called `sourcegraph`. This storage class must be configured before applying the base configuration to your cluster.

See [the official documentation](https://kubernetes.io/docs/tasks/administer-cluster/change-pv-reclaim-policy/#changing-the-reclaim-policy-of-a-persistentvolume) for more information about patching persistent volumes.

**Kubernetes 1.19 and higher required**

### Google Cloud Platform

1. Please read and follow the [official documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver) for enabling the persistent disk CSI driver on a [new](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver#enabling_the_on_a_new_cluster) or [existing](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver#enabling_the_on_an_existing_cluster) cluster.


2. Add the gcp storage class component to the kustomization.yaml file for your Kustomize Overlay:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/storage-class/gcp
```

[Additional documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver).

### Amazon Web Services

1. Follow the [official instructions](https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html) to deploy the Amazon Elastic Block Store (Amazon EBS) Container Storage Interface (CSI) driver.

2. Add the aws storage class component to the kustomization.yaml file for your Kustomize Overlay:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/storage-class/aws
```

[Additional documentation](https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html).

### Azure

> WARNING: If you are deploying on Azure, you **must** ensure that your cluster is created with support for CSI storage drivers [(link)](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers)). This **can not** be enabled after the fact

1. Follow the [official instructions](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers) to deploy the Amazon Elastic Block Store (Amazon EBS) Container Storage Interface (CSI) driver.

2. Add the azure storage class component to the kustomization.yaml file for your Kustomize Overlay:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/storage-class/azure
```

[Additional documentation](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers).


### Trident

If you are using Trident as your storage orchestrator, you must have [fsType](https://docs.netapp.com/us-en/trident/trident-reference/objects.html#storage-pool-selection-attributes) defined in your storageClass for it to respect the volume ownership required by Sourcegraph. When [fsType](https://docs.netapp.com/us-en/trident/trident-reference/objects.html#storage-pool-selection-attributes) is not set, all the files within the cluster will be owned by user 99 (NOBODY), resulting in permission issues for all Sourcegraph databases.

Add one of the available `storage-class/trident/$FSTYPE` components to your overlay as a component based on your fsType:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/storage-class/trident/ext3
- ../../components/storage-class/trident/ext4
- ../../components/storage-class/trident/xfs
```

### Other cloud providers

Add the `new/components/storage-class/cloud` to your overlay:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/storage-class/cloud
```

Update the following variables inside the overlay.config config file for your overlay. Replace them with the correct values according to the instruction provided by your cloud provider:

```
DEPLOY_SOURCEGRAPH_STORAGECLASS_NAME=STORAGECLASS_NAME
DEPLOY_SOURCEGRAPH_STORAGECLASS_PROVISIONER=STORAGECLASS_PROVISIONER
DEPLOY_SOURCEGRAPH_STORAGECLASS_PARAM_TYPE=STORAGECLASS_PARAM_TYPE
```

## Network access

To enable external users to access the main web server, it is necessary to configure it to be reachable over the network.

As part of our base configuration, the [sourcegraph-frontend](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/frontend/sourcegraph-frontend.Ingress.yaml) has implemented ingress with the default ingress rules.

### Ingress controller

We **recommend** using the [ingress-nginx](https://kubernetes.github.io/ingress-nginx/) [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) production environments.

To utilize the sourcegraph-frontend ingress, you'll need to install a NGINX ingress controller (ingress-nginx) in your cluster by following the official instructions at https://kubernetes.github.io/ingress-nginx/deploy/.

You can also install it using one of our ingress-nginx-controller overlays if applicable:

  ```bash
  # aws
  $ kubectl apply -k new/quick-start/ingress-nginx-controller/aws
  # or other cloud providers
  $ kubectl apply -k new/quick-start/ingress-nginx-controller/cloud
  ```

Once the controller has been set up successfully, you can check the external address by running the following command and look for the `LoadBalancer` entry:

```bash
kubectl -n ingress-nginx get svc
```

If you are having trouble accessing Sourcegraph, ensure ingress-nginx IP is accessible above. Otherwise see [Troubleshooting ingress-nginx](https://kubernetes.github.io/ingress-nginx/troubleshooting/). The namespace of the ingress-controller is `ingress-nginx`.

Once you have [installed Sourcegraph](./index.md#installation), run the following command to confirm an IP address has been assigned to your ingress resource before trying to access Sourcegraph in your browser using the IP address or configured URL.

```sh
kubectl get ingress sourcegraph-frontend

NAME                   CLASS    HOSTS             ADDRESS     PORTS     AGE
sourcegraph-frontend   <none>   sourcegraph.com   8.8.8.8     80, 443   1d
```

### Configuration

`ingress-nginx` has extensive configuration documented at [NGINX Configuration](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/). We expect most administrators to modify [ingress-nginx annotations](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/) in [sourcegraph-frontend.Ingress.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/frontend/sourcegraph-frontend.Ingress.yaml), which can be done by adding the following to your overlay:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/...
...
patchesJson6902:
- target:
    version: v1
    kind: Ingress
    name: sourcegraph-frontend
  path: config/frontend-ingress.annotations.yaml
```

Create a file named `frontend.annotations.yaml` within the `config` directory inside your overlay, and add the additional annotations below at the end:

```yaml
# new/overlays/$OVERLAY_NAME/config/frontend.annotations.yaml
- op: add
  path: /metadata/annotations
  value:
    # includes the default annotations
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/proxy-body-size: 150m
    # add the additional annotations below
    # ex: nginx.ingress.kubernetes.io/auth-type: basic
```

### TLS

It is recommended to enable TLS and configure a certificate properly on your Ingress. You can utilize managed certificate solutions provided by Cloud providers, or your own method.

#### Enable manually

To configure TLS certificate via [TLS Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets) manually:

Step 1: Move the tls.cert and tls.key to the env folder inside the overlay directory. Ex. `new/overlays/$OVERLAY_NAME/config`.

Step 2: Generate secrets with the provided files using secretGenerator by adding the following to your kustomization file:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
...
secretGenerator:
- name: sourcegraph-frontend-tls
  behavior: create
  files:
  - env/tls.crt
  - env/tls.key
...
```

A new Secret resource named sourcegraph-frontend-tls will be created with the cert and key encoded:

```yaml
# the data is abbreviated in this example
apiVersion: v1
kind: Secret
type: kubernetes.io/tls
metadata:
  name: sourcegraph-frontend-tls-99dh8g92m5
  namespace: $YOUR_NAMESPACE
data:
  tls.crt: |
    LS...FUlRJRklDQVRFLS0tLS0=
  tls.key: |
    LS...SSVZBVEUgS0VZLS0tLS0=
```

Step 3: Configure the TLS settings on your Ingress using variables

Add the following variables to your overlay.config file:

- **TLS_HOST**: your domain name
- **TLS_INGRESS_CLASS_NAME**: ingress class name required by your cluster-issuer
- **TLS_CLUSTER_ISSUER**: name of the cluster-issuer

Example:

```yaml
# new/overlays/$OVERLAY_NAME/config/overlay.config
TLS_HOST=sourcegraph.company.com
TLS_INGRESS_CLASS_NAME=example-ingress-class-name
TLS_CLUSTER_ISSUER=letsencrypt
```

Step 4: Add the tls component to your kustomization file:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/tls
```

#### Let’s Encrypt

Alternatively, you may consider configuring [cert-manager with Let’s Encrypt](https://cert-manager.io/docs/configuration/acme/) in your cluster, and then configure frontend ingress using the same steps listed above, but replace `TLS_CLUSTER_ISSUER=letsencrypt` in your overlay.config file.

## Network rule

> NOTE: this setup path does not support TLS.

Add a network rule that allows ingress traffic to port 30080 (HTTP) on at least one node.

### Google Cloud Platform Firewall

- Expose the necessary ports.

```bash
gcloud compute --project=$PROJECT firewall-rules create sourcegraph-frontend-http --direction=INGRESS --priority=1000 --network=default --action=ALLOW --rules=tcp:30080
```

- Add the nodeport component to change the type of the `sourcegraph-frontend` service from `ClusterIP` to `NodePort` with the `nodeport` component:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/nodeport
```

- Directly applying this change to a running service [will fail](https://github.com/kubernetes/kubernetes/issues/42282). You must first delete the old service before redeploying a new one (with a few seconds of downtime):

```bash
kubectl delete svc sourcegraph-frontend
kubectl apply -k $PATH_TO_OVERLAY
```

- Find a node name.

```bash
kubectl get pods -l app=sourcegraph-frontend -o=custom-columns=NODE:.spec.nodeName
```

- Get the EXTERNAL-IP address (will be ephemeral unless you [make it static](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address#promote_ephemeral_ip)).

```bash
kubectl get node $NODE -o wide
```

Learn more about [Google Cloud Platform Firewall rules](https://cloud.google.com/compute/docs/vpc/using-firewalls).

### AWS Security Group

Sourcegraph should now be accessible at `$EXTERNAL_ADDR:30080`, where `$EXTERNAL_ADDR` is the address of _any_ node in the cluster.

Learn more about [AWS Security Group rules](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_SecurityGroups.html).

### Rancher Kubernetes Engine

If your [Rancher Kubernetes Engine (RKE)](https://rancher.com/docs/rke/latest/en/) cluster is configured to use [NodePort](https://docs.ranchermanager.rancher.io/v2.0-v2.4/how-to-guides/new-user-guides/migrate-from-v1.6-v2.x/expose-services#nodeport), use the [nodeport/custom component](https://github.com/sourcegraph/deploy-sourcegraph/tree/main/new/components/nodeport) to change the port type for `sourcegraph-frontend` service from `ClusterIP` to `NodePort` to use `nodePort: 30080`:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/nodeport
```

> NOTE: Check with your upstream admin for the correct nodePort value.

## NetworkPolicy

Network policy is a Kubernetes resource that defines how pods are allowed to communicate with each other and with
other network endpoints. If the cluster administration requires an associated NetworkPolicy when doing an installation,
then we recommend running Sourcegraph in a namespace as described in the [namespace section](#namespace) and then apply the 
[network-policy component](https://github.com/sourcegraph/deploy-sourcegraph/tree/main/new/components/network-policy) that utilizes the `namespaceSelector` to allow traffic between the Sourcegraph pods. 
The [NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) created by this component will 
only allow resources that are deployed to this specified namespace that has the `name: sourcegraph-prod` label.

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/network-policy
```

> NOTE: You will need to augment this example NetworkPolicy to allow traffic to external services
> you plan to use (like github.com) and ingress traffic from
> the outside to the frontend for the users of the Sourcegraph installation.
> Check out this [collection](https://github.com/ahmetb/kubernetes-network-policy-recipes) of NetworkPolicies to get started.

## External databases

We recommend utilizing an external database when deploying Sourcegraph to provide the most resilient and performant backend for your deployment. For more information on the specific requirements for Sourcegraph databases, see [this guide](../../postgres.md).

Simply add the relevant PostgreSQL environment variables (e.g. PGHOST, PGPORT, PGUSER, [etc.](http://www.postgresql.org/docs/current/static/libpq-envars.html)) to point them to your existing PostgreSQL instance inside the `new/overlays/$OVERLAY_NAME/env/frontend.env file`:

```sh
# new/overlays/$OVERLAY_NAME/env/frontend.env
...
PGHOST=NEW_PGHOST
PGPORT=NEW_PGPORT
```

## SSH for cloning

Sourcegraph will clone repositories using SSH credentials when the `id_rsa` and `known_hosts` files are mounted at `/home/sourcegraph/.ssh` (or `/root/.ssh` when the cluster is run by root users) in the `gitserver` deployment. 

**WARNING:** Do not commit the actual `id_rsa` and `known_hosts` files to any public repository.

To mount the files through Kustomize:

**Step 1:** Copy the required files to the `configs` folder at the same level as your overylay's kustomization.yaml file

**Step 2:** Add the follow code to your kustomization.yaml file to [generate secrets](https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-kustomize/) and base64 encoded the values in those files

  ```yaml
  # new/overlays/$OVERLAY_NAME/kustomization.yaml
  ...
  secretGenerator:
  - name: gitserver-ssh
    files:
    - configs/id_rsa
    - configs/known_hosts
  ...
  ```

**Step 3:** Add the following component to mount the [secret as a volume](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod) in [gitserver.StatefulSet.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/gitserver/gitserver.StatefulSet.yaml).

  ```yaml
  # new/overlays/$OVERLAY_NAME/kustomization.yaml
  components:
    - ../../components/ssh/non-root
    # OR use the ssh/root component when running with root users
    - ../../components/ssh/root
   ```

**Step 4:** Update code host configuration

Update the configuration file for your code host to enable ssh cloning. For example, set [gitURLType](../../../../admin/external_service/github.md#gitURLType) to ssh for GitHub. See the [external service docs](../../../../admin/external_service.md) for the correct setting for your code host.

## Custom Redis

Sourcegraph supports specifying a custom Redis server with these environment variables:

- **REDIS_CACHE_ENDPOINT**=[redis-cache:6379](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24++REDIS_CACHE_ENDPOINT+AND+REDIS_STORE_ENDPOINT+-file:doc+file:internal&patternType=literal) for caching information.
- **REDIS_STORE_ENDPOINT**=[redis-store:6379](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24++REDIS_CACHE_ENDPOINT+AND+REDIS_STORE_ENDPOINT+-file:doc+file:internal&patternType=literal) for storing information (session data and job queues). 

When using a custom Redis server, the corresponding environment variable must also be added to the following services:

<!-- Use ./dev/depgraph/depgraph summary internal/redispool to generate this -->

- `sourcegraph-frontend`
- `repo-updater`
- `gitserver`
- `searcher`
- `symbols`
- `worker`

Step 1: Add the following component to your overlay:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- .../components/redis
```

Step 2: Define the variables inside the `env/frontend.env` file that is located within your overlay directory:

```sh
# new/overlays/$OVERLAY_NAME/env/frontend.env
REDIS_CACHE_ENDPOINT=<REDIS_CACHE_DSN>
REDIS_STORE_ENDPOINT=<REDIS_STORE_DSN>
```

## OpenTelemetry Collector

Learn more about Sourcegraph's integrations with the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) in our [OpenTelemetry documentation](../../observability/opentelemetry.md).

Sourcegraph currently supports exporting tracing data to several backends. Refer to [OpenTelemetry](../../observability/opentelemetry.md) for detailed descriptions on how to configure your backend of choice.

### Configure a tracing backend

By default, the collector is [configured to export trace data by logging](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/docker-images/opentelemetry-collector/configs/logging.yaml). Follow these steps to add a config for a different backend:

1. Copy and paste the [base/otel-collector/otel-collector.ConfigMap.yaml file](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph@master/-/tree/base/otel-collector/otel-collector.ConfigMap.yaml) to the `config subdirectory` of your overlay
2. In the copied file, make the necessary changes to the `exporters` and `service` blocks to connect to your backend based on the documentation linked above
3. Add the following lines to your overlay:

  ```yaml
  # new/overlays/$OVERLAY_NAME/kustomization.yaml
  components:
  - ../../components/otel-collector/backend
  patchesStrategicMerge:
  - config/otel-collector.ConfigMap.yaml
  ...
  ```

The component will update the `command` for the `otel-collector` container to `"--config=/etc/otel-collector/conf/config.yaml"`, which is now point to the mounted config.

## Add license key

Sourcegraph's Kubernetes deployment [requires an Enterprise license key](https://about.sourcegraph.com/pricing).

Once you have a license key, add it to your [site configuration](https://docs.sourcegraph.com/admin/config/site_config).

## Environment variables

Update the environment variables for frontend inside the `env/frontend.env` file that is located within your overlay directory. For example:

```sh
# new/overlays/$OVERLAY_NAME/env/frontend.env
PGHOST=NEW_PGHOST
REDIS_STORE_ENDPOINT=NEW_EDIS_STORE_DSN
NEW_ENV_VAR=NEW_VALUE
```

The values will be automatically merge with the environment variables currently listed in your frontendend's ConfigMap.

## Filtering cAdvisor metrics

Due to how cAdvisor works, Sourcegraph's cAdvisor deployment can pick up metrics for services unrelated to the Sourcegraph deployment running on the same nodes as Sourcegraph services.
[Learn more](../../../dev/background-information/observability/cadvisor.md#identifying-containers). To work around this:

1. Copy and paste the `new/base/prometheus/prometheus.ConfigMap.yaml` file to the `config subdirectory` of your overlay
2. In the **new** prometheus.ConfigMap.yaml copy, uncomment the lines highlighted [here](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph@v4.3.1/-/blob/base/prometheus/prometheus.ConfigMap.yaml?L262-264).
3. Replace [ns-sourcegraph](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph@v4.3.1/-/blob/base/prometheus/prometheus.ConfigMap.yaml?L263) with your namespace
4. Add the following to your overlay file:

```yaml
 # new/overlays/$OVERLAY_NAME/kustomization.yaml
patchesStrategicMerge:
- config/prometheus.ConfigMap.yaml
```

This will cause Prometheus to drop all metrics *from cAdvisor* that are not from services in the desired namespace.

## Private registry

To update all image names with your private registry, eg. `index.docker.io/sourcegraph/service_name` to `your.private.registry.com/sourcegraph/service_name`, use the `private-registry` component:

```yaml
# new/overlays/$OVERLAY_NAME/kustomization.yaml
components:
- .../components/private-registry
```

Set the `DEPLOY_SOURCEGRAPH_PRIVATE_REGISTRY` variable in your overlay.config:

```yaml
DEPLOY_SOURCEGRAPH_PRIVATE_REGISTRY=your.private.registry.com
```

## Outbound Traffic

When working with an [Internet Gateway](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_Internet_Gateway.html) or VPC it may be necessary to expose ports for outbound network traffic. Sourcegraph must open port 443 for outbound traffic to codehosts, and to enable [telemetry](https://docs.sourcegraph.com/admin/pings) with Sourcegraph.com. Port 22 must also be opened to enable git SSH cloning by Sourcegraph. Take care to secure your cluster in a manner that meets your organization's security requirements.

## Troubleshooting

See the [Troubleshooting docs](../troubleshoot.md).
