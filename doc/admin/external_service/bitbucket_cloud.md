# Bitbucket Cloud

Site admins can sync Git repositories hosted on [Bitbucket Cloud](https://bitbucket.org) with Sourcegraph so that users can search and navigate the repositories.

To connect Bitbucket Cloud to Sourcegraph:

1. Go to **Site admin > Manage code hosts > Add repositories**.
2. Select **Bitbucket.org**.
3. Configure the connection to Bitbucket Cloud using the action buttons above the text field. Additional fields can be added using <kbd>Cmd/Ctrl+Space</kbd> for auto-completion. See the [configuration documentation below](#configuration).
4. Press **Add repositories**.

## Repository syncing

Currently, all repositories belonging to the user configured will be synced.

In addition, there is one more field for configuring which repositories are mirrored:

- [`teams`](bitbucket_cloud.md#configuration)<br>A list of teams that the configured user has access to whose repositories should be synced.
- [`exclude`](bitbucket_cloud.md#configuration)<br>A list of repositories to exclude, which takes precedence over the `teams` field.

### HTTPS cloning

Sourcegraph clones repositories from your Bitbucket Cloud via HTTP(S), using the [`username`](bitbucket_cloud.md#configuration) and [`appPassword`](bitbucket_cloud.md#configuration) required fields you provide in the configuration.

## Internal rate limits

Internal rate limiting can be configured to limit the rate at which requests are made from Sourcegraph to Bitbucket Cloud. 

If enabled, the default rate is set at 7200 per hour (2 per second), which can be configured via the `requestsPerHour` field (see below):

- For Sourcegraph <=3.38, if rate limiting is configured more than once for the same code host instance, the most restrictive limit will be used.
- For Sourcegraph >=3.39, rate limiting should be enabled and configured for each individual code host connection.

**NOTE** Internal rate limiting is only currently applied when synchronizing changesets in [batch changes](../../batch_changes/index.md), repository permissions, and repository metadata from code hosts.

## User authentication

To configure Bitbucket Cloud as an authentication provider (which will enable sign-in via Bitbucket Cloud), see the
[authentication documentation](../auth/index.md#bitbucket-cloud).

## Configuration

Bitbucket Cloud connections support the following configuration options, which are specified in the JSON editor in the site admin "Manage code hosts" area.

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/bitbucket_cloud.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/bitbucket_cloud) to see rendered content.</div>

## Webhooks

Using the `webhooks` property on the external service has been deprecated.

Please consult [this page](../config/webhooks.md) in order to configure webhooks.
