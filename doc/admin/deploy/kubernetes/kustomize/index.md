# Kustomize

[Kustomize](https://kustomize.io) is a tool that is integrated with kubectl, which enables users to customize Kubernetes objects using [kustomization.yaml files](#kustomization-yaml-template). These files allow users to configure untemplated YAML files by providing instructions defined in [overlays](#overlays), resulting in the generation of a new set of resources without changing the original source files. 

During its build process, Kustomize will first build the resources from the base layer of the application. If generators are used, it will then create ConfigMaps and Secrets. Next, Kustomize will apply patches specified by the components to selectively overwrite resources in the base layer. Finally, it will perform validation to create a customized deployment.

## Prerequisites

- [Kubernetes cluster access](https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/) with `kubectl`
- A [private clone](../../repositories.md) of the [Sourcegraph Kubernetes deployment repository](./index.md#deployment-repository)
- Determine your instance size using our [instance size chart](../../instance-size.md)
- A [configured](configure.md) Sourcegraph Kustomize Overlay
  - see details below to learn more about building a customized Overlay that can be reused across updates

## Configure

Please refer to our [configuration guides](configure.md) for detailed instructions on building an overlay for Sourcegraph.

## Deploy

**Step 1:** Follow the instructions below to build an overlay with [custom configurations](configure.md) for your deployment environment. Alternatively, you can use one of our [pre-built overlays](#pre-built-overlays) to deploy Sourcegraph into a pre-configured cluster.

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

---

## Overview

_Components are evaluated after the resources of the parent kustomization (overlay or component) have been accumulated, and on top of them. ([source](https://sourcegraph.com/github.com/kubernetes/enhancements@master/-/blob/keps/sig-cli/1802-kustomize-components/README.md#proposal))_

### Deployment repository

The [new/base](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/base) directory in our [deployment repository](https://github.com/sourcegraph/deploy-sourcegraph) contains the Kubernetes manifests for Sourcegraph that form the **base** layer, which is maintained by Sourcegraph. The [new/components](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/base) directory contains various pre-configured reusable components that can be used to customize the base for different purposes while leaving the base files untouched. These components include additional resources and patches that can be applied to the base component to customize for a particular environment or use case. Most of the components can be used together to create a customized deployment on top of the original files for different environments based on where and how they will be deployed. 

The [new/overlays/deploy](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/overlays/deploy) directory is the recommended path to create and store a customized overlay for your deployment. If you would like to set up two seperated instances (ex. create two overlays for `production` and `staging` purposes), it is recommended to create a seperated directory within the [new/overlays/](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/overlays/deploy) directory using the files in [new/overlays/template](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/overlays/deploy).

### kustomization.yaml template

Below is a `kustomization` file we use as a template to deploy the default Sourcegraph instance. You can build on top of this template by adding components listed in our [configuration guides](configure.md) to create a tailored Sourcegraph instance.

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

### Overlays

An [overlay](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/#bases-and-overlays) acts as a **customization layer** that contains a [kustomization  file](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/) where **components** are defined as the **configuration layer** to include a set of instructions for Kustomize to generate and apply configurations to the **base layer**

#### Pre-built overlays

We have a sets of pre-built overlays that are ready-to-use for clusters that do not require additional configurations for Sourcegraph to be deployed. You can find the complete list of pre-built overlays inside the `new/overlays/quick-start` directory.

Please see our [configuration docs for Kustomize](./configure.md) on using components to build a overlay that is tailored to your specific need.

### Components

To understand an overlay is to examine its components, which listed under the `components` field inside the `kustomization` file of an overlay. Our Kustomize components are a set of configurations that utilize [transformers](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/#everything-is-a-transformer) to apply customization to the base layers. Some of them are designed to be reusable for different environments and use cases.

#### Pre-built components

The components located inside the [new/components](ttps://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/components) directory are pre-configured and ready-to-use.

If you would like to modify a component from this directory, it is strongly recommend to do it in a duplicate of the said component inside the [new/config directory](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/config). which is the destinated directory for components that require additional configurations. Adding them to your overlay without following the instructions listed in our configuration docs could result in errors during the build stage or deploy stage.

### Examples

Here is an example `kustomization` file from one of our pre-built overlays:

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

See the complete list of ready-to-use overlays inside the `new/overlays/quick-start` directory to learn more about combining different pre-configured components to build an overlay to configure your Sourcegraph instance to work in your environment.

### Remote build

Remote build allows you to deploy an overlay using a git URL to build and deploy; however, it does not support custom configurations as the resources are hosted remotely.

To create manifests using a remote overlay:

```bash
# Replace the $REMOTE_OVERLAY_URL with a URL of an overlays.
$ kubectl kustomize $REMOTE_OVERLAY_URL -o preview-cluster.yaml
```

The manifests will be grouped into a single file `preview-cluster.yaml` once the overlay is built and validated. You can then preview the resources before running the apply command to deploy using the remote overlay:

```bash
$ kubectl apply -k --prune -l deploy=sourcegraph -f $REMOTE_OVERLAY_URL
```

### Preview manifests

Run the following command from the root directory to build a customized deployment using your overlay. 

A new set of manifests with the customization applied using the overlay can then be found inside the [new/preview-cluster](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/preview-cluster) directory.

```bash
$ kubectl kustomize $PATH_TO_OVERLAY -o new/preview-cluster
```

The $PATH_TO_OVERLAY can be a local path or remote path, for example:

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
