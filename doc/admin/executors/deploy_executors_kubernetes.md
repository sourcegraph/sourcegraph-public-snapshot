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
  1. `helm upgrade --install -f values.yaml --version 5.0.2 sg-executor sourcegraph/sourcegraph-executor`
5. Confirm executors are working are working by checking the _Executors_ page under **Site admin > Executors > Instances** .


For more information on the components being deployed see the [Executors readme](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/configure/executors/README.md).

## Note

Executors deployed in kubernetes do not use [Firecracker](index.md#how-it-works), meaning they require [privileged access](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) to the docker daemon running in a sidecar alongside the executor pod.

If you have security concerns, consider deploying via [terraform](deploy_executors_terraform.md) or [installing the binary](deploy_executors_binary.md) directly.


<!--
# Deploying Sourcegraph executors on Kubernetes

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This deployment is experimental and may change in the future.
</p>

<p><b>We're very much looking for input and feedback on this feature.</b> You can either <a href="https://about.sourcegraph.com/contact">contact us directly</a>, <a href="https://github.com/sourcegraph/sourcegraph">file an issue</a>, or <a href="https://twitter.com/sourcegraph">tweet at us</a>.</p>
</aside>

> NOTE: This feature is available in Sourcegraph 5.0.2 and later.

[Kubernetes manifests](https://github.com/sourcegraph/deploy-sourcegraph) are provided to deploy Sourcegraph Executors
on a running Kubernetes cluster. If you are deploying Sourcegraph with helm, charts are
available [here](https://github.com/sourcegraph/deploy-sourcegraph-helm).

## Requirements

### RBAC Roles

Executors interact with the Kubernetes API to handle jobs. The following are the RBAC Roles needed to run Executors on
Kubernetes.

| API Groups | Resources          | Verbs                     | Reason                                                                                    |
|------------|--------------------|---------------------------|-------------------------------------------------------------------------------------------|
| `batch`    | `jobs`             | `create`, `delete`, `get` | Executors create Job pods to run processes. Once Jobs are completed, they are cleaned up. |
|            | `pods`, `pods/log` | `get`, `list`, `watch`    | Executors need to look up and steam logs from the Job Pods.                               |

See
the [example Role YAML](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/blob/enterprise/cmd/executor/kubernetes/executor-batches.Role.yml)
for details.

### Docker Image

The Executor Docker image is available on Docker Hub
at [`sourcegraph/executor-kubernetes`](https://hub.docker.com/r/sourcegraph/executor-kubernetes/tags).

### Environment Variables

The following are Environment Variables that are specific to the Kubernetes runtime. These environment variables can be
set on the Executor `Deployment` and will configure the `Job`s that it spawns.

| Name                                                         | Default Value     | Description                                                                                                                            |
|--------------------------------------------------------------|:------------------|----------------------------------------------------------------------------------------------------------------------------------------|
| EXECUTOR_KUBERNETES_CONFIG_PATH                              | N/A               | The path to the Kubernetes configuration file. If not specified, the in cluster config is used.                                        |
| EXECUTOR_KUBERNETES_NODE_NAME                                | N/A               | The name of the Kubernetes Node to create Jobs in. If not specified, the Pods are created in the first available node.                 |
| EXECUTOR_KUBERNETES_NODE_SELECTOR                            | N/A               | A comma separated list of values to use as a node selector for Kubernetes Jobs. e.g. `foo=bar,app=my-app`                              |
| EXECUTOR_KUBERNETES_NODE_REQUIRED_AFFINITY_MATCH_EXPRESSIONS | N/A               | The JSON encoded required affinity match expressions for Kubernetes Jobs. e.g. `[{"key": "foo", "operator": "In", "values": ["bar"]}]` |
| EXECUTOR_KUBERNETES_NODE_REQUIRED_AFFINITY_MATCH_FIELDS      | N/A               | The JSON encoded required affinity match fields for Kubernetes Jobs. e.g. `[{"key": "foo", "operator": "In", "values": ["bar"]}]`      |
| EXECUTOR_KUBERNETES_NAMESPACE                                | `default`         | The namespace to create the Jobs in.                                                                                                   |
| EXECUTOR_KUBERNETES_PERSISTENCE_VOLUME_NAME                  | `sg-executor-pvc` | The name of the Executor Persistence Volume. Must match the `PersistentVolumeClaim` configured for the instance.                       |
| EXECUTOR_KUBERNETES_RESOURCE_LIMIT_CPU                       | N/A               | The maximum CPU resource for Kubernetes Jobs.                                                                                          |
| EXECUTOR_KUBERNETES_RESOURCE_LIMIT_MEMORY                    | `12Gi`            | The maximum memory resource for Kubernetes Jobs.                                                                                       |
| EXECUTOR_KUBERNETES_RESOURCE_REQUEST_CPU                     | N/A               | The minimum CPU resource for Kubernetes Jobs.                                                                                          |
| EXECUTOR_KUBERNETES_RESOURCE_REQUEST_MEMORY                  | `12Gi`            | The minimum memory resource for Kubernetes Jobs.                                                                                       |
| KUBERNETES_JOB_RETRY_BACKOFF_LIMIT                           | `600`             | The number of attempts to try when checking if a Job has completed.                                                                    |
| KUBERNETES_JOB_RETRY_BACKOFF_DURATION                        | `100ms`           | The duration of the backoff when checking if a Job has completed.                                                                      |
| KUBERNETES_KEEP_JOBS                                         | `false`           | If true, Kubernetes jobs will not be deleted after they complete. Useful for debugging.                                                |

See other possible Environment Variables [here](./deploy_executors_binary.md#step-2-setup-environment-variables).

### Firewall Rules

The following are Firewall rules that are _highly recommended_ when running Executors on Kubernetes in a Cloud
Environment.

- Disable access to internal resources.
- Disable access to `5.0.2.254` (AWS / GCP Instance Metadata Service).

### Batch Changes

To run [Batch Changes](../../batch_changes/index.md) on
Kubernetes, [Native Server-Side Batch Changes](./native_server_side_batch_changes.md) must be enabled.

### Example Configuration YAML

See
the [local development YAMLs](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/blob/enterprise/cmd/executor/kubernetes)
for an example of how to configure the Executor in Kubernetes.

## Deployment

Executors on Kubernetes require specific RBAC rules to be configured in order to operate correctly.
See [RBAC Roles](#rbac-roles) for more information.

### Step-by-step Guide

Ensure you have the following tools installed.

- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)

#### Deployment via kubectl (Kubernetes manifests)

1. Clone the [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) repository to your local machine.
2. Run `cd deploy-sourcegraph/configure/executors`.
3. Configure
   the [Executor environment variables](https://docs.sourcegraph.com/admin/deploy_executors_binary#step-2-setup-environment-variables)
   in the `executor/executor.deployment.yaml` file.
4. Run  `kubectl apply -f . --recursive` to deploy all components.
5. Confirm executors are working by checking the _Executors_ page under _Site Admin_ > _Executors_ > _Instances_ .

## Note

Executors deployed on Kubernetes do not use [Firecracker](executors.md#how-it-works).

If you have security concerns, consider deploying via [terraform](deploy_executors_terraform.md)
or [installing the binary](deploy_executors_binary.md) directly.

-->
