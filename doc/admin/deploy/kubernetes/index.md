# Sourcegraph with Kubernetes

<p class="lead">
Deploying Sourcegraph on Kubernetes is for organizations that need highly scalable and available code search and code intelligence.
</p>

> NOTE: Sourcegraph recommends [using Helm to deploy Sourcegraph](helm.md) if possible.
> This page covers a more manual Kubernetes deployment, using `kubectl` to deploy manifests. This is only recommended if Helm cannot be used in your Kubernetes enviroment. See the Helm guide for more information on why Helm is preferable.

<div class="cta-group">
<a class="btn btn-primary" href="#installation">★ Installation</a>
<a class="btn" href="operations">Operations guides</a>
<a class="btn" href="#about">About Kubernetes</a>
<a class="btn" href="../../../#get-help">Get help</a>
</div>

## Requirements for using Kubernetes

Our Kubernetes support has the following requirements:

- [Sourcegraph Enterprise license](configure.md#add-license-key). _You can run through these instructions without one, but you must obtain a license for instances of more than 10 users_
- Minimum Kubernetes version: [v1.19](https://kubernetes.io/blog/2020/08/26/kubernetes-release-1.19-accentuate-the-paw-sitive/) and [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) v1.19 or later (check kubectl docs for backward and forward compatibility with Kubernetes versions)
- Support for Persistent Volumes (SSDs recommended)

We also recommend some familiarity with the following Kubernetes concepts before proceeding:

- [Kubernetes Objects](https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/)
  - [Namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)
- [Role Based Access Control](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)

Not sure if Kubernetes is the right choice for you? Learn more about the various [Sourcegraph installation options](../index.md).

## Installation

Before starting, we recommend reading the [configuration guide](configure.md#getting-started), ensuring you have prepared the items below so that you're ready to start your installation:

- [Customization](./configure.md#customizations)
- [Storage class](./configure.md#configure-a-storage-class)
- [Network Access](./configure.md#configure-network-access)
- [PostgreSQL Database](./configure.md#sourcegraph-databases)
- [Scaling services](./scale.md#tuning-replica-counts-for-horizontal-scalability)
- [Cluster role administrator access](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)

> WARNING: If you are deploying on Azure, you **must** ensure that [your cluster is created with support for CSI storage drivers](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers). This **can not** be enabled after the fact.

Once you are all set up, either [install Sourcegraph directly](#direct-installation) or [deploy Sourcegraph to a cloud of your choice](#cloud-installation).

### Reference repository

Sourcegraph for Kubernetes is configured using our [`sourcegraph/deploy-sourcegraph` reference repository](https://github.com/sourcegraph/deploy-sourcegraph/). This repository contains everything you need to [spin up](#installation) and [configure](./configure.md) a Sourcegraph deployment on Kubernetes.

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

### ARM / ARM64 support

> WARNING: Running Sourcegraph on ARM / ARM64 images is not supported for production deployments.
