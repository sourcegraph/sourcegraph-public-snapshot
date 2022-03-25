# Sourcegraph Helm Chart

## Requirements

* [Helm 3 CLI](https://helm.sh/docs/intro/install/)
* Kubernetes 1.19 or greater

## Quickstart

To use the Helm chart, add Sourcegraph helm repository:
 
```sh
helm repo add sourcegraph https://sourcegraph.github.io/deploy-sourcegraph-helm/
```

Install the Sourcegraph chart using default values:

```sh
helm install --version 0.7.0 sourcegraph sourcegraph/sourcegraph
```

## Configuration 

The Sourcegraph chart is highly customizable to support a wide range of environments. Please review the default values from [values.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/values.yaml) and all [supported options](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph#configuration-options). Customizations can be applied using an override file. Using an override file allows customizations to persist through upgrades without needing to manage merge conflicts.

To customize configuration settings with an override file, create an empty yaml file (e.g. `override.yaml`) and configure overrides.

> WARNING: __DO NOT__ copy the [default values file](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/values.yaml) as a boilerplate for your override file. You risk having outdated values during upgrades.

Example overrides can be found in the [examples](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples) folder. Please take a look at our examples before providing your own configuration and consider using them as boilerplates.

Provide the override file to helm:

```sh
helm upgrade --install --values ./override.yaml --version 0.7.0 sourcegraph sourcegraph/sourcegraph
```

### Using external PostgreSQL databases

__TODO__

### Using external Redis instances

__TODO__

### Using external Object Storage

__TODO__

### Cloud providers guides

This section is aimed at providing high-level guidance on configuring Ingress and Storage Class for Sourcegraph deployment on major Cloud providers. In general, you need the following to get started:

- A working Kubernetes cluster 1.19 and higher
- The cluster should have Block Storage CSI storage drivers installed
- The cluster should have Ingress Controller installed. We recommend the use of platform native ingress controller.
- You can have control over your `company.com` domain to create DNS records for Sourcegraph, e.g. `sourcegraph.company.com`

#### Configure Sourcegraph on Google Kubernetes Engine (GKE)

#### Prerequisites

You need to have a __public__ VPC-native GKE cluster (>=1.19) with the following addons enabled:

- [x] HTTP Load Balancing
- [x] Compute Engine persistent disk CSI Driver

You account should have sufficient access equivalent to the `cluster-admin` ClusterRole.

#### Steps

Create an override file with the following value. We configure Ingress to use [Container-native load balancing] to expose Sourcegraph publically and Storage Class to use [Compute Engine persistent disk].

[override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/gcp/override.yaml)
```yaml
frontend:
  serviceType: ClusterIP
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: null
    ingressClassName: gce
    host: sourcegraph.company.com # Replace with your actual domain
  serviceAnnotations:
    cloud.google.com/neg: '{"ingress": true}'
    # reference the `BackendConfig` CR we will be configuring at a later step
    beta.cloud.google.com/backend-config: '{"default": "sourcegraph-frontend"}'

storageClass:
  create: true
  type: pd-ssd # This configures SSDs (recommended).
  provisioner: pd.csi.storage.gke.io
  volumeBindingMode: WaitForFirstConsumer
  reclaimPolicy: Retain
```

Install the chart

```sh
helm upgrade --install --values ./override.yaml --version 0.7.0 sourcegraph sourcegraph/sourcegraph
```

You need to deploy the [BackendConfig] CRD to properly expose Sourcegraph publically. The [BackendConfig] CR should be deployed in the same namespace where Sourcegraph chart is installed.

[sourcegraph-frontend.BackendConfig.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/michael/improve-gcp-example/charts/sourcegraph/examples/gcp/sourcegraph-frontend.BackendConfig.yaml)
```yaml
apiVersion: cloud.google.com/v1
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

```sh
kubectl apply -f sourcegraph-frontend.BackendConfig.yaml
```

It will take around 10 mintues for the load balancer to be fully ready, you may check on the status and obtain the load balancer IP:

```sh
kubectl describe ingress sourcegraph-frontend
```

Upon obtaining the allocated IP address of the load balancer, you should create an A record for the `sourcegraph.company.com` domain. Finally, it is recommended to enable TLS and you can learn more from about how to use [Google-managed certificate](https://cloud.google.com/kubernetes-engine/docs/how-to/managed-certs) in GKE.

#### Configure Sourcegraph on Elastic Kubernetes Service (EKS)

#### Prerequisites

You need to have a EKS cluster (>=1.19) with the following addons enabled:

- [x] [AWS Load Balancer Controller]
- [x] [AWS EBS CSI driver]

You account should have sufficient access equivalent to the `cluster-admin` ClusterRole.

#### Steps

Create an override file with the following value. We configure Ingress to use [AWS Load Balancer Controller] to expose Sourcegraph publically and Storage Class to use [AWS EBS CSI driver].

[override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/aws/override.yaml)
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
  provisioner: ebs.csi.aws.com
  volumeBindingMode: WaitForFirstConsumer
  reclaimPolicy: Retain
```

Install the chart

```sh
helm upgrade --install --values ./override.yaml --version 0.7.0 sourcegraph sourcegraph/sourcegraph
```

It will take some time for the load balancer to be fully ready, you may check on the status and obtain the load balancer address:

```sh
kubectl describe ingress sourcegraph-frontend
```

Upon obtaining the allocated address of the load balancer, you should create a DNS record for the `sourcegraph.company.com` domain that resolves to the load balancer address.

It is recommended to enable TLS and configure certificate properly on your load balancer. You may consider using [AWS-managed certificate](https://docs.aws.amazon.com/acm/latest/userguide/acm-overview.html) and add the following annotations to Ingress.

```yaml
frontend:
  ingress:
    annotations:
      kubernetes.io/ingress.class: alb
      # ARN of the AWS-managed TLS certificate
      alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:us-west-2:xxxxx:certificate/xxxxxxx
```

#### References

- [Enable TLS with AWS-managed certificate](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations/#ssl)
- [Supported AWS load balancer annotations](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations)

#### Configure Sourcegraph on Azure Managed Kubernetes Service (AKS)

#### Prerequisites

You need to have a AKS cluster (>=1.19) with the following addons enabled:

- [x] [Azure Application Gateway Ingress Controller](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-install-new)
- [x] [Azure Disk CSI driver](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers)

You account should have sufficient access equivalent to the `cluster-admin` ClusterRole.

#### Steps

Create an override file with the following value. We configure Ingress to use [Application Gateway](https://azure.microsoft.com/en-us/services/application-gateway) to expose Sourcegraph publically and Storage Class to use [Azure Disk CSI driver](https://docs.microsoft.com/en-us/azure/aks/azure-disk-csi).

[override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/azure/override.yaml)
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

Install the chart

```sh
helm upgrade --install --values ./override.yaml --version 0.7.0 sourcegraph sourcegraph/sourcegraph
```

It will take some time for the load balancer to be fully ready, you may check on the status and obtain the load balancer address:

```sh
kubectl describe ingress sourcegraph-frontend
```

Upon obtaining the allocated address of the load balancer, you should create a DNS record for the `sourcegraph.company.com` domain that resolves to the load balancer address.

It is recommended to enable TLS and configure certificate properly on your load balancer. You may consider using an [Azure-managed certificate](https://azure.github.io/application-gateway-kubernetes-ingress/features/appgw-ssl-certificate/) and add the following annotations to Ingress.

```yaml
frontend:
  ingress:
    annotations:
      kubernetes.io/ingress.class: azure/application-gateway
      # Name of the Azure-managed TLS certificate
      appgw.ingress.kubernetes.io/appgw-ssl-certificate: azure-key-vault-managed-ssl-cert
```

#### References

- [Expose an AKS service over HTTP or HTTPS using Application Gateway](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-expose-service-over-http-https)
- [Supported Azure Application Gateway Ingress Controller annotations](https://azure.github.io/application-gateway-kubernetes-ingress/annotations/)
- [What is Application Gateway Ingress Controller?](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-overview)


#### Configure Sourcegraph on other Cloud providers or on-prem

#### Prerequisites

You need to have a Kubernetes cluster (>=1.19) with the following components installed:

- [x] Ingress Controller, e.g. Cloud providers-native solution, [NGINX Ingress Controller](https://github.com/kubernetes/ingress-nginx)
- [x] Block Storage CSI driver

You account should have sufficient access equivalent to the `cluster-admin` ClusterRole.

#### Steps

Read <https://kubernetes.io/docs/concepts/storage/storage-classes/> to configure the `storageClass.provisioner` and `storageClass.parameters` fields for your cloud provider or consult documentation of the storage solution in your on-prem environment.

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

Install the chart

```sh
helm upgrade --install --values ./override.yaml --version 0.7.0 sourcegraph sourcegraph/sourcegraph
```

Depending how your Ingress Controller work, you may be able to check on status and obtain the public address of your Ingress.

```sh
kubectl describe ingress sourcegraph-frontend
```

You should create a DNS record for the `sourcegraph.company.com` domain that resolves to the Ingress public address.

It is recommended to enable TLS and configure certificate properly on your Ingress. You are encouraged to utilize managed certificate solution provided by Cloud providers. Alternatively, you may consider configuring [cert-manager with Let's Encrypt](https://cert-manager.io/docs/configuration/acme/) in your cluster and add the following override to Ingress.

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

As a last resort, you may manually configure TLS certificate via [TLS Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets).

`sourcegraph-frontend-tls.Secret.yaml`
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

### Advanced configuration

#### Integrate Kustomize with Helm chart

The Helm chart is new and still under active development, and we may not cover all of your use cases. 

Please contact [support@sourcegraph.com](mailto:support@sourcegraph.com) or your Customer Engineer directly to discuss your specific need.

For advanced users who are looking for a temporary workaround, we __recommend__ applying [Kustomize](https://kustomize.io) on the rendered manifests from our chart. Please __do not__ maintain your own fork of our chart, this may impact our ability to support you if you run into issues.

You can learn more about how to integrate Kustomize with Helm from our [example](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/kustomize-chart).

#### Sub-chart

__TODO__

## Upgrading Sourcegraph

__TODO__

[backendconfig]: https://cloud.google.com/kubernetes-engine/docs/how-to/ingress-features#create_backendconfig
[azure application gateway]: https://docs.microsoft.com/en-us/azure/application-gateway/overview
[Container-native load balancing]: https://cloud.google.com/kubernetes-engine/docs/how-to/container-native-load-balancing
[Compute Engine persistent disk]: https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver
[AWS Load Balancer Controller]: https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html
[AWS EBS CSI driver]: https://docs.aws.amazon.com/eks/latest/userguide/managing-ebs-csi.html
