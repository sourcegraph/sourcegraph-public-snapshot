# Configure Sourcegraph on Google Kubernetes Engine (GKE)

## Prerequisites {#gke-prerequisites}

1. You need to have a GKE cluster (>=1.19) with the `HTTP Load Balancing` addon enabled. Alternatively, you can use your own choice of Ingress Controller and disable the `HTTP Load Balancing` add-on, [learn more](https://cloud.google.com/kubernetes-engine/docs/how-to/custom-ingress-controller).
1. Your account should have sufficient access rights, equivalent to the `cluster-admin` ClusterRole.
1. Connect to your cluster (via either the console or the command line using `gcloud`) and ensure the cluster is up and running by running: `kubectl get nodes` (several `ready` nodes should be listed)
1. Have the [Helm CLI](https://helm.sh/docs/intro/install/) installed and run the following command to link to the Sourcegraph helm repository (on the machine used to interact with your cluster):

```sh
helm repo add sourcegraph https://helm.sourcegraph.com/release
```

## Steps {#gke-steps}

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
helm upgrade --install --values ./override.yaml --version 3.41.0 sourcegraph sourcegraph/sourcegraph
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