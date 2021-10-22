# Configure Sourcegraph with Kubernetes

Configuring a [Sourcegraph Kubernetes cluster](./index.md) is done by applying manifest files and with simple
`kubectl` commands. You can configure Sourcegraph as flexibly as you need to meet the requirements
of your deployment environment.

## Featured guides

<div class="getting-started">
  <a href="#getting-started" class="btn btn-primary" alt="Configure">
   <span>Getting started</span>
   </br>
   Get started with configuring Sourcegraph with Kubernetes.
  </a>

  <a href="#overlays" class="btn" alt="Overlays">
   <span>Overlays</span>
   </br>
   Learn about Kustomize, how to use our provided overlays, and how to create your own.
  </a>

  <!-- <a href="#configure-external-databases" class="btn" alt="Configure external databases">
   <span>External databases</span>
   </br>
   Learn about setting up an external database for Sourcegraph with Kubernetes.
  </a> -->
</div>

## Getting started

We **strongly** recommend you fork the [Sourcegraph with Kubernetes reference repository](./index.md#reference-repository) to track your configuration changes in Git.
**This will make upgrades far easier** and is a good practice not just for Sourcegraph, but for any Kubernetes application.

- Create a fork of the [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) repository.

    > WARNING: Set your fork to **private** if you plan to store secrets (SSL certificates, external Postgres credentials, etc.) within the repository.

    <span class="virtual-br"></span>

    > NOTE: We do not recommend storing secrets in the repository itself - instead, consider leveraging [Kubernetes's Secret objects](https://kubernetes.io/docs/concepts/configuration/secret).

- Clone your fork using the repository's URL.

    > NOTE: The `docker-compose.yaml` file currently depends on configuration files which live in the repository, so you must have the entire repository cloned onto your server.

  ```bash
  git clone $FORK_URL
  ```

- Add the [reference repository](./index.md#reference-repository) as an `upstream` remote so that you can [get updates](update.md).

  ```bash
  git remote add upstream https://github.com/sourcegraph/deploy-sourcegraph
  ```

- Create a `release` branch to track all of your customizations to Sourcegraph. This branch will be used to [upgrade Sourcegraph](update.md) and [install your Sourcegraph instance](./index.md#installation).

  ```bash
  export SOURCEGRAPH_VERSION="v3.33.0"
  git checkout $SOURCEGRAPH_VERSION -b release
  ```

Some of the following instructions require cluster access. Ensure you can [access your Kubernetes cluster](https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/) with `kubectl`.

### Customizations

To make customizations to the Sourcegraph deployment such as resources, replicas or other changes, we recommend using [Kustomize](./index.md#kustomize) and [overlays](./index.md#overlays).
This means that you define your customizations as patches, and generate a manifest from our provided manifests to [apply](./operations.md#applying-manifests).

In general, we recommend that customizations work like this:

1. [Create, customize, and apply overlays](#overlays) for your deployment
2. Ensure the services came up correctly, then commit all the customizations to the new branch

  ```sh
  git add /overlays/$MY_OVERLAY/*
  # Keeping all overlays contained to a single commit allows for easier cherry-picking
  git commit amend -m "overlays: update $MY_OVERLAY"
  ```

See the [overlays guide](#overlays) to learn about the [overlays we provide](#provided-overlays) and [how to create your own overlays](#custom-overlays).

### Applying manifests

To deploy your configuration changes, [apply your Kubernetes manifests](./operations.md#applying-manifests).

## Overlays

Kustomize overlays are our recommended way to [customize Sourcegraph with Kubernetes](#customization).

> NOTE: If you have not worked with [Kustomize](./index.md#kustomize) or overlays before, please refer to our [Kustomize introduction](./index.md#kustomize).

To generate Kubernetes manifests from an overlay, run the `overlay-generate-cluster.sh` with two arguments:

- the name of the overlay
- and a path to an output directory where the generated manifests will be

For example:

```sh
#                overlay directory name    output directory
#                                 |             |
./overlay-generate-cluster.sh my-overlay generated-cluster
```

After executing the script you can apply the generated manifests from the `generated-cluster` directory:

```sh
kubectl apply --prune -l deploy=sourcegraph -f generated-cluster --recursive
```

We recommend that you:

- [Update the `./overlay-generate-cluster` script](./operations.md#applying-manifests) to apply the generated manifests from the `generated-cluster` directory with something like the above snippet
- Commit your overlays changes separately - see our [customization guide](#customizations) for more details.

You can now get started with using overlays:

- [Provided overlays](#provided-overlays)
- [Custom overlays](#custom-overlays)

### Provided overlays

Overlays provided out-of-the-box are in the subdirectories of [`deploy-sourcegraph/overlays`](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/overlays) and are documented here.

#### Namespaced overlay

This overlay adds a namespace declaration to all the manifests.

1. Change the namespace by replacing `ns-sourcegraph` to the name of your choice everywhere within the
[overlays/namespaced/](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/overlays/namespaced/) directory. 

1. Generate the overlay by running this command from the `root` directory:

    ```
    ./overlay-generate-cluster.sh namespaced generated-cluster
    ```

1. Create the namespace if it doesn't exist yet:

    ```
    kubectl create namespace ns-<EXAMPLE NAMESPACE>
    kubectl label namespace ns-<EXAMPLE NAMESPACE> name=ns-sourcegraph
    ```

1. Apply the generated manifests (from the `generated-cluster` directory) by running this command from the `root` directory:

  ```
  kubectl apply -n ns-<EXAMPLE NAMESPACE> --prune -l deploy=sourcegraph -f generated-cluster --recursive
  ```

1. Check for the namespaces and their status with:

  ```
  kubectl get pods -A
  ```

#### Non-privileged create cluster overlay

This kustomization is for Sourcegraph installations in clusters with security restrictions. It runs all containers as a non root users, as well removing cluster roles and cluster role bindings and does all the rolebinding in a namespace. It configures Prometheus to work in the namespace and not require ClusterRole wide privileges when doing service discovery for scraping targets. It also disables cAdvisor.

This version and `non-privileged` need to stay in sync. This version is only used for cluster creation.

To use it, execute the following command from the `root` directory:

```
./overlay-generate-cluster.sh non-privileged-create-cluster generated-cluster
```

After executing the script you can apply the generated manifests from the generated-cluster directory:

```
kubectl create namespace ns-sourcegraph
kubectl apply -n ns-sourcegraph --prune -l deploy=sourcegraph -f generated-cluster --recursive
```

#### Non-privileged overlay

This overlay is for continued use after you have successfully deployed the `non-privileged-create-cluster`. It runs all containers as a non root users, as well removing cluster roles and cluster role bindings and does all the rolebinding in a namespace. It configures Prometheus to work in the namespace and not require ClusterRole wide privileges when doing service discovery for scraping targets. It also disables cAdvisor.

To use it, execute the following command from the `root` directory:

```shell script
./overlay-generate-cluster.sh non-privileged generated-cluster
```

After executing the script you can apply the generated manifests from the generated-cluster directory:

```shell script
kubectl apply -n ns-sourcegraph --prune -l deploy=sourcegraph -f generated-cluster --recursive
```

If you are starting a fresh installation use the overlay `non-privileged-create-cluster`. After creation you can use the overlay
`non-privileged`.

#### Migrate-to-nonprivileged overlay

If you already are running a Sourcegraph instance using user `root` and want to convert to running with non-root user then
you need to apply a migration step that will change the permissions of all persistent volumes so that the volumes can be
used by the non-root user. This migration is provided as overlay `migrate-to-nonprivileged`. After the migration you can use
overlay `non-privileged`. If you have previously deployed your cluster in a non-default namespace, be sure to edit the `kustomization.yaml` file in the overlays directly to ensure the files are generated with the correct namespace. 

This kustomization injects initContainers in all pods with persistent volumes to transfer ownership of directories to specified non-root users. It is used for migrating existing installations to a non-privileged environment.

```
./overlay-generate-cluster.sh migrate-to-nonprivileged generated-cluster
```

After executing the script you can apply the generated manifests from the generated-cluster directory:

```
kubectl apply --prune -l deploy=sourcegraph -f generated-cluster --recursive
```

#### minikube overlay

This kustomization deletes resource declarations and storage classnames to enable running Sourcegraph on minikube.

To use it, execute the following command from the `root` directory:

```sh
./overlay-generate-cluster.sh minikube generated-cluster
```

After executing the script you can apply the generated manifests from the generated-cluster directory:

```sh
minikube start
kubectl create namespace ns-sourcegraph
kubectl -n ns-sourcegraph apply --prune -l deploy=sourcegraph -f generated-cluster --recursive
kubectl -n ns-sourcegraph expose deployment sourcegraph-frontend --type=NodePort --name sourcegraph --port=3080 --target-port=3080
minikube service list
```

To tear it down:

```sh
kubectl delete namespaces ns-sourcegraph
minikube stop
```

### Custom overlays

To create your own [overlays](#overlays), first [set up your deployment reference repository to enable customizations](#getting-started).

Then, within the `overlays` directory of the [reference repository](./index.md#reference-repository), create a new directory for your overlay along with a `kustomization.yaml`.

```text
deploy-sourcegraph
 |-- overlays
 |    |-- my-new-overlay
 |    |    +-- kustomization.yaml
 |    |-- bases
 |    +-- ...
 +-- ...
```

Within `kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
# Only include resources from 'overlays/bases' you are interested in modifying
# To learn more about bases: https://kubectl.docs.kubernetes.io/references/kustomize/glossary/#base
resources:
  - ../bases/deployments
  - ../bases/rbac-roles
  - ../bases/pvcs
```

You can then define patches, transformations, and more. A complete reference is available [here](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/).
To get started, we recommend you explore writing your own [`patches`](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/patches/), or the more specific variants:

- [`patchesStrategicMerge`](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/patchesstrategicmerge/)
- [`patchesJson6902`](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/patchesjson6902/)

To avoid complications with reference cycles an overlay can only reference resources inside the directory subtree of the directory it resides in (symlinks are not allowed either).

Learn more in the [`kustomization` documentation](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/).
You can also explore how our [provided overlays](#provided-overlays) use patches, for reference: [`deploy-sourcegraph` usage of patches](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/deploy-sourcegraph%24+lang:YAML+patches:+:%5B_%5D+OR+patchesStrategicMerge:+:%5B_%5D+OR+patchesJson6902:+:%5B_%5D+count:999&patternType=structural).

Once you have created your overlays, refer to our [overlays guide](#overlays).

<br />

## Configure a storage class

Sourcegraph by default requires a storage class for all persisent volumes claims. By default this storage class is called `sourcegraph`. This storage class must be configured before applying the base configuration to your cluster.

- Create `base/sourcegraph.StorageClass.yaml` with the appropriate configuration for your cloud provider and commit the file to your fork.

- The sourcegraph StorageClass will retain any persistent volumes created in the event of an accidental deletion of a persistent volume claim.

- The sourcegraph StorageClass also allows the persistent volumes to expand their storage capacity by increasing the
 size of the related persistent volume claim.

- This cannot be changed once the storage class has been created. Persistent volumes not created with the reclaimPolicy set to `Retain` can be patched with the following command:

```bash
kubectl patch pv <your-pv-name> -p '{"spec":{"persistentVolumeReclaimPolicy":"Retain"}}'
```

See [the official documentation](https://kubernetes.io/docs/tasks/administer-cluster/change-pv-reclaim-policy/#changing-the-reclaim-policy-of-a-persistentvolume) for more information about patching persistent volumes.

### Google Cloud Platform (GCP)

#### Kubernetes 1.18 and higher

1. Please read and follow the [official documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver) for enabling the persistent disk CSI driver on a [new](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver#enabling_the_on_a_new_cluster) or [existing](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver#enabling_the_on_an_existing_cluster) cluster.


2. Add the following Kubernetes manifest to the `base` directory of your fork:

```yaml
# base/sourcegraph.StorageClass.yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: sourcegraph
  labels:
    deploy: sourcegraph
provisioner: pd.csi.storage.gke.io
parameters:
  type: pd-ssd # This configures SSDs (recommended).
reclaimPolicy: Retain
allowVolumeExpansion: true
volumeBindingMode: WaitForFirstConsumer
```

[Additional documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver).

#### Kubernetes 1.17 and below

```yaml
# base/sourcegraph.StorageClass.yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: sourcegraph
  labels:
    deploy: sourcegraph
provisioner: kubernetes.io/gce-pd
parameters:
  type: pd-ssd # This configures SSDs (recommended).
reclaimPolicy: Retain
allowVolumeExpansion: true
```

[Additional documentation](https://kubernetes.io/docs/concepts/storage/storage-classes/#gce-pd).

### Amazon Web Services (AWS)

#### Kubernetes 1.17 and higher

1. Follow the [official instructions](https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html) to deploy the Amazon Elastic Block Store (Amazon EBS) Container Storage Interface (CSI) driver.

1. Add the following Kubernetes manifest to the `base` directory of your fork:


```yaml
# base/sourcegraph.StorageClass.yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: sourcegraph
  labels:
    deploy: sourcegraph
provisioner: ebs.csi.aws.com
parameters:
  type: gp2 # This configures SSDs (recommended).
reclaimPolicy: Retain
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
```

[Additional documentation](https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html).

#### Kubernetes 1.16 and below


```yaml
# base/sourcegraph.StorageClass.yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: sourcegraph
  labels:
    deploy: sourcegraph
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp2 # This configures SSDs (recommended).
reclaimPolicy: Retain
allowVolumeExpansion: true
```

[Additional documentation](https://kubernetes.io/docs/concepts/storage/storage-classes/#aws-ebs).

### Azure

#### Kubernetes 1.18 and higher

> WARNING: If you are deploying on Azure, you **must** ensure that your cluster is created with support for CSI storage drivers [(link)](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers)). This **can not** be enabled after the fact

1. Follow the [official instructions](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers) to deploy the Amazon Elastic Block Store (Amazon EBS) Container Storage Interface (CSI) driver.

2. Add the following Kubernetes manifest to the `base` directory of your fork:


```yaml
# base/sourcegraph.StorageClass.yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: sourcegraph
  labels:
    deploy: sourcegraph
provisioner: disk.csi.azure.com
parameters:
  storageaccounttype: Premium_LRS # This configures SSDs (recommended). A Premium VM is required.
reclaimPolicy: Retain
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
```


[Additional documentation](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers).

#### Kubernetes 1.17 and below


```yaml
# base/sourcegraph.StorageClass.yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: sourcegraph
  labels:
    deploy: sourcegraph
provisioner: kubernetes.io/azure-disk
parameters:
  storageaccounttype: Premium_LRS # This configures SSDs (recommended). A Premium VM is required.
reclaimPolicy: Retain
allowVolumeExpansion: true
```

[Additional documentation](https://kubernetes.io/docs/concepts/storage/storage-classes/#azure-disk).

### Other cloud providers

```yaml
# base/sourcegraph.StorageClass.yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: sourcegraph
  labels:
    deploy: sourcegraph
reclaimPolicy: Retain
allowVolumeExpansion: true
# Read https://kubernetes.io/docs/concepts/storage/storage-classes/ to configure the "provisioner" and "parameters" fields for your cloud provider.
# SSDs are highly recommended!
# provisioner:
# parameters:
```

### Using a storage class with an alternate name

If you wish to use a different storage class for Sourcegraph, then you need to update all persistent volume claims with the name of the desired storage class. Convenience script:

```bash
#!/usr/bin/env bash

# This script requires https://github.com/mikefarah/yq v4 or greater

# Set SC to your storage class name
SC=

PVC=()
STS=()
mapfile -t PVC < <(fd --absolute-path --extension yaml "PersistentVolumeClaim" base)
mapfile -t STS < <(fd --absolute-path --extension yaml "StatefulSet" base)

for p in "${PVC[@]}"; do yq eval -i ".spec.storageClassName|=\"$SC\"" "$p"; done
for s in "${STS[@]}"; do yq eval -i ".spec.volumeClaimTemplates.[].spec.storageClassName|=\"$SC\"" "$s"; done
```

## Configure network access

You need to make the main web server accessible over the network to external users.

There are a few approaches, but using an ingress controller is recommended.

### Ingress controller (recommended)

For production environments, we recommend using the [ingress-nginx](https://kubernetes.github.io/ingress-nginx/) [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/).

- As part of our base configuration, we install an ingress for [sourcegraph-frontend](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/frontend/sourcegraph-frontend.Ingress.yaml). It installs rules for the default ingress, see comments to restrict it to a specific host.

- In addition to the sourcegraph-frontend ingress, you'll need to install the NGINX ingress controller (ingress-nginx).

- Follow the instructions at https://kubernetes.github.io/ingress-nginx/deploy/ to create the ingress controller.

- Add the files to [configure/ingress-nginx](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/configure/ingress-nginx), including an [install.sh](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/configure/ingress-nginx/install.sh) file which applies the relevant manifests.

- We include sample generic-cloud manifests as part of this repository, but please follow the official instructions for your cloud provider.

- Add the [configure/ingress-nginx/install.sh](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/configure/ingress-nginx/install.sh) command to [create-new-cluster.sh](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/create-new-cluster.sh) and commit the change:

```shell
echo ./configure/ingress-nginx/install.sh >> create-new-cluster.sh
```

- Once the ingress has acquired an external address, you should be able to access Sourcegraph using that.

- You can check the external address by running the following command and looking for the `LoadBalancer` entry:

```bash
kubectl -n ingress-nginx get svc
```

If you are having trouble accessing Sourcegraph, ensure ingress-nginx IP is accessible above. Otherwise see [Troubleshooting ingress-nginx](https://kubernetes.github.io/ingress-nginx/troubleshooting/). The namespace of the ingress-controller is `ingress-nginx`.

Once you have [installed Sourcegraph](./index.md#installation), run the following command, and ensure an IP address has been assigned to your ingress resource. Then browse to the IP or configured URL.

```sh
kubectl get ingress sourcegraph-frontend

NAME                   CLASS    HOSTS             ADDRESS     PORTS     AGE
sourcegraph-frontend   <none>   sourcegraph.com   8.8.8.8     80, 443   1d
```

#### Configuration

`ingress-nginx` has extensive configuration documented at [NGINX Configuration](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/). We expect most administrators to modify [ingress-nginx annotations](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/) in [sourcegraph-frontend.Ingress.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/frontend/sourcegraph-frontend.Ingress.yaml). Some settings are modified globally (such as HSTS). In that case we expect administrators to modify the [ingress-nginx configmap](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/) in [configure/ingress-nginx/mandatory.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/configure/ingress-nginx/mandatory.yaml).

### NGINX service

In cases where ingress controllers cannot be created, creating an explicit NGINX service is a viable
alternative. See the files in the [configure/nginx-svc](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/configure/nginx-svc) folder for an
example of how to do this via a NodePort service (any other type of Kubernetes service will also
work):

- Modify [configure/nginx-svc/nginx.ConfigMap.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/configure/nginx-svc/nginx.ConfigMap.yaml) to
   contain the TLS certificate and key for your domain.

- `kubectl apply -f configure/nginx-svc` to create the NGINX service.

- Update [create-new-cluster.sh](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/create-new-cluster.sh) with the previous command.

   ```
   echo kubectl apply -f configure/nginx-svc >> create-new-cluster.sh
   ```

### Network rule

> NOTE: this setup path does not support TLS.

Add a network rule that allows ingress traffic to port 30080 (HTTP) on at least one node.

#### [Google Cloud Platform Firewall rules](https://cloud.google.com/compute/docs/vpc/using-firewalls).

- Expose the necessary ports.

```bash
gcloud compute --project=$PROJECT firewall-rules create sourcegraph-frontend-http --direction=INGRESS --priority=1000 --network=default --action=ALLOW --rules=tcp:30080
```

- Change the type of the `sourcegraph-frontend` service in [base/frontend/sourcegraph-frontend.Service.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/frontend/sourcegraph-frontend.Service.yaml) from `ClusterIP` to `NodePort`:

```diff
spec:
  ports:
  - name: http
    port: 30080
+    nodePort: 30080
-  type: ClusterIP
+  type: NodePort
```

- Directly applying this change to the service [will fail](https://github.com/kubernetes/kubernetes/issues/42282). Instead, you must delete the old service and then create the new one (this will result in a few seconds of downtime):

```shell
kubectl delete svc sourcegraph-frontend
kubectl apply -f base/frontend/sourcegraph-frontend.Service.yaml
```

- Find a node name.

```bash
kubectl get pods -l app=sourcegraph-frontend -o=custom-columns=NODE:.spec.nodeName
```

- Get the EXTERNAL-IP address (will be ephemeral unless you [make it static](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address#promote_ephemeral_ip)).

```bash
kubectl get node $NODE -o wide
```

#### [AWS Security Group rules](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_SecurityGroups.html).

Sourcegraph should now be accessible at `$EXTERNAL_ADDR:30080`, where `$EXTERNAL_ADDR` is the address of _any_ node in the cluster.

### Using NetworkPolicy

Network policy is a Kubernetes resource that defines how pods are allowed to communicate with each other and with
other network endpoints. If the cluster administration requires an associated NetworkPolicy when doing an installation,
then we recommend running Sourcegraph in a namespace (as described in our [Overlays guide](#overlays) or below in the
[Using NetworkPolicy with Namespaced Overlay Example](#using-networkpolicy-with-namespaced-overlay)).
You can then use the `namespaceSelector` to allow traffic between the Sourcegraph pods.
When you create the namespace you need to give it a label so it can be used in a `matchLabels` clause.

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ns-sourcegraph
  labels:
    name: ns-sourcegraph
```

If the namespace already exists you can still label it like so

```shell script
kubectl label namespace ns-sourcegraph name=ns-sourcegraph
```

> NOTE: You will need to augment this example NetworkPolicy to allow traffic to external services
> you plan to use (like github.com) and ingress traffic from
> the outside to the frontend for the users of the Sourcegraph installation.
> Check out this [collection](https://github.com/ahmetb/kubernetes-network-policy-recipes) of NetworkPolicies to get started.

```yaml
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: np-sourcegraph
  namespace: ns-sourcegraph
spec:
  # For all pods with the label "deploy: sourcegraph"
  podSelector:
    matchLabels:
      deploy: sourcegraph
  policyTypes:
  - Ingress
  - Egress
  # Allow all traffic inside the ns-sourcegraph namespace
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ns-sourcegraph
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: ns-sourcegraph
```

## Configure external databases

We recommend utilizing an external database when deploying Sourcegraph to provide the most resilient and performant backend for your deployment. For more information on the specific requirements for Sourcegraph databases, see [this guide](../../postgres.md).

Simply edit the relevant PostgreSQL environment variables (e.g. PGHOST, PGPORT, PGUSER, [etc.](http://www.postgresql.org/docs/current/static/libpq-envars.html)) in [base/frontend/sourcegraph-frontend.Deployment.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/frontend/sourcegraph-frontend.Deployment.yaml) to point to your existing PostgreSQL instance.

If you do not have an external database available, configuration is provided to deploy PostgreSQL on Kubernetes. 


## Configure repository cloning via SSH

Sourcegraph will clone repositories using SSH credentials if they are mounted at `/home/sourcegraph/.ssh` in the `gitserver` deployment.

[Create a secret](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-environment-variables) that contains the base64 encoded contents of your SSH private key (_make sure it doesn't require a password_) and known_hosts file.

   ```bash
   kubectl create secret generic gitserver-ssh \
    --from-file id_rsa=${HOME}/.ssh/id_rsa \
    --from-file known_hosts=${HOME}/.ssh/known_hosts
   ```

Update [create-new-cluster.sh](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/create-new-cluster.sh) with the previous command.

   ```bash
   echo kubectl create secret generic gitserver-ssh \
    --from-file id_rsa=${HOME}/.ssh/id_rsa \
    --from-file known_hosts=${HOME}/.ssh/known_hosts >> create-new-cluster.sh
   ```

Mount the [secret as a volume](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod) in [gitserver.StatefulSet.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/gitserver/gitserver.StatefulSet.yaml).

   For example:

   ```yaml
   # base/gitserver/gitserver.StatefulSet.yaml
   spec:
     containers:
       volumeMounts:
         - mountPath: /root/.ssh
           name: ssh
     volumes:
       - name: ssh
         secret:
           defaultMode: 0644
           secretName: gitserver-ssh
   ```

   Convenience script:

   ```bash
   # This script requires https://github.com/sourcegraph/jy and https://github.com/sourcegraph/yj
   GS=base/gitserver/gitserver.StatefulSet.yaml
   cat $GS | yj | jq '.spec.template.spec.containers[].volumeMounts += [{mountPath: "/root/.ssh", name: "ssh"}]' | jy -o $GS
   cat $GS | yj | jq '.spec.template.spec.volumes += [{name: "ssh", secret: {defaultMode: 384, secretName:"gitserver-ssh"}}]' | jy -o $GS
   ```

   If you run your installation with non-root users (the non-root overlay) then use the mount path `/home/sourcegraph/.ssh` instead of `/root/.ssh`:

   ```yaml
   # base/gitserver/gitserver.StatefulSet.yaml
   spec:
     containers:
       volumeMounts:
         - mountPath: /home/sourcegraph/.ssh
           name: ssh
     volumes:
       - name: ssh
         secret:
           defaultMode: 0644
           secretName: gitserver-ssh
   ```

   Convenience script:

   ```bash
   # This script requires https://github.com/sourcegraph/jy and https://github.com/sourcegraph/yj
   GS=base/gitserver/gitserver.StatefulSet.yaml
   cat $GS | yj | jq '.spec.template.spec.containers[].volumeMounts += [{mountPath: "/home/sourcegraph/.ssh", name: "ssh"}]' | jy -o $GS
   cat $GS | yj | jq '.spec.template.spec.volumes += [{name: "ssh", secret: {defaultMode: 384, secretName:"gitserver-ssh"}}]' | jy -o $GS
   ```


3. Apply the updated `gitserver` configuration to your cluster.

  ```bash
  ./kubectl-apply-all.sh
  ```

**WARNING:** Do NOT commit the actual `id_rsa` and `known_hosts` files to your fork (unless
your fork is private **and** you are okay with storing secrets in it).

## Configure custom Redis

Sourcegraph supports specifying a custom Redis server for:

- caching information (specified via the `REDIS_CACHE_ENDPOINT` environment variable)
- storing information (session data and job queues) (specified via the `REDIS_STORE_ENDPOINT` environment variable)

If you want to specify a custom Redis server, you'll need specify the corresponding environment variable for each of the following deployments:

- `sourcegraph-frontend`
- `repo-updater`

## Connect to an external Jaeger instance

If you have an existing Jaeger instance you would like to connect Sourcegraph to (instead of running the Jaeger instance inside the Sourcegraph cluster), do:

1. Remove the `base/jaeger` directory: `rm -rf base/jaeger`
1. Update the Jaeger agent containers to point to your Jaeger collector.
   1. Find all instances of Jaeger agent (`grep -R 'jaegertracing/jaeger-agent'`).
   1. Update the `args` field of the Jaeger agent container configuration to point to the external
      collector. E.g.,
      ```
      args:
        - --reporter.grpc.host-port=external-jaeger-collector-host:14250
        - --reporter.type=grpc
      ```
1. Apply these changes to the cluster.

### Disable Jaeger entirely

To disable Jaeger entirely, do:

1. Update the Sourcegraph [site
   configuration](https://docs.sourcegraph.com/admin/config/site_config) to remove the
   `observability.tracing` field.
1. Remove the `base/jaeger` directory: `rm -rf base/jaeger`
1. Remove the jaeger agent containers from each `*.Deployment.yaml` and `*.StatefulSet.yaml` file.
1. Apply these changes to the cluster.

## Install without cluster-wide RBAC

Sourcegraph communicates with the Kubernetes API for service discovery. It also has some janitor DaemonSets that clean up temporary cache data. To do that we need to create RBAC resources.

If using cluster roles and cluster rolebinding RBAC is not an option, then you can use the [non-privileged](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/overlays/non-privileged) overlay to generate modified manifests. Read the [Overlays](#overlays) section below about overlays.
## Add license key

Sourcegraph's Kubernetes deployment [requires an Enterprise license key](https://about.sourcegraph.com/pricing).

- Create an account on or sign in to sourcegraph.com, and go to https://sourcegraph.com/subscriptions/new to obtain a license key.

- Once you have a license key, add it to your [site configuration](https://docs.sourcegraph.com/admin/config/site_config).

## Environment variables

Update the environment variables in the appropriate deployment manifest.
For example, the following [patch](#overlays) will update `PGUSER` to have the value `bob`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sourcegraph-frontend
spec:
  template:
    spec:
      containers:
        - name: frontend
          env:
            - name: PGUSER
              value: bob
```

## Filtering cAdvisor metrics

Due to how cAdvisor works, Sourcegraph's cAdvisor deployment can pick up metrics for services unrelated to the Sourcegraph deployment running on the same nodes as Sourcegraph services.
[Learn more](../../../dev/background-information/observability/cadvisor.md#identifying-containers).

To work around this, update your `prometheus.ConfigMap.yaml` to target your [namespaced Sourcegraph deployment](#namespaced-overlay) by uncommenting the below `metric_relabel_configs` entry and updating it with the appropriate namespace.
This will cause Prometheus to drop all metrics *from cAdvisor* that are not from services in the desired namespace.

```yaml
apiVersion: v1
data:
  prometheus.yml: |
    # ...

      metric_relabel_configs:
      # cAdvisor-specific customization. Drop container metrics exported by cAdvisor
      # not in the same namespace as Sourcegraph.
      # Uncomment this if you have problems with certain dashboards or cAdvisor itself
      # picking up non-Sourcegraph services. Ensure all Sourcegraph services are running
      # within the Sourcegraph namespace you have defined.
      # The regex must keep matches on '^$' (empty string) to ensure other metrics do not
      # get dropped.
      - source_labels: [container_label_io_kubernetes_pod_namespace]
        regex: ^$|ns-sourcegraph # ensure this matches with namespace declarations
        action: keep

    # ...
```

## Outbound Traffic

When working with an [Internet Gateway](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_Internet_Gateway.html) or VPC it may be necessary to expose ports for outbound network traffic. Sourcegraph must open port 443 for outbound traffic to codehosts, and to enable [telemetry](https://docs.sourcegraph.com/admin/pings) with Sourcegraph.com. Port 22 must also be opened to enable git SSH cloning by Sourcegraph. Take care to secure your cluster in a manner that meets your organization's security requirements.

## Troubleshooting

See the [Troubleshooting docs](troubleshoot.md).
