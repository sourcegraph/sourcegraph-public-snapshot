# Configure Sourcegraph on Azure Managed Kubernetes Service (AKS)

To install Sourcegraph on Azure Managed Kubernetes Service, you must deploy onto a supported machine type and use a persistent standard disk or a persistent SSD.

## Prerequisites {#aks-prerequisites}

- You need to have a AKS cluster (>=1.19) with the following addons enabled:
  - [Azure Application Gateway Ingress Controller](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-install-new)
  - [Azure Disk CSI driver](https://docs.microsoft.com/en-us/azure/aks/csi-storage-drivers)

> [Learn more](https://docs.microsoft.com/en-us/azure/aks/ingress-basic) about deploying your own Ingress Controller instead of Application Gateway.

- You need to have an account with sufficient access equivalent to the `cluster-admin` ClusterRole.
- You need to be able to connect to your cluster (via either the console or the command line using the Azure CLI) and ensure the cluster is up and running. You should see several `ready` nodes listed when you run: `kubectl get nodes`.
- You need to have the [Helm CLI](https://helm.sh/docs/intro/install/) installed and run the following command to link to the Sourcegraph helm repository. This should be ran on the machine used to interact with your cluster:

```sh
helm repo add sourcegraph https://helm.sourcegraph.com/release
```

## Hardware & service requirements

Before beginning the deployment we recommend reviewing the required hardware and service resource requirements.

Use the [resource estimator](../resource_estimator.md) to determine the resource requirements for your environment. You will use this information to set up the instance and configure the override file in the steps below.

## Steps {#aks-steps}

### Create override file & add deployment configurations

Create your override file and add in any configuration override settings you need - see [configuration](./helm/#configuration) documentation for more information on override files and the options for custom configurations.

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

> ℹ️ You can review the changes using one of the [three mechanisms](./helm#reviewing-changes) to assess the customizations made. This is not required, but may be useful the first time you deploy Sourcegraph.

### Install the Sourcegraph Helm chart

Install the Sourcegraph Helm chart by running the following command:

```sh
helm upgrade --install --values ./override.yaml --version 3.41.0 sourcegraph sourcegraph/sourcegraph
```

It will take some time for the load balancer to be fully ready. Use the following command to check on the status and obtain the load balancer IP address once available:

```sh
kubectl describe ingress sourcegraph-frontend
```

### Create a DNS record

Once you have obtained the allocated address of the load balancer, you should create a DNS record for the `sourcegraph.company.com` domain that resolves to the load balancer address.

It is recommended to enable TLS and configure the certificate on your load balancer. We recommend using an [Azure-managed certificate](https://azure.github.io/application-gateway-kubernetes-ingress/features/appgw-ssl-certificate/) and add the following annotations to Ingress.

```yaml
frontend:
  ingress:
    annotations:
      kubernetes.io/ingress.class: azure/application-gateway
      # Name of the Azure-managed TLS certificate
      appgw.ingress.kubernetes.io/appgw-ssl-certificate: azure-key-vault-managed-ssl-cert
```

### Validate the deployment

Sourcegraph should now be available via the address set.

Navigate to the URL in your browser to ensure you now have access to the Sourcegraph UI to create the initial administrator account.

### Sourcegraph configuration

At this stage the deployment is considered to be complete. You are now ready to configure your Sourcegraph instance (site configuration, code host configuration, search configuration etc). Please see our [Configuring Sourcegraph](../../config/index.md) documentation for guidance.

## References {#aks-references}

- [Expose an AKS service over HTTP or HTTPS using Application Gateway](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-expose-service-over-http-https)
- [Supported Azure Application Gateway Ingress Controller annotations](https://azure.github.io/application-gateway-kubernetes-ingress/annotations/)
- [What is Application Gateway Ingress Controller?](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-overview)