# Automatic multi-version upgrades

From **Sourcegraph 5.1 and later**, multi-version upgrades can be performed **automatically** as if they were a standard upgrade for the same deployment type. Automatic multi-version upgrades take the following general form:

1. Determine if your instance is ready to Upgrade (check upgrade notes)
2. Merge the latest Sourcegraph release into your deployment manifests
3. Enable automatic upgrade for this upgrade by either:
	a. In the Updates section in site-admin (available on instances of Sourcegraph 5.1 and later), or
	b. set the `SRC_AUTOUPGRADE` environment variable to `true` on the migrator and frontend 5.1 or later deployment manifests (if the instance is of a version earlier than Sourcegraph 5.1)
4. With upstream changes to your manifests merged, start the new instance

The upgrade magic now happens when the new version is booted. In more detail, starting a new `frontend` container will:

1. Detect that a previous version of Sourcegraph was/is currently running
2. Plan the and persist the upgrade (which migrations to apply) based on the previously running version, the new target version, and the database state
3. Start a new internal server that sends poison pills to disconnect old services from the databases and prevents new services from connecting before the upgrade completes
4. Start a status server in place of the primary exposed port
5. Runs the migration plan (performing the same steps as `migration upgrade ...` )
6. Shuts down the internal and status servers and continues to boot normally

Note that if you have unresolved schema drift, the upgrade will refuse to continue to avoid future migration failures which are more difficult to resolve. Drift should be resolved prior to the beginning of an upgrade. Drift can also be explicitly ignored (which might be required when an upgrade is *resumed* after a failure) by setting the `SRC_AUTOUPGRADE_IGNORE_DRIFT` envvar to true on the migrator and frontend containers.

## Viewing progress

During an automatic multi-version upgrade, we'll attempt to boot a status server in the frontend container that is running (or blocking on) an active upgrade attempt. If there is an upgrade failure that affects the frontend, this status page will not be available and the `frontend` container logs should be viewed. Optimistically, the status server will also be unreachable in the case that an upgrade performs quickly enough that there's no time for the status server to start.

In the case that there's a migration failure, or an unfinished out-of-band migration that needs to be complete, the status server will be served instead of the normal Sourcegraph React app. The following screenshots show an upgrade from Sourcegraph v3.37.1 to Sourcegraph 5.0, in which the `frontend` schema is applying (or waiting to apply) a set of schema migrations, the `codeintel` schema has a pair of schema migration failures, and a single unfinished out-of-band migration is still actively being performed to completion.

![An example in-progress upgrade with schema migrations queued for application](https://storage.googleapis.com/sourcegraph-assets/docs/images/upgrades/5.1/queued.png)
![An example in-progress upgrade with a few schema migration failures](https://storage.googleapis.com/sourcegraph-assets/docs/images/upgrades/5.1/failed.png)
![An example in-progress upgrade with unfinished out-of-band migrations](https://storage.googleapis.com/sourcegraph-assets/docs/images/upgrades/5.1/oobmigrations.png)
