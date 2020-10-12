# Updating a single-image Sourcegraph instance (`sourcegraph/server`)

This document describes the exact changes needed to update a single-node Sourcegraph instance.

A new version of Sourcegraph is released every month (with patch releases in between, released as needed). Check the [Sourcegraph blog](https://about.sourcegraph.com/blog) or the site admin updates page to learn about updates. We actively maintain the two most recent monthly releases of Sourcegraph.

Upgrades should happen across consecutive minor versions of Sourcegraph. For example, if you are running Sourcegraph 3.1 and want to upgrade to 3.3, you should upgrade to 3.2 and then 3.3.

**Always refer to this page before upgrading Sourcegraph,** as it comprehensively describes the steps needed to upgrade, and any manual migration steps you must perform.

## 3.20 -> 3.21.0

If you had LSIF data uploaded prior to upgrading to 3.21.0, there is a background migration that moves all existing LSIF data into the `codeintel-db`. Once this process completes, the `/lsif-storage/dbs` directory on the precise-code-intel-bundle-manager volume should be empty, and the bundle manager should print the following log message:

> Migration to Postgres has completed. All existing LSIF bundles have moved to the path /lsif-storage/db-backups and can be removed from the filesystem to reclaim space.

Once this message has been printed, you are free to delete the bundle files moved into the `/lsif-storage/db-backups` directory on the bundle-manager volume.

> Warning: In order to ensure there is no data loss, **do not upgrade out of the 3.21.x release branch** until you have seen this log message, or verified that the `/lsif-storage/dbs` directory on the precise-code-intel-bundle-manager volume is empty.

### Standard upgrade procedure

To update, just use the newer `sourcegraph/server:N.N.N` Docker image (where `N.N.N` is the version number) in place of the older one, using the same Docker volumes. Your server's data will be migrated automatically if needed.

You can always find the version number of the latest release at [docs.sourcegraph.com](https://docs.sourcegraph.com) in the `docker run` command's image tag.

- As a precaution, before updating, we recommend backing up the contents of the Docker volumes used by Sourcegraph.
- If you need a HA deployment, use the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph).
- There is currently no automated way to downgrade to an older version after you have updated. [Contact support](https://about.sourcegraph.com/contact) for help.
