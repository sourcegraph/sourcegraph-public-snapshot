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
  <a href="./helm" class="btn" alt="Overlays">
   <span>Helm</span>
   </br>
   Deploy Sourcegraph with Helm
  </a>
</div>

<div class="getting-started">
<a class="btn btn-primary text-center" href="#installation">★ Installation</a>
<a class="btn text-center" href="kustomize/configure">Configuration</a>
<a class="btn text-center" href="../instance-size">Instance Sizes</a>
<a class="btn text-center" href="operations">Operations</a>
</div>

> WARNING: If you are currently on Sourcegraph version 4.4.1 or below, please refer to the [deprecated deployment docs for Kubernetes](../deprecated/index.md).

## Prerequisites

* [Sourcegraph Enterprise license](kustomize/configure.md#add-license-key) for instances with more than 10 users
* A [Kubernetes](https://kubernetes.io/) cluster
   - Minimum Kubernetes version: [v1.19](https://kubernetes.io/blog/2020/08/26/kubernetes-release-1.19-accentuate-the-paw-sitive/) with [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) v1.19 or later
   - [Kustomize](https://kustomize.io/) (built into `kubectl` in version >= 1.14) or [Helm 3 CLI](https://helm.sh/docs/intro/install/)
   - Support for Persistent Volumes (SSDs recommended)
   - [Cluster role administrator access](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
* A private local copy of the [Sourcegraph reference repository for Kubernetes](#reference-repository)
   - Follow our [reference repository docs](../repositories.md) to create one
* Determine your instance size using our [instance size chart](../instance-size.md)

## Reference repository

To deploy Sourcegraph on Kubernetes with Kustomize, please follow our [reference repository docs](../repositories.md) to create a private copy of the reference repository that contains everything you need to [configure](kustomize/configure.md) and [deploy](kustomize#deploy) a Sourcegraph instance to a Kubernetes cluster.

- For Kustomize: [sourcegraph/deploy-sourcegraph-k8s](https://github.com/sourcegraph/deploy-sourcegraph-k8s/)
- For Helm: [sourcegraph/deploy-sourcegraph-helm](https://github.com/sourcegraph/deploy-sourcegraph-helm/)

## Configure

The default deployment includes the necessary services to start Sourcegraph. It does not includes services or configurations that your cluster needs to run Sourcegraph successfully. As a result, additional configuration might be required in order to deploy Sourcegraph to your Kubernetes cluster successfully.
Common configurations include:

- Adjust resources
- Create storage class
- Configure network settings
- Set up an external PostgreSQL Database
- Set up SSH connection for cloning repositories

Please see the [configuration guide for Kustomize](kustomize/configure.md) or the [configuration guide for Helm](helm.md#configuration) for more configuration options.

## Learn more

Not sure if Kubernetes is the best choice for you? Check out our [deployment documentations](../index.md) to learn about other available deployment options.
