# Kustomize

[Kustomize](https://kustomize.io) is a tool that is integrated with kubectl, which enables users to customize Kubernetes objects using configuration files named `kustomization.yaml`. These files can be found in all [overlays](#overlays). They contain a set of instructions for Kustomize to to configure untemplated YAML files, resulting in the generation of a new set of resources without changing the original source files. 

During its build process, Kustomize will first build the **resources** from the base layer of the application. If generators are used, it will then create ConfigMaps and Secrets. Next, Kustomize will apply patches specified by the components to selectively overwrite resources in the base layer. Finally, it will perform validation to create a customized deployment.

## Prerequisites

See [prerequisites for Kubernetes](../index.md#prerequisites).

## Configure

Please refer to our [configuration guides](configure.md) for detailed instructions on building an overlay for a tailored Sourcegraph deployment.

## Deploy

Once you have met all the [prerequisites](../index.md#prerequisites):

**Step 0:** Install an ingress controller for your cluster as instructed in our [configuration guide](configure.md#ingress-controller) if applicable.

**Step 1:** Follow the instructions below to build an overlay with [custom configurations](configure.md) for your deployment environment. Alternatively, you can use one of our [pre-built overlays](#pre-built-overlays) to deploy Sourcegraph into a pre-configured cluster.

**Step 2:** Build the deployment manifests with the overlay prepared from step 1 for review

  ```bash
  $ kubectl kustomize $PATH_TO_OVERLAY -o new/generated-cluster.yaml
  ```

**Step 3:** Make sure the manifests in the output file `new/generated-cluster.yaml` are generated correctly

**Step 4:** Run the following command from the root of the cloned repository to apply the manifests generated from step 2 inside the `new/generated-cluster.yaml` file

  ```bash
  $ kubectl apply -k --prune -l deploy=sourcegraph -f new/generated-cluster.yaml
  ```

## Upgrade

To upgrade your instance with Kustomize, please refer to our [upgrade docs](../update.md#upgrades).

---

## Overview

_Components are evaluated after the resources of the parent kustomization (overlay or component) have been accumulated, and on top of them. ([source](https://sourcegraph.com/github.com/kubernetes/enhancements@master/-/blob/keps/sig-cli/1802-kustomize-components/README.md#proposal))_

### Deployment repository

Here is the layout of our **new** kustomize directory within our reference repository:

```txt
new
├── base
  └── sourcegraph
  └── monitoring
├── components
└── overlays
  └── template
    └── config
    ├── utils
    └── frontend.env
    └── kustomize.env
    └── kustomization.yaml
```

The [new/base](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/base) directory in our [deployment repository](https://github.com/sourcegraph/deploy-sourcegraph) contains the Kubernetes manifests for Sourcegraph that form the **base** layer, which is maintained by Sourcegraph. The [new/components](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/base) directory contains various pre-configured reusable components that can be used to customize the base for different purposes while leaving the base files untouched. These components include additional resources and patches that can be applied to the base component to customize for a particular environment or use case. Most of the components can be used together to create a customized deployment on top of the original files for different environments based on where and how they will be deployed. 

### Overlays

An [overlay](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/#bases-and-overlays) acts as a **customization layer** that contains a [kustomization  file](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/) where **components** are defined as the **configuration layer** to include a set of instructions for Kustomize to generate and apply configurations to the **base layer**

#### Pre-built overlays

We have a sets of pre-built overlays that are ready-to-use for clusters that do not require additional configurations for Sourcegraph to be deployed. You can find the complete list of pre-built overlays inside the `new/quick-start` directory.

### Overlay

#### Template

An overlay is a directory containing various files to configure your deployment for a specific scenario. The overlays should be stored in the `new/overlays directory`, where you can find the [new/overlays/template folder](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/overlays/template) containing a selection of necessary files to construct an overlay for Sourcegraph. The [new/overlays/template folder](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/overlays/template) can be duplicated as needed to create new overlays for deploying Sourcegraph.

#### Configuration files

Below are the configuration files defined by Kustomize and Sourcegraph:

##### kustomization.yaml

A kustomization.yaml file located in an overlay directory is used to specify how the resources defined in the base manifests should be customized or configured, following the instructions detailed in the [configuration docs for Kustomize](./configure.md).

##### frontend.env

The frontend.env is used to update environment variables for sourcegraph-frontend.

Update the file only if instructed by the component defined in your overlay.

##### kustomize.env

Certain components necessitate additional input from users to construct the overlay. The `kustomize.env` file is where the configurations needed by these components should be inputted.

Update the file only if instructed by the component defined in your overlay.


### Components

To understand an overlay is to examine its components, which listed under the `components` field inside the `kustomization` file of an overlay. Our Kustomize components are a set of configurations that utilize [transformers](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/#everything-is-a-transformer) to apply customization to the base layers. Some of them are designed to be reusable for different environments and use cases.

### Examples

Here is an example `kustomization` file from our quick-start overlay that is configured for deploying a size XS instance to a k3s cluster

```yaml
# new/quick-starts/k3s/xs/kustomization.yaml
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

See the complete list of ready-to-use overlays inside the `new/quick-start` directory to learn more about combining different pre-configured components to build an overlay to configure your Sourcegraph instance to work in your environment.

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

A new set of manifests with the customization applied using the overlay can then be found inside the [new/generated-cluster.yaml](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/generated-cluster.yaml) directory.

```bash
$ kubectl kustomize $PATH_TO_OVERLAY -o new/generated-cluster.yaml
```

The $PATH_TO_OVERLAY can be a local path or remote path, for example:

```bash
# Local
$ kubectl kustomize new/overlays/deploy -o preview-cluster.yaml
# Remote
$ kubectl kustomize https://github.com/sourcegraph/deploy-sourcegraph/new/quick-start/k3s/xs?ref=v4.4.0 -o preview-cluster.yaml
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
    <(kustomize build new/quick-start/k3s/xs) \
    <(kustomize build new/quick-start/k3s/xl) |\
    more
```

Example 2: compare diff between the old base cluster and the old base cluster built with the new Kustomize setup (both with default values):

```bash
$ diff \
    <(kustomize build new/quick-start/current) \
    <(kustomize build new/quick-start/old-base/default) |\
    more
```

#### Between an overlay and a running cluster

Compare diff between the manifests generated by an overlay and the resources that are being used by the running cluster connected to the kubectl tool:

```bash
kubectl kustomize $PATH_TO_OVERLAY | kubectl diff -f  -
```

Example: compare diff between the k3s overlay for size xl instance and the instance that is connected with `kubectl`:

```bash
kubectl kustomize new/quick-start/k3s/xl | kubectl diff -f  -
```

### Kustomize with Helm

Kustomize can be used **with** Helm to configure Sourcegraph (see [this guidance](helm.md#integrate-kustomize-with-helm-chart)) but this is only recommended as a temporary workaround while Sourcegraph adds to the Helm chart to support previously unsupported customizations.

## Deprecated

The previous Kustomize structure we built for our Kubernetes deployments depends on scripting to create deployment manifests. It does not provide flexibility and requires direct changes made to the base manifests that can now be avoided with using the new Kustomize we have introduced in this documentation.

The previous version of the Sourcegraph Kustomize Overlays are still supported but should not be used for any new Kubernetes deployment.

> NOTE: The latest version of our Kustomize overlays does not work on instances that are v4.4.0 or older.

❌ See the [docs for the soon-to-be-deprecated version of Kustomize for Sourcegraph](deprecated.md).


### Migration from deploy scripts

Prior to v4.4.0, custom scripts are used for deploying Sourcegraph with Kubernetes, which is now [deprecated](deprecated.md). 

The transition from the older deployment scripts to the new Sourcegraph Kustomize setup is straightforward, since the older scripts utilize Kustomize internally. However, it's crucial to note that both tools are utilized for generating manifests for deployment and **do not alter existing resources in an active cluster**. Therefore, the objective is to produce a new overlay that generates a similar set of resources as the ones currently utilized in the running cluster.

#### Old overlays vs new overlays

The new Sourcegraph base cluster now runs in `non-privileged` mode by default. It was created using the previous [non-privileged](https://github.com/sourcegraph/deploy-sourcegraph/tree/v4.3.0/overlays/non-privileged) and [non-privileged-create-cluster](https://github.com/sourcegraph/deploy-sourcegraph/tree/v4.3.0/overlays/non-privileged-create-cluster) overlays with other improvement.

If RBAC is currently enabled in your cluster, please refer to the configuration docs on how to deploy Sourcegraph with privileged access. 

#### Migration process

Step 1: Create a new release branch from your current release branch.

Step 2: Start creating a new overlay for Sourcegraph using the instructions detailed in our [configuration docs for Kustomize](./configure.md).

Step 3: [Compare the manifests](#between-an-overlay-and-a-running-cluster) generated by your new overlay with the ones in your running cluster:

```bash
kubectl kustomize $PATH_TO_OVERLAY | kubectl diff -f  -
```

Review the changes to make sure the manifests generated with your new overlay is similiar to the ones that are being used by your active cluster. 

Step 4: Once you are satisfy with the overlay output, you can now deploy using the new overlay:

```bash
kustomize build $PATH_TO_OVERLAY/. | kubectl apply --prune -l deploy=sourcegraph -f -
```
