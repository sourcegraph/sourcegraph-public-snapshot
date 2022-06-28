# Sourcegraph with Kubernetes & Helm


<div class="cta-group">
<a class="btn" href="#prerequisites">Prerequisites</a>
<a class="btn btn-primary" href="#installing-sourcegraph-with-kubernetes">Installation</a>
<a class="btn" href="#configuration">Configuration</a>
<a class="btn" href="#upgrading-sourcegraph">Upgrading</a>
</div>

> WARNING: Sourcegraph currently does not support migration from an existing Kubernetes Sourcegraph deployment without Helm to a Sourcegraph deployment using Helm. This guide is recommended for new installations of Sourcegraph. We are currently working to provide migration guidance from a non-Helm deployment. If you are inquiring about performing such a migration please email <support@sourcegraph.com>

## Why use Helm

Sourcegraph's Helm chart is the recommended way to install and configure Sourcegraph on Kubernetes. Helm charts make it simple to package and deploy applications on Kubernetes. Our Helm chart offers a lot of defaults in the `values.yaml` which makes customizations much easier than using Kustomize or manually editing Sourcegraph's manifest files. When using Helm chart override files to make customizations, you _never_ have to deal with merge conflicts during upgrades.

To deploy Sourcegraph with Kubernetes and Helm you will typically follow these steps:

1. Configure a cloud or bare metal instance and ensure you have access to launch persistent volumes.
2. Add the Sourcegraph Helm repository.
3. Prepare any additional required customizations. Sourcegraph offers out of the box defaults but most environments will likely to need to implement their own customizations. For additional guidance see the [Configuration](#configuration) section below.
4. Review the changes. We offer guidance on [three mechanisms](#reviewing-changes) that can be used to review customizations. This is an optional step, but may be useful the first time you deploy Sourcegraph.
5. Install the Helm chart

## Prerequisites

Deploying Sourcegraph with Kubernetes with Helm has the following requirements:

- You must have a [Sourcegraph Enterprise license](configure.md#add-license-key) if your instance will have more than 10 users.
- You must have a basic understanding of [Helm charts and how to create them.](https://helm.sh/)
- You must have a running Kubernetes cluster ***[TODO link to docs]***
- You must be using a minimum Kubernetes version of [v1.19](https://kubernetes.io/blog/2020/08/26/kubernetes-release-1.19-accentuate-the-paw-sitive/) 
- You must have the [kubectl command line](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed and are using v1.19 or later
- You must have the [Helm 3 CLI](https://helm.sh/docs/intro/install/) installed
- You have a cloud account (EKS, AKS, GKE etc.) or access to a bare metal instance with ability to launch instances and persistent volumes (SSDs recommended).

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

- [Deploy Sourcegraph with Kubernetes on Amazon Elastic Kubernetes Service (EKS)](/k8-eks.md)
- [Deploy Sourcegraph with Kubernetes on Google Kubernetes Engine (GKE)](/k8-gke.md)
- [Deploy Sourcegraph with Kubernetes on Azure Managed Kubernetes Service (AKS)](/k8-aks.md)
- [Deploy Sourcegraph with Kubernetes on other Cloud providers or on-prem](/k8-other-cloud.md)

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

(The configuration override file can be created in advance of deployment, and the configuration override settings can be populated in preparation.)
 ***[TODO what does this mean. Ie. customers should do this first? What is the specific step we are referring to here?]***

### Common Helm Chart Configurations

This section outlines a few common scenarios for creating custom configurations. For an exhaustive list of configuration options please see our [Sourcegraph Helm Chart README.md in Github.](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/README.md)

- [Using external PostgreSQL databases](#using-external-postgresql-databases)
- [Using external Redis instances](#using-external-redis-instances)
- [Using external Object Storage](#using-external-object-storage)
- [Using SSH to clone repositories](#using-ssh-to-clone-repositories)

If your customization needs are not covered below please email <support@sourcegraph.com> for assistance.

#### Using external PostgreSQL databases

The default Sourcecgraph deployment ships three separate Postgres instances: [codeinsights-db.StatefulSet.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/templates/codeinsights-db/codeinsights-db.StatefulSet.yaml), [codeintel-db.StatefulSet.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/templates/codeintel-db/codeintel-db.StatefulSet.yaml), and [pgsqlStatefulSet.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/templates/pgsql/pgsql.StatefulSet.yaml). All three can be disabled individually and replaced with external Postgres instances.

To use external PostgreSQL databases, first review our [general recommendations](https://docs.sourcegraph.com/admin/external_services/postgres#using-your-own-postgresql-server) and [required postgres permissions](https://docs.sourcegraph.com/admin/external_services/postgres#postgres-permissions-and-database-migrations).

> An example of this approach can be found in our Helm chart [using external databases](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/external-databases) example on Github.

We recommend storing the database credentials in [Secrets] created outside of the helm chart and managed in a secure manner. The Secrets should be deployed to the same namespace as the existing Sourcegraph deployment. Each database requires its own Secret and should follow the following format. The Secret name can be customized as desired:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: pgsql-credentials
data:
  # notes: secrets data has to be base64-encoded
  database: ""
  host: "" # example: pgsql.database.example.com
  password: ""
  port: ""
  user: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: codeintel-db-credentials
data:
  # notes: secrets data has to be base64-encoded
  database: ""
  host: ""
  password: ""
  port: ""
  user: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: codeinsights-db-credentials
data:
  # notes: secrets data has to be base64-encoded
  database: ""
  host: ""
  password: ""
  port: ""
  user: ""
```

Set the Secret name your `override.yaml` by configuring the `auth.existingSecret` key for each database. A full example can be seen in this [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/examples/external-databases/override.yaml)

```yaml
codeIntelDB:
  enabled: false # disables deployment of the database
  auth:
    existingSecret: codeintel-db-credentials

codeInsightsDB:
  enabled: false
  auth:
    existingSecret: codeinsights-db-credentials

pgsql:
  enabled: false
  auth:
    existingSecret: pgsql-credentials
```

Although not recommended, credentials can also be configured directly in the helm chart. For example, add the following to your override.yaml to customize pgsql credentials:

```yaml
pgsql:
  enabled: false # disable internal pgsql database
  auth:
    database: "customdb"
    host: pgsql.database.company.com # external pgsql host
    user: "newuser"
    password: "newpassword"
    port: "5432"
```

#### Using external Redis instances

The default Sourcecgraph deployment ships two separate Redis instances: [redis-cache.Deployment.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/templates/redis/redis-cache.Deployment.yaml) and [redis-store.Deployment.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/templates/redis/redis-store.Deployment.yaml).

To use external Redis instances, first review our [general recommendations](https://docs.sourcegraph.com/admin/external_services/redis). When using external Redis instances, you’ll need to specify the new endpoint for each. You can specify the endpoint directly in the values file, or by referencing an existing secret.

> An example of this approach can be found in our Helm chart [using external redis](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/external-redis) example on Github.

If your external Redis instances do not require authentication, you can configure access in your [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/examples/external-redis/override.yaml) with the `endpoint` setting:

```yaml
redisCache:
  enabled: false
  connection:
    endpoint: redis://redis-cache.example.com:6379 # use a dedicated Redis, recommended

redisStore:
  enabled: false
  connection:
    endpoint: redis://redis-shared.example.com:6379/2 # shared Redis, not recommended
```

If your endpoints do require authentication, we recommend storing the credentials in [Secrets] created outside of the helm chart and managed in a secure manner.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: redis-cache-connection
data:
  # notes: secrets data has to be base64-encoded
  endpoint: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: redis-store-connection
data:
  # notes: secrets data has to be base64-encoded
  endpoint: ""
```

You can reference this secret in your [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/examples/external-redis/override-secret.yaml) by configuring the `connection.existingSecret` key:

```yaml
redisCache:
  enabled: false
  connection:
    existingSecret: redis-cache-connection

redisStore:
  enabled: false
  connection:
    existingSecret: redis-store-connection
```

#### Using external Object Storage

By default Sourcegraph uses a MinIO server bundled with the instance to temporarily store precise code intelligence indexes uploaded by users. If you prefer, you can configure your instance to store this data in an S3 or GCS bucket. To use an external Object Storage service (S3-compatible services, or GCS), first review our [general recommendations](https://docs.sourcegraph.com/admin/external_services/object_storage). Then review the following example and adjust to your use case.

To target a managed object storage service, you will need to set a handful of environment variables for configuration and authentication to the target service.

> An example of this approach can be found in our Helm chart [using external object storage](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/external-object-storage/override.yaml) example on Github.
> The example assumes the use of AWS S3. You may configure the environment variables accordingly for your own use case based on our [general recommendations](https://docs.sourcegraph.com/admin/external_services/object_storage).

If you provide credentials with an access key / secret key, we recommend storing the credentials in [Secrets] created outside of the helm chart and managed in a secure manner. An example Secret is shown here:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sourcegraph-s3-credentials
data:
  # notes: secrets data has to be base64-encoded
  PRECISE_CODE_INTEL_UPLOAD_AWS_ACCESS_KEY_ID: ""
  PRECISE_CODE_INTEL_UPLOAD_AWS_SECRET_ACCESS_KEY: ""
```

In your [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/external-object-storage/override.yaml), reference this Secret and the necessary environment variables:

```yaml

minio:
  enabled: false # Disable deployment of the built-in object storage

# we use YAML anchors and alias to keep override file clean
objectStorageEnv: &objectStorageEnv
  PRECISE_CODE_INTEL_UPLOAD_BACKEND:
    value: S3 # external object storage type, either "S3" or "GCS"
  PRECISE_CODE_INTEL_UPLOAD_BUCKET:
    value: lsif-uploads # external object storage bucket name
  PRECISE_CODE_INTEL_UPLOAD_AWS_ENDPOINT:
    value: https://s3.us-east-1.amazonaws.com
  PRECISE_CODE_INTEL_UPLOAD_AWS_REGION:
    value: us-east-1
  PRECISE_CODE_INTEL_UPLOAD_AWS_ACCESS_KEY_ID:
    secretKeyRef: # Pre-existing secret, not created by this chart
      name: sourcegraph-s3-credentials
      key: PRECISE_CODE_INTEL_UPLOAD_AWS_ACCESS_KEY_ID
  PRECISE_CODE_INTEL_UPLOAD_AWS_SECRET_ACCESS_KEY:
    secretKeyRef: # Pre-existing secret, not created by this chart
      name: sourcegraph-s3-credentials
      key: PRECISE_CODE_INTEL_UPLOAD_AWS_SECRET_ACCESS_KEY

frontend:
  env:
    <<: *objectStorageEnv

preciseCodeIntel:
  env:
    <<: *objectStorageEnv
```

#### Using SSH to clone repositories

If repository authentication is required to `git clone` a repository then you must provide credentials to the container.

Create a [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) that contains the base64 encoded contents of your SSH private key and `known_hosts` file. The SSH private key should not require a passphrase. The Secret will be mounted in the `gitserver` deployment to authenticate with your code host.

**Option 1: Create Secret with Local SSH Keys**
If you have access to the ssh keys locally, you can run the command below to create the secret:

```sh
kubectl create secret generic gitserver-ssh \
	    --from-file id_rsa=${HOME}/.ssh/id_rsa \
	    --from-file known_hosts=${HOME}/.ssh/known_hosts
```

**Option 2: Create Secret from Manifest File**
Alternatively, you may manually create the secret from a manifest file.

> WARNING: For security purposes, do NOT commit the secret manifest into your Git repository unless you are comfortable storing sensitive information in plaintext and your repository is private.

Create a file with the following and save it as `gitserver-ssh.Secret.yaml`

 ***[TODO Where are they creating and storing this file? In root?]***

```sh
apiVersion: v1
kind: Secret
metadata:
  name: gitserver-ssh
data:
  # notes: secrets data has to be base64-encoded
  id_rsa: ""
  known_hosts: ""
```

Add the following values to your override file to reference the Secret:

```yaml
gitserver:
  sshSecret: gitserver-ssh
```

Apply the created Secret to your Kubernetes instance with the command below:

```sh
kubectl apply -f gitserver-ssh.Secret.yaml
```

### Advanced Helm Chart Configurations

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

> WARNING: Skipping minor version upgrades of Sourcegraph is not supported. You must upgrade one minor version at a time - e.g. v3.26 –> v3.27 –> v3.28.

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

### Database Migrations

By default, database migrations will be performed during application startup by a `migrator` init container running prior to the `frontend` deployment. These migrations **must** succeed before Sourcegraph will become available. If the databases are large, these migrations may take a long time.

In some situations, administrators may wish to migrate their databases before upgrading the rest of the system to reduce downtime. Sourcegraph guarantees database backward compatibility to the most recent minor point release so the database can safely be upgraded before the application code.

To execute the database migrations independently, you can use the [Sourcegraph Migrator] helm chart.

## Reviewing Changes

When configuring an override file or performing an upgrade, we recommend reviewing the changes before applying them.

### Using helm template

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

### Using helm upgrade --dry-run

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

### Using Helm Diff plugin

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
