# Deploying Sourcegraph executors on Kubernetes

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This deployment is experimental and may change in the future.
</p>

<p><b>We're very much looking for input and feedback on this feature.</b> You can either <a href="https://about.sourcegraph.com/contact">contact us directly</a>, <a href="https://github.com/sourcegraph/sourcegraph">file an issue</a>, or <a href="https://twitter.com/sourcegraph">tweet at us</a>.</p>
</aside>

> NOTE: This feature is available in Sourcegraph 5.1.0 and later.

[Kubernetes manifests](https://github.com/sourcegraph/deploy-sourcegraph) are provided to deploy Sourcegraph Executors
on a running Kubernetes cluster. If you are deploying Sourcegraph with helm, charts are
available [here](https://github.com/sourcegraph/deploy-sourcegraph-helm).

## Requirements

### RBAC Roles

Executors interact with the Kubernetes' API to handle jobs. The following are the RBAC Roles needed to run Executors on Kubernetes.

| API Groups | Resources          | Verbs                     | Reason                                                                                    |
|------------|--------------------|---------------------------|-------------------------------------------------------------------------------------------|
| `batch`    | `jobs`             | `create`, `delete`, `get` | Executors create Job pods to run processes. Once Jobs are completed, they are cleaned up. |
|            | `pods`, `pods/log` | `get`, `list`, `watch`    | Executors need to look up and steam logs from the Job Pods.                               |


### Environment Variables

The following are Environment Variables that are specific to the Kubernetes runtime. See other possible Environment
Variables [here](./deploy_executors_binary.md#step-2-setup-environment-variables).

| Name                                        | Default Value  | Description                                                                                                            |
|---------------------------------------------|:---------------|------------------------------------------------------------------------------------------------------------------------|
| EXECUTOR_KUBERNETES_CONFIG_PATH             | N/A            | The path to the Kubernetes configuration file. If not specified, the in cluster config is used.                        |
| EXECUTOR_KUBERNETES_NODE_NAME               | N/A            | The name of the Kubernetes Node to create Jobs in. If not specified, the Pods are created in the first available node. |
| EXECUTOR_KUBERNETES_NAMESPACE               | `default`      | The namespace of to assign the Jobs to.                                                                                |
| EXECUTOR_KUBERNETES_PERSISTENCE_VOLUME_NAME | `executor-pvc` | The name of the Executor Persistence Volume. Must match the `PersistentVolumeClaim` configured for the instance.       |
| EXECUTOR_KUBERNETES_RESOURCE_LIMIT_CPU      | `1`            | The CPU resource limit for Kubernetes Jobs.                                                                            |
| EXECUTOR_KUBERNETES_RESOURCE_LIMIT_MEMORY   | `1Gi`          | The memory resource limit for Kubernetes Jobs.                                                                         |
| EXECUTOR_KUBERNETES_RESOURCE_REQUEST_CPU    | `1`            | The maximum CPU resource limit for Kubernetes Jobs.                                                                    |
| EXECUTOR_KUBERNETES_RESOURCE_REQUEST_MEMORY | `1Gi`          | The maximum memory resource limit for Kubernetes Jobs.                                                                 |

### Example Configuration YAML

The following is an example YAML file that can be used to deploy an Executor instance on Kubernetes.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: executor
spec:
  selector:
    matchLabels:
      app: executor
  template:
    metadata:
      labels:
        app: executor
    spec:
      hostNetwork: true
      serviceAccountName: executor-service-account
      containers:
        - name: executor
          # Update the image tag to the version that matches your Sourcegraph instance.
          image: sourcegraph/executor-kubernetes:5.1.0
          imagePullPolicy: Never
          ports:
            - containerPort: 8080
          env:
            - name: EXECUTOR_FRONTEND_URL
              # The URL of the Sourcegraph instance to connect to.
              value: https://my.sourcegraph.com
            - name: EXECUTOR_FRONTEND_PASSWORD
              # The shared secret between the executor and the Sourcegraph instance.
              value: my-password
            - name: EXECUTOR_QUEUE_NAME
              # The name of the queue to pull jobs from. Either batches or codeintel.
              value: batches
            - name: EXECUTOR_MAXIMUM_NUM_JOBS
              # The maximum number of jobs to run concurrently.
              value: "10"
          volumeMounts:
            - mountPath: /data
              name: executor-volume
      volumes:
        - name: executor-volume
          persistentVolumeClaim:
            claimName: executor-pvc
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: executor-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      # Configure this to the possible cumulative size of the Repositories that can be worked at any given time.
      storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: executor
  labels:
    app: executor
spec:
  selector:
    app: executor
  ports:
    - name: http
      port: 8080
      targetPort: 8080
  type: LoadBalancer
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: executor-service-account
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: default
  name: executor-job-role
rules:
  - apiGroups: ["batch"]
    resources: ["jobs"]
    verbs: ["get", "create", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: executor-log-role
rules:
  - apiGroups: [""]
    resources: ["pods", "pods/log"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: executor-job-role-binding
  namespace: default
subjects:
  - kind: ServiceAccount
    name: executor-service-account
    namespace: default
roleRef:
  kind: Role
  name: executor-job-role
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: executor-log-role-binding
subjects:
  - kind: ServiceAccount
    name: executor-service-account
roleRef:
  kind: Role
  name: executor-log-role
  apiGroup: rbac.authorization.k8s.io
```

## Deployment

Executors on Kubernetes require specific RBAC rules to be configured in order to operate correctly.
See [RBAC Roles](#rbac-roles) for more information.

### Step-by-step Guide

Ensure you have the following tools installed.
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
- [Helm](https://helm.sh/) if you're installing Sourcegraph with helm.

#### Deployment via kubectl (Kubernetes manifests)

1. Clone the [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) repository to your local machine.
2. Run `cd deploy-sourcegraph/configure/executors`.
3. Configure the [Executor environment variables](https://docs.sourcegraph.com/admin/deploy_executors_binary#step-2-setup-environment-variables) in the `executor/executor.deployment.yaml` file.
4. Run  `kubectl apply -f . --recursive` to deploy all components.
5. Confirm executors are working by checking the _Executors_ page under _Site Admin_ > _Executors_ > _Instances_ .

#### Deployment via Helm

1. Clone the [deploy-sourcegraph-helm](https://github.com/sourcegraph/deploy-sourcegraph-helm) repository to your local machine.
2. Run `cd deploy-sourcegraph-helm/charts/sourcegraph-executor`.
3. Edit the `values.yaml` with any other customizations you may require.
4. Run the following command:
   1. `helm upgrade --install -f values.yaml --version 5.1.0 sg-executor sourcegraph/sourcegraph-executor`
5. Confirm executors are working by checking the _Executors_ page under _Site Admin_ > _Executors_ > _Instances_ .


For more information on the components being deployed see the [Executors readme](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/configure/executors/README.md).

## Note

Executors deployed in kubernetes do not use [Firecracker](executors.md#how-it-works).

If you have security concerns, consider deploying via [terraform](deploy_executors_terraform.md) or [installing the binary](deploy_executors_binary.md) directly.


