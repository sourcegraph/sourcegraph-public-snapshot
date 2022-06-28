# Sourcegraph with Kubernetes & Helm


<div class="cta-group">
<a class="btn btn-primary" href="#prerequisites">Prerequisites</a>
<a class="btn" href="#installing-sourcegraph-with-kubernetes">Installation</a>
<a class="btn" href="#configuration">Configuration</a>
<a class="btn" href="#upgrading-sourcegraph">Upgrading</a>
</div>

> ⚠️ WARNING: Sourcegraph currently does not support migration from an existing Kubernetes Sourcegraph deployment without Helm to a Sourcegraph deployment using Helm. This guide is recommended for new installations of Sourcegraph. We are currently working to provide migration guidance from a non-Helm deployment. If you are inquiring about performing such a migration please email <support@sourcegraph.com>

## Why Helm?

Sourcegraph's Helm chart is the recommended way to install and configure Sourcegraph on Kubernetes. Helm charts make it simple to package and deploy applications on Kubernetes. Our Helm chart offers a lot of defaults in the `values.yaml` which makes customizations much easier than using Kustomize or manually editing Sourcegraph's manifest files. When using Helm chart override files to make customizations, you _never_ have to deal with merge conflicts during upgrades.

To deploy Sourcegraph with Kubernetes and Helm you will typically follow these steps:

1. Configure a cloud or bare metal instance and ensure you have access to launch persistent volumes.
2. Add the Sourcegraph Helm repository.
3. Prepare any additional required customizations. Sourcegraph offers out of the box defaults but most environments will likely to need to implement their own customizations. For additional guidance see the [Configuration](#configuration) section below.
4. Review the changes. We offer guidance on [three mechanisms](#reviewing-changes) that can be used to review customizations. This is an optional step, but may be useful the first time you deploy Sourcegraph.
5. Install the Helm chart

## Prerequisites

Deploying Sourcegraph with Kubernetes with Helm has the following requirements:

- A [Sourcegraph Enterprise license](configure.md#add-license-key) if your instance will have more than 10 users.
- A basic understanding of [Helm charts and how to create them.](https://helm.sh/)
- Access to a running Kubernetes cluster
  - Using a minimum Kubernetes version of [v1.19](https://kubernetes.io/blog/2020/08/26/kubernetes-release-1.19-accentuate-the-paw-sitive/)
  - Have the [kubectl command line](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed and are using v1.19 or later
- Have the [Helm 3 CLI](https://helm.sh/docs/intro/install/) installed
- A cloud account (EKS, AKS, GKE etc.) or access to a bare metal instance with ability to launch instances and persistent volumes (SSDs recommended).

## Quickstart

> ℹ️ The quickstart guide should be used by those already familiar with Helm and who want to quickly deploy Sourcegraph with Helm with the default configuration. If the default configuration settings do not fit your needs, read more about applying customizations in [Configuring Sourcegraph](../../config/index.md) below before installing the Helm chart.

To use the Helm chart, add the Sourcegraph Helm chart repository on the machine used to interact with your cluster:

```sh
helm repo add sourcegraph https://helm.sourcegraph.com/release
```

Install the latest release of the Sourcegraph Helm chart using default values:

```sh
helm install --version 3.41.0 sourcegraph sourcegraph/sourcegraph
```

Sourcegraph should now be available via the address set. Navigating to the URL will provide access to the Sourcegraph UI to create the initial administrator account.

## Installing Sourcegraph with Kubernetes

This section provides high-level guidance on deploying Sourcegraph via Kubernetes with Helm on major Cloud providers. In general, you need the following to get started:

- A working Kubernetes cluster, v1.19 or higher
- The ability to provision persistent volumes (e.g. have Block Storage [CSI storage driver](https://kubernetes-csi.github.io/docs/drivers.html) installed)
- An Ingress Controller installed (e.g. platform native ingress controller, [NGINX Ingress Controller])
- The ability to create DNS records for Sourcegraph (e.g. `sourcegraph.company.com`)

You can install Sourcegraph on the supported virtualization platform of your choice. Follow these links for cloud-specific guides on preparing the environment and installing Sourcegraph:

- [Deploy Sourcegraph with Kubernetes on Amazon Elastic Kubernetes Service (EKS)](./helm-cloud-deployment-guides/eks.md)
- [Deploy Sourcegraph with Kubernetes on Google Kubernetes Engine (GKE)](./helm-cloud-deployment-guides/gke.md)
- [Deploy Sourcegraph with Kubernetes on Azure Managed Kubernetes Service (AKS)](./helm-cloud-deployment-guides/aks.md)
- [Deploy Sourcegraph with Kubernetes on other Cloud providers or on-prem](./helm-cloud-deployment-guides/other-cloud.md)

## Configuration

The Sourcegraph Helm chart is highly customizable to support a wide range of environments. We highly recommend that customizations be applied using an override file, which allows customizations to persist through upgrades without needing to manage merge conflicts.

The default configuration values can be viewed in the [values.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/values.yaml) file along with all [supported options](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph#configuration-options).

To create customized configurations

1. Create an empty yaml file (e.g. `override.yaml`) to store the settings. We recommend that the override file be maintained in a version control system such as GitHub, but for testing, this can be created on the machine from which the Helm deployment commands will be run.
2. Use our example overrides as a boilerplate for your override file. A list of examples can be found in the [examples](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples) folder in Github.

> WARNING: __DO NOT__ copy the [default values file](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/values.yaml) as a boilerplate for your override file. You risk having outdated values during future upgrades. Instead, only include the configuration that you need to change and override.

3. Adapt any customizations from the example overrides.
4. Review the changes that will be applied - see [Reviewing Changes](#reviewing-changes).
5. Apply the changes by installing the chart with the custom values in your override file. Provide the override file to Helm by adding the values flag and the name of the override file in the following command:

```sh
helm upgrade --install --values ./override.yaml --version 3.41.0 sourcegraph sourcegraph/sourcegraph
```

For more information our common Helm chart configurations see our [Examples](./helm-examples.md).

### Advanced Helm chart configurations

Sourcegraph's Helm chart is new and our defaults in our `values.yaml` and customizations available via an override file may not cover every need. Equally, some changes are environment or customer-specific, and so will never be part of the default Sourcegraph Helm chart.

The following guidance for using Kustomize with Helm and Helm Subcharts covers both of these scenarios.

 ***[TODO Do we really want to recommend Kustomize with Helm? If we recommend customers reach out to CE or support do we really want this section to be externally documented?]***

> ⚠️ All deployment changes that are not covered by Sourcegraph's documentation or existing functionality should be discussed with your Customer Engineer or Support. Please contact your Customer Engineer before proceeding to ensure the changes proposed can be supported by Sourcegraph.

#### Integrate Kustomize with Helm chart

For advanced users who are looking for a temporary workaround, we __recommend__ applying [Kustomize](https://kustomize.io) on the rendered manifests from our chart. 
>  ⚠️ __Do not__ maintain your own fork of our chart as this may negatively impact our ability to provide support if you run into issues.

You can learn more about how to integrate Kustomize with Helm from our [example](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/kustomize-chart) on Github.

#### Helm subcharts

[Helm subcharts](https://helm.sh/docs/chart_template_guide/subcharts_and_globals/) can be used for permanent customizations to the official Sourcegraph helm chart. This is useful for changes such as adding a new resource unique to your deployment (PodSecurityPolicy, NetworkPolicy, additional services, etc.). These are long-lived customizations that shouldn't be contributed back to the Sourcegraph helm chart.

With a subchart, you create your own helm chart and specify the Sourcegraph chart as a dependency. Any resources you place in the templates folder of your chart will be deployed, as well as the Sourcegraph resources, allowing you to extend the Sourcegraph chart without maintaining a fork.

An example of a subchart can be found in the [examples/subchart](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples) folder in Github. More details on how to create and configure a subchart can be found in the [helm documentation](https://helm.sh/docs/chart_template_guide/subcharts_and_globals).



## Upgrading Sourcegraph

A new version of Sourcegraph is released every month (with patch releases in between, released as needed). Check the [Sourcegraph blog](https://about.sourcegraph.com/blog) for release announcements.

> ⚠️ WARNING: Skipping minor version upgrades of Sourcegraph is not supported. You must upgrade one minor version at a time - e.g. v3.26 –> v3.27 –> v3.28.

### Upgrading

1. Review [Helm Changelog] and [Sourcegraph Changelog] and select the most recent version compatible with your current Sourcegraph version.

> ⚠️ You can only upgrade one minor version of Sourcegraph at a time.

1. Update your copy of the Sourcegraph Helm repo to ensure you have all the latest versions:

   ```bash
      helm repo update sourcegraph
   ```

1. (Optional) Review the changes that will be applied - see [Reviewing Changes](#reviewing-changes) for options.

1.  Install the new version:

   ```bash
      helm upgrade --install -f override.yaml --version 3.41.0 sourcegraph sourcegraph/sourcegraph
   ```

1.  Verify the installation has started:

   ```bash
      kubectl get pods --watch
   ```

   When all pods have restarted and show as Running, you can browse to your Sourcegraph deployment and login to verify the instance is working as expected. For troubleshooting, refer to the [Operations guide](https://docs.sourcegraph.com/admin/install/kubernetes/operations) for common commands to gather more information about failures.

### Rollback

You can revert to a previous version with the following command:

   ```bash
      helm rollback sourcegraph
   ```

Sourcegraph only supports rolling back one minor version, due to database compatibility guarantees.

### Database migrations

By default, database migrations will be performed during application startup by a `migrator` init container running prior to the `frontend` deployment. These migrations **must** succeed before Sourcegraph will become available. If the databases are large, these migrations may take a long time.

In some situations, administrators may wish to migrate their databases before upgrading the rest of the system to reduce downtime. Sourcegraph guarantees database backward compatibility to the most recent minor point release so the database can safely be upgraded before the application code.

To execute the database migrations independently, you can use the [Sourcegraph Migrator] helm chart.

## Reviewing changes

When configuring an override file or performing an upgrade, we recommend reviewing the changes before applying them.

### Using `helm template`

The helm template command can be used to render manifests for review and comparison. This is particularly useful to confirm the effect of changes to your override file. This approach does not require access to the Kubernetes server.

For example:

1. Render the initial manifests from your existing deployment setup to an output file:

   ```bash
      CHART_VERSION=0.7.0 # Currently deployed version
      helm template sourcegraph -f override.yaml --version $CHART_VERSION sourcegraph sourcegraph/sourcegraph > original_manifests
   ```

1. Make changes to your override file, and/or update the chart version, then render that output:

   ```bash
      CHART_VERSION=3.39.0 # Not yet deployed version
      helm template sourcegraph -f override.yaml --version $CHART_VERSION sourcegraph sourcegraph/sourcegraph > new_manifests
   ```

1. Compare the two outputs:

   ```bash
      diff original_manifests new_manifests
   ```

### Using `helm upgrade --dry-run`

Similar to `helm template`, the `helm upgrade --dry-run` command can be used to render manifests for review and comparison. This requires access to the Kubernetes server but has the benefit of validating the Kubernetes manifests.

The following command will render and validate the manifests:

   ```bash
      helm upgrade --install --dry-run -f override.yaml sourcegraph sourcegraph/sourcegraph
   ```

Any validation errors will be displayed instead of the rendered manifests.

If you are having difficulty tracking down the cause of an issue, add the `--debug` flag to enable verbose logging:

   ```bash
      helm upgrade --install --dry-run --debug -f override.yaml sourcegraph sourcegraph/sourcegraph
   ```

The `--debug` flag will enable verbose logging and additional context, including the computed values used by the chart. This is useful when confirming your overrides have been interpreted correctly.

### Using `helm diff` plugin

The [Helm Diff] plugin can provide a diff against a deployed chart. It is similar to the `helm upgrade --dry-run` option but can run against the live deployment. This requires access to the Kubernetes server.

To install the plugin, run:

   ```bash
      helm plugin install https://github.com/databus23/helm-diff
   ```

Then, display a diff between a live deployment and an upgrade, with 5 lines of context:

   ```bash
      helm diff upgrade -f override.yaml sourcegraph sourcegraph/sourcegraph -C 5
   ```

For more examples and configuration options, reference the [Helm Diff] plugin documentation.

## Uninstalling Sourcegraph

Sourcegraph can be uninstalled by running the following command:

```sh
helm uninstall sourcegraph
```

Some Persistent Volumes may be retained after the uninstall is complete. In your cloud provider, check for unattached disks and delete them as necessary.

[backendconfig]: https://cloud.google.com/kubernetes-engine/docs/how-to/ingress-features#create_backendconfig
[azure application gateway]: https://docs.microsoft.com/en-us/azure/application-gateway/overview
[Container-native load balancing]: https://cloud.google.com/kubernetes-engine/docs/how-to/container-native-load-balancing
[Compute Engine persistent disk]: https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver
[AWS Load Balancer Controller]: https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html
[AWS EBS CSI driver]: https://docs.aws.amazon.com/eks/latest/userguide/managing-ebs-csi.html
[NGINX Ingress Controller]: https://github.com/kubernetes/ingress-nginx
[Helm Changelog]: https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/CHANGELOG.md
[Sourcegraph Changelog]: https://github.com/sourcegraph/sourcegraph/blob/main/CHANGELOG.md
[Sourcegraph Migrator]: https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph-migrator
[Helm Diff]: https://github.com/databus23/helm-diff
[Secrets]: https://kubernetes.io/docs/concepts/configuration/secret/
