# Updating a single-image Sourcegraph instance (`sourcegraph/server`)

This document describes the exact changes needed to update a single-node Sourcegraph instance.
**Always refer to this page before upgrading Sourcegraph,** as it comprehensively describes the steps needed to upgrade, and any manual migration steps you must perform.

1. Read our [update policy](index.md#update-policy) to learn about Sourcegraph updates.
2. Find the relevant entry for your update in the update notes on this page.
3. After checking the relevant update notes, refer to the [standard upgrade procedure](../install/docker/operations.md#upgrade) to upgrade your instance.

## 3.32 -> 3.33

Follow the [standard upgrade procedure](../install/docker/operations.md#upgrade).

## 3.31 -> 3.32

Follow the [standard upgrade procedure](../install/docker/operations.md#upgrade).

## 3.30 -> 3.31

The **built-in** main Postgres (`pgsql`) and codeintel (`codeintel-db`) databases have switched to an alpine-based Docker image. Upon upgrading, Sourcegraph will need to re-index the entire database.

All users that use our bundled (built-in) database instances **must** read through the [3.31 upgrade guide](../migration/3_31.md) _before_ upgrading.

> NOTE: The above does not apply to users that use external databases (e.x: Amazon RDS, Google Cloud SQL, etc.).

## 3.29 -> 3.30.3

> WARNING: **Users on 3.29.x are advised to upgrade directly to 3.30.3**. If you have already upgraded to 3.30.0, 3.30.1, or 3.30.2 please follow [this migration guide](../migration/3_30.md).

To update, just use the newer `sourcegraph/server:N.N.N` Docker image (where `N.N.N` is the version number) in place of the older one, using the same Docker volumes. Your server's data will be migrated automatically if needed.

You can always find the version number of the latest release at [docs.sourcegraph.com](https://docs.sourcegraph.com) in the `docker run` command's image tag.

## 3.28 -> 3.29

Follow the [standard upgrade procedure](../install/docker/operations.md#upgrade).

## 3.27 -> 3.28

Follow the [standard upgrade procedure](../install/docker/operations.md#upgrade).

## 3.26 -> 3.27

> Warning: ⚠️ Sourcegraph 3.27 now requires **Postgres 12+**.

If you are using an external database, [upgrade your database](https://docs.sourcegraph.com/admin/postgres#upgrading-external-postgresql-instances) to Postgres 12 or above prior to upgrading Sourcegraph. If you are using the embedded database, [prepare your data for migration](https://docs.sourcegraph.com/admin/postgres#upgrading-single-node-docker-deployments) prior to upgrading Sourcegraph.

## 3.25 -> 3.26

Follow the [standard upgrade procedure](../install/docker/operations.md#upgrade).

> NOTE: ⚠️ From **3.27** onwards we will only support PostgreSQL versions **starting from 12**.

## 3.24 -> 3.25

- Go `1.15` introduced changes to SSL/TLS connection validation which requires certificates to include a `SAN`. This field was not included in older certificates and clients relied on the `CN` field. You might see an error like `x509: certificate relies on legacy Common Name field`. We recommend that customers using Sourcegraph with an external database and and connecting to it using SSL/TLS check whether the certificate is up to date.
  - AWS RDS customers please reference [AWS' documentation on updating the SSL/TLS certificate](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.SSL-certificate-rotation.html) for steps to rotate your certificate.

## 3.20 -> 3.21

- A [bug](https://github.com/sourcegraph/customer/issues/144) exists in the version requiring upgrade from a patch release. **When upgrading please upgrade from v3.20.0 -> v3.20.1 -> v3.21.x**

This release introduces a second database instance, `codeintel-db`. If you have configured Sourcegraph with an external database, then update the `CODEINTEL_PG*` environment variables to point to a new external database as described in the [external database documentation](../external_services/postgres.md). Again, these must not point to the same database or the Sourcegraph instance will refuse to start.

### If you wish to keep existing LSIF data

> WARNING: **Do not upgrade out of the 3.21.x release branch** until you have seen the log message indicating the completion of the LSIF data migration, or verified that the `/lsif-storage/dbs` directory on the precise-code-intel-bundle-manager volume is empty. Otherwise, you risk data loss for precise code intelligence.

If you had LSIF data uploaded prior to upgrading to 3.21.0, there is a background migration that moves all existing LSIF data into the `codeintel-db` upon upgrade. Once this process completes, the `/lsif-storage/dbs` directory on the precise-code-intel-bundle-manager volume should be empty, and the bundle manager should print the following log message:

> Migration to Postgres has completed. All existing LSIF bundles have moved to the path /lsif-storage/db-backups and can be removed from the filesystem to reclaim space.

**Wait for the above message to be printed in `docker logs precise-code-intel-bundle-manager` before upgrading to the next Sourcegraph version**.

### Turn off database secrets encryption

> WARNING: Please delete the secret key file `/var/lib/sourcegraph/token` inside the container before attempting to upgrade to 3.21.x.

In Sourcegraph version 3.20, we would automatically generate a secret key file (`/var/lib/sourcegraph/token`) inside the container for encrypting secrets stored in the database. However, it is not yet ready for general use and format of the secret key file might change. Therefore, it is best to delete the secret key file (`/var/lib/sourcegraph/token`) and turn off the database secrets encryption.

## 3.16 -> 3.17

- There was [a bug](https://github.com/sourcegraph/sourcegraph/issues/11618) in release that caused the version displayed on the `site-admin/update` page to be `0.0.0+dev` instead of `3.17.0`. This issue [was fixed](https://github.com/sourcegraph/sourcegraph/pull/11633) in the `3.17.2` release. We recommend that you avoid this issue by upgrading past `3.17.0` to `3.17.2` using the [Standard upgrade procedure](../install/docker/operations.md#upgrade) listed below.
