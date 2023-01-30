# Sourcegraph with Kubernetes

<p class="lead">
Deploying Sourcegraph on Kubernetes is for organizations that need highly scalable and available code search and code navigation. We recommend deploying Sourcegraph on Kubernetes with Kustomize.
</p>

<div class="getting-started">
  <a href="./kustomize" class="btn btn-primary" alt="Configure">
   <span>★ Kustomize</span>
   </br>
   Deploy Sourcegraph with simple kubectl commands
  </a>
  <a href="./helm" class="btn" alt="instance">
   <span>Helm</span>
   </br>
   Deploy Sourcegraph with Helm
  </a>
</div>

<div class="getting-started">
<a class="btn btn-primary text-center" href="#prerequisites">★ Installation</a>
<a class="btn text-center" href="kustomize/configure">Configuration</a>
<a class="btn text-center" href="../instance-size">Instance Sizes</a>
<a class="btn text-center" href="operations">Operations</a>
</div>

> WARNING: If you are currently on Sourcegraph version 4.5.0 or below, please refer to the [deprecated deployment docs for Kubernetes](../deprecated/index.md).

### Prerequisites

* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) v1.19 or later
   - [Kustomize](https://kustomize.io/) (built into `kubectl` in version >= 1.14)
* A [Kubernetes](https://kubernetes.io/) cluster
   - Minimum Kubernetes version: [v1.19](https://kubernetes.io/blog/2020/08/26/kubernetes-release-1.19-accentuate-the-paw-sitive/)
   - Support for Persistent Volumes (SSDs recommended)

> NOTE: Alternatively, you can use our [terraform configs](https://github.com/sourcegraph/tf-k8s-configs) to quickly setup a cluster that will support a standard Sourcegraph instance on [Google Cloud Platform (GKE)](https://github.com/sourcegraph/tf-k8s-configs/tree/main/gcp),  [Amazon Web Services (EKS)](https://github.com/sourcegraph/tf-k8s-configs/tree/main/aws), or [Azure AKS](https://github.com/sourcegraph/tf-k8s-configs/tree/main/azure).

---

### **Step 1**: Setup the reference repository

Clone [Sourcegraph reference repository for Kubernetes](https://github.com/sourcegraph/deploy-sourcegraph-k8s) for config files and components needed to deploy Sourcegraph to a Kubernetes cluster using Kustomize.

```bash
$ git clone https://github.com/sourcegraph/deploy-sourcegraph-k8s.git
```

>NOTE: Please refer to the [reference repository docs](../repositories.md) on setting up a private copy of the reference repository.

### **Step 2**: Setup an overlay

From the root of your repository clone, create a new directory for `$INSTANCE_NAME`, a specific Sourcegraph instance (e.g. `instances/dev`) using the branch for the latest version.

```bash
$ export INSTANCE_NAME=dev
$ git checkout v4.4.1
$ git checkout -b $INSTANCE_NAME
$ mkdir instance/$INSTANCE_NAME
```

### **Step 3**: Setup a configuration file

Copy [instances/kustomization.template.yaml](intro.md#template) to `instances/$INSTANCE_NAME` as `kustomization.yaml` for deployment config.

```bash
# The new kustomization.yaml file will be used to configure your deployment.
$ cp instances/kustomization.template.yaml instances/$INSTANCE_NAME/kustomization.yaml
```

### **Step 4**: Configure namespace

You can specify the namespace for the Sourcegraph deployment in the `kustomization.yaml` file. The `namespace` field can be set to match an existing namespace in the cluster.

You can also create a new namespace ([cluster role administrator access required](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)) by uncommenting the `resources/namespace` component under the components list:

  ```yaml
  # instances/dev/kustomization.yaml
  apiVersion: kustomize.config.k8s.io/v1beta1
  kind: Kustomization
  resources:
    # Resources for Sourcegraph base cluster with default settings
    - ../../base/sourcegraph
  # -- Set namespace to namespace values for all resources
  namespace: sg-dev # -- [ACTION] Set namespace value here
  components:
    # This create a new namespace with the namespace value input above
    # Leave this component commented if you do not need to create one
    - ../../components/resources/namespace
  ```

### **Step 5**: Configure storage class

Sourcegraph requires a storage class that supports SSDs for proper storage and instance performance. You can either use an [exisitng storage class](#existing-storage-class) pre-configured for Sourcegraph, or [create a new storage class](#create-a-new-storage-class) ([cluster role administrator access required](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)).

You can also chooses to use the default storage class without SSD support for non-prod and configure storage class during the [post-installation step](#post-install-configure) later. However, you must recreate cluster with SSDs for production.

#### Existing storage class

For Sourcegraph to use an exisiting storage class in your cluster:

1. Include the `storage-class/update-class-name` component under the components list in the `kustomization.yaml` file.
2. Enter the value for `storageClassName` under the **literals** list in the *configMapGenerator* section using the `STORAGECLASS_NAME` config key

Example, add `STORAGECLASS_NAME=sourcegraph` if `sourcegraph` is the name for the existing storage class:

  ```yaml
  # instances/dev/kustomization.yaml
  apiVersion: kustomize.config.k8s.io/v1beta1
  kind: Kustomization
  resources:
    - ../../base/sourcegraph
  namespace: sg-dev
  components:
    - ../../components/resources/namespace
    # Update storageClassName to the STORAGECLASS_NAME value set below
    - ../../components/storage-class/update-class-name
    
  configMapGenerator:
  - name: sourcegraph-kustomize-env
    behavior: merge
    literals:
      - STORAGECLASS_NAME=sourcegraph # -- [ACTION] Set storage class name here
  ```
The `storage-class/update-class-name` component updates the `storageClassName` field for all associated resources to the `STORAGECLASS_NAME` value set in step 2.

#### Create a new storage class

Alternatively, you choose one of the preconfigured components to create a new storage class named `sourcegraph` for your Sourcegraph deployment. 

Example for Google Kubernetes Engine (GKE): Include the `storage-class/gke` component under the components list in the `kustomization.yaml` file. The component takes care of creating a new storage class named sourcegraph with the following configurations:

- Provisioner: pd.csi.storage.gke.io
- SSD: types: pd-ssd

It also update the storage class name for all resources to `sourcegraph`.

  ```yaml
  # instances/dev/kustomization.yaml
  apiVersion: kustomize.config.k8s.io/v1beta1
  kind: Kustomization
  resources:
    - ../../base/sourcegraph
  namespace: sg-dev
  components:
    - ../../components/resources/namespace
    - ../../components/storage-class/gke
  ```

Please refer to the configurations guide for a complete list of available storage class components and other specific configuration options.

### **Step 6**: Build manifests with Kustomize

Generate a new set of manifests using the configuration applied to the `dev` directory without applying to the cluster.

  ```bash
  # instances/dev/kustomization.yaml
  $ kubectl kustomize instances/dev -o cluster.yaml
  ```

### **Step 7**: Review manifests

Review the generated manifests to ensure they match your intended configuration.

  ```bash
  $ less cluster.yaml
  ```

### **Step 8**: Deploy the generated manifests

Apply the manifests from the ouput file cluster.yaml to your cluster:

  ```bash
  $ kubectl apply --prune -l deploy=sourcegraph -f cluster.yaml
  ```

### **Step 9**: Monitor the deployment

Monitor the deployment status to ensure all components are running properly.

  ```bash
  $ kubectl get pods -A -o wide --watch
  ```

### **Step 10**: Access Sourcegraph in Browser

To access Sourcegraph in a web browser, forward the remote port temporarily.

  ```bash
  $ kubectl port-forward svc/sourcegraph-frontend 3080:30080
  ```

You can then access your new Sourcegraph instance at http://localhost:3080  🎉

  ```bash
  $ open http://localhost:3080
  ```


## Configure

After the initial deployment, additional configuration might be required for Sourcegraph to customize your deployment to suit your specific needs:

- [Add monitoring](configure.md#monitoring-stack)
- [Allocate resources based on your instance size](configure.md#instance-size-based-resources) (refer to our [instance size chart](../instance-size.md))
- [Enable TLS](configure.md#tls)
- [Configure network](configure.md#network-access)
- [Set up an external PostgreSQL Database](configure.md#external-postgres)
- [Set up SSH connection for cloning repositories](configure.md#ssh-for-cloning)

This can all be done by commenting or uncommenting specific components in the kustomization file. Please see the [configuration guide for Kustomize](kustomize/configure.md) for more configuration options.

## Learn more

Please refer to the deployment examples below for each cloud provider:

- [Amazon EKS](kustomize/eks.md)
- [Google GKE](kustomize/gke.md)


Not sure if Kubernetes is the best choice for you? Check out our [deployment documentations](../index.md) to learn about other available deployment options.
