# Install Sourcegraph with Kubernetes

Deploying Sourcegraph into a Kubernetes cluster is for organizations that need highly scalable and
available code search and code intelligence.

The Kubernetes manifests for a Sourcegraph on Kubernetes installation are in the repository
 [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph).

## Requirements

- [Sourcegraph Enterprise license](configure.md#add-license-key). _You can run through these instructions without one, but you must obtain a license for instances of more than 10 users._
- [Kubernetes](https://kubernetes.io/) v1.15
  - Verify that you have enough capacity by following our [resource allocation guidelines](scale.md)
  - Sourcegraph requires an SSD backed storage class
  - [Cluster role administrator access](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) v1.15 or later

> WARNING: You need to create a [fork of our deployment reference.](configure.md#fork-this-repository)

## Steps

- Make sure you have configured `kubectl` to [access your cluster](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/).

   - If you are using GCP, you'll need to give your user the ability to create roles in Kubernetes [(see GCP's documentation)](https://cloud.google.com/kubernetes-engine/docs/how-to/role-based-access-control#prerequisites_for_using_role-based_access_control):

       ```bash
       kubectl create clusterrolebinding cluster-admin-binding --clusterrole cluster-admin --user $(gcloud config get-value account)
       ```

- Clone the [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) repository and check out the version tag you wish to deploy.

   ```bash
   # 🚨 The master branch tracks development. Use the branch of this repository corresponding to the version of Sourcegraph you wish to deploy, e.g. git checkout 3.24

   git clone https://github.com/sourcegraph/deploy-sourcegraph
   cd deploy-sourcegraph
   SOURCEGRAPH_VERSION="v3.24.1"
   git checkout $SOURCEGRAPH_VERSION
   ```

- Configure the `sourcegraph` storage class for the cluster by reading through ["Configure a storage class"](./configure.md#configure-a-storage-class).

- If you want to add a large number of repositories to your instance, you should [configure the number of gitserver replicas](configure.md#configure-gitserver-replica-count) and [the number of indexed-search replicas](configure.md#configure-indexed-search-replica-count) _before_ you continue with the next step. (See ["Tuning replica counts for horizontal scalability"](scale.md#tuning-replica-counts-for-horizontal-scalability) for guidelines.)

- Deploy the desired version of Sourcegraph to your cluster:

   ```bash
   ./kubectl-apply-all.sh
   ```

- Monitor the status of the deployment.

   ```bash
   watch kubectl get pods -o wide
   ```

- Once the deployment completes, verify Sourcegraph is running by temporarily making the frontend port accessible:

   kubectl 1.9.x:

   ```bash
   kubectl port-forward $(kubectl get pod -l app=sourcegraph-frontend -o template --template="{{(index .items 0).metadata.name}}") 3080
   ```

   kubectl 1.10.0 or later:

   ```
   kubectl port-forward svc/sourcegraph-frontend 3080:30080
   ```

   Open http://localhost:3080 in your browser and you will see a setup page. Congrats, you have Sourcegraph up and running!

- Now [configure your deployment](configure.md).

### Configuration

See the [configuration docs](configure.md).

### Troubleshooting

See the [Troubleshooting docs](troubleshoot.md).

### Updating

See the [Upgrading Howto](update.md) on how to upgrade.
See the [Upgrading docs](../../updates/kubernetes.md) for details on what changed in a version and if manual migration steps
are necessary.

### Cluster-admin privileges

> Note: Not all organizations have this split in admin privileges. If your organization does not then you don't need to
> change anything and can ignore this section.

The default installation has a few manifests that require cluster-admin privileges to apply. We have labelled all resources
with a label indicating if they require cluster-admin privileges or not. This allows cluster admins to install the
manifests that cannot be installed otherwise.

- Manifests deployed by cluster-admin

   ```bash
   ./kubectl-apply-all.sh -l sourcegraph-resource-requires=cluster-admin
   ```

- Manifests deployed by non-cluster-admin

   ```bash
   ./kubectl-apply-all.sh -l sourcegraph-resource-requires=no-cluster-admin
   ```

We also provide an [overlay](configure.md#non-privileged-overlay) that generates a version of the manifests that does not
require cluster-admin privileges.

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
