# Local Development

The following is a guide to setting up a local development environment for the executor within a Kubernetes cluster.

## Prerequisites

- Install [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

### Docker Desktop

- Install Docker Desktop
- Enable Kubernetes in Docker Desktop
  - Settings > Kubernetes > Enable Kubernetes

### MiniKube

- Install [MiniKube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
- Start MiniKube
  - `minikube start`

## Build Image

Run the following command to build the image.

```bash
IMAGE=executor-kubernetes ./build.sh
```

## Secrets

The frontend password should be stored in a Kubernetes secret. Run the following command to create the secret.

```bash
kubectl create secret generic executor-frontend-password --from-literal=EXECUTOR_FRONTEND_PASSWORD=hunter2hunter2hunter2
```

## Deploy

Run the following command to deploy the executor.

```bash
kubectl apply -f executor-batches.yml
```

## Verify

An executor pod should now be created. Confirm this by running the following command.

```bash
kubectl get pods
```

You should see an executor pod in the `Running` state.

You can also check the **Site admin** Page to see the registered executor.

## Cleanup

Run the following command to delete the executor.

```bash
kubectl delete -f executor-batches.yml
```
