# Kustomize

[Kustomize](https://kustomize.io) is a Kubernetes configuration tool integrated with kubectl. It allows users to customize Kubernetes objects using [kustomization files](#kustomizationyaml) that contain a set of instructions for Kustomize to generate a new set of resources without altering the original source files.

<div class="getting-started">
<a class="btn btn-primary text-center" href="#">★ Installation</a>
<a class="btn text-center" href="intro">Introduction</a>
<a class="btn text-center" href="configure">Configuration</a>
<a class="btn text-center" href="../operations">Operation</a>
</div>

## Prerequisites

You must have all the [prerequisites](../index.md#prerequisites) installed and configured to deploy Sourcegraph with Kustomize.

## Configure

Please follow our [configuration guides](configure.md) to build an overlay for a tailored Sourcegraph deployment.

## Deploy

Follow the steps below to deploy Sourcegraph on a Kubernetes cluster.

**Step 0:** Ensure all the [prerequisites](../index.md#prerequisites) have been met and are properly configured.

**Step 1:** Create an overlay by following the instructions in our [configuration guides](configure.md). As an alternative, you can choose to use one of our [pre-built overlays](#pre-built-overlays) that are already configured for your specific deployment environment.

**Step 2:** Use the overlay created in step 1 to generate a new set of manifests.

  ```bash
  $ kubectl kustomize $PATH_TO_OVERLAY -o new/generated-cluster.yaml
  ```

  A new set of manifests will be generated and grouped into a single output file `new/generated-cluster.yaml` for your review without applying to the cluster.

**Step 3:** Review the generated manifests to ensure they match your intended configuration.

**Step 4:**  Deploy the generated manifests

  ```bash
  $ kubectl apply --prune -l deploy=sourcegraph -f new/generated-cluster.yaml
  ```

### Examples

See the quick start examples for: [Amazon EKS](eks.md), [Google GKE](gke.md), and [other cloud providers](../index.md#quick-start).

## Upgrade

To upgrade your Sourcegraph Kubernetes instance, please refer to our [upgrade docs for Sourcegraph with Kubernetes](../update.md#upgrades).


