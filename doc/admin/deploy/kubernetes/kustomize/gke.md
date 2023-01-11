# Installation Guide - Google Kubernetes Engine (GKE)

This section is aimed at providing high-level guidance on deploying Sourcegraph using a Kustomize overlay on GKE. 

## Overview

The overlay 

- A [BackendConfig](https://cloud.google.com/kubernetes-engine/docs/how-to/ingress-configuration#create_backendconfig) CRD. This is necessary to instruct the GCP load balancer on how to perform health checks on our deployment.
- Ingress to use [Container-native load balancing](https://cloud.google.com/kubernetes-engine/docs/how-to/container-native-load-balancing) to expose Sourcegraph publicly on a domain of your choosing and
- Storage Class to use [Compute Engine persistent disk](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver).

## Prerequisites

- Minimum Kubernetes version: [v1.19](https://kubernetes.io/blog/2020/08/26/kubernetes-release-1.19-accentuate-the-paw-sitive/) with [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) v1.19 or later
- [Kustomize](https://kustomize.io/) (built into `kubectl` in version >= 1.14)
- Support for Persistent Volumes (SSDs recommended)
- A running cluster with the following configurations:
  - **Enable HTTP load balancing** in Networking
  - **SSD persistent disk** as book disk type 

## Quick Start

Once you have created a cluster that meets all the prerequisites listed above...

### Step 1: Deploy Sourcegraph

Deploy Sourcegraph main app without the monitoring stacks to your cluster:

```bash
$ kubectl apply --prune -l deploy=sourcegraph -k https://github.com/sourcegraph/deploy-sourcegraph/new/quick-start/gke/base?ref=v4.3.1
```

Monitor the deployment status to make sure everything is up and running:

```bash
kubectl get pods -o wide --watch
```

### Step 2: Access Sourcegraph in Browser

To check the status of the load balancer and obtain its IP:

```bash
$ kubectl describe ingress sourcegraph-frontend
```

From you output, look for the IP address of the load balancer, which is listed under `Address`.

```bash
# Sample output:
Name:             sourcegraph-frontend
Namespace:        default
Address:          12.345.678.0
```

Once the load balancer is ready, you can access your new Sourcegraph instance at the returned IP address in your browser via HTTP. Accessing the IP address with HTTPS returns errors because TLS must be enabled first.

It might take about 10 minutes for the load balancer to be fully ready. In the meantime, you can access Sourcegraph using the port forward method as described below.

#### Port forward

Forward the remote port so that you can access Sourcegraph without network configuration temporarily.

```bash
kubectl port-forward svc/sourcegraph-frontend 3080:30080
```

You should now be able to access your new Sourcegraph instance at http://localhost:3080  🎉

### Optional Step: Deploy monitoring stacks 

**IMPORTANT**: RBAC is required for the monitoring stacks to work properly.

If RBAC is enabled in your cluster, we strongly recommend you to deploy the monitoring stacks for Sourcegraph.

```bash
$ kubectl apply -l deploy=sourcegraph -k https://github.com/sourcegraph/deploy-sourcegraph/new/quick-start/monitoring?ref=v4.3.1
```

### Further configuration

The steps above have guided you to deploy Sourcegraph using the [quick-start/gke/base](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/new/quick-start/gke/base) overlay preconfigured by us.

If you would like to make other configurations to your existing instance, you can create a new overlay using its kustomization.yaml file shown below and build on top of it. For example, you can upgrade your instance from size XS to L, or add the monitoring stacks.

```yaml
# new/overlays/your_gke_deployment/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: default
resources:
# Sourcegraph Main Stacks
- ../../base/sourcegraph
components:
# Use resources for a size-XS instance
- ../../components/sizes/xs
# Apply configurations for GKE
- ../../components/gke/configure
```

#### Enable TLS

Once you have created a new overlay using the kustomization file from our quick-start overlay for gke, we strongly recommend you to: 
- create a DNS A record for your Sourcegraph instance domain
- enable TLS is highly recommended. 

If you would like to enable TLS with your own certificate, please read the [TLS configuration guide](configure.md#tls) for detailed instructions.

##### Google-managed certificate

Step 1: Add the `gke ingress` component to your overlay:

```yaml
# new/overlays/your_gke_deployment/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: default
resources:
- ../../base/sourcegraph
components:
- ../../components/sizes/xs
- ../../components/gke/configure
- ../../components/gke/mange-cert
```

Step 2: Add your Google-managed certificate name to the `overlay.config` file using the `GKE_MANAGED_CERT_NAME` variable:

```yaml
# new/overlays/your_gke_deployment/config/overlay.config
GKE_MANAGED_CERT_NAME=your-managed-cert-name
```
