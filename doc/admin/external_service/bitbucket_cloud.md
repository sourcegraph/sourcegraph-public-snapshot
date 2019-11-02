# Bitbucket Cloud

Site admins can sync Git repositories hosted on [Bitbucket Cloud](https://bitbucket.org) with Sourcegraph so that users can search and navigate the repositories.

To set this up, add Bitbucket Cloud as an external service to Sourcegraph:

1. Go to **User menu > Site admin**.
1. Open the **External services** page.
1. Press **+ Add external service**.
1. In the list, select **Bitbucket.org repositories**.
1. Enter a **Display name** (using "Bitbucket Cloud" is OK).
1. Configure the connection to Bitbucket Cloud in the JSON editor. Use Cmd/Ctrl+Space for completion, and [see configuration documentation below](#configuration).
1. Press **Add external service**.

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
