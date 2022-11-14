---
title: Install Sourcegraph locally with minikube
---

# Install Sourcegraph with minikube

This guide will take you through how to set up a Sourcegraph instance locally with [minikube](https://minikube.sigs.k8s.io/docs/), a tool that lets you run a single-node Kubernetes cluster on your local machine, using our [minikube overlay](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph@master/-/tree/overlays/minikube).

## Sourcegraph minikube overlay

The [Sourcegraph minikube overlay](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph@master/-/tree/overlays/minikube) deletes resource declarations and storage classnames to enable running Sourcegraph on minikube locally with less resources, as it normally takes a lot more of resources to run Sourcegraph at a production level. See our docs on creating [custom overlays](../kubernetes/kustomize.md#overlays) if you would like to customize the overlay.

## Prerequisites

Following are the prerequisites for running Sourcegraph with minikube on your local machine.

- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [minikube](https://minikube.sigs.k8s.io/docs/start/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- Enable Kubernetes in Docker Desktop. See the [official docs](https://docs.docker.com/desktop/kubernetes/#enable-kubernetes) for detailed instruction.

> NOTE: Running Sourcegraph on minikube locally requires a minimum of **8 CPU** and **32GB memory** assigned to your Kubernetes instance in Docker.

## Deploy

1\. Start a minikube cluster

```sh
# Docker must be running in the background
$ minikube start
```

2\. Create a clone of our Kubernetes deployment repository: [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph)

```sh
$ git clone https://github.com/sourcegraph/deploy-sourcegraph.git
```

3\. Check out the branch of the version you would like to deploy

```sh
# Example: git checkout v4.1.0
$ git checkout $VERSION-NUMBER
```

4\. Apply the [Sourcegraph minikube overlay](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph@master/-/tree/overlays/minikube) by running the following command in the root directory of your local copy of the [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) repository

```sh
$ ./overlay-generate-cluster.sh minikube generated-cluster
```

5\. Create the `ns-sourcegraph` namespace

```sh
$ kubectl create namespace ns-sourcegraph
```

6\. Apply the generated manifests from the `generated-cluster` directory

```sh
$ kubectl -n ns-sourcegraph apply --prune -l deploy=sourcegraph -f generated-cluster --recursive
```

7\. Make sure all the pods are up and running before moving to the next step

```sh
$ kubectl get pods -A
```

<img class="screenshot w-100" src="https://user-images.githubusercontent.com/68532117/141348352-a38dec9e-7166-40d7-a64e-019339732248.png" alt="minikube-terminal-screen"/>

> WARNING: The deployment time depends on how much resources have been assigned to your instance. You will need to add more resources to your Kubernetes instance through the Docker Desktop dashboard if some of your pods are stucks in the `Pending` state.

8\. Create a Service object that exposes the deployment

```sh
$ kubectl -n ns-sourcegraph expose deployment sourcegraph-frontend --type=NodePort --name sourcegraph --port=3080 --target-port=3080
```

9\.  Get the minikube node IP to access the nodePorts

```sh
$ minikube ip
```

10\.  Get the minikube service endpoint

```sh
# To get the service endpoint
$ minikube service list
# To get the service endpoint for Sourcegraph specifically
$ minikube service sourcegraph -n ns-sourcegraph
# To get the service endpoint URL for Sourcegraph specificlly
$ minikube service --url sourcegraph -n ns-sourcegraph

# Example return: http://127.0.0.1:32034
```

If you are on **Linux**, an URL will then be displayed in the service list if the instance has been deployed successfully

```
|----------------|-------------------------------|--------------|---------------------------|
|   NAMESPACE    |             NAME              | TARGET PORT  |            URL            |
|----------------|-------------------------------|--------------|---------------------------|
| default        | kubernetes                    | No node port |
| kube-system    | kube-dns                      | No node port |
| ns-sourcegraph | backend                       | No node port |
| ns-sourcegraph | codeinsights-db               | No node port |
| ns-sourcegraph | codeintel-db                  | No node port |
| ns-sourcegraph | github-proxy                  | No node port |
| ns-sourcegraph | gitserver                     | No node port |
| ns-sourcegraph | grafana                       | No node port |
| ns-sourcegraph | indexed-search                | No node port |
| ns-sourcegraph | indexed-search-indexer        | No node port |
| ns-sourcegraph | jaeger-collector              | No node port |
| ns-sourcegraph | jaeger-query                  | No node port |
| ns-sourcegraph | minio                         | No node port |
| ns-sourcegraph | pgsql                         | No node port |
| ns-sourcegraph | precise-code-intel-worker     | No node port |
| ns-sourcegraph | prometheus                    | No node port |
| ns-sourcegraph | query-runner                  | No node port |
| ns-sourcegraph | redis-cache                   | No node port |
| ns-sourcegraph | redis-store                   | No node port |
| ns-sourcegraph | repo-updater                  | No node port |
| ns-sourcegraph | searcher                      | No node port |
| ns-sourcegraph | sourcegraph                   |         3080 | http://127.0.0.1:32034 |
| ns-sourcegraph | sourcegraph-frontend          | No node port |
| ns-sourcegraph | sourcegraph-frontend-internal | No node port |
| ns-sourcegraph | symbols                       | No node port |
| ns-sourcegraph | syntect-server                | No node port |
| ns-sourcegraph | worker                        | No node port |
|----------------|-------------------------------|--------------|---------------------------|
```

That's it! You can now access the local Sourcegraph instance in browser using the URL or IP address retrieved from the previous steps 🎉

<img class="screenshot" src="https://user-images.githubusercontent.com/68532117/141357183-905d0dbe-2d40-4dec-98b1-0a1cb13b0cf4.png" alt="minikube-startup-screen"/>

## Upgrade

Please refer to the [upgrade docs for all Sourcegraph kubernetes instances](../kubernetes/operations.md#upgrade).

## Downgrade

Same instruction as upgrades.

## Uninstall

Steps to remove your Sourcegraph minikube instance:

1\. Delete the `ns-sourcegraph` namespace

```sh
$ kubectl delete namespaces ns-sourcegraph
```

2\. Stop the minikube cluster

```sh
$ minikube stop
```

## Other userful commands

#### Un-expose sourcegraph

```sh
$ kubectl delete service sourcegraph -n ns-sourcegraph
```

#### Gets a list of deployed services and cluster IP

```sh
$ kubectl get svc -n ns-sourcegraph
```

#### Deletes the minikube cluster

```sh
$ minikube delete
```

## Resources

- [Customizations](https://docs.sourcegraph.com/admin/install/kubernetes/configure#customizations)
- [Introduction to Kubectl and Kustomize](https://kubectl.docs.kubernetes.io/guides/introduction/)
- [List of commonly used Kubernetes commands](https://sourcegraph.github.io/support-generator/)
