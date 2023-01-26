# Sourcegraph with Kubernetes

<p class="lead">
Deploying Sourcegraph on Kubernetes is for organizations that need highly scalable and available code search and code navigation.
</p>

<div class="getting-started">
  <a href="./kustomize" class="btn btn-primary" alt="Configure">
   <span>Kustomize</span>
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

## Prerequisites

Not sure if Kubernetes is the right choice for you? Learn more about other [Sourcegraph installation options](../index.md).

1. [Sourcegraph Enterprise license](kustomize/configure.md#add-license-key) for instances with more than 10 users
2. A [Kubernetes](https://kubernetes.io/) cluster
   - Minimum Kubernetes version: [v1.19](https://kubernetes.io/blog/2020/08/26/kubernetes-release-1.19-accentuate-the-paw-sitive/) with [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) v1.19 or later
   - [Kustomize](https://kustomize.io/) (built into `kubectl` in version >= 1.14)
   - Support for Persistent Volumes (SSDs recommended)
3. [Cluster role administrator access](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
4. A private local copy of the [Sourcegraph reference repository for Kubernetes](#deployment-repository)
   - Follow our [reference repository docs](../repositories.md) to create one
5. Determine your instance size using our [instance size chart](../instance-size.md)

## Deployment repository

Follow our [reference repository docs](../repositories.md) to create a private copy of the [`sourcegraph/deploy-sourcegraph-k8s`](https://github.com/sourcegraph/deploy-sourcegraph-k8s/) repository, which contains everything you need to [configure](kustomize/configure.md) and [deploy](kustomize#deploy) a Sourcegraph Kubernetes instance using [Kustomize](kustomize/index.md).

## Configure

The default deployment includes the necessary services to start Sourcegraph. It does not includes services or configurations that your cluster needs to run Sourcegraph successfully. As a result, additional configuration might be required in order to deploy Sourcegraph to your Kubernetes cluster successfully.
Common configurations include:

- Adjust resources [Kustomize](kustomize/configure.md#resources) / [Helm](helm.md#configuration)
- Create storage class [Kustomize](kustomize/configure.md#storage-class) / [Helm](helm.md#cloud-providers-guides)
- Configure network settings [Kustomize](kustomize/configure.md#ingress-controller) / [Helm](helm.md#helm-subcharts)
- Set up an external PostgreSQL Database [Kustomize](kustomize/configure.md#external-services) / [Helm](helm.md#using-external-postgresql-databases)
- Set up SSH connection for cloning repositories [Kustomize](kustomize/configure.md#ssh-for-cloning) / [Helm](helm.md#using-ssh-to-clone-repositories)

For more information, please read the [configuration guide for Kustomize](kustomize/configure.md) or the [configuration guide for Helm](helm.md#configuration) before installing Sourcegraph.
