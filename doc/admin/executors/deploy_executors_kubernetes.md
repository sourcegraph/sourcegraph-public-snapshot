# Deploying Sourcegraph executors on Kubernetes

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This deployment is experimental and may change in the future.
</p>

<p><b>We're very much looking for input and feedback on this feature.</b> You can either <a href="https://about.sourcegraph.com/contact">contact us directly</a>, <a href="https://github.com/sourcegraph/sourcegraph">file an issue</a>, or <a href="https://twitter.com/sourcegraph">tweet at us</a>.</p>
</aside>

[Kubernetes manifests](https://github.com/sourcegraph/deploy-sourcegraph) are provided to deploy Sourcegraph Executors on a running Kubernetes cluster. If you are deploying Sourcegraph with helm, charts are available [here](https://github.com/sourcegraph/deploy-sourcegraph-helm).

## Deployment

Executors on kubernetes machines require privileged access to a container runtime daemon in order to operate correctly. In order to ensure maximum capability across Kubernetes versions and container runtimes, a [Docker in Docker](https://www.docker.com/blog/docker-can-now-run-within-docker/) side car is deployed with each executor pod to avoid accessing the host container runtime directly.

### Step-by-step Guide

Ensure you have the following tools installed:

  - [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
  - [Helm](https://helm.sh/) if you're installing Sourcegraph with helm.

#### Deployment via kubectl (Kubernetes manifests)

1. Clone the [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) repository to your local machine.
2. Run `cd deploy-sourcegraph/configure/executors`.
3. Configure the [Executor environment variables](https://docs.sourcegraph.com/admin/deploy_executors_binary#step-2-setup-environment-variables) in the `executor/executor.deployment.yaml` file.
4. Run  `kubectl apply -f . --recursive` to deploy all components.
5. Confirm executors are working are working by checking the _Executors_ page under **Site admin > Executors > Instances** .

#### Deployment via Helm

1. Clone the [deploy-sourcegraph-helm](https://github.com/sourcegraph/deploy-sourcegraph-helm) repository to your local machine.
2. Run `cd deploy-sourcegraph-helm/charts/sourcegraph-executor`.
3. Edit the `values.yaml` with any other customizations you may require.
4. Run the following command:
   1. `helm upgrade --install -f values.yaml --version 5.0.0 sg-executor sourcegraph/sourcegraph-executor`
5. Confirm executors are working are working by checking the _Executors_ page under **Site admin > Executors > Instances** .


For more information on the components being deployed see the [Executors readme](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/configure/executors/README.md).

## Note

Executors deployed in kubernetes do not use [Firecracker](executors.md#how-it-works), meaning they require [privileged access](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) to the docker daemon running in a sidecar alongside the executor pod.

If you have security concerns, consider deploying via [terraform](deploy_executors_terraform.md) or [installing the binary](deploy_executors_binary.md) directly.


