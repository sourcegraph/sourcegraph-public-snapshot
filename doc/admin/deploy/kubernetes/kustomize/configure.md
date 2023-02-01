# Configure Sourcegraph with Kustomize

This guide will demonstrate how to customize your Sourcegraph deployment by creating a [Kustomize overlay](#create-an-overlay-for-sourcegraph).

<div class="getting-started">
  <a class="btn text-center" href="../">Installation</a>
  <a class="btn text-center" href="../kustomize">Introduction</a>
  <a class="btn text-center btn-primary" href="#">★ Configuration</a>
  <a class="btn text-center" href="../operations">Maintenance</a>
</div>

## Overview

To ensure optimal performance and functionality of your Sourcegraph deployment, it is recommended to use components pre-configured by us in your overlay. These components include settings that have been specifically designed and tested for Sourcegraph and do not require any additional configuration changes.

Understand how each component works and its dependencies before combining components in your overlay file. Review component configuration settings and resources to ensure components are compatible and work together properly. Disable components carefully, as some may depend on others. 

Following these guidelines will help you create a seamless deployment and avoid conflicts.

> NOTE: All commands in this guide should be run from the root of the reference repository.


## Base cluster

The base resources in Sourcegraph include the services that make up the main Sourcegraph apps as well as the monitoring services. These services are configured to run as non-root users without privileges, ensuring a secure deployment.

The base resources also include a set of services that are responsible for providing metrics about the Sourcegraph cluster, such as CPU and memory usage. To enable cluster metrics, you will need to deploy Cadvisor, which is a container resource usage monitoring service. Cadvisor is configured for Sourcegraph and can be deployed using one of the provided components. This component contains RBAC resources and must be run with privileges to ensure that it has the necessary permissions to access the container metrics.

### RBAC

Sourcegraph has removed all the Role-Based Access Control (RBAC) resources from the default base cluster. This means service discovery is not available by default, and the endpoints for each service replica must be manually input into the frontend ConfigMap, which is automatically done by one of the component defined in the [kustomization file](index.md#kustomization-yaml) built for Sourcegraph.

### Non-Privileged

By default, all Sourcegraph services are deployed in a non-root and non-privileged mode, as defined in the [base](index.md#base) cluster.

### Privileged

To deploy a High Availability (HA) configured Sourcegraph instance to an RBAC-enabled cluster, you can include the [privileged](#privileged) and [monitoring](#monitoring-stack) components in your components list. 

This will enable Kubernetes service discovery for the frontend and also provide privileged access and run all Sourcegraph services as the root user by adding [cadvisor component](#deploy-cadvisor) in the list.

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
# Deploy monitoring services for Sourcegraph
- ../../components/monitoring
# Run Sourcegraph main stack with privilege and root
- ../../components/privileged
# Run monitoring services with privilege and root
# This also adds RBAC resources to the monitoring stack
- ../../components/monitoring/privileged
# Add resources for cadvisor
# cadvisor includes RBAC resources and must be run with privileges
- ../../components/monitoring/cadvisor
```

> NOTE: When including the [privileged component](#privileged), please remove the [enable/service-discovery component](#service-discovery) to avoid duplication of configurations.

### Service discovery

RBAC must be enabled in your cluster for frontend to communicate with other services through the Kubernetes API. To enable service discovery for the frontend service, Include the following component as the **last** component in your `kustomization.yaml` file:

```yaml
components:
...
# IMPORTANT: Include as the last component
- ../../components/enable/service-discovery
```

This will allow the frontend service to discover endpoints for each service replica and communicate with them through the Kubernetes API. Note that this component should only be added if RBAC is enabled in your cluster.

## Monitoring stack

The monitoring stack for Sourcegraph, similar to the main stack, does not include RBAC (Role-Based Access Control) resources by default. As a result, some dashboards may not display any data unless cAdvisor is deployed seperately with privileged access.

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
# Deploy monitoring services for Sourcegraph
- ../../components/monitoring
```

If RBAC is enabled in your cluster, it is highly recommended to deploy cAdvisor with privileged access to your cluster. With privileged access, cAdvisor will have the necessary permissions to gather and display detailed information about the resources used by your Sourcegraph instance. It's a considered as the key component for monitoring and troubleshooting. See [Deploy cAdvisor](#deploy-cadvisor) below for more information.

### Deploy cAdvisor

cAdvisor requires a service account and certain permissions to access and gather information about the Kubernetes cluster in order to display key metrics such as resource usage and performance data. Removing the service account for cAdvisor could impede its ability to collect this information, resulting in missing data on Grafana dashboards and potentially impacting visibility and monitoring capabilities for the cluster and its pods. This could negatively impact the level of monitoring and visibility into the cluster's state that cAdvisor is able to provide.

To deploy cAdvisor with privileged access, include the monitoring/cadvisor component **in addition to** the [monitoring component](#monitoring-stack) in your overlay.

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
# Deploy monitoring services for Sourcegraph
- ../../components/monitoring
# Run Sourcegraph main stack with privilege and root
- ../../components/privileged
# Run monitoring services with privilege and root
# This also adds RBAC resources to the monitoring stack
- ../../components/monitoring/privileged
# Add resources for cadvisor
# cadvisor includes RBAC resources and must be run with privileges
- ../../components/monitoring/cadvisor
```

> NOTE: Make sure to exclude `cAdvisor` from your components as it contains DaemonSet.

ℹ️ If the `monitoring component` is not included in your overlay, adding the `remove/daemonset component` would result in errors because there will be no daemonsets to remove.

## Namespace

Follow the steps below to add namespace to all your Sourcegraph resources:

1. Open the `kustomization.yaml` file located in your overlay directory
2. Include the namespace field in the file, and set it to an exisiting namespace in your cluster. For example:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: ns-sourcegraph
```

> NOTE: This step assumes that the namespace already exists in your cluster. If it does not, you will need to create one before applying the overlay.

### Create a namespace

To create a new namespace, include the [utils/namespace](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph-k8s/-/tree/components/resources/namespace) component in your overlay.

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/resources/namespace
```

## Resources

Properly allocating resources is crucial for ensuring optimal performance of your Sourcegraph instance. To ensure this, it is recommended to use one of the provided [sizes components](#instance-size-based-resources) for resource allocation, specifically designed for your [instance size](../../instance-size.md). These components have been tested and optimized based on load test results, and are designed to work seamlessly with Sourcegraph's design and functionality.

### Instance-size-based resources

To allocate resources based on your [instance size](../../instance-size.md):

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
# Include ONE of the sizes component based on your instance size
- ../../components/sizes/xs
- ../../components/sizes/s
- ../../components/sizes/m
- ../../components/sizes/l
- ../../components/sizes/xl
```

### Custom resources allocation

In cases where custom resource allocation is necessary, it is important to follow the instructions provided below:

**Step 1**: Create a copy of the `components/custom/resources` directory inside your overlay directory, and name it `custom-resources`:

```bash
# rename the directory from 'custom/resources' to 'custom-resources'
$ cp -R components/custom/resources instances/$INSTANCE_NAME/custom-resources
```

**Step 2**: In the copied version of the `resources.yaml` file, located in the `configs subdirectory` of the `custom-resources` directory, uncomment the desired service and update the resource values as necessary.

**Step 3**: In the `instances/$INSTANCE_NAME/kustomization.yaml` file, include the `custom-resources component` in the components list.

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- custom-resources
```

> WARNING: If service-discovery is not enabled for the sourcegraph-frontend service, the endpoint-update.yaml file within the patches subdirectory is responsible for updating the relevant variables for the frontend to generate the endpoint addresses for each service replica. It should not be removed at any point.

### Remove securityContext

The `remove/security-context` component removes all the `securityContext` configurations pre-defined in base.

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/remove/security-context
```

### Remove DaemonSets

If you do not have permission to deploy DaemonSets, you can include the remove/daemonset component to remove all services with DaemonSets resources (e.g. node-exporter and otel) from the [monitoring component](#monitoring-stack):

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
# monitoring component
- ../../components/monitoring
# component to remove all daemonsets from the monitoring stack
- ../../components/remove/daemonset
# Make sure the cadvisor is excluded from the components list
# - ../../components/monitoring/cadvisor
```

## Storage class

A storage class is required for all persistent volume claims by default. It must be configured and created before deploying Sourcegraph to your cluster.

See [the official documentation](https://kubernetes.io/docs/tasks/administer-cluster/change-pv-reclaim-policy/#changing-the-reclaim-policy-of-a-persistentvolume) for more information about configuring persistent volumes.

**Kubernetes 1.19 and higher required**

### Google Cloud Platform

1. Read and follow the [official documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver) for enabling the persistent disk CSI driver on a [new](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver#enabling_the_on_a_new_cluster) or [existing](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver#enabling_the_on_an_existing_cluster) cluster.

2. Include the GCP storage class component to the `kustomization.yaml` file for your Kustomize overlay:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/storage-class/gcp
```

The component takes care of creating a new storage class named `sourcegraph` with the following configurations:

- Provisioner: pd.csi.storage.gke.io
- SSD: types: pd-ssd

It also update the storage class name for all resources to `sourcegraph`.

[Additional documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver) for more information.

### Amazon Web Services

1. Follow the [official instructions](https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html) to deploy the [Amazon Elastic Block Store (Amazon EBS) Container Storage Interface (CSI) driver](https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html).

2. Include one of the AWS storage class components in your overlay: [storage-class/aws/eks](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph-k8s/-/tree/components/storage-class/aws/eks) or [storage-class/aws/ebs](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph-k8s/-/tree/components/storage-class/aws/ebs)
   * The [storage-class/aws/eks](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph-k8s/-/tree/components/storage-class/aws/eks) component is configured with the `ebs.csi.aws.com` storage class provisioner for clusters with self-managed Amazon EBS Container Storage Interface driver installed
   * The [storage-class/aws/ebs](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph-k8s/-/tree/components/storage-class/aws/ebs) component is configured with the `kubernetes.io/aws-ebs` storage class provisioner for clusters with the [AWS EBS CSI driver installed as Amazon EKS add-on](https://docs.aws.amazon.com/eks/latest/userguide/managing-ebs-csi.html)


```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/storage-class/aws/eks
- ../../components/storage-class/aws/ebs
```

[Additional documentation](https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html) for more information.

### Azure

> WARNING: If you are deploying on Azure, you **must** ensure that your cluster is created with support for CSI storage drivers [(link)](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers)). This **can not** be enabled after the fact

1. Follow the [official instructions](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers) to deploy the [Container Storage Interface (CSI) drivers](https://learn.microsoft.com/en-us/azure/aks/csi-storage-drivers).

2. Include the azure storage class component to the `kustomization.yaml` file for your Kustomize overlay:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/storage-class/azure
```

[Additional documentation](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers) for more information.

### k3s

Configure to use the default storage class `local-path` in a k3s cluster:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/storage-class/k3s
```

### Trident

If you are using Trident as your storage orchestrator, you must have [fsType](https://docs.netapp.com/us-en/trident/trident-reference/objects.html#storage-pool-selection-attributes) defined in your storageClass for it to respect the volume ownership required by Sourcegraph. When [fsType](https://docs.netapp.com/us-en/trident/trident-reference/objects.html#storage-pool-selection-attributes) is not set, all the files within the cluster will be owned by user 99 (NOBODY), resulting in permission issues for all Sourcegraph databases.

Add one of the available `storage-class/trident/$FSTYPE` components to the `kustomization.yaml` file for your Kustomize overlay based on your fsType:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
# -- fsType: ext3
- ../../components/storage-class/trident/ext3
# -- fsType: ext4
- ../../components/storage-class/trident/ext4
# -- fsType: xfs
- ../../components/storage-class/trident/xfs
```

### Other cloud providers

To use an existing storage class provided by other cloud providers:

1. Include the `storage-class/update-class-name` component to your overlay:

2. Enter the value for the existing storage class under the **literals** list in the *configMapGenerator* section using the `STORAGECLASS_NAME` config key

The `storage-class/update-class-name` component updates the `storageClassName` field for all associated resources to the `STORAGECLASS_NAME` value set in step 2.

Example, add `STORAGECLASS_NAME=sourcegraph` if `sourcegraph` is the name for the existing storage class:

  ```yaml
  # instances/$INSTANCE_NAME/kustomization.yaml
  components:
    # Update storageClassName to the STORAGECLASS_NAME value set below
    - ../../components/storage-class/update-class-name
    
  configMapGenerator:
  - name: sourcegraph-kustomize-env
    behavior: merge
    literals:
      - STORAGECLASS_NAME=sourcegraph # -- [ACTION] Set storage class name here
  ```


### Create a custom storage class

To create a custom storage class:

1. Include the `storage-class/cloud` component to your overlay:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/storage-class/cloud
```

Update the following variables under the [BUILD CONFIGURATIONS](index.md#build-configurations) section in your overlay. Replace them with the correct values according to the instructions provided by your cloud provider:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml > [BUILD CONFIGURATIONS]
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

To utilize the `sourcegraph-frontend` ingress, you'll need to install a NGINX ingress controller (ingress-nginx) in your cluster. Follow the official instructions at https://kubernetes.github.io/ingress-nginx/deploy/ to install the ingress-nginx controller in your cluster.

Alternatively, you can build the manifests using one of our pre-configured ingress-nginx-controller overlays:

  ```bash
  # Build manifests for AWS
  $ kubectl kustomize examples/ingress-controller/aws -o ingress-controller.yaml
  # Build manifests for other cloud providers
  $ kubectl kustomize examples/ingress-controller/cloud -o ingress-controller.yaml
  # Deploy to cluster after reviewing the manifests in ingress-controller.yaml
  $ kubectl apply -f ingress-controller.yaml
  ```

Check the external address by running the following command and look for the `LoadBalancer` entry:

  ```bash
  $ kubectl -n ingress-nginx get svc
  ```

Verify that the ingress-nginx IP is accessible. If you are having trouble accessing Sourcegraph, see [Troubleshooting ingress-nginx](https://kubernetes.github.io/ingress-nginx/troubleshooting/) for further assistance. The namespace of the ingress-controller is ingress-nginx.

Once you have completed the installation process for Sourcegraph, run the following command to check if an IP address has been assigned to your ingress resource. This IP address or the configured URL can then be used to access Sourcegraph in your browser.

```sh
kubectl get ingress sourcegraph-frontend

NAME                   CLASS    HOSTS             ADDRESS     PORTS     AGE
sourcegraph-frontend   <none>   sourcegraph.com   8.8.8.8     80, 443   1d
```

### TLS

To ensure secure communication, it is recommended to enable Transport Layer Security (TLS) and properly configure a certificate on your Ingress. This can be done by utilizing managed certificate solutions provided by cloud providers or by manually configuring a certificate.

To manually configure a certificate via [TLS Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets), follow these steps:

**Step 1**: Move the `tls.crt` and `tls.key` files to the root of your overlay directory (e.g. `instances/$INSTANCE_NAME`).

**Step 2**: Include the following lines in your overlay to generate secrets with the provided files:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml > [SECRETS GENERATOR]
secretGenerator:
- name: sourcegraph-frontend-tls
  behavior: create
  files:
  - tls.crt
  - tls.key
```

This will create a new Secret resource named sourcegraph-frontend-tls that contains the encoded cert and key.

```yaml
# cluster.yaml - output file after running build
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
# the data is abbreviated in this example
```

**Step 3**: Configure the TLS settings on your Ingress by adding the following variables under the [BUILD CONFIGURATIONS](index.md#build-configurations) section:

- **TLS_HOST**: your domain name
- **TLS_INGRESS_CLASS_NAME**: ingress class name required by your cluster-issuer
- **TLS_CLUSTER_ISSUER**: name of the cluster-issuer

Example:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml > [BUILD CONFIGURATIONS]
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

Step 4: Include the tls component:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/network/tls
```

### TLS with Let’s Encrypt

Alternatively, you can configure [cert-manager with Let’s Encrypt](https://cert-manager.io/docs/configuration/acme/) in your cluster. Then, follow the steps listed above for configuring TLS certificate via TLS Secrets manually. However, when adding the variables to the BUILD CONFIGURATIONS section, set **TLS_CLUSTER_ISSUER=letsencrypt** to include the cert-manager with Let's Encrypt.

## Ingress

Configuration options for ingress installed for sourcegraph-frontend.

### AWS ALB

Component to configure Ingress to use [AWS Load Balancer Controller](https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html) to expose Sourcegraph publicly by updating annotation to `kubernetes.io/ingress.class: alb` in frontend ingress. 

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/ingress/alb
```

### GKE

Component to configure network access for GKE clusters with HTTP load balancing enabled.

It also adds a [BackendConfig CRD](https://cloud.google.com/kubernetes-engine/docs/how-to/ingress-configuration#create_backendconfig). This is necessary to instruct the GCP load balancer on how to perform health checks on our deployment.

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/ingress/gke
```

### k3s

Component to configure Ingress to use the default HTTP reverse proxy and load balancer `traefik` in k3s clusters.

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/ingress/k3s
```

### Hostname

To configure the hostname for your Sourcegraph ingress, follow these steps:

**Step 1**: Under the [BUILD CONFIGURATIONS](index.md#build-configurations) section, include the `HOST_DOMAIN` variable and set it to your desired hostname, for example:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml > [BUILD CONFIGURATIONS]
configMapGenerator:
  # Handle updating configs using env vars for kustomize
  - name: sourcegraph-kustomize-env
    behavior: merge
    literals:
      - HOST_DOMAIN=sourcegraph.company.com
```

**Step 2**: Include the hostname component in your components.

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/ingress/hostname
```

This will configure the hostname for the ingress resource, allowing external users to access Sourcegraph using the specified hostname.

### ClusterIP

The Sourcegraph frontend service is configured as a [ClusterIP](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types) by default, which allows it to be accessed within the Kubernetes cluster using a stable IP address provided by Kubernetes. If you want to make the frontend service accessible from outside the cluster, you can use the [network/nodeport](#nodeport) or [network/loadbalancer](#loadbalancer) components.

### NodePort

The `network/nodeport` component creates a frontend service of [type NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport), making it accessible by using the IP address of any node in the cluster, along with the specified nodePort value (30080).

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/network/nodeport/30080
```

### LoadBalancer

The `network/loadbalancer` component sets the type of the frontend service as [LoadBalancer](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer), which provisions a load balancer and makes the service accessible from outside the cluster.

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/network/loadbalancer
```

### Annotations

To configure ingress-nginx annotations for the Sourcegraph frontend ingress:

**Step 1**: Create a subdirectory called 'patches' within the directory of your overlay

```bash
$ mkdir -p instances/$INSTANCE_NAME/patches
```

**Step 2**: Copy the `frontend-ingress-annotations.yaml` patch file from the components/patches directory to the new [patches subdirectory](index.md#patches-directory)

```bash
$ cp components/patches/frontend-ingress-annotations.yaml instances/$INSTANCE_NAME/patches/frontend-ingress-annotations.yaml
```

**Step 3**: Add the additional annotations at the end of the new patch file 

**Step 4**: Include the patch file in your overlay under `patchesStrategicMerge`:
   
  ```yaml
  # instances/$INSTANCE_NAME/kustomization.yaml
  components:
  - ../../components/...
  ...
  patchesStrategicMerge:
    - patches/frontend-ingress.annotations.yaml
  ```

This will add the annotations specified in your copy of the [frontend-ingress-annotations.yaml](https://github.com/sourcegraph/deploy-sourcegraph-k8s/blob/master/base/frontend/sourcegraph-frontend.Ingress.yaml) file to the sourcegraph-frontend ingress resource. For more information on [ingress-nginx annotations](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/), refer to the [NGINX Configuration documentation](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/).

### NetworkPolicy

To configure network policy for your Sourcegraph installation, you will need to follow these steps:

1. Create a namespace for your Sourcegraph deployment as described in the [namespace section](#namespace).

2. Include the network-policy component:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/network-policy
```

3. Apply the network-policy component to your cluster. This will create a NetworkPolicy resource that only allows traffic between pods in the namespace labeled with name: sourcegraph-prod

4. If you need to allow traffic to external services or ingress traffic from the outside to the frontend, you will need to augment the example NetworkPolicy. You can check out this [collection](https://github.com/ahmetb/kubernetes-network-policy-recipes) of NetworkPolicies to get started.

> NOTE: You should check with your cluster administrator to ensure that NetworkPolicy is supported in your cluster.

## Network rule

Add a network rule that allows incoming traffic on port 30080 (HTTP) to at least one node. Note that this configuration does not include support for Transport Layer Security (TLS).

### Google Cloud Platform Firewall

- Expose the necessary ports.

```bash
$ gcloud compute --project=$PROJECT firewall-rules create sourcegraph-frontend-http --direction=INGRESS --priority=1000 --network=default --action=ALLOW --rules=tcp:30080
```

- Include the nodeport component to change the type of the `sourcegraph-frontend` service from `ClusterIP` to `NodePort` with the `nodeport` component:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/network/nodeport/30080
```

- Directly applying this change to a running service [will fail](https://github.com/kubernetes/kubernetes/issues/42282). You must first delete the old service before redeploying a new one (with a few seconds of downtime):

```bash
$ kubectl delete svc sourcegraph-frontend
$ kubectl apply -k $PATH_TO_OVERLAY
```

- Find a node name.

```bash
$ kubectl get pods -l app=sourcegraph-frontend -o=custom-columns=NODE:.spec.nodeName
```

- Get the EXTERNAL-IP address (will be ephemeral unless you [make it static](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address#promote_ephemeral_ip)).

```bash
$ kubectl get node $NODE -o wide
```

Learn more about [Google Cloud Platform Firewall rules](https://cloud.google.com/compute/docs/vpc/using-firewalls).

### AWS Security Group

Sourcegraph should now be accessible at `$EXTERNAL_ADDR:30080`, where `$EXTERNAL_ADDR` is the address of _any_ node in the cluster.

Learn more about [AWS Security Group rules](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_SecurityGroups.html).

### Rancher Kubernetes Engine

If your [Rancher Kubernetes Engine (RKE)](https://rancher.com/docs/rke/latest/en/) cluster is configured to use [NodePort](https://docs.ranchermanager.rancher.io/v2.0-v2.4/how-to-guides/new-user-guides/migrate-from-v1.6-v2.x/expose-services#nodeport), include the [network/nodeport/custom component](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph-k8s/-/tree/components/network/nodeport/custom) to change the port type for `sourcegraph-frontend` service from `ClusterIP` to `NodePort` to use `nodePort: 30080`:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/network/nodeport/30080
```

> NOTE: Check with your upstream admin for the correct nodePort value.

## Service mesh

There are a few [known issues](../troubleshoot.md#service-mesh) when running Sourcegraph with service mesh. We recommend including the `network/envoy` component in your components list to bypass the issue where Envoy, the proxy used by Istio, breaks Sourcegraph search function by dropping proxied trailers for requests made over HTTP/1.1 protocol.

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/network/envoy
```

## Environment variables

To update the environment variables for the **sourcegraph-frontend** service, edit the [FRONTEND ENV VARS](index.md#frontend-env-vars) section at the bottom of your [kustomization file](index.md#kustomizationyaml). For example:

```yaml
# instances/$INSTANCE_NAME/patches/frontend-vars.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sourcegraph-frontend-env
data:
  DEPLOY_TYPE: kustomize
  NEW_ENV_VAR: NEW_VALUE
```

These values will be automatically merged with the environment variables currently listed in the ConfigMap for frontend.

> WARNING: You must restart frontend for the updated values to be activiated

## External services

You can use an external or managed version of PostgreSQL and Redis with your Sourcegraph instance. For detailed information as well as the requirements for each service, please see our docs on [using external services with Sourcegraph](../../../external_service/index.md).

### External Postgres

For optimal performance and resilience, it is recommended to use an external database when deploying Sourcegraph. For more information on database requirements, please refer to the [Postgres guide](../../postgres.md).

To connect Sourcegraph to an existing PostgreSQL instance, add the relevant environment variables ([such as PGHOST, PGPORT, PGUSER, etc.](http://www.postgresql.org/docs/current/static/libpq-envars.html)) to the frontend ConfigMap with the following steps:

**Step 1**: Copy the `frontend-vars.yaml` patch file from the `components/patches` directory to the [patches subdirectory](index.md#patches-directory) in your overlay

```bash
$ cp components/patches/frontend-vars.yaml instances/$INSTANCE_NAME/patches/frontend-vars.yaml
```

**Step 2**: Add environment variables at the end of the new patch file `frontend-vars.yaml`. For example:

```yaml
# instances/$INSTANCE_NAME/patches/frontend-vars.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sourcegraph-frontend-env
data:
  DEPLOY_TYPE: kustomize
  PGHOST: NEW_PGHOST
  PGPORT: NEW_PGPORT
```

**Step 3**: Include the patch file in your overlay under `patchesStrategicMerge`:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
patchesStrategicMerge:
  - patches/frontend-vars.yaml
```

> WARNING: You must restart frontend for the updated values to be activiated

### External Redis

Sourcegraph supports specifying an external Redis server with these environment variables:

- **REDIS_CACHE_ENDPOINT**=[redis-cache:6379](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24++REDIS_CACHE_ENDPOINT+AND+REDIS_STORE_ENDPOINT+-file:doc+file:internal&patternType=literal) for caching information.
- **REDIS_STORE_ENDPOINT**=[redis-store:6379](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24++REDIS_CACHE_ENDPOINT+AND+REDIS_STORE_ENDPOINT+-file:doc+file:internal&patternType=literal) for storing information (session data and job queues). 

When using an external Redis server, the corresponding environment variable must also be added to the following services:

<!-- Use ./dev/depgraph/depgraph summary internal/redispool to generate this -->

- `sourcegraph-frontend`
- `repo-updater`
- `gitserver`
- `searcher`
- `symbols`
- `worker`

**Step 1**: Include the `services/redis` component in your components:

This adds the new environment variables for redis to the services listed above.

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/services/redis
```

**Step 2**: Copy the `frontend-vars.yaml` patch file from the `components/patches` directory to the [patches subdirectory](index.md#patches-directory) in your overlay

```bash
$ cp components/patches/frontend-vars.yaml instances/$INSTANCE_NAME/patches/frontend-vars.yaml
```

**Step 3**: Add the additional annotations at the end of the new patch file `frontend-vars.yaml`. For example:

```yaml
# instances/$INSTANCE_NAME/patches/frontend-vars.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sourcegraph-frontend-env
data:
  DEPLOY_TYPE: kustomize
  REDIS_CACHE_ENDPOINT: REDIS_CACHE_DSN
  REDIS_STORE_ENDPOINT: REDIS_STORE_DSN
```

**Step 4**: Include the patch file in your overlay under `patchesStrategicMerge`:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
patchesStrategicMerge:
  - patches/frontend-vars.yaml
```

## SSH for cloning

Sourcegraph will clone repositories using SSH credentials when the `id_rsa` and `known_hosts` files are mounted at `/home/sourcegraph/.ssh` (or `/root/.ssh` when the cluster is run by root users) in the `gitserver` deployment. 

**WARNING:** Do not commit the actual `id_rsa` and `known_hosts` files to any public repository.

To mount the files through Kustomize:

**Step 1:** Copy the required files to the `configs` folder at the same level as your overylay's kustomization.yaml file

**Step 2:** Include the following in your overlay to [generate secrets](https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-kustomize/) that base64 encoded the values in those files

  ```yaml
  # instances/$INSTANCE_NAME/kustomization.yaml > [SECRETS GENERATOR]
  secretGenerator:
  - name: gitserver-ssh
    files:
    - configs/id_rsa
    - configs/known_hosts
  ```

**Step 3:** Include the following component to mount the [secret as a volume](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod) in [gitserver.StatefulSet.yaml](https://github.com/sourcegraph/deploy-sourcegraph-k8s/blob/master/base/gitserver/gitserver.StatefulSet.yaml).

  ```yaml
  # instances/$INSTANCE_NAME/kustomization.yaml
  components:
    # Enable SSH to clon repositories as non-root user
    - ../../components/enable/ssh/non-root
    # Enable SSH to clon repositories as root user
    - ../../components/enable/ssh/root
   ```

> NOTE: If you are running Sourcegraph with privileges using our [privileged component](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph-k8s/-/tree/components/privileged), you are most likely running Sourcegraph with `root` access and should use the [ssh/root component](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph-k8s/-/tree/components/enable/ssh/root) to enable SSH cloning.

**Step 4:** Update code host configuration

Update your [code host configuration file](../../../external_service/index.md#full-code-host-docs) to enable ssh cloning. For example, set [gitURLType](../../../../admin/external_service/github.md#gitURLType) to `ssh` for [GitHub](../../../external_service/github.md). See the [external service docs](../../../admin/external_service.md) for the correct setting for your code host.

## Openshift

### Arbitrary users

Our Postgres databases can only be run with the UIDs defined by the upstream images, for example, [UID 70 and GID 70](https://sourcegraph.com/github.com/docker-library/postgres@master/-/blob/12/alpine/Dockerfile?L9-15#L4)--this could cause permission issues for clusters that run pods with arbitrary users. This can be resolved with the `utils/uid` component created based on one of the solutions suggested by Postgres on their [official docker page](https://hub.docker.com/_/postgres). 

The `utils/uid` component bind-mount `/etc/passwd` as read-only through hostpath so that you can run the containers with valid users on your host.

```yaml
components:
- ../../components/utils/uid
```

## OpenTelemetry Collector

Learn more about Sourcegraph's integrations with the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) in our [OpenTelemetry documentation](../../observability/opentelemetry.md).

Sourcegraph currently supports exporting tracing data to several backends. Refer to [OpenTelemetry](../../observability/opentelemetry.md) for detailed descriptions on how to configure your backend of choice.

### Configure a tracing backend

By default, the collector is [configured to export trace data by logging](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/docker-images/opentelemetry-collector/configs/logging.yaml). Follow these steps to add a config for a different backend:

1. Create a subdirectory called 'patches' within the directory of your overlay
2. Copy and paste the [base/otel-collector/otel-collector.ConfigMap.yaml file](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph-k8s@master/-/tree/base/otel-collector/otel-collector.ConfigMap.yaml) to the new [patches subdirectory](index.md#patches-directory)
3. In the copied file, make the necessary changes to the `exporters` and `service` blocks to connect to your backend based on the documentation linked above
4. Include the following in your overlay:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/otel-collector/backend
...
patchesStrategicMerge:
- patches/otel-collector.ConfigMap.yaml
```

The component will update the `command` for the `otel-collector` container to `"--config=/etc/otel-collector/conf/config.yaml"`, which is now point to the mounted config.

## Add license key

Sourcegraph's Kubernetes deployment [requires an Enterprise license key](https://about.sourcegraph.com/pricing).

Once you have a license key, add it to your [site configuration](https://docs.sourcegraph.com/admin/patches/site_config).

## Filtering cAdvisor metrics

cAdvisor can pick up metrics for services unrelated to the Sourcegraph deployment running on the same nodes
([Learn more](../../../dev/background-information/observability/cadvisor.md#identifying-containers)). To work around this:

1. Create a subdirectory called 'patches' within the directory of your overlay
2. Copy and paste the `base/prometheus/prometheus.ConfigMap.yaml` file to the new [patches subdirectory](index.md#patches-directory)
2. In the copied file, include the lines highlighted [here](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph@v4.3.1/-/blob/base/prometheus/prometheus.ConfigMap.yaml?L262-264).
3. Replace [ns-sourcegraph](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph@v4.3.1/-/blob/base/prometheus/prometheus.ConfigMap.yaml?L263) with your namespace
4. Include the following in your overlay:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
patchesStrategicMerge:
- patches/prometheus.ConfigMap.yaml
```

This will cause Prometheus to drop all metrics *from cAdvisor* that are not from services in the desired namespace.

## Private registry

To update all image names with your private registry, eg. `index.docker.io/sourcegraph/service_name` to `your.private.registry.com/sourcegraph/service_name`, include the `private-registry` component:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/enable/private-registry
```

Set the `PRIVATE_REGISTRY` variable under the [BUILD CONFIGURATIONS](index.md#build-configurations) section:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml > [BUILD CONFIGURATIONS]
configMapGenerator:
  # Handle updating configs using env vars for kustomize
  - name: sourcegraph-kustomize-env
    behavior: merge
    literals:
      ...
      - PRIVATE_REGISTRY=your.private.registry.com
```

## Multi-version upgrade

In order to perform a [multi-version upgrade](../../../updates/index.md#multi-version-upgrades), all pods must be scaled down to 0 except databases, which can be handled by including the `utils/multi-version-upgrade` component:

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/utils/multi-version-upgrade
```

After completing the multi-version-upgrade process, exclude the component to allow the pods to scale back to their original number as defined in your overlay.

## Migrate from Privileged to Non-privileged

To migrate an existing deployment from root to non-root environment, you must first transfer ownership of all data directories to the specified non-root users for each service using the utils/migrate-to-nonprivileged component.

After transferring ownerships, you can redeploy the instance with non-privileged configurations.

```yaml
# instances/$INSTANCE_NAME/kustomization.yaml
components:
- ../../components/utils/migrate-to-nonprivileged
```

## Outbound Traffic

When working with an [Internet Gateway](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_Internet_Gateway.html) or VPC it may be necessary to expose ports for outbound network traffic. Sourcegraph must open port 443 for outbound traffic to codehosts, and to enable [telemetry](https://docs.sourcegraph.com/admin/pings) with Sourcegraph.com. Port 22 must also be opened to enable git SSH cloning by Sourcegraph. In addition, please make sure to apply other required changes to secure your cluster in a manner that meets your organization's security requirements.

## Troubleshooting

See the [Troubleshooting docs](../troubleshoot.md).
