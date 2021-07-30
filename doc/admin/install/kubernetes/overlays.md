# Overlays

- [Overlays](#overlays)
  - [Overlay basic principles](#overlay-basic-principles)
  - [Handling overlays](#handling-overlays)
  - [Git Strategies when using overlays to reduce conflicts](#git-strategies-with-overlays)
    - [Steps](#general-steps)
      - [Namespaced overlay](#namespaced-overlay)
      - [Non-privileged create cluster overlay](#non-privileged-create-cluster-overlay)
      - [Non-privileged overlay](#non-privileged-overlay)
      - [Migrate-to-nonprivileged overlay](#migrate-to-nonprivileged-overlay)
      - [minikube overlay](#minikube-overlay)
    - [Upgrading Sourcegraph with an overlay](#upgrading-sourcegraph-with-an-overlay)
  - [Troubleshooting](#troubleshooting)

## Customizing your Sourcegraph Deployment

The configuration provided in the `deploy-sourcegraph` repository contains all the defaults to deploy and run Sourcegraph. However, should you require changes to the manifests provided, 
we support the use of [Kustomize](https://kustomize.io/) to create overlays. If you're unfamiliar with overlays, read on to find out more about how to use overlays. 

## Overlay basic principles

An overlay specifies customizations for a base directory of Kubernetes manifests. The base has no knowledge of the overlay.
Overlays can be used for example to change the number of replicas, change a namespace, add a label etc. Overlays can refer to
other overlays that eventually refer to the base forming a directed acyclic graph with the base as the root.

An overlay is defined in a `kustomization.yaml` file (the name of the file is fixed and there can be only one kustomization
file in one directory). To avoid complications with reference cycles an overlay can only reference resources inside the
directory subtree of the directory it resides in (symlinks are not allowed either).

For more details about overlays please consult the `kustomize` [documentation](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/).

Using overlays and applying them to the cluster
can be done in two ways: by using `kubectl` or with the `kustomize` tool.

Starting with `kubectl` client version 1.14 `kubectl` can handle `kustomization.yaml` files directly.
When using `kubectl` there is no intermediate step that generates actual manifest files. Instead the combined resources from the
overlays and the base are directly sent to the cluster. This is done with the `kubectl apply -k` command. The argument to the
command is a directory containing a `kustomization.yaml` file.

The second way to use overlays is with the `kustomize` tool. This does generate manifest files that are then applied
in the conventional way using `kubectl apply -f`.


## Handling overlays

The overlays provided in our [overlays directory](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/overlays) rely on the `kustomize` tool and the `overlay-generate-cluster.sh` script in the
`root` directory to generate the manifests. There are two reasons why it was set up like this:

- It avoids having to put a `kustomization.yaml` file in the `base` directory and forcing users that don't use overlays
  to deal with it (unfortunately `kubectl apply -f` doesn't work if a `kustomization.yaml` file is in the directory).

- It generates manifests instead of applying them directly. This provides opportunity to additionally validate the files
  and also allows using `kubectl apply -f` with `--prune` flag turned on (`apply -k` with `--prune` does not work correctly).

To generate the manifests run the `overlay-generate-cluster.sh` with two arguments:

- the name of the overlay

- and a path to an output directory where the generated manifests will be

Example (assuming you are in the `root` directory):

```shell script
./overlay-generate-cluster.sh non-root generated-cluster
```

After executing the script you can apply the generated manifests from the `generated-cluster` directory:

```shell script
kubectl apply --prune -l deploy=sourcegraph -f generated-cluster --recursive
```

Available overlays are the subdirectories of `overlays` (only give the name of the subdirectory, not the full path as an argument).

You only need to apply one of the three overlays, each builds on the overlay listed before. So, for example, using the non-root overlay will also install Sourcegraph in a non-default namespace.


# Git strategies with overlays

One benefit of generating manifest from base instead of modifying base directly is that it reduces the odds of encountering a merge conflict when upgrading.

[Bases and Overlays](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/#bases-and-overlays) allow you to separate your unique changes from the upstream bases. 


## General Steps

1. Create a new branch for the customizations from the current release branch

  ```
  export SOURCEGRAPH_VERSION="v3.30.3"
  git checkout $SOURCEGRAPH_VERSION
  git checkout -b $SOURCEGRAPH_VERSION-kustomize   
  ```
  
1. Create and customize the overlays for your deployment

1. Generate the overlays with the `./overlay-generate-cluster` script

1. apply the generated manifests from the `generated-cluster` directory using `kubectl apply` 

1. Ensure the services came up correctly, then commit all the customizations to the new branch

    ```
    git add /Overalys/$MY_OVERLAYS/*
    git commit amend -m "Message" # Keeping all overlays contained to a single commit allows for easier cherry-picking
    ```

1. Start Sourcegraph on your local machine by temporarily making the frontend port accessible:

    ```
    kubectl port-forward svc/sourcegraph-frontend 3080:30080
    ```

1. Open http://localhost:3080 in your browser and you will see a setup page. 

1. 🎉 Congrats, you have Sourcegraph up and running! Now [configure your deployment](configure.md).


### Namespaced overlay

This overlay adds a namespace declaration to all the manifests. 

1. Create a new branch for the customizations from the current release branch

    ```
    # EXAMPLE
    export SOURCEGRAPH_VERSION="v3.30.3"
    git checkout $SOURCEGRAPH_VERSION
    git checkout -b $SOURCEGRAPH_VERSION-kustomize   
    ```

1. Change the namespace by replacing `ns-sourcegraph` to the name of your choice everywhere within the
[overlays/namespaced/](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/overlays/namespaced/) directory. 

1. Generate the overlay by running this command from the `root` directory:

    ```
    ./overlay-generate-cluster.sh namespaced generated-cluster
    ```

1. Create the namespace if it doesn't exist yet:

    ```
    kubectl create namespace ns-<EXAMPLE NAMESPACE>
    kubectl label namespace ns-<EXAMPLE NAMESPACE> name=ns-sourcegraph
    ```

1. Apply the generated manifests (from the `generated-cluster` directory) by running this command from the `root` directory:

  ```
  kubectl apply -n ns-<EXAMPLE NAMESPACE> --prune -l deploy=sourcegraph -f generated-cluster --recursive
  ```

1. Check for the namespaces and their status with:

  ```
  kubectl get pods -A
  ```

### Non-privileged create cluster overlay
This kustomization is for Sourcegraph installations in clusters with security restrictions. It runs all containers as a non root users, as well removing cluster roles and cluster role bindings and does all the rolebinding in a namespace. It configures Prometheus to work in the namespace and not require ClusterRole wide privileges when doing service discovery for scraping targets. It also disables cAdvisor.

This version and `non-privileged` need to stay in sync. This version is only used for cluster creation.

To use it, execute the following command from the `root` directory:

```
./overlay-generate-cluster.sh non-privileged-create-cluster generated-cluster
```

After executing the script you can apply the generated manifests from the generated-cluster directory:

```
kubectl create namespace ns-sourcegraph
kubectl apply -n ns-sourcegraph --prune -l deploy=sourcegraph -f generated-cluster --recursive
```

### Non-privileged overlay

This overlay is for continued use after you have successfully deployed the `non-privileged-create-cluster`. It runs all containers as a non root users, as well removing cluster roles and cluster role bindings and does all the rolebinding in a namespace. It configures Prometheus to work in the namespace and not require ClusterRole wide privileges when doing service discovery for scraping targets. It also disables cAdvisor.

To use it, execute the following command from the `root` directory:

```shell script
./overlay-generate-cluster.sh non-privileged generated-cluster
```

After executing the script you can apply the generated manifests from the generated-cluster directory:

```shell script
kubectl apply -n ns-sourcegraph --prune -l deploy=sourcegraph -f generated-cluster --recursive
```

If you are starting a fresh installation use the overlay `non-privileged-create-cluster`. After creation you can use the overlay
`non-privileged`.

### Migrate-to-nonprivileged overlay

If you already are running a Sourcegraph instance using user `root` and want to convert to running with non-root user then
you need to apply a migration step that will change the permissions of all persistent volumes so that the volumes can be
used by the non-root user. This migration is provided as overlay `migrate-to-nonprivileged`. After the migration you can use
overlay `non-privileged`. If you have previously deployed your cluster in a non-default namespace, be sure to edit the `kustomization.yaml` file in the overlays directly to ensure the files are generated with the correct namespace. 

This kustomization injects initContainers in all pods with persistent volumes to transfer ownership of directories to specified non-root users. It is used for migrating existing installations to a non-privileged environment.

```
./overlay-generate-cluster.sh migrate-to-nonprivileged generated-cluster
```

After executing the script you can apply the generated manifests from the generated-cluster directory:

```
kubectl apply --prune -l deploy=sourcegraph -f generated-cluster --recursive
```

### minikube overlay

This kustomization deletes resource declarations and storage classnames to enable running Sourcegraph on minikube.

To use it, execute the following command from the `root` directory:

```
./overlay-generate-cluster.sh minikube generated-cluster
```

After executing the script you can apply the generated manifests from the generated-cluster directory:

```
minikube start
kubectl create namespace ns-sourcegraph
kubectl -n ns-sourcegraph apply --prune -l deploy=sourcegraph -f generated-cluster --recursive
kubectl -n ns-sourcegraph expose deployment sourcegraph-frontend --type=NodePort --name sourcegraph --port=3080 --target-port=3080
minikube service list
```

To tear it down:

```
kubectl delete namespaces ns-sourcegraph
minikube stop
```


## Upgrading Sourcegraph with an overlay

1. Create a new branch from the origin branch to the version upgrading to

    ```
    export SOURCEGRAPH_VERSION="v3.30.3"
    git checkout $SOURCEGRAPH_VERSION
    ```

1. Create a new branch for this specific version

    ```
    git checkout -b "nameofbranch/version"
    ```

1. Cherry pick the customizations for this version

1. git log <name of previous branch> (Pick the latest SHA from this log)

1. git cherry-pick <Latest SHA> (Always cherry-pick from the latest minor version)

1. Generate the overlays from the provided script `kubectl apply`


# Troubleshooting

See the [Troubleshooting docs](troubleshoot.md).
