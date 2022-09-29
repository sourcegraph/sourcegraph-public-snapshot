# Updating a Kubernetes Sourcegraph instance

**Always refer to this page before upgrading Sourcegraph,** as it comprehensively describes any special manual migration steps you must perform per-version.

## Upgrade procedure

1. Read our [update policy](index.md#update-policy) to learn about Sourcegraph updates.
1. Find the relevant entry for your update in the update notes on this page.
1. After checking the relevant update notes, refer to either of the following guides to upgrade your instance:
    * [Kubernetes with Helm upgrade guide](../deploy/kubernetes/helm.md#standard-upgrades)
    * [Kubernetes without Helm upgrade guide](../deploy/kubernetes/update.md#standard-upgrades)

## Multi-version upgrade procedure

1. Read our [update policy](index.md#update-policy) to learn about Sourcegraph updates.
1. Find the relevant entry for your update in the update notes on this page. These notes may contain relevant information about the infrastructure update such as resource requirement changes or versions of depencies (Docker, Kubernetes, externalized databases).
1. After checking the relevant update notes, refer to either of the following guides to upgrade your instance:
    * [Kubernetes with Helm upgrade guide](../deploy/kubernetes/helm.md#multi-version-upgrades)
    * [Kubernetes without Helm upgrade guide](../deploy/kubernetes/update.md#multi-version-upgrades)

<!-- GENERATE UPGRADE GUIDE ON RELEASE (release tooling uses this to add entries) -->

## Unreleased

<!-- Add changes changes to this section before release. -->

TODO - replace me

## 3.43 -> 4.0.1

* `jaeger-agent` sidecars have been removed in favor of an  [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) DaemonSet + Deployment configuration. See [Configure a tracing backend section.](#configure-a-tracing-backend)
* Exporting traces to an external observability backend is now available. Read the [documentation](../deploy/kubernetes/configure.md#configure-a-tracing-backend) to configure.
* The bundled Jaeger instance is now disabled by default. It can be [enabled](../deploy/kubernetes/configure.md#enable-the-bundled-jaeger-deployment) if you do not wish to utilise your own external tracing backend.

Follow the [steps](#upgrade-procedure) outlined at the top of this page to upgrade.

## 3.42 -> 3.43.2

Follow the [standard upgrade procedure](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.41 -> 3.42.2

Follow the [standard upgrade procedure](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.40 -> 3.41

- The Postgres DBs `frontend` and `codeintel-db` are now given 1 hour to begin accepting connections before Kubernetes restarts the containers. [#4136](https://github.com/sourcegraph/deploy-sourcegraph/pull/4136)

Follow the [standard upgrade procedure](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.39 -> 3.40.2

- `cadvisor` now defaults to run in `privileged` mode. This allows `cadvisor` to collect out of memory events happening to containers which can be used to discover underprovisoned resources. This is disabled by default in `non-privileged` overlay. [#4126](https://github.com/sourcegraph/deploy-sourcegraph/pull/4126)
- Updated the Nginx ingress controller to v1.2.0. Previously this image originated from quay.io, now it is pulled from the official k8s repository. A redeployment of the ingress
 controller may be necessary if your deployment used the manifests provided in `configure/ingress-nginx`. [#4128](https://github.com/sourcegraph/deploy-sourcegraph/pull/4128)
- The alpine-3.12 docker images used as init containers for some deployments have been replaced with images based on alpine-3.14. [#4129](https://github.com/sourcegraph/deploy-sourcegraph/pull/4129)

Follow the [standard upgrade procedure](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.38 -> 3.39

The`codeinsights-db` container no longer uses TimescaleDB and is now based on the standard Postgres image [sourcegraph/deploy-sourcegraph#4103](https://github.com/sourcegraph/deploy-sourcegraph/pull/4103). Metrics scraping is also enabled.

**CAUTION** If you use a custom Code Insights postgres config, you must update the `shared_preload_libraries` list to remove timescaledb. The [above PR](https://github.com/sourcegraph/deploy-sourcegraph/pull/4103/files#diff-e5f8d6e46f8c9335c489c0d8e9ae9be4f4655f878f3ac569c73ebb3865b0eeeeL695-R688) demonstrates this change.

To upgrade your deployment follow either:
  * [Kubernetes with Helm upgrade guide](../deploy/kubernetes/helm.md#upgrading-sourcegraph)
  * [Kubernetes without Helm upgrade guide](../deploy/kubernetes/update.md) to upgrade your instance.

## 3.37 -> 3.38

Follow the [standard upgrade procedure](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.36 -> 3.37

This release adds a new `migrator` initContainer to the frontend deployment to run database migrations. Confirm the environment variables on this new container match your database settings. [Docs](https://docs.sourcegraph.com/admin/deploy/kubernetes/update#database-migrations)

Follow the [standard upgrade procedure](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.35 -> 3.36

The `backend` service has been removed, so if you deploy with a method other than `kubectl-apply-all.sh`, a manual removal of the service may be necessary.

Follow the [standard upgrade procedure](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.34 -> 3.35.1

**Due to issues related to Code Insights on the 3.35.0 release, Users are advised to upgrade directly to 3.35.1.**

The query-runner deployment has been removed, so if you deploy with a method other than the `kubectl-apply-all.sh`, a manual removal of the deployment may be necessary.
Follow the [standard upgrade procedure](../deploy/kubernetes/update.md) to upgrade your deployment.

There is a [known issue](../../code_insights/how-tos/Troubleshooting.md#oob-migration-has-made-progress-but-is-stuck-before-reaching-100) with the Code Insights out-of-band settings migration not reaching 100% complete when encountering deleted users or organizations.

## 3.33 -> 3.34

No manual migration is required - follow the [standard upgrade procedure](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.32 -> 3.33

No manual migration is required - follow the [standard upgrade procedure](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.31 -> 3.32

No manual migration is required - follow the [standard upgrade procedure](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.30.3 -> 3.31

The **built-in** main Postgres (`pgsql`) and codeintel (`codeintel-db`) databases have switched to an alpine-based Docker image. Upon upgrading, Sourcegraph will need to re-index the entire database.

If you have already upgraded to 3.30.3, which uses the new alpine-based Docker images, all users that use our bundled (built-in) database instances should have already performed [the necessary re-indexing](../migration/3_31.md).

> NOTE: The above does not apply to users that use external databases (e.x: Amazon RDS, Google Cloud SQL, etc.).

## 3.30.x -> 3.31

The **built-in** main Postgres (`pgsql`) and codeintel (`codeintel-db`) databases have switched to an alpine-based Docker image. Upon upgrading, Sourcegraph will need to re-index the entire database.

All users that use our bundled (built-in) database instances **must** read through the [3.31 upgrade guide](../migration/3_31.md) _before_ upgrading.

> NOTE: The above does not apply to users that use external databases (e.x: Amazon RDS, Google Cloud SQL, etc.).

## 3.29 -> 3.30.3

> WARNING: **Users on 3.29.x are advised to upgrade directly to 3.30.3**. If you have already upgraded to 3.30.0, 3.30.1, or 3.30.2 please follow [this migration guide](../migration/3_30.md).

This upgrade removes the `non-root` overlay, in favor of using only the `non-privileged` overlay for deploying Sourcegraph in secure environments. If you were
previously deploying using the `non-root` overlay, you should now generate overlays using the `non-privileged` overlay.

No other manual migration is required, follow the [standard upgrade method](../deploy/kubernetes/update.md) to upgrade your
deployment.

## 3.28 -> 3.29

This upgrade adds a new `worker` service that runs a number of background jobs that were previously run in the `frontend` service. See [notes on deploying workers](../workers.md#deploying-workers) for additional details. Good initial values for CPU and memory resources allocated to this new service should match the `frontend` service.

## 3.27 -> 3.28

- All Sourcegraph images now have a registry prefix. [#2901](https://github.com/sourcegraph/deploy-sourcegraph/pull/2901)
- The memory requirements for `redis-cache` and `redis-store` have been increased by 1GB. See https://github.com/sourcegraph/deploy-sourcegraph/pull/2898 for more context.

## 3.26 -> 3.27

> WARNING: Sourcegraph 3.27 now requires **Postgres 12+**.

If you are using an external
database, [upgrade your database](https://docs.sourcegraph.com/admin/postgres#upgrading-external-postgresql-instances)
to Postgres 12 or above prior to upgrading Sourcegraph. No action is required if you are using the supplied database
images.

> NOTE: The Postgres 12 database migration scales with the size of your database, and the resources provided to the container.
> Expect to have downtime relative to the size of your database. Additionally, you must ensure that have enough storage
> space to accommodate the migration. A rough guide would be 2x the current on-disk database size

<!---->

> WARNING: We have updated the default replicas for `sourcegraph-frontend` and `precise-code-intel-worker` to `2`. If you use a custom value, make sure you do not merge the replica change.

Afterwards, follow the [standard upgrade method](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.25 -> 3.26

No manual migration required, follow the [standard upgrade method](../deploy/kubernetes/update.md) to upgrade your
deployment.

> NOTE: From **3.27** onwards we will only support PostgreSQL versions **starting from 12**.

## 3.24 -> 3.25

- Go `1.15` introduced changes to SSL/TLS connection validation which requires certificates to include a `SAN`. This field was not included in older certificates and clients relied on the `CN` field. You might see an error like `x509: certificate relies on legacy Common Name field`. We recommend that customers using Sourcegraph with an external database and and connecting to it using SSL/TLS check whether the certificate is up to date.
  - AWS RDS customers please reference [AWS' documentation on updating the SSL/TLS certificate](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.SSL-certificate-rotation.html) for steps to rotate your certificate.

## 3.23 -> 3.24

No manual migration required, follow the [standard upgrade method](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.22 -> 3.23

No manual migration is required, follow the [standard upgrade method](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.21 -> 3.22

No manual migration is required, follow the [standard upgrade method](../deploy/kubernetes/update.md) to upgrade your deployment.

This upgrade removes the `code intel bundle manager`. This service has been deprecated and all references to it have been removed.

This upgrade also adds a MinIO container that doesn't require any custom configuration. You can find more detailed documentation in https://docs.sourcegraph.com/admin/external_services/object_storage.

## 3.20 -> 3.21

Follow the [standard upgrade method](../deploy/kubernetes/update.md) to upgrade your deployment.

This release introduces a second database instance, `codeintel-db`. If you have configured Sourcegraph with an external database, then update the `CODEINTEL_PG*` environment variables to point to a new external database as described in the [external database documentation](../external_services/postgres.md). Again, these must not point to the same database or the Sourcegraph instance will refuse to start.

## 3.20

No manual migration is required, follow the [standard upgrade method](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.19

No manual migration is required, follow the [standard upgrade method](../deploy/kubernetes/update.md) to upgrade your deployment.

> WARNING: If you use an overlay that does not reference one of the provided overlays, please add `- ../bases/pvcs` as an additional base
to your `kustomization.yaml` file. Otherwise the PVCs could be pruned if `kubectl apply -prune` is used.

## 3.18

No manual migration is required, follow the [standard upgrade method](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.17

No manual migration is required, follow the [standard upgrade method](../deploy/kubernetes/update.md) to upgrade your deployment.

## 3.16

No manual migration is required, follow the [standard upgrade method](../deploy/kubernetes/update.md) to upgrade your deployment.

Note: The following deployments have had their `strategy` changed from `rolling` to `recreate`:

- redis-cache
- redis-store
- pgsql
- precise-code-intel-bundle-manager
- prometheus

This change was made to avoid two pods writing to the same volume and causing corruption. No special action is needed to apply the change.

## 3.15

### Note: Prometheus and Grafana resource requirements increase

Resource _requests and limits_ for Grafana and Prometheus are now equal to the following:

- Grafana 100Mi -> 512Mi
- Prometheus: 500M -> 3G

This change was made to ensure that even if another Sourcegraph service starts consuming more memory than expected and the Kubernetes node has been over-provisioned, that Sourcegraph's monitoring will still have enough memory to run and monitor / send alerts to the site admin. For additional information see [#638](https://github.com/sourcegraph/deploy-sourcegraph/pull/638)

You may run the following commands to remove the now unused resources:

```shell script
kubectl delete svc lsif-server
kubectl delete deployment lsif-server
kubectl delete pvc lsif-server
```

**Migrating**

The lsif-server service has been replaced by a trio of services defined in [precise-code-intel](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/precise-code-intel),
and the persistent volume claim in which lsif-server  stored converted LSIF uploads has been replaced by
[bundle storage](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/precise-code-intel/bundle-storage.PersistentVolume.yaml).

Upgrading to 3.15 will create a new empty volume for LSIF data. Without any action, the LSIF data previously uploaded
to the instance will be lost. To retain old LSIF data, perform the following migration steps. This will cause some
temporary downtime for precise code navigation.

**Migrating**

1. Deploy 3.15. This will create a `bundle-manager` persistent volume claim.
2. Release the claims to old and new persistent volumes by taking down `lsif-server` and `precise-code-intel-bundle-manager`.

```shell script
kubectl delete svc lsif-server
kubectl delete deployment lsif-server
kubectl delete deployment precise-code-intel-bundle-manager
```

3. Deploy the `lsif-server-migrator` deployment to transfer the data from the old volume to the new volume.

```shell script
kubectl apply -f configure/lsif-server-migrator/lsif-server-migrator.Deployment.yaml
```

4. Watch the output of the `lsif-server-migrator` until the copy completes (`'Copy complete!'`).

```shell script
kubectl logs lsif-server-migrator
```

5. Tear down the deployment and re-create the bundle manager deployment.

```shell script
kubectl delete deployment lsif-server-migrator
./kubectl-apply-all.sh
```

6. Remove the old persistent volume claim.

```shell script
kubectl delete pvc lsif-server
```

### Configuration

In Sourcegraph 3.0 all site configuration has been moved out of the `config-file.ConfigMap.yaml` and into the PostgreSQL database. We have an automatic migration if you use version 3.2 or before. Please do not upgrade directly from 2.x to 3.3 or higher.

After running 3.0, you should visit the configuration page (`/site-admin/configuration`) and [the management console](https://docs.sourcegraph.com/admin/management_console) and ensure that your configuration is as expected. In some rare cases, automatic migration may not be able to properly carry over some settings and you may need to reconfigure them.

### `sourcegraph-frontend` service type

The type of the `sourcegraph-frontend` service ([base/frontend/sourcegraph-frontend.Service.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/frontend/sourcegraph-frontend.Service.yaml)) has changed
from `NodePort` to `ClusterIP`. Directly applying this change [will
fail](https://github.com/kubernetes/kubernetes/issues/42282). Instead, you must delete the old
service and then create the new one (this will result in a few seconds of downtime):

```shell
kubectl delete svc sourcegraph-frontend
kubectl apply -f base/frontend/sourcegraph-frontend.Service.yaml
```

### HTTPS / TLS

Sourcegraph 3.0 removed HTTPS / TLS features from Sourcegraph in favor of relying on [Kubernetes Ingress Resources](https://kubernetes.io/docs/concepts/services-networking/ingress/). As a consequence, Sourcegraph 3.0 does not expose TLS as the NodePort 30433. Instead you need to ensure you have setup and configured either an ingress controller (recommended) or an explicit NGINX service. See [ingress controller documentation](../deploy/kubernetes/configure.md#ingress-controller-recommended), [NGINX service documentation](../deploy/kubernetes/configure.md#nginx-service), and [configure TLS/SSL documentation](../deploy/kubernetes/configure.md#configure-tlsssl).

If you previously configured `TLS_KEY` and `TLS_CERT` environment variables, you can remove them from [base/frontend/sourcegraph-frontend.Deployment.yaml](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/frontend/sourcegraph-frontend.Deployment.yaml)

### Postgres 11.1

Sourcegraph 3.0 ships with Postgres 11.1. The upgrade procedure is mostly automatic. Please read [this page](https://docs.sourcegraph.com/admin/postgres) for detailed information.
