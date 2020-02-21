# Bitbucket Cloud

Site admins can sync Git repositories hosted on [Bitbucket Cloud](https://bitbucket.org) with Sourcegraph so that users can search and navigate the repositories.

To connect Bitbucket Cloud to Sourcegraph:

1. Go to **Site admin > Manage repositories > Add repositories**
1. Select **Bitbucket.org**.
1. Configure the connection to Bitbucket Cloud using the action buttons above the text field, and additional fields can be added using <kbd>Cmd/Ctrl+Space</kbd> for auto-completion. See the [configuration documentation below](#configuration).
1. Press **Add repositories**.

## Repository syncing

Currently, all repositories belonging the user configured will be synced.

In addition, there is one more field for configuring which repositories are mirrored:

- [`teams`](bitbucket_cloud.md#configuration)<br>A list of teams that the configured user has access to whose repositories should be synced.
- [`exclude`](bitbucket_cloud.md#configuration)<br>A list of repositories to exclude which takes precedence over the `teams` field.

### HTTPS cloning

Sourcegraph clones repositories from your Bitbucket Cloud via HTTP(S), using the [`username`](bitbucket_cloud.md#configuration) and [`appPassword`](bitbucket_cloud.md#configuration) required fields you provide in the configuration.

## Configuration

Bitbucket Cloud external service connections support the following configuration options, which are specified in the JSON editor in the site admin external services area.

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/bitbucket_cloud.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/bitbucket_cloud) to see rendered content.</div>
