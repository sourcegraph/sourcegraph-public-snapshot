# Kustomize

> WARNING: Kustomize can be used **with** Helm to configure Sourcegraph (see [this guidance](helm.md#integrate-kustomize-with-helm-chart)) but this is only recommended as a temporary workaround while Sourcegraph adds to the Helm chart to support previously unsupported customizations. 
> If you have yet to deploy Sourcegraph, it is highly recommended to us Helm for the deployment and configuration ([Using Helm with Sourcegraph](helm.md)). 

Sourcegraph supports the use of [Kustomize](https://kustomize.io) to modify and customize our Kubernetes manifests. Kustomize is a template-free way to customize configuration with a simple configuration file.

Some benefits of using Kustomize to generate manifests instead of modifying the base directly include:

- Reduce the odds of encountering a merge conflict when [updating Sourcegraph](update.md) - they allow you to separate your unique changes from the upstream base files Sourcegraph provides.
- Better enable Sourcegraph to support you if you run into issues, because how your deployment varies from our defaults is encapsulated in a small set of files.

## Using Kustomize

### General premise

In general, we recommend that customizations work like this:

1. Create, customize, and apply overlays for your deployment
2. Ensure the services came up correctly, then commit all the customizations to the new branch

  ```sh
  git add /overlays/$MY_OVERLAY/*
  # Keeping all overlays contained to a single commit allows for easier cherry-picking
  git commit amend -m "overlays: update $MY_OVERLAY"
  ```

See the [overlays guide](#overlays) to learn about the [overlays we provide](#provided-overlays) and [how to create your own overlays](#custom-overlays).

## Overlays

An [*overlay*](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/#bases-and-overlays) specifies customizations for a base directory of Kubernetes manifests, in this case the `base/` directory in the [reference repository](#reference-repository). 

Overlays can:

- Be used for example to change the number of replicas, change a namespace, add a label, etc
- Refer to other overlays that eventually refer to the base (forming a directed acyclic graph with the base as the root)

### Using overlays

Overlays can be used in one of two ways:

- With `kubectl`: Starting with `kubectl` client version 1.14 `kubectl` can handle `kustomization.yaml` files directly. When using `kubectl` there is no intermediate step that generates actual manifest files. Instead the combined resources from the overlays and the base are directly sent to the cluster. This is done with the `kubectl apply -k` command. The argument to the command is a directory containing a `kustomization.yaml` file.
- With`kustomize`: This generates manifest files that are then applied in the conventional way using `kubectl apply -f`.

The overlays provided in our [overlays directory](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/overlays) rely on the `kustomize` tool and the `overlay-generate-cluster.sh` script in the `root` directory to generate the manifests. There are two reasons why it was set up like this:

- It avoids having to put a `kustomization.yaml` file in the `base` directory and forcing users that don't use overlays
  to deal with it (unfortunately `kubectl apply -f` doesn't work if a `kustomization.yaml` file is in the directory).
- It generates manifests instead of applying them directly. This provides opportunity to additionally validate the files
  and also allows using `kubectl apply -f` with `--prune` flag turned on (`apply -k` with `--prune` does not work correctly).

### Generating Manifests

To generate Kubernetes manifests from an overlay, run the `overlay-generate-cluster.sh` with two arguments:

- the name of the overlay
- and a path to an output directory where the generated manifests will be

For example:
```sh
#                overlay directory name    output directory
#                                 |             |
./overlay-generate-cluster.sh my-overlay generated-cluster
```

After executing the script you can apply the generated manifests from the `generated-cluster` directory:

```sh
kubectl apply --prune -l deploy=sourcegraph -f generated-cluster --recursive
```

We recommend that you:

- [Update the `./overlay-generate-cluster` script](./operations.md#applying-manifests) to apply the generated manifests from the `generated-cluster` directory with something like the above snippet
- Commit your overlays changes separately - see our [customization guide](#customizations) for more details.

You can now get started with using overlays:

- [Provided overlays](#provided-overlays)
- [Custom overlays](#custom-overlays)

### Provided overlays

Overlays provided out-of-the-box are in the subdirectories of [`deploy-sourcegraph/overlays`](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/overlays) and are documented here.

#### Namespaced overlay

This overlay adds a namespace declaration to all the manifests.

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

#### Storageclass

By default Sourcegraph is configured to use a storage class called `sourcegraph`. If you wish to use an alternate name, you can use this overlay to change all `storageClass` references in the manifests. 

You need to create the storageclass if it doesn't exist yet. See [these docs](./configure.md#configure-a-storage-class) for more instructions.

1. To use it, update the following two files, `replace-storageclass-name-pvc.yaml` and `replace-storageclass-name-sts.yaml` in the `deploy-sourcegraph/overlays/storageclass` directory with your storageclass name.

1. To generate to the cluster, execute the following command:
```shell script
./overlay-generate-cluster.sh storageclass generated-cluster
```

1. After executing the script you can apply the generated manifests from the `generated-cluster` directory:

```shell script
kubectl apply --prune -l deploy=sourcegraph -f generated-cluster --recursive
```

1. Ensure the persistent volumes have been created in the correct storage class by running the following command and inspecting the output: 

```shell script
kubectl get pvc
```

#### Non-privileged create cluster overlay

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

#### Non-privileged overlay

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

#### Migrate-to-nonprivileged overlay

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

#### minikube overlay

This kustomization deletes resource declarations and storage classnames to enable running Sourcegraph on minikube.

To use it, execute the following command from the `root` directory:

```sh
./overlay-generate-cluster.sh minikube generated-cluster
```

After executing the script you can apply the generated manifests from the generated-cluster directory:

```sh
minikube start
kubectl create namespace ns-sourcegraph
kubectl -n ns-sourcegraph apply --prune -l deploy=sourcegraph -f generated-cluster --recursive
kubectl -n ns-sourcegraph expose deployment sourcegraph-frontend --type=NodePort --name sourcegraph --port=3080 --target-port=3080
minikube service list
```

> NOTE: For Mac Users, run `minikube service sourcegraph -n ns-sourcegraph` to open the newly deployed Sourcegraph in your browser

To tear it down:

```sh
kubectl delete namespaces ns-sourcegraph
minikube stop
```

### Custom overlays

To create your own [overlays](#overlays), first [set up your deployment reference repository to enable customizations](./configure.md#getting-started).

Then, within the `overlays` directory of the [reference repository](./index.md#reference-repository), create a new directory for your overlay along with a `kustomization.yaml`.

```text
deploy-sourcegraph
 |-- overlays
 |    |-- my-new-overlay
 |    |    +-- kustomization.yaml
 |    |-- bases
 |    +-- ...
 +-- ...
```

Within `kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
# Only include resources from 'overlays/bases' you are interested in modifying
# To learn more about bases: https://kubectl.docs.kubernetes.io/references/kustomize/glossary/#base
resources:
  - ../bases/deployments
  - ../bases/rbac-roles
  - ../bases/pvcs
```

You can then define patches, transformations, and more. A complete reference is available [here](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/).
To get started, we recommend you explore writing your own [`patches`](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/patches/), or the more specific variants:

- [`patchesStrategicMerge`](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/patchesstrategicmerge/)
- [`patchesJson6902`](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/patchesjson6902/)

To avoid complications with reference cycles an overlay can only reference resources inside the directory subtree of the directory it resides in (symlinks are not allowed either).

Learn more in the [`kustomization` documentation](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/).
You can also explore how our [provided overlays](#provided-overlays) use patches, for reference: [`deploy-sourcegraph` usage of patches](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/deploy-sourcegraph%24+lang:YAML+patches:+:%5B_%5D+OR+patchesStrategicMerge:+:%5B_%5D+OR+patchesJson6902:+:%5B_%5D+count:999&patternType=structural).

Once you have created your overlays, refer to our [overlays guide](#generating-manifests) to generate and apply your changes.
