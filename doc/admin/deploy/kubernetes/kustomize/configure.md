# Configure Sourcegraph with Kustomize

This guide will demonstrate how to customize your Sourcegraph deployment by creating a [Kustomize overlay](#create-an-overlay-for-sourcegraph). 

## Create an overlay for Sourcegraph

**IMPORTANT NOTE**: It is recommended to create your own overlays within the [overlays](index.md#overlays) directory of your reference repository and avoid making changes to other directories to prevent potential merge conflicts during future updates.

1. Ensure you have met [all the prerequisites](../kustomize/index.md#prerequisites) and have your private reference repository available locally.
2. Understand the structure of the [reference repository](intro.md#overview) as well as the [custom kustomization file](intro.md#kustomization-yaml) provided by Sourcegraph
3. Create a new directory within the `overlays` subdirectory.
   * The name of this directory `$OVERLAY_NAME` would be the name of your overlay
   * e.g. dev, prod, staging, etc
4. Copy the `kustomization.template.yaml file` from the [overlays/template](intro.md#template) directory to the new directory
   * The template includes a list of components configured for Sourcegraph deployments
   * You can enable a component by leaving it uncommented
   * Commenting a component will disable it
5. Begin configuring the `kustomization file` for your overlay following the instructions provided below

>NOTE: All the configuration should take place within the `$OVERLAY_NAME` overlay directory created in step 3

## Deploy Sourcegraph with an overlay

Generate a new set of manifests for your deployment with `$OVERLAY_NAME`

```bash
$ kubectl kustomize overlays/$OVERLAY_NAME -o cluster.yaml
```

A new set of manifests will be generated and grouped into a single output file `cluster.yaml` for your review. Once you have reviewed the manifests, you can then deploy Sourcegraph to your Kubernetes cluster with a simple kubectl command:

```bash
$ kubectl apply --prune -l deploy=sourcegraph -f cluster.yaml
```

---

## Non-Privileged

By default, all Sourcegraph services are deployed in a **non-root and non-privileged** mode, as defined in the [base](intro.md#base) layer.

## RBAC

Sourcegraph has removed all the Role-Based Access Control (RBAC) resources from base. This means service discovery is not available by default, and the endpoints for each service replica must be manually input into the frontend ConfigMap, which is automatically done by one of the component defined in the [kustomization file](intro.md#kustomization-yaml) built for Sourcegraph.

### Privileged

If you want to deploy a Sourcegraph instance configured for High Availability (HA) to an RBAC-enabled cluster, you can add the privileged component along with the monitoring component as well as cadvisor to your components list. This will enable Kubernetes service discovery for the frontend, and all Sourcegraph services will run with privileged access and as the root user.

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
# Deploy monitoring services for Sourcegraph
- ../../components/monitoring
# Run Sourcegraph main stack with privilege and root
- ../../components/privileged
# Run monitoring services with privilege and root
# This also adds RBAC resources to the monitoring stack
- ../../components/monitoring/privileged
# Add resources for cadvisor
# Require RBAC resources
- ../../components/monitoring/cadvisor
```

> NOTE: When using the [privileged component](#privileged), it is not necessary to include the [enable/service-discovery component](#service-discovery) to avoid duplication of configurations.

### Service discovery

To enable communication between the frontend and other services through the Kubernetes API, it is necessary to have RBAC enabled in your cluster. To enable service discovery for the frontend service, add the following component as the last component in your `kustomization.yaml` file:

```yaml
components:
...
# Add as the last component
- ../../components/enable/service-discovery
```

This will allow the frontend service to discover endpoints for each service replica and communicate with them through the Kubernetes API. Note that this component should only be added if RBAC is enabled in your cluster.

## Monitoring Stack

The monitoring stack for Sourcegraph, similar to the main stack, does not include RBAC (Role-Based Access Control) resources by default. As a result, some dashboards may not display any data unless cAdvisor is deployed with privileged access.

If RBAC is enabled in your cluster, it is highly recommended to deploy cAdvisor with privileged access. With privileged access, cAdvisor will have the necessary permissions to gather and display detailed information about the resources used by your Sourcegraph instance. It's a considered as the key component for monitoring and troubleshooting.

### Deploy cAdvisor

To deploy cAdvisor with privileged access, include the cadvisor component with monitoring in your overlay.

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/monitoring
- ../../components/cadvisor
```

> IMPORTANT: RBAC must be enabled to add RBAC resources.

## Remove securityContext

The `remove/security-context` component removes all the `securityContext` configurations pre-defined in base.

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/remove/security-context
```

## Remove DaemonSets

If you do not have permission to deploy DaemonSets, you can include the remove/daemonset component to remove all services with DaemonSets (node-exporter and otel) from the **monitoring component**:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
# monitoring component
- ../../components/monitoring
# component to remove all daemonsets from the monitoring stack
- ../../components/remove/daemonset
# ❌ Make sure the cadvisor is not included in your components
# - ../../components/cadvisor
```

> NOTE: Make sure to exclude `cAdvisor` from your components as it contains DaemonSet.

ℹ️ If the `monitoring component` is not included in your overlay, adding the `remove/daemonset component` would result in errors because there will be no daemonsets to remove.

## Namespace

To configure the namespace for your Sourcegraph deployment, you can follow these steps:

1. Open the kustomization.yaml file located in your overlay directory (overlays/$OVERLAY_NAME/kustomization.yaml)
2. Add the namespace field in the file, and set it to the desired namespace in your cluster. For example:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: default
```

> NOTE: This step assumes that the namespace already exists in your cluster. If it does not, you will need to create one before applying the overlay.

### Create a namespace

To create a new namespace, you can add the namespace component to your kustomization.yaml file.

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/namespace
```

## Resources

The standard practice is to configure resource allocation using the `sizes component`, which we have preconfigured based on load test results for each [instance size](../../instance-size.md). This will ensure that your instance size is properly allocated

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
# Include the component for your instance size
# Pick ONE only
- ../../components/sizes/xs
- ../../components/sizes/s
- ../../components/sizes/m
- ../../components/sizes/l
- ../../components/sizes/xl
```

## Storage class

A storage class is required for all persistent volume claims by default. The default storage class is called `sourcegraph`. This storage class must be configured before applying the base configuration to your cluster.

See [the official documentation](https://kubernetes.io/docs/tasks/administer-cluster/change-pv-reclaim-policy/#changing-the-reclaim-policy-of-a-persistentvolume) for more information about configuring persistent volumes.

**Kubernetes 1.19 and higher required**

### Google Cloud Platform

1. Read and follow the [official documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver) for enabling the persistent disk CSI driver on a [new](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver#enabling_the_on_a_new_cluster) or [existing](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver#enabling_the_on_an_existing_cluster) cluster.

2. Add the GCP storage class component to the `kustomization.yaml` file for your Kustomize overlay:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/storage-class/gcp
```

[Additional documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver) for more information.

### Amazon Web Services

1. Follow the [official instructions](https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html) to deploy the [Amazon Elastic Block Store (Amazon EBS) Container Storage Interface (CSI) driver](https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html).

2. Add the aws storage class component to the `kustomization.yaml` file for your Kustomize overlay:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/storage-class/aws
```

[Additional documentation](https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html) for more information.

### Azure

> WARNING: If you are deploying on Azure, you **must** ensure that your cluster is created with support for CSI storage drivers [(link)](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers)). This **can not** be enabled after the fact

1. Follow the [official instructions](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers) to deploy the [Container Storage Interface (CSI) drivers](https://learn.microsoft.com/en-us/azure/aks/csi-storage-drivers).

2. Add the azure storage class component to the `kustomization.yaml` file for your Kustomize overlay:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/storage-class/azure
```

[Additional documentation](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers) for more information.

### Trident

If you are using Trident as your storage orchestrator, you must have [fsType](https://docs.netapp.com/us-en/trident/trident-reference/objects.html#storage-pool-selection-attributes) defined in your storageClass for it to respect the volume ownership required by Sourcegraph. When [fsType](https://docs.netapp.com/us-en/trident/trident-reference/objects.html#storage-pool-selection-attributes) is not set, all the files within the cluster will be owned by user 99 (NOBODY), resulting in permission issues for all Sourcegraph databases.

Add one of the available `storage-class/trident/$FSTYPE` components to the `kustomization.yaml` file for your Kustomize overlay based on your fsType:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/storage-class/trident/ext3
- ../../components/storage-class/trident/ext4
- ../../components/storage-class/trident/xfs
```

### Other cloud providers

To use a storage class provided by other cloud providers, follow these steps:

1. Add the `components/storage-class/cloud` component to your overlay:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/storage-class/cloud
```

Update the following variables under the [OVERLAY CONFIGURATIONS](intro.md#overlay-configurations) section in your overlay. Replace them with the correct values according to the instructions provided by your cloud provider:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
########################################################## 
# OVERLAY CONFIGURATIONS
########################################################## 
configMapGenerator:
  # Handle updating configs using env vars for kustomize
  - name: sourcegraph-kustomize-env
    behavior: merge
    literals:
      ...
      - STORAGECLASS_NAME=STORAGECLASS_NAME
      - STORAGECLASS_PROVISIONER=STORAGECLASS_PROVISIONER
      - STORAGECLASS_PARAM_TYPE=STORAGECLASS_PARAM_TYPE
```

> IMPORTANT: Make sure to create the storage class in your cluster before deploying Sourcegraph

## Network access

To allow external users to access the main web server, you need to configure it to be reachable over the network. 

We recommend using the [ingress-nginx](https://kubernetes.github.io/ingress-nginx/) [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) controller for production environments.

### Ingress controller

To utilize the sourcegraph-frontend ingress, you'll need to install a NGINX ingress controller (ingress-nginx) in your cluster by following the official instructions at https://kubernetes.github.io/ingress-nginx/deploy/.

Follow the official instructions at https://kubernetes.github.io/ingress-nginx/deploy/ to install the ingress-nginx controller in your cluster.

Alternatively, you can use one of our pre-configured ingress-nginx-controller overlays:

  ```bash
  # aws
  $ kubectl apply -k quick-start/ingress-nginx-controller/aws
  # or other cloud providers
  $ kubectl apply -k quick-start/ingress-nginx-controller/cloud
  ```

Check the external address by running the following command and look for the `LoadBalancer` entry:

```bash
kubectl -n ingress-nginx get svc
```

Verify that the ingress-nginx IP is accessible. If you are having trouble accessing Sourcegraph, see[Troubleshooting ingress-nginx](https://kubernetes.github.io/ingress-nginx/troubleshooting/) for further assistance. The namespace of the ingress-controller is ingress-nginx.

Once you have completed the installation process for Sourcegraph, run the following command to check if an IP address has been assigned to your ingress resource. This IP address or the configured URL can then be used to access Sourcegraph in your browser.

```sh
kubectl get ingress sourcegraph-frontend

NAME                   CLASS    HOSTS             ADDRESS     PORTS     AGE
sourcegraph-frontend   <none>   sourcegraph.com   8.8.8.8     80, 443   1d
```

### Hostname

To configure the hostname for your Sourcegraph ingress, follow these steps:

**Step 1**: Under the [OVERLAY CONFIGURATIONS](intro.md#overlay-configurations) section, add the `HOST_DOMAIN` variable and set it to your desired hostname, for example:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
########################################################## 
# OVERLAY CONFIGURATIONS
########################################################## 
configMapGenerator:
  # Handle updating configs using env vars for kustomize
  - name: sourcegraph-kustomize-env
    behavior: merge
    literals:
      ...
      - HOST_DOMAIN=sourcegraph.company.com
```

**Step 2**: Include the hostname component in your components.

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/hostname
```

This will configure the hostname for the ingress resource, allowing external users to access Sourcegraph using the specified hostname.

### Annotations

To configure ingress-nginx annotations for the Sourcegraph frontend:

**Step 1**: Create a subdirectory called 'patches' within the directory of your overlay

**Step 2**: Create a file named `frontend-ingress-annotations.yaml` within the new [patches subdirectory](intro.md#patches-directory)

**Step 3**: Copy and paste the context below to the new file

```yaml
# overlays/$OVERLAY_NAME/patches/frontend.annotations.yaml
- op: add
  path: /metadata/annotations
  value:
    # includes the default annotations
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/proxy-body-size: 150m
    # add the additional annotations below
    # ex: nginx.ingress.kubernetes.io/auth-type: basic
```

**Step 4**: Add the additional annotations at the end of the file

**Step 5**: Include the following in your overlay:
   
  ```yaml
  # overlays/$OVERLAY_NAME/kustomization.yaml
  components:
  - ../../components/...
  ...
  patchesJson6902:
  - target:
      version: v1
      kind: Ingress
      name: sourcegraph-frontend
    path: patches/frontend-ingress.annotations.yaml
  ```

This will add the annotations specified in the [frontend-ingress-annotations.yaml](https://github.com/sourcegraph/deploy-sourcegraph-k8s/blob/master/base/frontend/sourcegraph-frontend.Ingress.yaml) file to the sourcegraph-frontend ingress resource. For more information on [ingress-nginx annotations](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/), refer to the [NGINX Configuration documentation](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/).

### TLS

To ensure secure communication, it is recommended to enable Transport Layer Security (TLS) and properly configure a certificate on your Ingress. This can be done by utilizing managed certificate solutions provided by cloud providers or by manually configuring a certificate.

#### Enable manually

To manually configure a certificate via [TLS Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets), follow these steps:

**Step 1**: Move the `tls.crt` and `tls.key` files to the `config` folder within your overlay directory (e.g. `overlays/$OVERLAY_NAME/config`).

**Step 2**: include the `secretGenerator` to generate secrets with the provided files by adding the following to your kustomization file:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
...
secretGenerator:
- name: sourcegraph-frontend-tls
  behavior: create
  files:
  - env/tls.crt
  - env/tls.key
...
```

This will create a new Secret resource named sourcegraph-frontend-tls that contains the encoded cert and key.

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

**Step 3**: Configure the TLS settings on your Ingress by adding the following variables under the [OVERLAY CONFIGURATIONS](intro.md#overlay-configurations) section:

- **TLS_HOST**: your domain name
- **TLS_INGRESS_CLASS_NAME**: ingress class name required by your cluster-issuer
- **TLS_CLUSTER_ISSUER**: name of the cluster-issuer

Example:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
########################################################## 
# OVERLAY CONFIGURATIONS
########################################################## 
configMapGenerator:
  # Handle updating configs using env vars for kustomize
  - name: sourcegraph-kustomize-env
    behavior: merge
    literals:
      ...
      - TLS_HOST=sourcegraph.company.com
      - TLS_INGRESS_CLASS_NAME=example-ingress-class-name
      - TLS_CLUSTER_ISSUER=letsencrypt
```

Step 4: Add the tls component to your kustomization file:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/tls
```

#### Let’s Encrypt

Alternatively, you can configure [cert-manager with Let’s Encrypt](https://cert-manager.io/docs/configuration/acme/) in your cluster. Then, follow the steps listed above for configuring TLS certificate via TLS Secrets manually. However, when adding the variables to the OVERLAY CONFIGURATIONS section, replace **TLS_CLUSTER_ISSUER=letsencrypt** to include the cert-manager with Let's Encrypt.

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
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/nodeport/30080
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

If your [Rancher Kubernetes Engine (RKE)](https://rancher.com/docs/rke/latest/en/) cluster is configured to use [NodePort](https://docs.ranchermanager.rancher.io/v2.0-v2.4/how-to-guides/new-user-guides/migrate-from-v1.6-v2.x/expose-services#nodeport), include the [nodeport/custom component](https://github.com/sourcegraph/deploy-sourcegraph-k8s/tree/main/components/nodeport) to change the port type for `sourcegraph-frontend` service from `ClusterIP` to `NodePort` to use `nodePort: 30080`:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/nodeport
```

> NOTE: Check with your upstream admin for the correct nodePort value.

## NetworkPolicy

To configure network policy for your Sourcegraph installation, you will need to follow these steps:

1. Create a namespace for your Sourcegraph deployment as described in the [namespace section](#namespace).

2. Add the network-policy component to your kustomization file:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/network-policy
```

3. Apply the network-policy component to your cluster. This will create a NetworkPolicy resource that only allows traffic between pods in the namespace labeled with name: sourcegraph-prod

4. If you need to allow traffic to external services or ingress traffic from the outside to the frontend, you will need to augment the example NetworkPolicy. You can check out this [collection](https://github.com/ahmetb/kubernetes-network-policy-recipes) of NetworkPolicies to get started.

> NOTE: You should check with your cluster administrator to ensure that NetworkPolicy is supported in your cluster.

## Environment variables

To update the environment variables for the **sourcegraph-frontend** service, edit the [FRONTEND ENV VARS](intro.md#frontend-env-vars) section at the bottom of your [kustomization file](intro.md#kustomizationyaml). For example:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
########################################################## 
# FRONTEND ENV VARS
########################################################## 
  - name: sourcegraph-frontend-env
    behavior: merge
    literals:
      - DEPLOY_TYPE=kubernetes
      - PGHOST=NEW_PGHOST
      - PGPORT=NEW_PGPORT
      - REDIS_STORE_ENDPOINT=NEW_EDIS_STORE_DSN
      - NEW_ENV_VAR=NEW_VALUE
      ...
```

These values will be automatically merged with the environment variables currently listed in your frontendend's ConfigMap.

## External databases

For optimal performance and resilience, it is recommended to use an external database when deploying Sourcegraph. For more information on database requirements, refer to [this guide](../../postgres.md).

To connect to an existing PostgreSQL instance, add the relevant environment variables ([such as PGHOST, PGPORT, PGUSER, etc.](http://www.postgresql.org/docs/current/static/libpq-envars.html)) under the [FRONTEND ENV VARS](intro.md#frontend-env-vars) section at the bottom of your [kustomization file](intro.md#kustomizationyaml):

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
########################################################## 
# FRONTEND ENV VARS
########################################################## 
  - name: sourcegraph-frontend-env
    behavior: merge
    literals:
      - DEPLOY_TYPE=kubernetes
      ...
      - PGHOST=NEW_PGHOST
      - PGPORT=NEW_PGPORT
```

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

**Step 1**: Add the following component to your overlay:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- .../components/redis
```

**Step 2**: Define the variables under the [FRONTEND ENV VARS](intro.md#frontend-env-vars) section at the bottom of your [kustomization file](intro.md#kustomizationyaml):

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
########################################################## 
# FRONTEND ENV VARS
########################################################## 
  - name: sourcegraph-frontend-env
    behavior: merge
    literals:
      - DEPLOY_TYPE=kubernetes
      ...
      - REDIS_CACHE_ENDPOINT=<REDIS_CACHE_DSN>
      - REDIS_STORE_ENDPOINT=<REDIS_STORE_DSN>
```

## SSH for cloning

Sourcegraph will clone repositories using SSH credentials when the `id_rsa` and `known_hosts` files are mounted at `/home/sourcegraph/.ssh` (or `/root/.ssh` when the cluster is run by root users) in the `gitserver` deployment. 

**WARNING:** Do not commit the actual `id_rsa` and `known_hosts` files to any public repository.

To mount the files through Kustomize:

**Step 1:** Copy the required files to the `configs` folder at the same level as your overylay's kustomization.yaml file

**Step 2:** Add the follow code to your kustomization.yaml file to [generate secrets](https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-kustomize/) and base64 encoded the values in those files

  ```yaml
  # overlays/$OVERLAY_NAME/kustomization.yaml
  ...
  secretGenerator:
  - name: gitserver-ssh
    files:
    - configs/id_rsa
    - configs/known_hosts
  ...
  ```

**Step 3:** Add the following component to mount the [secret as a volume](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod) in [gitserver.StatefulSet.yaml](https://github.com/sourcegraph/deploy-sourcegraph-k8s/blob/master/base/gitserver/gitserver.StatefulSet.yaml).

  ```yaml
  # overlays/$OVERLAY_NAME/kustomization.yaml
  components:
    - ../../components/ssh/non-root
    # OR include the ssh/root component when running with root users
    - ../../components/ssh/root
   ```

**Step 4:** Update code host configuration

Update the configuration file for your code host to enable ssh cloning. For example, set [gitURLType](../../../../admin/external_service/github.md#gitURLType) to ssh for GitHub. See the [external service docs](../../../../admin/external_service.md) for the correct setting for your code host.

## OpenTelemetry Collector

Learn more about Sourcegraph's integrations with the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) in our [OpenTelemetry documentation](../../observability/opentelemetry.md).

Sourcegraph currently supports exporting tracing data to several backends. Refer to [OpenTelemetry](../../observability/opentelemetry.md) for detailed descriptions on how to configure your backend of choice.

### Configure a tracing backend

By default, the collector is [configured to export trace data by logging](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/docker-images/opentelemetry-collector/configs/logging.yaml). Follow these steps to add a config for a different backend:

1. Create a subdirectory called 'patches' within the directory of your overlay
2. Copy and paste the [base/otel-collector/otel-collector.ConfigMap.yaml file](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph-k8s@master/-/tree/base/otel-collector/otel-collector.ConfigMap.yaml) to the new [patches subdirectory](intro.md#patches-directory)
3. In the copied file, make the necessary changes to the `exporters` and `service` blocks to connect to your backend based on the documentation linked above
4. Include the following in your overlay:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/otel-collector/backend
...
patchesStrategicMerge:
- patches/otel-collector.ConfigMap.yaml
...
```

The component will update the `command` for the `otel-collector` container to `"--config=/etc/otel-collector/conf/config.yaml"`, which is now point to the mounted config.

## Add license key

Sourcegraph's Kubernetes deployment [requires an Enterprise license key](https://about.sourcegraph.com/pricing).

Once you have a license key, add it to your [site configuration](https://docs.sourcegraph.com/admin/patches/site_config).

## Filtering cAdvisor metrics

cAdvisor can pick up metrics for services unrelated to the Sourcegraph deployment running on the same nodes
([Learn more](../../../dev/background-information/observability/cadvisor.md#identifying-containers)). To work around this:

1. Create a subdirectory called 'patches' within the directory of your overlay
2. Copy and paste the `base/prometheus/prometheus.ConfigMap.yaml` file to the new [patches subdirectory](intro.md#patches-directory)
2. In the copied file, add the lines highlighted [here](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph@v4.3.1/-/blob/base/prometheus/prometheus.ConfigMap.yaml?L262-264).
3. Replace [ns-sourcegraph](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph@v4.3.1/-/blob/base/prometheus/prometheus.ConfigMap.yaml?L263) with your namespace
4. Include the following in your overlay:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
patchesStrategicMerge:
- patches/prometheus.ConfigMap.yaml
```

This will cause Prometheus to drop all metrics *from cAdvisor* that are not from services in the desired namespace.

## Private registry

To update all image names with your private registry, eg. `index.docker.io/sourcegraph/service_name` to `your.private.registry.com/sourcegraph/service_name`, include the `private-registry` component:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
components:
- ../../components/private-registry
```

Set the `PRIVATE_REGISTRY` variable under the [OVERLAY CONFIGURATIONS](intro.md#overlay-configurations) section:

```yaml
# overlays/$OVERLAY_NAME/kustomization.yaml
########################################################## 
# OVERLAY CONFIGURATIONS
########################################################## 
configMapGenerator:
  # Handle updating configs using env vars for kustomize
  - name: sourcegraph-kustomize-env
    behavior: merge
    literals:
      ...
      - PRIVATE_REGISTRY=your.private.registry.com
```

## Outbound Traffic

When working with an [Internet Gateway](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_Internet_Gateway.html) or VPC it may be necessary to expose ports for outbound network traffic. Sourcegraph must open port 443 for outbound traffic to codehosts, and to enable [telemetry](https://docs.sourcegraph.com/admin/pings) with Sourcegraph.com. Port 22 must also be opened to enable git SSH cloning by Sourcegraph. In addition, please make sure to apply other required changes to secure your cluster in a manner that meets your organization's security requirements.

## Troubleshooting

See the [Troubleshooting docs](../troubleshoot.md).
