<style>
.estimator label {
    display: flex;
}

.estimator .radioInput label {
    display: inline-flex;
    align-items: center;
    margin-left: .5rem;
}

.estimator .radioInput label span {
    margin-left: .25rem;
    margin-right: .25rem;
}

.estimator input[type=range] {
    width: 15rem;
}

.estimator .post-label {
    font-size: 16px;
    margin-left: 0.5rem;
}

.estimator .copy-as-markdown {
    width: 100%;
    height: 8rem;
}

.estimator a[title]:hover:after {
  content: attr(title);
  background: red;
  position: relative;
  z-index: 1000;
  top: 16px;
  left: 0;
}

</style>

<script src="https://storage.googleapis.com/sourcegraph-resource-estimator/go_1_18_wasm_exec.js"></script>
<script src="https://storage.googleapis.com/sourcegraph-resource-estimator/launch_script.js?v2" version="dbc14f6"></script>

# Sourcegraph resource estimator

Updating the form below will recalculate an estimate for the resources you can use to configure your Sourcegraph deployment.

The output is estimated based on existing data we collected from current running deployments.

<form id="root"></form>

## Additional information

#### How to apply these changes to your deployment?

- For docker-compose deployments, edit your [docker-compose.yml file](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml) and set cpus and mem_limit to the limits shown above.
- For Helm deployments, create an [override file](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/examples/common-modifications/override.yaml) (or update your existing override file) with the new values shown above.
- For non-Helm Kubernetes deployments, we recommend using Kustomize to generate manifests with the values shown above. Please refer to our [Kustomize docs](https://docs.sourcegraph.com/admin/deploy/kubernetes/kustomize#kustomize) on how to use Kustomize.

#### What is the default deployment size?

- Our default deployment should support ~1000 users and ~1000 repositories with one monorepo that is less than 5GB.

#### What is engagement rate?

- Engagement rate refers to the percentage of users who use Sourcegraph regularly. It is generally used for existing deployments to estimate resources.

#### What is the recommended deployment type?

- We recommend Kubernetes for any deployments requiring > 1 service replica, but docker-compose does support service replicas and can scale up with multiple replicas as long as you can provision a sufficiently large single machine.

#### If you plan to enforce repository permissions on Sourcegraph

- Repository permissions on Sourcegraph can have a noticeable impact on search performance if you have a large number of users and/or repositories on your code host. We suggest setting your authorization ttl values as high as you are comfortable setting it in order to reduce the chance of this (e.g. to 72h) in the repository permission configuration.

#### What kind of data can be regenerated without backup?

- See our docs on [Persistent data backup in Kubernetes](https://docs.sourcegraph.com/admin/deploy/migrate-backup#persistent-data-backup-in-kubernetes) for more detail.

#### How does Sourcegraph scale?

- [Click here to learn more about how Sourcegraph scales.](https://docs.sourcegraph.com/admin/install/kubernetes/scale)
