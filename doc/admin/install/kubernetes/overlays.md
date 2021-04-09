# Overlays

- [Overlays](#overlays)
  - [Overlay basic principles](#overlay-basic-principles)
  - [Handling overlays in this repository](#handling-overlays-in-this-repository)
  - [Namespaced overlay](#namespaced-overlay)
  - [Non-root overlay](#non-root-overlay)
  - [Non-privileged overlay](#non-privileged-overlay)
- [Git Strategies when using overlays to reduce conflicts](#git-strategies-with-overlays)

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

## Handling overlays in this repository

The overlays provided in this repository rely on the `kustomize` tool and the `overlay-generate-cluster.sh` script in the
root directory of this repository to generate the manifests. There are two reasons why it was set up like this:

- It avoids having to put a `kustomization.yaml` file in the `base` directory and forcing users that don't use overlays
  to deal with it (unfortunately `kubectl apply -f` doesn't work if a `kustomization.yaml` file is in the directory).

- It generates manifests instead of applying them directly. This provides opportunity to additionally validate the files
  and also allows using `kubectl apply -f` with `--prune` flag turned on (`apply -k` with `--prune` does not work correctly).

To generate the manifests run the `overlay-generate-cluster.sh` with two arguments: 
- the name of the overlay
- and a path to an output directory where the generated manifests will be

Example (assuming you are in the root directory of this repository):

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


## Steps to setup overlay

1. Create a new branch for the customizations from the current release branch

  ```shell
  git checkout 3.26
  git checkout -b 3.26-kustomize   
  ```
  
2. Create and customize the overlays for your deployment

3. Generate the overlays with the `./overlay-generate-cluster` script

4. `kubectl apply` the generated cluster:

5. Ensure the services came up correctly, then commit all the customizations to the new branch

   ```shell
   git add /Overalys/$MY_OVERLAYS/*
   git commit amend -m "Message" # Keeping all overlays contained to a single commit allows for easier cherry-picking
   ```


### Namespaced overlay

This overlay adds a namespace declaration to all the manifests. 

1. Change the namespace by replacing `ns-sourcegraph` to name of your choice (`<REPLACE THIS>` in this example) in the
[overlays/namespaced/kustomization.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/overlays/namespaced/kustomization.yaml) file.

1. Execute this from the root directory of the repository to generate:

  ```shell script
  ./overlay-generate-cluster.sh namespaced generated-cluster
  ```

1. Create the namespace if it doesn't exist yet:

  ```kubectl create namespace ns-<REPLACE THIS>
  kubectl label namespace ns-<REPLACE THIS> name=ns-sourcegraph
  ```

1. Execute this from the root directory of the repository to apply the generated manifests from the `generated-cluster` directory:

  ```kubectl apply -n ns-<REPLACE THIS> --prune -l deploy=sourcegraph -f generated-cluster --recursive
  ```

1. Run `kubectl get pods -A` to check for the namespaces and their status --it should now be up and running 🎉


### Non-root overlay

The manifests in the `base` directory specify user `root` for all containers. This overlay changes the specification to be
a non-root user.

If you are starting a fresh installation use the overlay `non-root-create-cluster`. After creation you can use the overlay
`non-root`.

If you already are running a Sourcegraph instance using user `root` and want to convert to running with non-root user then
you need to apply a migration step that will change the permissions of all persistent volumes so that the volumes can be
used by the non-root user. This migration is provided as overlay `migrate-to-nonroot`. After the migration you can use
overlay `non-root`.

### Non-privileged overlay

This overlays goes one step further than the `non-root` overlay by also removing cluster roles and cluster role bindings.

If you are starting a fresh installation use the overlay `non-privileged-create-cluster`. After creation you can use the overlay
`non-privileged`.


## Upgrading sourcegraph with an overlay

1. Create a new branch from the origin branch to the version upgrading to
  ```shell
  git checkout 3.23
  ```

1. Create a new branch for this specific version
  ```shell
  git checkout -b "nameofbranch/version"
  ```

1. Cherry pick the customizations for this version

1. git log <name of previous branch> (Pick the latest SHA from this log)

1. git cherry-pick <Latest SHA> (Always cherry-pick from the latest minor version)

1. Generate the overlays from the provided script `kubectl apply`


# Troubleshooting

> error: error retrieving RESTMappings to prune: invalid resource networking.k8s.io/v1, Kind=Ingress, Namespaced=true: no matches for kind "Ingress" in version "networking.k8s.io/v1"

- See the ["Configure network access"](https://docs.sourcegraph.com/admin/install/kubernetes/configure#configure-network-access)
- Check for duplicate `sourcegraph-frontend` using `kubectl get ingresses -A`
- Delete duplicate using `kubectl delete ingress sourcegraph-frontend -n default`