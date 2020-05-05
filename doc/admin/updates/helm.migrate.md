# Migrating from the legacy Sourcegraph Helm chart (2.10.x and prior)

Two things have changed in 2.11.x that require migration:

- Gitserver is now configured using [StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/).
- We have [a new deployment strategy](#why-is-there-a-new-deployment-strategy).

## Migrating

These steps will uninstall Sourcegraph from your cluster while preserving your data. Then you will be able to deploy Sourcegraph using the new process. If you would like help with this process, please reach out to support@sourcegraph.com.

**Please read through all instructions first before starting the migration so you know what is involved**

1. Make a backup of the yaml deployed to your cluster.

   ```bash
   kubectl get all --export -o yaml > backup.yaml
   ```

2. Set the reclaim policy for your existing deployments to `retained`.

   ```bash
   kubectl get pv -o json | jq --raw-output  ".items | map(select(.spec.claimRef.name)) | .[] | \"kubectl patch pv -p '{\\\"spec\\\":{\\\"persistentVolumeReclaimPolicy\\\":\\\"Retain\\\"}}' \\(.metadata.name)\"" | bash
   ```

3. (**Downtime starts here**) Delete the `sourcegraph` release from your cluster.

   ```bash
   helm del --purge sourcegraph
   ```

4. Remove `tiller` from your cluster

   ```bash
   helm reset
   ```

5. Update the old persistent volumes so they can be reused by the new deployment

   ```bash
   # mark all persistent volumes as claimable by the new deployments

   kubectl get pv -o json | jq --raw-output ".items | map(select(.spec.claimRef.name)) | .[] | \"kubectl patch pv -p '{\\\"spec\\\":{\\\"claimRef\\\":{\\\"uid\\\":null}}}' \\(.metadata.name)\"" | bash

   # rename the `gitserver` persistent volumes so that the new `gitserver` stateful set can re-use it

   kubectl get pv -o json | jq --raw-output ".items | map(select(.spec.claimRef.name | contains(\"gitserver-\"))) | .[] | \"kubectl patch pv -p '{\\\"spec\\\":{\\\"claimRef\\\":{\\\"name\\\":\\\"repos-gitserver-\\(.spec.claimRef.name | ltrimstr(\"gitserver-\") | tonumber - 1)\\\"}}}' \\(.metadata.name)\""  | bash
   ```

6. Proceed with the normal [installation steps](../install/kubernetes/index.md). 🚨 When following the instructions for [configuring a storage class](../install/kubernetes/configure.md#configure-a-storage-class), you need to make sure that the newly configured storage class has the same configuration as the one that you were using in the legacy helm deployment. Steps:

   1. When creating the new storage class, use the same `cluster.storageClass.name` and `cluster.storageClass.zone` fields that were in your old [values.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/helm-legacy/values.yaml).

   1. Use the convenience script in ["Using a storage class with an alternate name"](../install/kubernetes/configure.md#using-a-storage-class-with-an-alternate-name) to update all the `storageClassName` references in the PVCs to refer to the old `cluster.storageClass.name` field.

7. The previous step produces a fresh base state, so you will need to reconfigure your cluster by following the relevant steps in [configure.md](../install/kubernetes/configure.md) (e.g. exposing ports, applying your site config, enabling other services like language servers, Prometheus, Alertmanager, Jaeger, etc.).

   **Downtime ends once installation and configuration is complete**

## Why is there a new deployment strategy?

2.10.x and prior was deployed by configuring `values.yaml` and using `helm` to generate the final yaml to deploy to a cluster.

There were a few downsides with this approach:

- `values.yaml` was a custom configuration format defined by us which implicitly made configuring certain Kubernetes settings special cases. We didn't want this to grow over time into an unmaintainable/unusable mess.
- If customers wanted to configure things not supported in `values.yaml`, then we would either need to add support or the customer would need to make further modifications to the generated yaml.
- Writing Go templates inside of yaml was error prone and hard to maintain. It was too easy to make a silly mistake and generate invalid yaml. Our editors could not help us because Go template logic made the yaml templates not valid yaml.
- It required using `helm` to generate templates even though some customers don't care to use `helm` to deploy the yaml.

Our new approach is simpler and more flexible.

- We have removed our dependency on `helm`. It is no longer needed to generate templates, and we no longer recommend it as the easiest way to deploy our yaml to a cluster. You are still free to use `helm` to deploy to your cluster if you wish.
- Our base config is pure yaml which can be deployed directly to a cluster. It is easier for you to use, and also easier for us to maintain.
- You can configure our base yaml using whatever process best for you (Git ops, [Kustomize](https://github.com/kubernetes-sigs/kustomize), custom scripts, etc.). We provide [documentation and recipies for common customizations](../install/kubernetes/configure.md).
