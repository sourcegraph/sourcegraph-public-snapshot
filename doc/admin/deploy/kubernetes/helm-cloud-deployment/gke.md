# Configure Sourcegraph on Google Kubernetes Engine (GKE)

To install Sourcegraph on Google Kubernetes Engine, you must deploy onto a supported machine type and use a persistent standard disk or a persistent SSD.

## Prerequisites {#gke-prerequisites}

- You need to have a GKE cluster (>=1.19) with the `HTTP Load Balancing` addon enabled.

> [Learn more](https://cloud.google.com/kubernetes-engine/docs/how-to/custom-ingress-controller) about deploying your own Ingress Controller instead of disabling the `HTTP Load Balancing` add-on.

- You need to have an account with sufficient access equivalent to the `cluster-admin` ClusterRole.
- You need to be able to connect to your cluster (via either the console or the command line using `gcloud`) and ensure the cluster is up and running. You should see several `ready` nodes listed when you run: `kubectl get nodes`.
- You need to have the [Helm CLI](https://helm.sh/docs/intro/install/) installed and run the following command to link to the Sourcegraph helm repository. This should be ran on the machine used to interact with your cluster:

```sh
helm repo add sourcegraph https://helm.sourcegraph.com/release
```

## Hardware & service requirements

Before beginning the deployment we recommend reviewing the required hardware and service resource requirements.

Use the [resource estimator](../resource_estimator.md) to determine the resource requirements for your environment. You will use this information to set up the instance and configure the override file in the steps below.

## Steps {#gke-steps}

### Create override file & add deployment configurations

Create your override file and add in any configuration override settings you need - see [configuration](./helm/#configuration) documentation for more information on override files and the options for custom configurations.

Add into your override file the below values to configure both your ingress hostname and your storage class. We recommend configuring Ingress to use [Container-native load balancing](https://cloud.google.com/kubernetes-engine/docs/how-to/container-native-load-balancing) to expose Sourcegraph publicly on a domain of your choosing and setting the Storage Class to use [Compute Engine persistent disk](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver#:~:text=Google%20Kubernetes%20Engine%20(GKE)%20provides,tied%20to%20GKE%20version%20numbers.). For an example, see [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/gcp/override.yaml).

> ℹ️ [Container-native load balancing](https://cloud.google.com/kubernetes-engine/docs/how-to/container-native-load-balancing) is only available on VPC-native clusters. [Learn more](https://cloud.google.com/kubernetes-engine/docs/how-to/load-balance-ingress) about ingress load balancing for legacy clusters.

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

The override file includes a [BackendConfig](https://cloud.google.com/kubernetes-engine/docs/how-to/ingress-features#associating_backendconfig_with_your_ingress) CRD. This is required to instruct the GCP load balancer on how to perform health checks on our deployment.

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

Once you have obtained the allocated address of the load balancer, you should create a DNS A record for the `sourcegraph.company.com` domain. It is recommended to enable TLS and we recommend using [Google-managed certificate](https://cloud.google.com/kubernetes-engine/docs/how-to/managed-certs) in GKE.

If using a GKE managed certificate, add the following annotations to Ingress:

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
      kubernetes.io/ingress.class: gce
    tlsSecret: sourcegraph-frontend-tls # reference the created TLS Secret
    # replace with your actual domain
    host: sourcegraph.company.com
```

### Validate the deployment

Sourcegraph should now be available via the address set.

Navigate to the URL in your browser to ensure you now have access to the Sourcegraph UI to create the initial administrator account.

### Sourcegraph Configuration

At this stage the deployment is considered to be complete. You are now ready to configure your Sourcegraph instance (site configuration, code host configuration, search configuration etc). Please see our [Configuring Sourcegraph](../../config/index.md) documentation for guidance.