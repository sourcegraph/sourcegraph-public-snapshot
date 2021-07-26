# Updating Sourcegraph

## Update policy

A new version of Sourcegraph is released every month (with patch releases in between, released as needed). Check the [Sourcegraph blog](https://about.sourcegraph.com/blog) or the site admin updates page to learn about updates. We actively maintain the two most recent monthly releases of Sourcegraph.

⚠️ **Regardless of your deployment type:** ⚠️

- **Upgrade one minor version at a time**, e.g. v3.26 --> v3.27 --> v3.28.
  - Patches (e.g. vX.X.4 vs. vX.X.5) do not have to be adopted when moving between vX.X versions.
- **Check the [update notes for your deployment type](#update-notes) for any required manual actions** before updating.
- Check your [out of band migration status](../migration/index.md) prior to upgrade to avoid a necessary rollback while the migration finishes.

## Update notes

Please see the instructions for your deployment type:

- [Single-container `sourcegraph/server` deployments](server.md)
- [Docker Compose single-machine deployments](docker_compose.md)
- [Kubernetes cluster deployments](kubernetes.md)
- [pure-Docker custom deployments](pure_docker.md)

For product update notes, please refer to the [changelog](../../CHANGELOG.md).

## Migrating to a new deployment type

See [this page](../install/index.md) to get advice on which deployment type you should be running.

- [Migrate to Docker Compose](../install/docker-compose/migrate.md) for improved stability and performance if you are using a single-container `sourcegraph/server` deployment.
- [Migrate to a Kubernetes cluster](../install/kubernetes/index.md) if you exceed the limits of a single machine Docker Compose deployment.
