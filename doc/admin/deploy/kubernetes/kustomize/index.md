# Kustomize

Sourcegraph supports the use of [Kustomize](https://kustomize.io), which allows users to modify and customize our [default Kubernetes manifests](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/base) without the need to modify or create copies of the original files.

We have created a new set of Sourcegraph Kustomize overlays that align with Kustomize best practices, which provides more flexibility in creating an overlay that suits your deployments, while eliminating the need to clone the deployment repository. You can find the complete list of overlays in our [sourcegraph/kustomize](https://github.com/sourcegraph/kustomize) repository.

> NOTE: If you have yet to deploy Sourcegraph, it is highly recommended to use Helm for the deployment and configuration ([Using Helm with Sourcegraph](helm.md)).

## Overview

[Kustomize](https://kustomize.io) is a configuration management solution that leverages layering to preserve the default settings and components managed by the original source. It utilizes overlaying declarative yaml artifacts that selectively override default settings without needing to make direct changes to the original files. Using Kustomize has the following benefits:

- The [remote targets](https://github.com/kubernetes-sigs/kustomize/blob/master/examples/remoteBuild.md) feature allows users to refer to remote repositories (including their branches and directories) in their overlays, eliminating the need to clone the deployment manifests locally.
- Reduce the odds of encountering a merge conflict when [updating Sourcegraph](../update.md) - they allow you to separate your unique changes from the upstream base files Sourcegraph provides.
- Better enable Sourcegraph to support you if you run into issues because how your deployment varies from our defaults is encapsulated in a small set of files.

See the [official docs](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/#overview-of-kustomize) to learn more about Kustomize and overlays.

## Overlays

An [*overlay*](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/#bases-and-overlays) specifies customizations for a base directory of Kubernetes manifests, in this case, the `base/` directory in the [reference repository](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/base), where [components/](https://github.com/sourcegraph/kustomize/tree/main/components) directory in the [sourcegraph/kustomize repository](https://github.com/sourcegraph/kustomize) are where we specify the customizations.

Overlays can:

- Be used for example to change the number of replicas, change a namespace, add a label, etc
- Refer to other overlays that eventually refer to the base (forming a directed acyclic graph with the base as the root)
  
### Create overlays

To create an overlay is to create a new `kustomization.yaml` file where you will specify the base under the `resources` field, and the customizations under the `components` field.

For example, [the kustomization.yaml file for our non-privileged overlay](https://sourcegraph.com/github.com/sourcegraph/kustomize@main/-/blob/overlays/non-privileged/kustomization.yaml?L10) as shown below is using our reference repository as base, where `?ref=v` indicates the version number by specifying the name of the branch. Below the resources where we identify the base for the overylay, we include the components we want to build this overlays, which are the non-privileged components and non-privileged-create-cluster components from our [sourcegraph/kustomize repository](https://github.com/sourcegraph/kustomize):

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: np-sourcegraph
resources:
  - git::git@github.com:sourcegraph/deploy-sourcegraph/base?ref=v4.1.3
components:
  - git::git@github.com:sourcegraph/kustomize/components/non-privileged
  - git::git@github.com:sourcegraph/kustomize/components/non-privileged-create-cluster
```

All changes should be made within a new `kustomization.yaml` file inside the [overlays directory](https://github.com/sourcegraph/kustomize/tree/main/overlays). For instance, if you would like to use the storageclass overlay, you will need to update the storage class value inside [.storageclass.env file](https://github.com/sourcegraph/kustomize/tree/main/overlays/storageclass/.storageclass.env) instead of the files in the storageclass component.

> WARNING: Making changes directly in the existing components is not recommended. New components should be created if the existing components do not suit your needs.

### Build overlays

Once you have an overlay, run the following command in the directory where the `kustomization.yaml` file for your overlay is located to build the customized manifests for your deployment. The updated manifests with your customization can then be found inside the `.overlay_output.yaml`.

```bash
kustomize build . > .overlay_output.yaml
```

> NOTE: This command will build a new set of manifests based on your overlay. It does not affect your current deployment until you run the apply command.

### Apply overlays

In order to apply the customized manifests to your cluster, run the following command in the directory where the `kustomization.yaml` file for your overlay is located:

```bash
kustomize build . | kubectl apply -f -
```

### Kustomize with Helm

Kustomize can be used **with** Helm to configure Sourcegraph (see [this guidance](helm.md#integrate-kustomize-with-helm-chart)) but this is only recommended as a temporary workaround while Sourcegraph adds to the Helm chart to support previously unsupported customizations.

## Deprecated

The latest version of our Kustomize overlays does not work for instances that are on v4.1.3 or older.

The previous Kustomize structure we built for our Kubernetes deployments depended on additional scripting to create deployment manifests. It does not provide flexibility and requires further implementation and integrations that can now be avoided with the latest structure we have introduced in this documentation. 

❌ See the [docs for the deprecated version of kustomize](deprecated.md).
