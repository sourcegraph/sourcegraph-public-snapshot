# Configure Sourcegraph on other Cloud providers or on-prem

## Prerequisites {#others-prerequisites}

1. You need to have a Kubernetes cluster (>=1.19) with the following components installed:
   - [x] Ingress Controller, e.g. Cloud providers-native solution, [NGINX Ingress Controller]
   - [x] Block Storage CSI driver
1. Your account should have sufficient access privileges, equivalent to the `cluster-admin` ClusterRole.
1. Connect to your cluster (via either the console or the command line using the relevant CLI tool) and ensure the cluster is up and running using: `kubectl get nodes` (several `ready` nodes should be listed)
1. Have the [Helm CLI](https://helm.sh/docs/intro/install/) installed and run the following command to link to the Sourcegraph helm repository (on the machine used to interact with your cluster):

```sh
helm repo add sourcegraph https://helm.sourcegraph.com/release
```

## Steps {#others-steps}

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
helm upgrade --install --values ./override.yaml --version 3.41.0 sourcegraph sourcegraph/sourcegraph
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