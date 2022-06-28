# Configure Sourcegraph on other Cloud Providers or On-Prem

## Prerequisites {#others-prerequisites}

- You need to have a Kubernetes cluster (>=1.19) with the following components installed:
  - Ingress Controller, e.g. Cloud providers-native solution, [NGINX Ingress Controller]
  - Block Storage CSI driver
- You need to have an account with sufficient access equivalent to the `cluster-admin` ClusterRole.
- Connect to your cluster (via either the console or the command line using the relevant CLI tool) and ensure the cluster is up and running. You should see several `ready` nodes listed when you run: `kubectl get nodes`.
- You need to have the [Helm CLI](https://helm.sh/docs/intro/install/) installed and run the following command to link to the Sourcegraph helm repository. This should be ran on the machine used to interact with your cluster:

```sh
helm repo add sourcegraph https://helm.sourcegraph.com/release
```

## Hardware and Service Requirements

Before beginning the deployment we recommend reviewing the required hardware and service resource requirements.

Use the [resource estimator](../resource_estimator.md) to determine the resource requirements for your environment. You will use this information to set up the instance and configure the override file in the steps below.

## Steps {#others-steps}

### Create Override File & Add Deployment Configurations

Create your override file and add any configuration override settings you need. See the [configuration](./helm/#configuration) documentation for more information on override files and the options for custom configurations.

Reference the [Kubernetes Storage Class](https://kubernetes.io/docs/concepts/storage/storage-classes/) documentation for guidance on how to configure the `storageClass.provisioner` and `storageClass.parameters` fields for your cloud provider or consult documentation of the storage solution in your on-prem environment.

The following will need to be included in your `override.yaml`, once adapted the Storage Class to your environment.

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

> ℹ️ You can review the changes using one of the [three mechanisms](./helm#reviewing-changes) to assess the customizations made. This is not required, but may be useful the first time you deploy Sourcegraph.

### Install the Sourcegraph Helm chart

Install the Sourcegraph Helm chart by running the following command:

```sh
helm upgrade --install --values ./override.yaml --version 3.41.0 sourcegraph sourcegraph/sourcegraph
```

It will take some time for the your ingress to be fully ready. Depending on how your Ingress Controller is configured, you can use the following command to check on the status and obtain the load balancer IP address once available:

```sh
kubectl describe ingress sourcegraph-frontend
```

### Create a DNS Record

Once you have obtained the allocated address of your ingress, you should create a DNS A record for the `sourcegraph.company.com` domain that resolves to the Ingress public address.

It is recommended to enable TLS and configure a certificate on your Ingress. You can utilize managed certificate solutions provided by Cloud providers, or your own method.

We recommend configuring [cert-manager with Let's Encrypt](https://cert-manager.io/docs/configuration/acme/) in your cluster. You can do this by adding the following override to your Ingress:

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

If you prefer to use your own certificate, you can do so with [TLS Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets).

1. Create a file with the following and save it as `sourcegraph-frontend-tls.Secret.yaml`

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

2. Apply the configuration by running:

```sh
kubectl apply -f ./sourcegraph-frontend-tls.Secret.yaml
```

3. Add the following values to your override file:

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

### Validate the deployment

Sourcegraph should now be available via the address set.

Navigate to the URL in your browser to ensure you now have access to the Sourcegraph UI to create the initial administrator account.

### Sourcegraph Configuration

At this stage the deployment is considered to be complete. You are now ready to configure your Sourcegraph instance (site configuration, code host configuration, search configuration etc). Please see our [Configuring Sourcegraph](../../config/index.md) documentation for guidance.