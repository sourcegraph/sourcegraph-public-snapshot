# Sourcegraph with Kubernetes

<p class="lead">
Deploying Sourcegraph on Kubernetes is for organizations that need highly scalable and
available code search and code intelligence.
</p>

Not sure if Kubernetes is the right choice for you? Learn more about the various [Sourcegraph installation options](../index.md).

## Installation

Before you get started, we recommend [learning about how Sourcegraph with Kubernetes works](#about).

In addition to the requirements in the [about](#about) section, we recommend reading [configuration guide](configure.md) 
ensuring you have prepared the items below to ensure you're ready to start your installation. 

 - [Customizations](./overlays.md)
 - [Storage class](./configure.md#configure-a-storage-class)
 - [Network Acess](./configure.md#security-configure-network-access)
 - [PostgreSQL Database](./configure.md#sourcegraph-databases)
 - [Scaling services](./scale.md#tuning-replica-counts-for-horizontal-scalability)
 - [Cluster role administrator access](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)

### Steps

1. After meeting all the requirements, make sure you can [access your cluster](https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/) with `kubectl`.

```bash
kubectl version
Client Version: version.Info{Major:"1", Minor:"21", GitVersion:"v1.21.0", GitCommit:"cb303e613a121a29364f75cc67d3d580833a7479", GitTreeState:"clean", BuildDate:"2021-04-08T16:31:21Z", GoVersion:"go1.16.1", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"20+", GitVersion:"v1.20.8-gke.900", GitCommit:"28ab8501be88ea42e897ca8514d7cd0b436253d9", GitTreeState:"clean", BuildDate:"2021-06-30T09:23:36Z", GoVersion:"go1.15.13b5", Compiler:"gc", Platform:"linux/amd64"}
```

```bash
# Google Cloud Platform (GCP) users are required to give their user the ability to create roles in Kubernetes.
# See the [GCP's documentation: https://cloud.google.com/kubernetes-engine/docs/how-to/role-based-access-control#prerequisites_for_using_role-based_access_control
kubectl create clusterrolebinding cluster-admin-binding \
    --clusterrole cluster-admin --user $(gcloud config get-value account)
```

2) `cd` to the forked local copy of [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) repository previously setup during [configuration](./configure.md#getting-started):

```bash
# 🚨 The master branch tracks development.
# Use the branch of this repository corresponding to the version of Sourcegraph you wish to deploy, e.g. git checkout 3.30
cd deploy-sourcegraph
export SOURCEGRAPH_VERSION="v3.30.3"
git checkout $SOURCEGRAPH_VERSION -b release
```

3) Deploy the desired version of Sourcegraph to your cluster:

```
./kubectl-apply-all.sh
```

4) Monitor the status of the deployment:

```
kubectl get pods -o wide --watch
```

5) After deployment is completed, verify Sourcegraph is running by temporarily making the frontend port accessible:

```
kubectl port-forward svc/sourcegraph-frontend 3080:30080
```

6) Open http://localhost:3080 in your browser and you will see a setup page.


7) 🎉 Congrats, you have Sourcegraph up and running!

8) If you previously setup an `ingress-controller`, you can also access your deployment via the `sourcegraph-frontend-ingress`.

Run the following command, and ensure an IP address has been assigned to your ingress resource. Then browse to the IP or configured URL.
```
kubectl get ingress sourcegraph-frontend

NAME                   CLASS    HOSTS             ADDRESS     PORTS     AGE
sourcegraph-frontend   <none>   sourcegraph.com   8.8.8.8     80, 443   1d
```

## Troubleshooting

See the [Troubleshooting docs](troubleshoot.md).

## Updating

- See the [Updating Sourcegraph docs](update.md) on how to upgrade.<br/>
- See the [Updating a Kubernetes Sourcegraph instance docs](../../updates/kubernetes.md) for details on changes in each version to determine if manual migration steps are necessary.

## Cloud installation guides

>**Security note:** If you intend to set this up as a production instance, we recommend you create the cluster in a VPC
>or other secure network that restricts unauthenticated access from the public Internet. You can later expose the
>necessary ports via an
>[Internet Gateway](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_Internet_Gateway.html) or equivalent
>mechanism. Take care to secure your cluster in a manner that meets your organization's security requirements.

Follow the instructions linked in the table below to provision a Kubernetes cluster for the
infrastructure provider of your choice, using the recommended node and list types in the
table.

> Note: Sourcegraph can run on any Kubernetes cluster, so if your infrastructure provider is not
> listed, see the "Other" row. Pull requests to add rows for more infrastructure providers are
> welcome!

|Provider|Node type|Boot/ephemeral disk size|
|--- |--- |--- |
|Compute nodes| | |
|[Amazon EKS (better than plain EC2)](eks.md)|m5.4xlarge|N/A|
|[AWS EC2](https://kubernetes.io/docs/getting-started-guides/aws/)|m5.4xlarge|N/A|
|[Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine/docs/quickstart)|n1-standard-16|100 GB (default)|
|[Azure](azure.md)|D16 v3|100 GB (SSD preferred)|
|[Other](https://kubernetes.io/docs/setup/pick-right-solution/)|16 vCPU, 60 GiB memory per node|100 GB (SSD preferred)|


## About

### Kubernetes

Kubernetes is a portable, extensible, open-source platform for managing containerized workloads and services, that facilitates both declarative configuration and automation. Applications are deployed via set a of YAML files to configure the various components (storage, networking, containers). Learn more about Kubernetes [here](https://kubernetes.io/docs/concepts/overview/what-is-kubernetes/).

Our Kubernetes support has the following requirements:

- [Sourcegraph Enterprise license](configure.md#add-license-key). _You can run through these instructions without one, but you must obtain a license for instances of more than 10 users._
- Minimum Kubernetes version: [v1.15](https://kubernetes.io/blog/2019/06/19/kubernetes-1-15-release-announcement/) and [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) v1.15 or later.

We also recommend familiarizing yourself with the following before proceeding with the install steps:

- [Kubernetes Objects](https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/)
  - [Namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)
- [Role Based Access Control](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)

### Kustomize

We support the use of [Kustomize](https://kustomize.io) to modify and customize our Kubernetes manifests. Kustomize is a template free way to customize configuration, in a Kubernetes like way with a simple configuration file. For more information about how to use Kustomize with Sourcegraph, see our [overlays guide](./overlays.md). 

### Reference repository

Sourcegraph for Kubernetes is configured using our [`sourcegraph/deploy-sourcegraph` reference repository](https://github.com/sourcegraph/deploy-sourcegraph/). This repository contains everything you need to [spin up](#installation) and [configure](./configure.md) a Sourcegraph deployment on Kubernetes.
