# Kustomize

Sourcegraph supports the use of [Kustomize](https://kustomize.io), which allows users to modify and customize our [default Kubernetes manifests](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/base) without the need to modify or create copies of the original files.

We have created a new set of Sourcegraph Kustomize overlays that align with Kustomize best practices, which provides more flexibility in creating an overlay that suits your deployments, while eliminating the need to clone the deployment repository. You can find the complete list of overlays in our [sourcegraph/kustomize](https://github.com/sourcegraph/kustomize) repository.

> NOTE: If you have yet to deploy Sourcegraph, it is highly recommended to use Helm for the deployment and configuration ([Using Helm with Sourcegraph](helm.md)).

## Overview

[Kustomize](https://kustomize.io) is a configuration management solution that leverages layering to preserve the default settings and components managed by the original source. It utilizes overlaying declarative yaml artifacts that selectively override default settings without needing to make direct changes to the original files. Using the provided Kustomize overlays have the following benefits:

- The [remote targets](https://github.com/kubernetes-sigs/kustomize/blob/master/examples/remoteBuild.md) feature allows users to refer to remote repositories (including their branches and directories) in their overlays, eliminating the need to clone the deployment manifests locally.
- Reduce the odds of encountering a merge conflict when [updating Sourcegraph](../update.md) - they allow you to separate your unique changes from the upstream base files Sourcegraph provides.
- Better enable Sourcegraph to support you if you run into issues because how your deployment varies from our defaults is encapsulated in a small set of files.

### Overlays

  
## Tutorial

### Create overlays

All changes should be made within a kustomization.yaml file inside the [overlays directory](https://github.com/sourcegraph/kustomize/tree/main/overlays).

For example, if you would like to use the storageclass overlay, you will need to update the storage class value inside [.storageclass.env file](https://github.com/sourcegraph/kustomize/tree/main/overlays/storageclass/.storageclass.env) instead of the files in the storageclass component.

### Build overlays

To build an overlay for your deployment,

Once you have an overlay, run the following command in the directory where the kustomization.yaml file for your overlay is located:

```bash
kustomize build . > .overlay_output.yaml
```

This command will build a new set of manifests based on your overlay. It does not affect your current deployment until you run the apply command.

### Apply overlays

In order to apply the overlays to your cluster, run the following command in the directory where the kustomization.yaml file for your overlay is located:

```bash
# example: kustomize build . | kubectl apply -f -
kustomize build . | kubectl apply -f -
```

### Kustomize with Helm

Kustomize can be used **with** Helm to configure Sourcegraph (see [this guidance](helm.md#integrate-kustomize-with-helm-chart)) but this is only recommended as a temporary workaround while Sourcegraph adds to the Helm chart to support previously unsupported customizations.

## Deprecated

The latest version of our Kustomize overlays does not work for instances that are on v4.1.3 or older.

The previous Kustomize structure we built for our Kubernetes deployments depended on additional scripting to create deployment manifests. It does not provide flexibility and requires further implementation and integrations that can now be avoided with the latest structure we have introduced in this documentation. 

❌ [Deprecated version of kustomize](deprecated.md)
