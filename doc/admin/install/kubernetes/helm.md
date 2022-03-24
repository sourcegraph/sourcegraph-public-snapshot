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
# Installation
helm install --values ./override.yaml --version 0.7.0 sourcegraph sourcegraph/sourcegraph

# Upgrade
helm upgrade --values ./override.yaml --version 0.7.0 sourcegraph sourcegraph/sourcegraph
```

### Using external PostgreSQL databases

__TODO__

### Using external Redis instances

__TODO__

### Using external Object Storage

__TODO__

### Cloud providers guides

#### Configure Sourcegraph on Google Kubernetes Engine (GKE)

You should include these suggested values in your override file, learn more from [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/gcp/override.yaml).

> Looking to enable TLS with Google-managed certificates? [learn more](https://cloud.google.com/kubernetes-engine/docs/how-to/managed-certs)

In addition to the provided values, you need to deploy the [BackendConfig] CRD to properly expose Sourcegraph publically. The [BackendConfig] CR should be deployed in the same namespace where Sourcegraph chart is installed.

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

#### Configure Sourcegraph on Elastic Kubernetes Service (EKS)

You should include these suggested values in your override file, learn more from [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/aws/override.yaml).

In addition to the provided values, you should consult AWS documentation to learn how to configure the Ingress properly.

- [Applicatoin load balancing on Amazon EKS](https://docs.aws.amazon.com/eks/latest/userguide/alb-ingress.html)
- [Enable TLS with AWS-managed certificate](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations/#ssl)
- [Supported AWS load balancer annotations](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations)

```yaml
frontend:
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: alb
      # additional aws alb ingress controller supported annotations
      # ...
    # replace with your actual domain
    host: sourcegraph.company.com
```

#### Configure Sourcegraph on Azure Managed Kubernetes Service (AKS)

You should include these suggested values in your override file, learn more from [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/azure/override.yaml).

In addition to the provided values, you should configure Ingress to use [Azure Application Gateway] to expose Sourcegraph publically.

- [Expose an AKS service over HTTP or HTTPS using Application Gateway](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-expose-service-over-http-https)
- [Use certificates with LetsEncrypt.org on Application Gateway for AKS clusters](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-letsencrypt-certificate-application-gateway)
- [Supported Azure Application Gateway Ingress Controller annotations](https://azure.github.io/application-gateway-kubernetes-ingress/annotations/)
- [What is Application Gateway Ingress Controller?](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-overview)

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
```

#### Configure Sourcegraph on other Cloud providers or on-prem

Read <https://kubernetes.io/docs/concepts/storage/storage-classes/> to configure the `storageClass.provisioner` and `storageClass.parameters` fields for your cloud provider or your on-prem environment.

```yaml
storageClass:
  create: true
  provisioner: <REPLACE_ME>
  volumeBindingMode: WaitForFirstConsumer
  reclaimPolicy: Retain
  parameters:
    key1: value1
```

### Advanced configuration

The Helm chart is new and still under active development, and we may not cover all of your use cases. 

Please contact [support@sourcegraph.com](mailto:support@sourcegraph.com) or your Customer Engineer directly to discuss your specific need.

For advanced users who are looking for a temporary workaround, we __recommend__ applying [Kustomize](https://kustomize.io) on the rendered manifests from our chart. Please __do not__ maintain your own fork of our chart, this may impact our ability to support you if you run into issues.

You can learn more about how to integrate Kustomize with Helm from our [example](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/kustomize-chart).

## Upgrading Sourcegraph

__TODO__

[backendconfig]: https://cloud.google.com/kubernetes-engine/docs/how-to/ingress-features#create_backendconfig
[azure application gateway]: https://docs.microsoft.com/en-us/azure/application-gateway/overview
