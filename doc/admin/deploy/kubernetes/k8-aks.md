# Configure Sourcegraph on Azure Managed Kubernetes Service (AKS)

## Prerequisites {#aks-prerequisites}

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

## Steps {#aks-steps}

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
helm upgrade --install --values ./override.yaml --version 3.41.0 sourcegraph sourcegraph/sourcegraph
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

## References {#aks-references}

- [Expose an AKS service over HTTP or HTTPS using Application Gateway](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-expose-service-over-http-https)
- [Supported Azure Application Gateway Ingress Controller annotations](https://azure.github.io/application-gateway-kubernetes-ingress/annotations/)
- [What is Application Gateway Ingress Controller?](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-overview)