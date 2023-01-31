# Migration Docs for Kustomize

The old method of deploying Sourcegraph with custom scripts is now [outdated and deprecated](../deprecated/index.md). Instead, the new setup uses Kustomize, a Kubernetes-native tool, for configurations. This guide explains how to migrate from the old setup to the new one.

### Why migrate?

Here are the benefits of the new Kustomize setup compared to the old setup:

- Improved security defaults:
 * Runs in non-privileged mode
 * Uses non-root users
 * Does not require RBAC resources
- Streamlined resource allocation process:
 * Allocates resources based on the size of the instance
 * Optimized through load testing
 * The searcher and symbols use StatefulSets and do not require ephemeral storage
- Utilizes the Kubernetes-native tool Kustomize:
 * Built into kubectl
 * No additional scripting required
 * More extensible and composable
 * Highly reusable that enables creation of multiple instances with the same base resources and components
- Effortless configurations:
 * A comprehensive list of components pre-configured for different use cases
 * Designed to work seamlessly with Sourcegraph’s design and functionality
 * Prevents merge conflicts during upgrades

### Migration process

The migration process for transitioning from the Kustomize setup in [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) to [deploy-sourcegraph-k8s](https://github.com/sourcegraph/deploy-sourcegraph-k8s) involves the steps shown below. 

The goal of this migration process is to create a new overlay that will generate similiar resources as the current cluster, ensuring a smooth deployment process without disrupting existing resources. 

>NOTE: Both the old custom scripts and Kustomize only create manifests for deployment and don’t change any existing resources in an active cluster.

#### Step 0: Upgrade current instance

Upgrade your current instance to the latest version of Sourcegraph using the old deployment method in [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph).

#### Step 2: Set up a release branch

Set up a release branch from the default branch in your local forked copy of the [deploy-sourcegraph-k8s](https://github.com/sourcegraph/deploy-sourcegraph-k8s) repository.

```bash
$ git checkout -b release
```

#### Step 2: Create a new overlay

`cd` into the repository and create a new directory `instances/my-sourcegraph` for your Sourcegraph instance.

```bash
$ mkdir instance/my-sourcegraph
```

#### Step 3: Set up a configuration file

Copy [instances/kustomization.template.yaml](index.md#template) to the `instances/my-sourcegraph` subdirectory as `kustomization.yaml`.

```bash
$ cp instances/kustomization.template.yaml instances/my-sourcegraph/kustomization.yaml
```

The new `kustomization.yaml` file will be used to configure your deployment.


#### Step 4: Set namespace

Replace `sourcegraph` with a namespace that matches the existing namespace for your current instance. 

Set it to `default` to deploy to the default namespace.

  ```yaml
  # instances/my-sourcegraph/kustomization.yaml
  namespace: sourcegraph
  ```

#### Step 5: Set storage class

Add storage class name that your current instance is using for all associated resources:

1. Include the `storage-class/update-class-name` component under the components list
2. Input the storage class name by setting the value for `STORAGECLASS_NAME` under the configMapGenerator section
   
For example, set `STORAGECLASS_NAME=sourcegraph` if `sourcegraph` is the name of an existing storage class:

  ```yaml
  # instances/my-sourcegraph/kustomization.yaml
  components:
    # Update storageClassName to the STORAGECLASS_NAME value set below
    - ../../components/storage-class/update-class-name
  # ...
  configMapGenerator:
  - name: sourcegraph-kustomize-env
    behavior: merge
    literals:
      - STORAGECLASS_NAME=sourcegraph # Set STORAGECLASS_NAME value to 'sourcegraph'
  ```

#### Step 6: Recreate instance resources


Follow our [configuration guide](configure.md) to recreate your running instance with the provided components.

Please keep in mind that you should not introduce any changes to the characteristics during the migration process. For example, if your cluster is currently running in privileged mode with root users, deploying the instance in non-privileged mode will cause permission issues.

##### Privileged

If your Sourcegraph instance is currently running in privileged mode, you can build the overlay using the `clusters/old-base` component, which generates resources similar to the base cluster in deploy-sourcegraph.

```yaml
components:
  - ../../components/clusters/old-base
```

##### Non-privileged

The default cluster now runs in non-privileged mode.

If your instance was deployed using the non-privileged overlay, you can follow the [configuration guide](configure.md) without adding the `clusters/old-base` component.

Note: pgsql, codeinsights-db, searcher, symbols, and codeintel-db have been changed from Deployments to StatefulSets. However, redeploying these databases as StatefulSets should not affect your existing deployment as they are all configured with the same PVCs.

#### Step 7: Review new manifests

[Compare the manifests](#between-an-overlay-and-a-running-cluster) generated by your new overlay with the ones in your running cluster using the command below:

```bash
$ kubectl diff -f new-cluster.yaml
```

Review the changes to ensure that the manifests generated by your new overlay are similar to the ones currently being used by your active cluster.

#### Step 8: Deploy new manifests

Once you are satisfied with the overlay output, you can now deploy the new overlay using these commands:

```bash
# Build manifests again with overlay
$ kubectl kustomize $PATH_TO_OVERLAY -o cluster.yaml
# Apply manifests to cluster
$ kubectl apply --prune -l deploy=sourcegraph -f cluster.yaml
```

> WARNING: Make sure to test the new overlay and the migration process in a non-production environment before applying it to your production cluster.
