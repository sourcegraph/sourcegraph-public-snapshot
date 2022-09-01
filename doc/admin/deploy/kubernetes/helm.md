# Using Helm

## Requirements

* [Helm 3 CLI](https://helm.sh/docs/intro/install/)
* Kubernetes 1.19 or greater

<div class="cta-group">
<!--<a class="btn btn-primary" href="#installation">★ ...</a>-->
<a class="btn" href="#configuration">Configuration</a>
<a class="btn" href="#configure-sourcegraph-on-google-kubernetes-engine-gke">Google GKE</a>
<a class="btn" href="#configure-sourcegraph-on-elastic-kubernetes-service-eks">AWS EKS</a>
<a class="btn" href="#configure-sourcegraph-on-azure-managed-kubernetes-service-aks">Azure AKS</a>
<a class="btn" href="#configure-sourcegraph-on-other-cloud-providers-or-on-prem">Other or on-prem</a>
<a class="btn" href="#upgrading-sourcegraph">Upgrading</a>
</div>


> WARNING: Sourcegraph currently does not support migration from an existing Sourcegraph deployment without Helm to a Sourcegraph deployment using Helm. The information you are looking at is currently recommended for a new install of Sourcegraph. We are currently working to provide migration guidance from a non-Helm deployment. If you are inquiring about performing such a migration please email <support@sourcegraph.com>

## Why use Helm

Our Helm chart has a lot of sensible defaults baked into the values.yaml. Not only does this make customizations much easier (than either using Kustomize or manually editing Sourcegraph's manifest files) it also means that, when an override file is used to make the changes, you _never_ have to deal with merge conflicts during upgrades (see more about customizations in the [configuration](#configuration) section).


## High-level overview of how to use Helm with Sourcegraph

1. Prepare any required customizations
   - Most environments are likely to need changes from the defaults - use the guidance in [Configuration](#configuration).
1. Review the changes
   - There are [three mechanisms](#reviewing-changes) that can be used to review any customizations made, this is an optional step, but may be useful the first time you deploy Sourcegraph, for peace of mind.
1. Select your deployment method and follow the guidance:
   - [Google GKE](#configure-sourcegraph-on-google-kubernetes-engine-gke)
   - [AWS EKS](#configure-sourcegraph-on-elastic-kubernetes-service-eks)
   - [Azure AKS](#configure-sourcegraph-on-azure-managed-kubernetes-service-aks)
   - [Other cloud providers or on-prem](#configure-sourcegraph-on-other-cloud-providers-or-on-prem)


## Quickstart

> ℹ️ This quickstart guide is useful to those already familiar with Helm who have a good understanding of how to use Helm in the environment they want to deploy into, and who just want to quickly deploy Sourcegraph with Helm with the default configuration. If this doesn't cover what you need to know, see the links above for platform-specific guides.

To use the Helm chart, add the Sourcegraph helm repository (on the machine used to interact with your cluster):

```sh
helm repo add sourcegraph https://helm.sourcegraph.com/release
```

Install the Sourcegraph chart using default values:

```sh
helm install --version 3.43.1 sourcegraph sourcegraph/sourcegraph
```

Sourcegraph should now be available via the address set. Browsing to the url should now provide access to the Sourcegraph UI to create the initial administrator account.

More information on configuring the Sourcegraph application can be found here:
[Configuring Sourcegraph](../../config/index.md)


## Configuration

The Sourcegraph Helm chart is highly customizable to support a wide range of environments. We highly recommend that customizations be applied using an override file, which allows customizations to persist through upgrades without needing to manage merge conflicts.

The default configuration values can be viewed in the [values.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/values.yaml) file along with all [supported options](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph#configuration-options).

To customize configuration settings with an override file, begin by creating an empty yaml file (e.g. `override.yaml`).

(The configuration override file can be created in advance of deployment, and the configuration override settings can be populated in preparation.)

It's recommended that the override file be maintained in a version control system such as GitHub, but for testing, this can be created on the machine from which the Helm deployment commands will be run.

> WARNING: __DO NOT__ copy the [default values file](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/values.yaml) as a boilerplate for your override file. You risk having outdated values during future upgrades. Instead, only include the configuration that you need to change and override.

Example overrides can be found in the [examples](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples) folder. Please take a look at our examples – feel free to copy and adapt them for your own override file.

Providing the override file to Helm is done with the inclusion of the values flag and the name of the file:
```sh
helm upgrade --install --values ./override.yaml --version 3.43.1 sourcegraph sourcegraph/sourcegraph
```
When making configuration changes, it's recommended to review the changes that will be applied - see [Reviewing Changes](#reviewing-changes).

### Specific Configuration Scenarios

#### Using external PostgreSQL databases

To use external PostgreSQL databases, first review our [general recommendations](https://docs.sourcegraph.com/admin/external_services/postgres#using-your-own-postgresql-server) and [required postgres permissions](https://docs.sourcegraph.com/admin/external_services/postgres#postgres-permissions-and-database-migrations).

We recommend storing the credentials in [Secrets] created outside of the helm chart and managed in a secure manner. Each database requires its own Secret and should follow the following format. The Secret name can be customized as desired:

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

The above Secrets should be deployed to the same namespace as the existing Sourcegraph deployment.

You can reference the Secrets in your [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/examples/external-databases/override.yaml) by configuring the `existingSecret` key:

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

The [using external databases](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/external-databases) example demonstrates this approach.

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

To use external Redis instances, first review our [general recommendations](https://docs.sourcegraph.com/admin/external_services/redis).


If your external Redis instances do not require authentication, you can configure access in your [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/examples/external-redis/override.yaml) with the `endpoint` settings:

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

You can reference this secret in your [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/examples/external-redis/override-secret.yaml) by configuring the `existingSecret` key:

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

The [using your own Redis](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/external-redis) example demonstrates this approach.

#### Using external Object Storage

To use an external Object Storage service (S3-compatible services, or GCS), first review our [general recommendations](https://docs.sourcegraph.com/admin/external_services/object_storage). Then review the following example and adjust to your use case.

> See [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/external-object-storage/override.yaml) for an example override file.
>
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
    value: S3 # external object stoage type, one of "S3" or "GCS"
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

Create a [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) that contains the base64 encoded contents of your SSH private key (make sure it doesn’t require a passphrase) and known_hosts file. The [Secret] will be mounted in the `gitserver` deployment to authenticate with your code host.

If you have access to the ssh keys locally, you can run the command below to create the secret:

```sh
kubectl create secret generic gitserver-ssh \
      --from-file id_rsa=${HOME}/.ssh/id_rsa \
      --from-file known_hosts=${HOME}/.ssh/known_hosts
```

Alternatively, you may manually create the secret from a manifest file.

> WARNING: Do NOT commit the secret manifest into your Git repository unless you are okay with storing sensitive information in plaintext and your repository is private.

Create a file with the following and save it as `gitserver-ssh.Secret.yaml`
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

Apply the created [Secret] with the command below:

```sh
kubectl apply -f gitserver-ssh.Secret.yaml
```

You should add the following values to your override file to reference the [Secret] you created earlier.

```yaml
gitserver:
  sshSecret: gitserver-ssh
```

### Advanced Configuration Methods

The Helm chart is new and still under active development, and our values.yaml (and therefore the customization available to use via an override file) may not cover every need. Equally, some changes are environment or customer-specific, and so will never be part of the default Sourcegraph Helm chart.

The following guidance for using Kustomize with Helm and Helm Subcharts covers both of these scenarios.

> ⚠️ While both of these approaches are available, deployment changes that are not covered by Sourcegraph documentation should be discussed with either your Customer Engineer or Support contact before proceeding, to ensure the changes proposed can be supported by Sourcegraph. This also allows Sourcegraph to consider adding the required customizations to the Helm chart.

#### Integrate Kustomize with Helm chart

For advanced users who are looking for a temporary workaround, we __recommend__ applying [Kustomize](https://kustomize.io) on the rendered manifests from our chart. Please __do not__ maintain your own fork of our chart, this may impact our ability to support you if you run into issues.

You can learn more about how to integrate Kustomize with Helm from our [example](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/kustomize-chart).

#### Helm subcharts

[Helm subcharts](https://helm.sh/docs/chart_template_guide/subcharts_and_globals/) can be used for permanent customizations to the official Sourcegraph helm chart. This is useful for changes such as adding a new resource unique to your deployment (PodSecurityPolicy, NetworkPolicy, additional services, etc.). These are long-lived customizations that shouldn't be contributed back to the Sourcegraph helm chart.

With a subchart, you create your own helm chart and specify the Sourcegraph chart as a dependency. Any resources you place in the templates folder of your chart will be deployed, as well as the Sourcegraph resources, allowing you to extend the Sourcegraph chart without maintaining a fork.

An example of a subchart is shown in the [examples/subchart](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples) folder.

More details on how to create and configure a subchart can be found in the [helm documentation](https://helm.sh/docs/chart_template_guide/subcharts_and_globals).

## Cloud providers guides

This section is aimed at providing high-level guidance on deploying Sourcegraph via Helm on major Cloud providers. In general, you need the following to get started:

- A working Kubernetes cluster, v1.19 or higher
- The ability to provision persistent volumes, e.g. have Block Storage [CSI storage driver](https://kubernetes-csi.github.io/docs/drivers.html) installed
- An Ingress Controller installed, e.g. platform native ingress controller, [NGINX Ingress Controller].
- The ability to create DNS records for Sourcegraph, e.g. `sourcegraph.company.com`

### Configure Sourcegraph on Google Kubernetes Engine (GKE)

#### Prerequisites {#gke-prerequisites}

1. You need to have a GKE cluster (>=1.19) with the `HTTP Load Balancing` addon enabled. Alternatively, you can use your own choice of Ingress Controller and disable the `HTTP Load Balancing` add-on, [learn more](https://cloud.google.com/kubernetes-engine/docs/how-to/custom-ingress-controller).
1. Your account should have sufficient access rights, equivalent to the `cluster-admin` ClusterRole.
1. Connect to your cluster (via either the console or the command line using `gcloud`) and ensure the cluster is up and running by running: `kubectl get nodes` (several `ready` nodes should be listed)
1. Have the [Helm CLI](https://helm.sh/docs/intro/install/) installed and run the following command to link to the Sourcegraph helm repository (on the machine used to interact with your cluster):

```sh
helm repo add sourcegraph https://helm.sourcegraph.com/release
```

#### Steps {#gke-steps}

**1** – Create your override file and add in any configuration override settings you need - see [configuration](#configuration) for more information on override files and the options for what can be configured.

Add into your override file the below values to configure both your ingress hostname and your storage class. We recommend configuring Ingress to use [Container-native load balancing] to expose Sourcegraph publicly on a domain of your choosing and setting the Storage Class to use [Compute Engine persistent disk]. (For an example file see [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/gcp/override.yaml))

<!--[override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/gcp/override.yaml)-->
```yaml
frontend:
  serviceType: ClusterIP
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: gce
# To enable HTTPS using a self-managed certificate
#    tlsSecret: example-secret
#    host: sourcegraph.example.com
  serviceAnnotations:
    cloud.google.com/neg: '{"ingress": true}'
    # Reference the `BackendConfig` CR created below
    beta.cloud.google.com/backend-config: '{"default": "sourcegraph-frontend"}'

storageClass:
  create: true
  type: pd-ssd # This configures SSDs (recommended).
  provisioner: pd.csi.storage.gke.io
  volumeBindingMode: WaitForFirstConsumer
  reclaimPolicy: Retain

extraResources:
  - apiVersion: cloud.google.com/v1
    kind: BackendConfig
    metadata:
      name: sourcegraph-frontend
    spec:
      healthCheck:
        checkIntervalSec: 5
        timeoutSec: 5
        requestPath: /ready
        port: 6060 # we use a custom port to perform healthcheck
```

The override file includes a [BackendConfig] CRD. This is necessary to instruct the GCP load balancer on how to perform health checks on our deployment.

> ℹ️ [Container-native load balancing] is only available on VPC-native clusters. For legacy clusters, [learn more](https://cloud.google.com/kubernetes-engine/docs/how-to/load-balance-ingress).

> ℹ️ Optionally, you can review the changes using one of [three mechanisms](#reviewing-changes) that can be used to assess the customizations made. This is not required, but may be useful the first time you deploy Sourcegraph, for peace of mind.

**2** – Install the chart

```sh
helm upgrade --install --values ./override.yaml --version 3.43.1 sourcegraph sourcegraph/sourcegraph
```

It will take around 10 minutes for the load balancer to be fully ready, you may check on the status and obtain the load balancer IP using the following command:

```sh
kubectl describe ingress sourcegraph-frontend
```

**3** – Upon obtaining the allocated IP address of the load balancer, you should create a DNS A record for the `sourcegraph.company.com` domain. Finally, it is recommended to enable TLS and you may consider using [Google-managed certificate](https://cloud.google.com/kubernetes-engine/docs/how-to/managed-certs) in GKE or your own certificate.

If using a GKE manage certificate, add the following annotations to Ingress:

```yaml
frontend:
  ingress:
    annotations:
      kubernetes.io/ingress.class: null
      networking.gke.io/managed-certificates: managed-cert # replace with actual Google-managed certificate name
      # if you reserve a static IP, uncomment below and update ADDRESS_NAME
      # also, make changes to your DNS record accordingly
      # kubernetes.io/ingress.global-static-ip-name: ADDRESS_NAME
```

If using your own certificate, you can do so with [TLS Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets).

Create a file with the following and save it as `sourcegraph-frontend-tls.Secret.yaml`
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sourcegraph-frontend-tls
type: kubernetes.io/tls
data:
  # the data is abbreviated in this example
  tls.crt: |
    MIIC2DCCAcCgAwIBAgIBATANBgkqh ...
  tls.key: |
    MIIEpgIBAAKCAQEA7yn3bRHQ5FHMQ ...
```

```sh
kubectl apply -f ./sourcegraph-frontend-tls.Secret.yaml
```

Add the following values to your override file.

```yaml
frontend:
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: gce
    tlsSecret: sourcegraph-frontend-tls # reference the created TLS Secret
    # replace with your actual domain
    host: sourcegraph.company.com
```

**5** – Validate the deployment
Sourcegraph should now be available via the address set.
Browsing to the url should now provide access to the Sourcegraph UI to create the initial administrator account.

**6** – Further configuration

Now the deployment is complete, more information on configuring the Sourcegraph application can be found here:
[Configuring Sourcegraph](../../config/index.md)

### Configure Sourcegraph on Elastic Kubernetes Service (EKS)

#### Prerequisites {#eks-prerequisites}

1. You need to have a EKS cluster (>=1.19) with the following addons enabled:
   - [AWS Load Balancer Controller](https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html)
   - [AWS EBS CSI driver](https://docs.aws.amazon.com/eks/latest/userguide/managing-ebs-csi.html)
> You may consider deploying your own Ingress Controller instead of the ALB Ingress Controller, [learn more](https://kubernetes.github.io/ingress-nginx/)
1. Your account should have sufficient access equivalent to the `cluster-admin` ClusterRole.
1. Connect to your cluster (via either the console or the command line using `eksctl`) and ensure the cluster is up and running using: `kubectl get nodes` (several `ready` nodes should be listed)
1. Have the [Helm CLI](https://helm.sh/docs/intro/install/) installed and run the following command to link to the Sourcegraph helm repository (on the machine used to interact with your cluster):

```sh
helm repo add sourcegraph https://helm.sourcegraph.com/release
```

#### Steps {#eks-steps}

**1** – Create your override file and add in any configuration override settings you need - see [configuration](#configuration) for more information on override files and the options around what can be configured.

We recommend adding the following values into your override file to configure Ingress to use [AWS Load Balancer Controller] to expose Sourcegraph publicly on a domain of your choosing, and to configure the Storage Class to use [AWS EBS CSI driver]. For an example, see [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/aws/override.yaml).

Uncomment the `provisioner` that your Amazon EKS cluster implements.

<!--[override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/aws/override.yaml)-->
```yaml
frontend:
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: alb # aws load balancer controller ingressClass name
      # additional aws alb ingress controller supported annotations
      # ...
    # replace with your actual domain
    host: sourcegraph.company.com

storageClass:
  create: true
  type: gp2 # This configures SSDs (recommended).
#  provisioner: ebs.csi.aws.com # use this provisioner if using the self-managed Amazon EBS Container Storage Interface driver
#  provisioner: kubernetes.io/aws-ebs # use this provisioner if using the Amazon EKS add-on
  volumeBindingMode: WaitForFirstConsumer
  reclaimPolicy: Retain
```
> ℹ️ Optionally, you can review the changes using one of [three mechanisms](#reviewing-changes) that can be used to assess the customizations made. This is not required, but may be useful the first time you deploy Sourcegraph, for peace of mind.

**2** – Install the chart

```sh
helm upgrade --install --values ./override.yaml --version 3.43.1 sourcegraph sourcegraph/sourcegraph
```

It will take some time for the load balancer to be fully ready, use the following to check on the status and obtain the load balancer address (once available):

```sh
kubectl describe ingress sourcegraph-frontend
```

**3** – Upon obtaining the allocated address of the load balancer, you should create a DNS record for the `sourcegraph.company.com` domain that resolves to the load balancer address.

It is recommended to enable TLS and configure a certificate properly on your load balancer. You may consider using an [AWS-managed certificate](https://docs.aws.amazon.com/acm/latest/userguide/acm-overview.html) and add the following annotations to Ingress.

```yaml
frontend:
  ingress:
    annotations:
      kubernetes.io/ingress.class: alb
      # ARN of the AWS-managed TLS certificate
      alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:us-west-2:xxxxx:certificate/xxxxxxx
```

**4** – Validate the deployment
Sourcegraph should now be available via the address set.
Browsing to the url should now provide access to the Sourcegraph UI to create the initial administrator account.

**5** – Further configuration

Now the deployment is complete, more information on configuring the Sourcegraph application can be found here:
[Configuring Sourcegraph](../../config/index.md)

#### References {#eks-references}

- [Enable TLS with AWS-managed certificate](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations/#ssl)
- [Supported AWS load balancer annotations](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations)

### Configure Sourcegraph on Azure Managed Kubernetes Service (AKS)

#### Prerequisites {#aks-prerequisites}

1. You need to have a AKS cluster (>=1.19) with the following addons enabled:
   - [Azure Application Gateway Ingress Controller](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-install-new)
   - [Azure Disk CSI driver](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers)
> You may consider using your custom Ingress Controller instead of Application Gateway, [learn more](https://docs.microsoft.com/en-us/azure/aks/ingress-basic)
1. Your account should have sufficient access equivalent to the `cluster-admin` ClusterRole.
1. Connect to your cluster (via either the console or the command line using the Azure CLI) and ensure the cluster is up and running using: `kubectl get nodes` (several `ready` nodes should be listed)
1. Have the [Helm CLI](https://helm.sh/docs/intro/install/) installed and run the following command to link to the Sourcegraph helm repository (on the machine used to interact with your cluster):

```sh
helm repo add sourcegraph https://helm.sourcegraph.com/release
```

#### Steps {#aks-steps}

**1** – Create your override file and add in any configuration override settings you need - see [configuration](#configuration) for more information on override files and the options around what can be configured.

Add into your override file the below values to configure both your ingress hostname and your storage class. We recommend configuring Ingress to use [Application Gateway](https://azure.microsoft.com/en-us/services/application-gateway) to expose Sourcegraph publicly on a domain of your choosing and Storage Class to use [Azure Disk CSI driver](https://docs.microsoft.com/en-us/azure/aks/azure-disk-csi). For an example see [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/azure/override.yaml).

<!--[override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/azure/override.yaml)-->
```yaml
frontend:
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: azure/application-gateway
      # additional azure application gateway supported annotations
      # ...
    # replace with your actual domain
    host: sourcegraph.company.com

storageClass:
  create: true
  type: null
  provisioner: disk.csi.azure.com
  volumeBindingMode: WaitForFirstConsumer
  reclaimPolicy: Retain
  parameters:
    storageaccounttype: Premium_LRS # This configures SSDs (recommended). A Premium VM is required.
```

> ℹ️ Optionally, you can review the changes using one of [three mechanisms](#reviewing-changes) that can be used to assess the customizations made. This is not required, but may be useful the first time you deploy Sourcegraph, for peace of mind.

**2** – Install the chart

```sh
helm upgrade --install --values ./override.yaml --version 3.43.1 sourcegraph sourcegraph/sourcegraph
```

It will take some time for the load balancer to be fully ready, you can check on the status and obtain the load balancer address (when ready) using:

```sh
kubectl describe ingress sourcegraph-frontend
```

**3** – Upon obtaining the allocated address of the load balancer, you should create a DNS record for the `sourcegraph.company.com` domain that resolves to the load balancer address.

It is recommended to enable TLS and configure the certificate properly on your load balancer. You may consider using an [Azure-managed certificate](https://azure.github.io/application-gateway-kubernetes-ingress/features/appgw-ssl-certificate/) and add the following annotations to Ingress.

```yaml
frontend:
  ingress:
    annotations:
      kubernetes.io/ingress.class: azure/application-gateway
      # Name of the Azure-managed TLS certificate
      appgw.ingress.kubernetes.io/appgw-ssl-certificate: azure-key-vault-managed-ssl-cert
```

**4** – Validate the deployment
Sourcegraph should now be available via the address set.
Browsing to the url should now provide access to the Sourcegraph UI to create the initial administrator account.

**5** – Further configuration

Now the deployment is complete, more information on configuring the Sourcegraph application can be found here:
[Configuring Sourcegraph](../../config/index.md)

#### References {#aks-references}

- [Expose an AKS service over HTTP or HTTPS using Application Gateway](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-expose-service-over-http-https)
- [Supported Azure Application Gateway Ingress Controller annotations](https://azure.github.io/application-gateway-kubernetes-ingress/annotations/)
- [What is Application Gateway Ingress Controller?](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-overview)


### Configure Sourcegraph on other Cloud providers or on-prem

#### Prerequisites {#others-prerequisites}

1. You need to have a Kubernetes cluster (>=1.19) with the following components installed:
   - [x] Ingress Controller, e.g. Cloud providers-native solution, [NGINX Ingress Controller]
   - [x] Block Storage CSI driver
1. Your account should have sufficient access privileges, equivalent to the `cluster-admin` ClusterRole.
1. Connect to your cluster (via either the console or the command line using the relevant CLI tool) and ensure the cluster is up and running using: `kubectl get nodes` (several `ready` nodes should be listed)
1. Have the [Helm CLI](https://helm.sh/docs/intro/install/) installed and run the following command to link to the Sourcegraph helm repository (on the machine used to interact with your cluster):

```sh
helm repo add sourcegraph https://helm.sourcegraph.com/release
```

#### Steps {#others-steps}

**1** – Create your override file and add in any configuration override settings you need - see [configuration](#configuration) for more information on override files and the options around what can be configured.

Read <https://kubernetes.io/docs/concepts/storage/storage-classes/> to configure the `storageClass.provisioner` and `storageClass.parameters` fields for your cloud provider or consult documentation of the storage solution in your on-prem environment.

The following will need to be included in your `override.yaml`, once adapted to your environment.

```yaml
frontend:
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: ingress-class-name # replace with actual ingress class name
      # additional ingress controller supported annotations
      # ...
    # replace with your actual domain
    host: sourcegraph.company.com

storageClass:
  create: true
  provisioner: <REPLACE_ME>
  volumeBindingMode: WaitForFirstConsumer
  reclaimPolicy: Retain
  parameters:
    key1: value1
```

> ℹ️ Optionally, you can review the changes using one of [three mechanisms](#reviewing-changes) that can be used to assess the customizations made. This is not required, but may be useful the first time you deploy Sourcegraph, for peace of mind.

**2** – Install the chart

```sh
helm upgrade --install --values ./override.yaml --version 3.43.1 sourcegraph sourcegraph/sourcegraph
```

It may take some time before your ingress is up and ready to proceed. Depending on how your Ingress Controller works, you may be able to check on its status and obtain the public address of your Ingress using:

```sh
kubectl describe ingress sourcegraph-frontend
```

**3** – You should create a DNS record for the `sourcegraph.company.com` domain that resolves to the Ingress public address.

It is recommended to enable TLS and configure a certificate properly on your Ingress. You can utilize managed certificate solutions provided by Cloud providers, or your own method.

Alternatively, you may consider configuring [cert-manager with Let's Encrypt](https://cert-manager.io/docs/configuration/acme/) in your cluster and add the following override to Ingress.

```yaml
frontend:
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: ingress-class-name # replace with actual ingress class name
      # additional ingress controller supported annotations
      # ...
      # cert-managed annotations
      cert-manager.io/cluster-issuer: letsencrypt # replace with actual cluster-issuer name
    tlsSecret: sourcegraph-frontend-tls # cert-manager will store the created certificate in this secret.
    # replace with your actual domain
    host: sourcegraph.company.com
```

You also have the option to manually configure TLS certificate via [TLS Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets).

Create a file with the following and save it as `sourcegraph-frontend-tls.Secret.yaml`
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sourcegraph-frontend-tls
type: kubernetes.io/tls
data:
  # the data is abbreviated in this example
  tls.crt: |
    MIIC2DCCAcCgAwIBAgIBATANBgkqh ...
  tls.key: |
    MIIEpgIBAAKCAQEA7yn3bRHQ5FHMQ ...
```

```sh
kubectl apply -f ./sourcegraph-frontend-tls.Secret.yaml
```

Add the following values to your override file.

```yaml
frontend:
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: ingress-class-name # replace with actual ingress class name
      # additional ingress controller supported annotations
      # ...
    tlsSecret: sourcegraph-frontend-tls # reference the created TLS Secret
    # replace with your actual domain
    host: sourcegraph.company.com
```

**4** – Validate the deployment
Sourcegraph should now be available via the address set.
Browsing to the url should now provide access to the Sourcegraph UI to create the initial administrator account.

**5** – Further configuration

Now the deployment is complete, more information on configuring the Sourcegraph application can be found here:
[Configuring Sourcegraph](../../config/index.md)

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
      helm upgrade --install -f override.yaml --version 3.43.1 sourcegraph sourcegraph/sourcegraph
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
