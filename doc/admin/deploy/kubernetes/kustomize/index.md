# Kustomize

[Kustomize](https://kustomize.io), which is a configuration management solution that leverages layering to preserve the default settings and components managed by the original source. It utilizes overlaying declarative yaml artifacts that selectively override default settings without needing to make direct changes to the original files.

## Prerequisites

- [Kubernetes cluster access](https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/) with `kubectl`
- A [private clone](../../repositories.md) of the [Sourcegraph Kubernetes deployment repository](./index.md#deployment-repository)
- Determine your instance size using our [instance size chart](../../instance-size.md)
- A [configured](configure.md) Sourcegraph Kustomize Overlay
  - see details below to learn more about building a customized Overlay that can be reused across updates

## Configure

Please refer to our [configuration guides](configure.md) for detailed instructions to build a Kustomize Overlay for Sourcegraph.

## Deploy

**Step 1:** Follow the instructions below to build a Kustomize Overlay with [custom configurations](configure.md) made for your Sourcegraph deployment, or use one of our [pre-built overlays](#pre-built-overlays) to deploy Sourcegraph into a pre-configured cluster.

**Step 2:** Build the deployment manifests with the overlay prepared from step 1 for review

  ```bash
  $ kubectl kustomize $PATH_TO_OVERLAY -o new/preview-cluster
  ```

**Step 3:** Make sure the manifests in the output directory `new/preview-cluster` are generated correctly

**Step 4:** Run the following command from the root of the cloned repository to apply the manifests generated from step 2 inside the `new/preview-cluster` directory

  ```bash
  $ kubectl apply -k --prune -l deploy=sourcegraph -f new/preview-cluster
  ```

## Upgrade

To upgrade your instance with Kustomize:

**Step 1:** Merge your release branch with the upstream branch

**Step 2:** Build the deployment manifests with an overlay

**Step 3:** Review the output manifests to make sure they reflect the configurations made by the overlay

  ```bash
  $ kubectl kustomize $PATH_TO_OVERLAY -o new/preview-cluster
  ```

**Step 4:** Run the following command from the root of your deployment repository to apply the manifests inside the `new/preview-cluster directory` generated from step 2

  ```bash
  $ kubectl apply -k --prune -l deploy=sourcegraph -f new/preview-cluster
  ```

## Deployment repository

The `new/overlays/deploy` directory is the recommended path to create and store the customized overlay for your deployment. If you would like to set up two seperated instances for production and staging purpose, you may create a copy of the `new/overlays/template` folder and rename it to `deploy-staging` and configure it seperately. 

If you would like to set up two seperated instances for production and staging purpose, you may create a copy of the `new/overlays/template` folder and rename it to `deploy-staging` and configure it seperately.

### Overlays

An [*overlay*](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/#bases-and-overlays) includes resources from the source as the base cluster, with configurations made on top to create an instance that would work in a specific environment (ex. different cloud providers, network policy etc),  

In our [kubernetes deployment repository](https://github.com/sourcegraph/deploy-sourcegraph), the Kubernetes manifests located within the [new/base directory](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/base) are the base resources that should only be maintained by Sourcegraph, where the [new/components/ directory](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/components) is where the pre-configured components that are reusable without additional configurations required. Please see our [configuration docs for Kustomize](./configure.md) for details.

#### kustomization.yaml

Below is the `kustomization.yaml` file template for building a Kustomize Overlay to deploy Sourcegraph. All fields list below must be included in order to set up a functional Sourcegraph instance with the default configurations. You should then build on this template by adding different components that we will cover in this docs to apply additional changes to your Sourcegraph deployment.

The same file can also be found inside the `new/overlays/deploy` directory.

```yaml
# new/overlays/template/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
# update namespace value if needed
namespace: default
# Resources for Sourcegraph main stacks
resources:
- ../../base/sourcegraph
# Essential component for updating frontend env vars
configMapGenerator:
- name: sourcegraph-frontend-env
  behavior: merge
  env: configs/sourcegraph.env
components:
# Resources for Sourcegraph monitoring stacks - RBAC required
- ../../components/monitoring
```

See the [kustomize docs by kubectl](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/) to learn more about the kustomization file.

#### Pre-built overlays

We have a sets of pre-built overlays that are ready-to-use for clusters that do not require additional configurations for Sourcegraph to be deployed. You can find the complete list of pre-built overlays inside the `new/overlays/quick-start` directory.

### Components

A kustomize component is essentially a smaller unit of a regular kustomization file. They are designed to be reusable for different purposes and easy to deploy using the [remote build feature](#remote-build).

_Components are evaluated after the resources of the parent kustomization (overlay or component) have been accumulated, and on top of them. ([source](https://sourcegraph.com/github.com/kubernetes/enhancements@master/-/blob/keps/sig-cli/1802-kustomize-components/README.md#proposal))_

To understand an overlay is to examine the components it uses. The components are listed under the `components` field inside the `kustomization.yaml` file of an overlay.

#### Pre-built components

Components located inside the `new/config` directory are all functional and do not require additional configurations as they were already pre-configured for different purposes.

If you would like to modify a component from this directory, it is strongly recommend to create a copy of the said component inside the [new/config directory](../config/) where all the components that require additional configurations are located, and then make your changes there to avoid merge conflicts and confusions during upgrades.

#### Configurable components

Components that required additional configurations are listed inside the `new/config` directory. Using them without making customized changes would result in error when building or deploying with an Overlay that includes components from this directory.

### Example

Here is an example `kustomization.yaml` file inside from a Kustomize Overlay built for deploying Sourcegraph:

```yaml
# new/overlays/example/kustomization.yaml
# Note: this is the kustomization.yaml file from our quick-start overlay that is configured for deploying a size XS instance to a k3s cluster
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: ns-sourcegraph-example
resources:
# Add resources for Sourcegraph Main Stacks
- ../../new/base/sourcegraph
components:
# Add resources for Sourcegraph Monitoring Stacks
- ../../new/components/monitoring
# Configurs the deployment for k3s cluster
- ../../new/components/k3s
# Configurs the resources we added above for a size XS instance
- ../../new/components/sizes/xs
```

See the complete list of ready-to-use overlays inside the `new/overlays/quick-start` directory to learn more about combining different pre-configured components to build a new overlay to deploy Sourcegraph.

### Remote build

Remote build allows you to deploy an overlay using a URL for an overlay to build and deploy; however, it does not support custom configurations as the resources are hosted remotely.

Run the following command to generate manifests using the remote URL for one of our quick-start overlays. You can replace the $REMOTE_OVERLAY_URL with the remote URL for the overlay of your choice, and replace $PATH_TO_EXISITING_DIRECTORY with the path to an existing directory on your local machine where the newly generated manifests can be found.

```bash
$ kubectl kustomize $REMOTE_OVERLAY_URL -o $PATH_TO_EXISITING_DIRECTORY
```

The manifests will be grouped into a single file `preview-cluster.yaml` you can preview before running the apply command to deploy:

```bash
$ kubectl apply -k --prune -l deploy=sourcegraph -f $REMOTE_OVERLAY_URL
```

### Preview manifests

Run the following command in the directory where the `kustomization.yaml` file for your overlay is located to build a new set of manifests with your overlay. The updated manifests with your customization can then be found inside the `new/preview-cluster` directory.

```bash
$ kubectl kustomize $PATH_TO_OVERLAY -o new/preview-cluster
```

PATH_TO_OVERLAY can be a local path and remote path, for example:

```bash
# Local
$ kubectl kustomize new/overlays/deploy -o preview-cluster.yaml
# Remote
$ kubectl kustomize https://github.com/sourcegraph/deploy-sourcegraph/new/overlays/quick-start/k3s/xs?ref=v4.4.0 -o preview-cluster.yaml
```

> NOTE: This command will build a new set of manifests based on your overlay. It does not affect your current deployment until you run the apply command.

### Compare overlays

[kustomize v4.0.5](https://kubectl.docs.kubernetes.io/installation/kustomize/) is required

#### Between two overlays

To compare resources between two different Kustomize overlays:

```bash
$ diff \
    <(kustomize build $PATH_TO_OVERLAY_1) \
    <(kustomize build $PATH_TO_OVERLAY_2) |\
    more
```

Example 1: compare diff between resources generated by the k3s overlay for size xs instance and the k3s overlay for size xl instance:

```bash
$ diff \
    <(kustomize build new/overlays/quick-start/k3s/xs) \
    <(kustomize build new/overlays/quick-start/k3s/xl) |\
    more
```

Example 2: compare diff between the old base cluster and the old base cluster built with the new Kustomize setup (both with default values):

```bash
$ diff \
    <(kustomize build new/overlays/quick-start/current) \
    <(kustomize build new/overlays/quick-start/old-base/default) |\
    more
```

#### Between an overlay and a running cluster

Compare diff between the manifests generated by an overlay and the resources that are being used by the running cluster connected to the kubectl tool:

```bash
kubectl kustomize $PATH_TO_OVERLAY | kubectl diff -f  -
```

Example: compare diff between the k3s overlay for size xl instance and the instance that is connected with `kubectl`:

```bash
kubectl kustomize new/overlays/quick-start/k3s/xl | kubectl diff -f  -
```

### Kustomize with Helm

Kustomize can be used **with** Helm to configure Sourcegraph (see [this guidance](helm.md#integrate-kustomize-with-helm-chart)) but this is only recommended as a temporary workaround while Sourcegraph adds to the Helm chart to support previously unsupported customizations.

## Deprecated

The previous Kustomize structure we built for our Kubernetes deployments depends on scripting to create deployment manifests. It does not provide flexibility and requires direct changes made to the base manifests that can now be avoided with using the new Kustomize we have introduced in this documentation.

The previous version of the Sourcegraph Kustomize Overlays are still supported but should not be used for any new Kubernetes deployment.

> NOTE: The latest version of our Kustomize overlays does not work on instances that are v4.4.0 or older.

❌ See the [docs for the soon-to-be-deprecated version of Kustomize for Sourcegraph](deprecated.md).


### Migration from the old Kustomize

@TODO

## RBAC

@TODO
