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
   - You can optionally refer to our [terraform configurations](https://github.com/sourcegraph/tf-k8s-configs) for setting up clusters on:
     - [Amazon Web Services EKS](https://github.com/sourcegraph/tf-k8s-configs/tree/main/aws)
     - [Azure AKS](https://github.com/sourcegraph/tf-k8s-configs/tree/main/azure)
     - [Google Cloud Platform GKE](https://github.com/sourcegraph/tf-k8s-configs/tree/main/gcp)

---

### **Step 1**: Set up a release branch

Set up a release branch from the default branch in your local forked copy of the [deploy-sourcegraph-k8s](https://github.com/sourcegraph/deploy-sourcegraph-k8s) repository.

```bash
$ git checkout -b release
```

### **Step 2**: Set up an overlay

`cd` into the repository and create a new directory `instances/my-sourcegraph` for your Sourcegraph instance.

```bash
$ mkdir instances/my-sourcegraph
```

### **Step 3**: Set up a configuration file

Copy [instances/kustomization.template.yaml](kustomize/index.md#template) to the `instances/my-sourcegraph` subdirectory as `kustomization.yaml`.

```bash
$ cp instances/template/kustomization.template.yaml instances/my-sourcegraph/kustomization.yaml
```

The new `kustomization.yaml` file will be used to configure your deployment.

### **Step 4**: Set namespace

By default, the provided kustomization.yaml template deploys sourcegraph into the `sourcegraph` namespace. If you intend to deploy sourcegraph into a different namespace, replace `sourcegraph` with the name of the existing namespace in your cluster.

Set it to `default` to deploy to the default namespace.

  ```yaml
  # instances/my-sourcegraph/kustomization.yaml
  namespace: sourcegraph
  ```

### **Step 5**: Set storage class

Sourcegraph requires a storage class that supports SSDs for proper storage and instance performance.

Here are the options to set the storage class for your instance:

#### Option 1: Create a new storage class

If you have the ability to create a new storage class, choose a storage-class component that corresponds to your cluster provider:

  ```yaml
  # instances/my-sourcegraph/kustomization.yaml
  components:
    - ../../components/storage-class/aws
    - ../../components/storage-class/azure
    - ../../components/storage-class/gke
  ```

Please refer to our [configurations guide](kustomize/configure.md) for a complete list of available storage class components and other specific configuration options.

#### Option 2: Use an existing storage class

If you'd like to use an exisiting storage class:

1. Include the `storage-class/update-class-name` component under the components list
2. Input the storage class name by setting the value for `STORAGECLASS_NAME` under the configMapGenerator section
   
For example, set `STORAGECLASS_NAME=sourcegraph` if `sourcegraph` is the name of an existing storage class:

  ```yaml
  # instances/my-sourcegraph/kustomization.yaml
  components:
    # Update storageClassName to the STORAGECLASS_NAME value set below
    - ../../components/storage-class/update-class-name
  # ...
  configMapGenerator:
  - name: sourcegraph-kustomize-env
    behavior: merge
    literals:
      - STORAGECLASS_NAME=sourcegraph # Set STORAGECLASS_NAME value to 'sourcegraph'
  ```

#### Option 3: Use default storage class

Skip this step to use the default storage class without SSD support for non-prod environment; however, you must recreate the cluster with SSDs configured for a production environment later.

### **Step 6**: Build manifests with Kustomize

Generate a new set of manifests using the configuration applied to the `my-sourcegraph` subdirectory without applying to the cluster.

  ```bash
  $ kubectl kustomize instances/my-sourcegraph -o cluster.yaml
  ```

### **Step 7**: Review manifests

Review the generated manifests to ensure they match your intended configuration.

  ```bash
  $ less cluster.yaml
  ```

### **Step 8**: Deploy the generated manifests

Apply the manifests from the ouput file `cluster.yaml` to your cluster:

  ```bash
  $ kubectl apply --prune -l deploy=sourcegraph -f cluster.yaml
  ```

### **Step 9**: Monitor the deployment

Monitor the deployment status to ensure all components are running properly.

  ```bash
  $ kubectl get pods -A -o wide --watch
  ```

### **Step 10**: Access Sourcegraph in Browser

To verify that the deployment was successful, port-forward the frontend pod with the following command:

  ```bash
  $ kubectl port-forward svc/sourcegraph-frontend 3080:30080
  ```

Then access your new Sourcegraph instance at http://localhost:3080 to proceed to the site-admin setup step.

  ```bash
  $ open http://localhost:3080
  ```

---

## Configure

After the initial deployment, additional configuration might be required for Sourcegraph to customize your deployment to suit your specific needs.

Common configurations that are strongly recommended for all Sourcegraph deployments:

- [Enable the Sourcegraph monitoring stack](kustomize/configure.md#monitoring-stack)
- [Allocate resources based on your instance size](kustomize/configure.md#instance-size-based-resources) (refer to our [instance size chart](../instance-size.md))
- [Configure ingress](kustomize/configure.md#ingress)
- [Enable TLS](kustomize/configure.md#tls)

Other common configurations include:

- [Set up an external PostgreSQL Database](kustomize/configure.md#external-postgres)
- [Set up SSH connection for cloning repositories](kustomize/configure.md#ssh-for-cloning)

See the [configuration guide for Kustomize](kustomize/configure.md) for more configuration options.

## Learn more

- [Migrate from deploy-sourcegraph to deploy-sourcegraph-k8s](kustomize/migrate.md)
- Examples of deploying Sourcegraph to the cloud provider listed below:
  - [Amazon EKS](kustomize/eks.md)
  - [Google GKE](kustomize/gke.md)
  - [Minikube](../single-node/minikube.md)
- [Troubleshooting](troubleshoot.md)

Not sure if Kubernetes is the best choice for you? Check out our [deployment documentations](../index.md) to learn about other available deployment options.
