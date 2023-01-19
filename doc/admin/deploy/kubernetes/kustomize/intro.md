
# Introduction to Kustomize for Sourcegraph

 Kustomize enables us to decompose our **[base](#base)** application into smaller building blocks, with multiple versions of each block preconfigured as **[components](#components)** for various use cases. This modular approach enables the mixing and matching of the building blocks to construct a customized version of the application by creating **[overlays](#overlay)**. This feature provides a high degree of flexibility and facilitates the maintenance and evolution of the application over time.

## Build process

During the build process, Kustomize will first build the resources from the base layer of the application. If generators are used, it will then create ConfigMaps and Secrets. These resources can be generated from files, or from data stored in ConfigMaps, or from image metadata.

Next, Kustomize will apply patches specified by the components to selectively overwrite resources in the base layer. Patching allows you to modify the resources defined in the base layer without changing the original source files. This is useful for making small, targeted changes to the resources that are needed for your specific deployment.

Finally, Kustomize will perform validation to ensure that the modified resources are valid and conform to the Kubernetes API. This is to ensure that the customized deployment is ready for use. Once the validation is passed, the modified resources are grouped into a single file, known as the output. After that, you can use kubectl to apply the overlaid resources to your cluster.

## Deployment repository

All the Kustomize resources for Sourcegraph are located inside the **new** directory of our deployment repository.

Here is the file structure:

```bash
# github.com/sourcegraph/deploy-sourcegraph
~/new
├── base
│ ├── sourcegraph
│ └── monitoring
├── components
└── overlays
  └── template
    ├── config
    │  └── overlay.config
    ├── env
    │  └── frontend.env
    ├── utils
    └── kustomization.yaml
```

> WARNING: Please create your own sets of overlays within the 'overlays' directory and refrain from making changes to the other directories to prevent potential merge conflicts during future updates.

## Base

**Base** refers to a set of YAML files created for the purpose of deploying a Sourcegraph instance to a Kubernetes cluster. These files come preconfigured with default values that can be used to quickly deploy a Sourcegraph instance. 

However, deploying with these default settings may not be suitable for all environments and specific use cases. For example, the default resource allocation may not match the requirements for your specific instance size, or the default deployments may include RBAC resources that you would like to remove. To address these issues, creating a Kustomize overlay can be an effective solution. It allows you to customize the resources defined in the base layer to suit the specific requirements of the deployment.

## Overlays

An [overlay](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/#bases-and-overlays) acts as a **customization layer** that contains a [kustomization file](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/), where components are defined as the **configuration layer** to include a set of instructions for Kustomize to generate and apply configurations to the **base layer**.

Here is an example of a `kustomization` file from one of our [pre-built overlays](#pre-built-overlays) that is configured for deploying a size XS instance to a k3s cluster:

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

### Pre-built overlays

 In addition to providing a template for creating new overlays, we also provide a set of pre-built overlays that are pre-configured for different environments. These pre-built overlays can be found inside the [new/quick-start](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/quick-start) directory. 

## Overlay

In this section, we will take a quick look at the essential part that make up a Kustomize overlay tailored for Sourcegraph.

### Template

An overlay is a directory that contains various files used to configure a deployment for a specific scenario. These files include the kustomization.yaml file, which is used to specify how the resources defined in the base manifests should be customized and configured, as well as other files such as environment variable files, configuration files, and patches.

All overlays should be stored in the [new/overlays directory](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/overlays), where you can find the [new/overlays/template folder](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/overlays/template). This folder contains a selection of necessary files to construct an overlay for Sourcegraph. You can duplicate this folder as needed to create new overlays for deploying Sourcegraph.

### Configuration files

Below are the configuration files we have defined to make configuring Sourcegraph with an overlay easier.

#### kustomization.yaml

 The kustomization.yaml file is a key component of a Kustomize overlay. It is located in the root of an overlay directory and is used to specify how the resources defined in the base manifests should be customized and configured according to our [configuration docs](./configure.md).

#### env and config directories

The **env directory**  and **config directory** are designated locations for storing configuration files related to your deployment.

- The **env directory** is intended to contain files that will be used by the deployed instance, such as frontend.env, tls.cert, tls.key, and ssh files.
- The **config directory** is intended to contain files that are used to customize the deployment using Kustomize, such as Kustomize overlays, kustomization.yaml files, and other configuration files used to patch the deployment.

#### frontend.env

The frontend.env file is used to update environment variables for the sourcegraph-frontend service. These environment variables can be used to configure various aspects of the frontend service, such as authentication and authorization settings, feature flags, and more.

Only update the frontend.env file if instructed by the components defined in your overlay

#### overlay.config

Some components may require additional input from users to construct the overlay. These inputs are typically configurations that are specific to the user's environment, use case, or preferences. These configurations are typically stored in an overlay.config file, which is located in the same directory as the overlay's kustomization.yaml file.

It's important to only update the overlay.config file if instructed by the components defined in your overlay. This is because not all components require additional configurations, and some may even have default values that are suitable for most use cases. Updating the overlay.config file unnecessarily can cause errors or unexpected behavior. Always refer to the component's documentation or the comments in the kustomization.yaml file before making changes to the overlay.config file.

Also, it's worth to mention that you can store your configuration in different ways and not just in a file named overlay.config, It's depending on your preference, it can be stored in an environment variable, a configmap, or a secret.

## Components

An overlay in Kustomize is a set of configuration files that are used to customize the base resources. To understand an overlay, it's important to examine its components, which are listed under the components field inside the kustomization.yaml file of the overlay.

Most of our components are designed to be reusable for different environments and use cases. They can be used to add common labels and annotations, apply common configurations, or even generate resources. By using these components, you can minimize the amount of duplicated code in your overlays and make them more maintainable.


## Remote build

Remote build feature allows you to deploy an overlay using a git URL, but it should be noted that it does not support custom configurations as the resources are hosted remotely. 

To create manifests using a remote overlay, you can use the following command:

```bash
# Replace the $REMOTE_OVERLAY_URL with a URL of an overlays.
$ kubectl kustomize $REMOTE_OVERLAY_URL -o generated-cluster.yaml
```

This command will download the overlay specified in the $REMOTE_OVERLAY_URL and apply the customizations to the base resources and output the resulting customized manifests to the file generated-cluster.yaml. This command allows you to preview the resources before running the apply command below to deploy using the remote overlay.


```bash
$ kubectl apply -k --prune -l deploy=sourcegraph -f $REMOTE_OVERLAY_URL
```

## Preview manifests

To create a customized deployment using your overlay, run the following command from the root directory of your deployment repository.

```bash
$ kubectl kustomize $PATH_TO_OVERLAY -o new/generated-cluster.yaml
```

This command will apply the customizations specified in the overlay located at $PATH_TO_OVERLAY to the base resources and output the resulting customized manifests to the file `new/generated-cluster.yaml`.

The $PATH_TO_OVERLAY path can be a local path or remote path. For example:

```bash
# Local
$ kubectl kustomize new/overlays/deploy -o generated-cluster.yaml
# Remote
$ kubectl kustomize https://github.com/sourcegraph/deploy-sourcegraph/new/quick-start/k3s/xs?ref=v4.5.0 -o generated-cluster.yaml
```

> NOTE: This command will only generate the customized manifests and will not apply them to the cluster. . It does not affect your current deployment until you run the apply command.

## Compare overlays

[kustomize v4.0.5](https://kubectl.docs.kubernetes.io/installation/kustomize/) is required

Below are the commands that will output the differences between the two overlays, allowing you to review and compare the changes and ensure that the new overlay produces similar resources as the ones currently being used by the active cluster or another overlay you want to compare with, before applying the new overlay. 

### Between two overlays

To compare resources between two different Kustomize overlays:

```bash
$ diff \
    <(kubectl kustomize $PATH_TO_OVERLAY_1) \
    <(kubectl kustomize $PATH_TO_OVERLAY_2) |\
    more
```

Example 1: compare diff between resources generated by the k3s overlay for size xs instance and the k3s overlay for size xl instance:

```bash
$ diff \
    <(kubectl kustomize new/quick-start/k3s/xs) \
    <(kubectl kustomize new/quick-start/k3s/xl) |\
    more
```

Example 2: compare diff between the deprecated cluster and the deprecated cluster built with the new Kustomize setup (both with default values):

```bash
$ diff \
    <(kubectl kustomize new/quick-start/deprecated) \
    <(kubectl kustomize new/quick-start/old-base/default) |\
    more
```

Example 3: compare diff between the output files from two different overlay builds:

```bash
$ diff new/generated-cluster-old.yaml new/generated-cluster-new.yaml
```

### Between an overlay and a running cluster

To compare the difference between the manifests generated by an overlay and the resources that are being used by the running cluster connected to the kubectl tool:

```bash
kubectl kustomize $PATH_TO_OVERLAY | kubectl diff -f  -
```

The command will output the differences between the customizations specified in the overlay and the resources currently running in the cluster, allowing you to review the changes and ensure that the overlay produces similar resources as the ones currently being used by the active cluster before applying the new overlay.

Example: compare diff between the k3s overlay for size xl instance and the instance that is connected with `kubectl`:

```bash
kubectl kustomize new/quick-start/k3s/xl | kubectl diff -f  -
```

## Kustomize with Helm

Kustomize can be used in conjunction with Helm to configure Sourcegraph, as outlined in [this guidance](helm.md#integrate-kustomize-with-helm-chart). However, this approach is only recommended as a temporary workaround while Sourcegraph adds support for previously unsupported customizations in its Helm chart. This means that using Kustomize with Helm is not a long-term solution.

## Deprecated

The previous Kustomize structure we built for our Kubernetes deployments depends on scripting to create deployment manifests. It does not provide flexibility and requires direct changes made to the base manifests that can now be avoided with using the new Kustomize we have introduced in this documentation.

The previous version of the Sourcegraph Kustomize Overlays are still supported but should not be used for any new Kubernetes deployment.

> NOTE: The latest version of our Kustomize overlays does not work on instances that are v4.5.0 or older.

❌ See the [docs for the soon-to-be-deprecated version of Kustomize for Sourcegraph](deprecated.md).


## Migrating from deploy scripts

Prior to version 4.5.0, custom scripts were used for deploying Sourcegraph with Kubernetes. However, as of version 4.5.0, this method is now [deprecated](deprecated.md). . It is important to note that the transition from the older deployment scripts to the new Sourcegraph Kustomize setup is relatively straightforward, as the older scripts utilize Kustomize internally.

It's crucial to note that both tools are used for generating manifests for deployment and will not alter existing resources in an active cluster. The objective is to produce a new overlay that generates a similar set of resources as the ones currently used in the running cluster. This will ensure that the deployment process is smooth and does not disrupt the existing cluster resources. This change will provide an improved and more maintainable way of deploying Sourcegraph on Kubernetes clusters.

### Old cluster vs new cluster

As of the latest version, the new Sourcegraph cluster runs in non-root and non-privileged mode by default. This change was implemented by recreating the cluster using a modified version of the previous [non-privileged](https://github.com/sourcegraph/deploy-sourcegraph/tree/v4.3.0/overlays/non-privileged) overlay and [non-privileged-create-cluster](https://github.com/sourcegraph/deploy-sourcegraph/tree/v4.3.0/overlays/non-privileged-create-cluster) overlay. This modification was made to ensure that the Sourcegraph cluster is running in a more secure and stable environment.

For instructions on deploying Sourcegraph with privileged access, please refer to the [configuration docs](configure.md).

### Migration process

The migration process for transitioning to the new Sourcegraph Kustomize setup involves the following steps:

**Step 1**: Create a new release branch from your current release branch (must be on v4.5.0 or above)

**Step 2**: Use the instructions detailed in the [configuration docs for Kustomize](./configure.md) to create a new overlay for Sourcegraph

**Step 3**: [Compare the manifests](#between-an-overlay-and-a-running-cluster) generated by your new overlay with the ones in your running cluster using the command:

```bash
kubectl kustomize $PATH_TO_OVERLAY | kubectl diff -f  -
```

Review the changes to ensure that the manifests generated by your new overlay are similar to the ones currently being used by your active cluster.

**Step 4**: Once you are satisfied with the overlay output, you can now deploy the new overlay using the command:

```bash
kubectl kustomize $PATH_TO_OVERLAY/. | kubectl apply --prune -l deploy=sourcegraph -f -
```

It's important to review the changes, and ensure that the new overlay produces similar resources as the ones currently being used by the active cluster, before applying the new overlay.

Note: Make sure to test the new overlay and the migration process in a non-production environment before applying it to your production cluster.
