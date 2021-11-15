# Sourcegraph with Kubernetes

<p class="lead">
Deploying Sourcegraph on Kubernetes is for organizations that need highly scalable and
available code search and code intelligence.
</p>

Not sure if Kubernetes is the right choice for you? Learn more about the various [Sourcegraph installation options](../index.md).

<div class="cta-group">
<a class="btn btn-primary" href="#installation">★ Installation</a>
<a class="btn" href="operations">Operations guides</a>
<a class="btn" href="#about">About Kubernetes</a>
<a class="btn" href="../../../#get-help">Get help</a>
</div>

## Installation

Before you get started, we recommend [learning about how Sourcegraph with Kubernetes works](#about).

Additionally, we recommend reading the [configuration guide](configure.md#getting-started), ensuring you have prepared the items below so that you're ready to start your installation:

- [Customization](./configure.md#customizations)
- [Storage class](./configure.md#configure-a-storage-class)
- [Network Access](./configure.md#configure-network-access)
- [PostgreSQL Database](./configure.md#sourcegraph-databases)
- [Scaling services](./scale.md#tuning-replica-counts-for-horizontal-scalability)
- [Cluster role administrator access](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)

> WARNING: If you are deploying on Azure, you **must** ensure that [your cluster is created with support for CSI storage drivers](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers). This **can not** be enabled after the fact.

Once you are all set up, either [install Sourcegraph directly](#direct-installation) or [deploy Sourcegraph to a cloud of your choice](#cloud-installation).

### Direct installation

- After meeting all the requirements, make sure you can [access your cluster](https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/) with `kubectl`.
- `cd` to the forked local copy of the [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) repository previously set up during [configuration](./configure.md#getting-started).
- Deploy the desired version of Sourcegraph to your cluster by [applying the Kubernetes manifests](./configure.md#applying-manifests):

  ```sh
  ./kubectl-apply-all.sh
  ```

  > NOTE: Google Cloud Platform (GCP) users are required to give their user the ability to create roles in Kubernetes
  > ([Learn more](https://cloud.google.com/kubernetes-engine/docs/how-to/role-based-access-control#prerequisites_for_using_role-based_access_control)):
  >
  > `kubectl create clusterrolebinding cluster-admin-binding --clusterrole cluster-admin --user $(gcloud config get-value account)`

- Monitor the status of the deployment:

  ```sh
  kubectl get pods -o wide --watch
  ```

- After deployment is completed, verify Sourcegraph is running by temporarily making the frontend port accessible:

  ```sh
  kubectl port-forward svc/sourcegraph-frontend 3080:30080
  ```

- Open http://localhost:3080 in your browser and you will see a setup page. Congratulations, you have Sourcegraph up and running! 🎉 

> NOTE: If you previously [set up an `ingress-controller`](./configure.md#ingress-controller-recommended), you can now also access your deployment via the ingress.

### Cloud installation

> WARNING: If you intend to set this up as a production instance, we recommend you create the cluster in a VPC
> or other secure network that restricts unauthenticated access from the public Internet. You can later expose the
> necessary ports via an
> [Internet Gateway](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_Internet_Gateway.html) or equivalent
> mechanism. Note that SG must expose port 443 for outbound traffic to codehosts and to enable [telemetry](https://docs.sourcegraph.com/admin/pings) with 
> Sourcegraph.com. Additionally port 22 may be opened to enable git SSH cloning by Sourcegraph. Take care to secure your cluster in a manner that meets your 
> organization's security requirements.

Follow the instructions linked in the table below to provision a Kubernetes cluster for the
infrastructure provider of your choice, using the recommended node and list types in the
table.

|Provider|Node type|Boot/ephemeral disk size|
|--- |--- |--- |
|[Amazon EKS (better than plain EC2)](eks.md)|m5.4xlarge| 100 GB (SSD preferred) |
|[AWS EC2](https://kubernetes.io/docs/getting-started-guides/aws/)|m5.4xlarge|  100 GB (SSD preferred) |
|[Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine/docs/quickstart)|n1-standard-16|100 GB (default)|
|[Azure](azure.md)|D16 v3|100 GB (SSD preferred)|
|[Other](https://kubernetes.io/docs/setup/pick-right-solution/)|16 vCPU, 60 GiB memory per node|100 GB (SSD preferred)|

<span class="virtual-br"></span>

> NOTE: Sourcegraph can run on any Kubernetes cluster, so if your infrastructure provider is not
> listed, see the "Other" row. Pull requests to add rows for more infrastructure providers are
> welcome!

<span class="virtual-br"></span>

> WARNING: If you are deploying on Azure, you **must** ensure that [your cluster is created with support for CSI storage drivers](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers). This **can not** be enabled after the fact.

## About

### Kubernetes

Kubernetes is a portable, extensible, open-source platform for managing containerized workloads and services, that facilitates both declarative configuration and automation. Applications are deployed via a set of YAML files to configure the various components (storage, networking, containers). Learn more about Kubernetes [here](https://kubernetes.io/docs/concepts/overview/what-is-kubernetes/).

Our Kubernetes support has the following requirements:

- [Sourcegraph Enterprise license](configure.md#add-license-key). _You can run through these instructions without one, but you must obtain a license for instances of more than 10 users._
- Minimum Kubernetes version: [v1.18](https://kubernetes.io/blog/2020/03/25/kubernetes-1-18-release-announcement/) and [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) v1.18 or later (check kubectl docs for backward and forward compatibility with Kubernetes versions).
- Support for Persistent Volumes. 

We also recommend familiarizing yourself with the following before proceeding with the install steps:

- [Kubernetes Objects](https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/)
  - [Namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)
- [Role Based Access Control](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)

### Kustomize

We support the use of [Kustomize](https://kustomize.io) to modify and customize our Kubernetes manifests. Kustomize is a template free way to customize configuration, in a Kubernetes like way with a simple configuration file.

Some benefits of using Kustomize to generate manifests instead of modifying the base directly include:

- Reduce the odds of encountering a merge conflict when [upgrading](update.md) - they allow you to separate your unique changes from the upstream bases.
- Better enable us to support you if you run into issues, because how your deployment varies from our [reference deployment](#reference-repository) is encapsulated in a small set of files.

For more information about how to use Kustomize with Sourcegraph, see our [customization guide](./configure.md#customizations) and [introduction to overlays](#overlays).

#### Overlays

An [*overlay*](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/#bases-and-overlays) specifies customizations for a base directory of Kubernetes manifests, in this case the `base/` directory in the [reference repository](#reference-repository). Overlays can:

- be used for example to change the number of replicas, change a namespace, add a label, etc
- refer to other overlays that eventually refer to the base (forming a directed acyclic graph with the base as the root)

Overlays can be used in one of two ways:

- With `kubectl`: Starting with `kubectl` client version 1.14 `kubectl` can handle `kustomization.yaml` files directly.
When using `kubectl` there is no intermediate step that generates actual manifest files. Instead the combined resources from the
overlays and the base are directly sent to the cluster. This is done with the `kubectl apply -k` command. The argument to the
command is a directory containing a `kustomization.yaml` file.
- With`kustomize`: This generates manifest files that are then applied in the conventional way using `kubectl apply -f`.

The overlays provided in our [overlays directory](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/overlays) rely on the `kustomize` tool and the `overlay-generate-cluster.sh` script in the
`root` directory to generate the manifests. There are two reasons why it was set up like this:

- It avoids having to put a `kustomization.yaml` file in the `base` directory and forcing users that don't use overlays
  to deal with it (unfortunately `kubectl apply -f` doesn't work if a `kustomization.yaml` file is in the directory).
- It generates manifests instead of applying them directly. This provides opportunity to additionally validate the files
  and also allows using `kubectl apply -f` with `--prune` flag turned on (`apply -k` with `--prune` does not work correctly).

To learn about our available overlays and how to use them, please refer to our [overlays guides](./configure.md#overlays).

### Reference repository

Sourcegraph for Kubernetes is configured using our [`sourcegraph/deploy-sourcegraph` reference repository](https://github.com/sourcegraph/deploy-sourcegraph/). This repository contains everything you need to [spin up](#installation) and [configure](./configure.md) a Sourcegraph deployment on Kubernetes.
