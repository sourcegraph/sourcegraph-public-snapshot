# Bitbucket Server

Site admins can sync Git repositories hosted on [Bitbucket Server](https://www.atlassian.com/software/bitbucket/server) (and the [Bitbucket Data Center](https://www.atlassian.com/enterprise/data-center/bitbucket) deployment option) with Sourcegraph so that users can search and navigate the repositories.

To connect Bitbucket Server to Sourcegraph:

1. Go to **Site admin > Manage repositories > Add repositories**
1. Select **Bitbucket Server**.
1. Configure the connection to Bitbucket Server using the action buttons above the text field, and additional fields can be added using <kbd>Cmd/Ctrl+Space</kbd> for auto-completion. See the [configuration documentation below](#configuration).
1. Press **Add repositories**.

Also consider installing the [Sourcegraph Bitbucket Server plugin](../../integration/bitbucket_server.md#sourcegraph-bitbucket-server-plugin) which enables native code intelligence for every Bitbucket user when browsing code and reviewing pull requests, allows for faster permission syncing between Sourcegraph and Bitbucket Server and adds support for webhooks to Bitbucket Server.

## Repository syncing

There are four fields for configuring which repositories are mirrored:

- [`repos`](bitbucket_server.md#configuration)<br>A list of repositories in `projectKey/repositorySlug` format.
- [`repositoryQuery`](bitbucket_server.md#configuration)<br>A list of strings with some pre-defined options (`none`, `all`), and/or a [Bitbucket Server Repo Search Request Query Parameters](https://docs.atlassian.com/bitbucket-server/rest/6.1.2/bitbucket-rest.html#idp355).
- [`exclude`](bitbucket_server.md#configuration)<br>A list of repositories to exclude which takes precedence over the `repos`, and `repositoryQuery` fields.
- ['excludePersonalRepositories'](bitbucket_server.md#configuration)<br>With this enabled, Sourcegraph will exclude any personal repositories from being imported, even if it has access to them.

## Webhooks

The [Sourcegraph Bitbucket Server plugin](../../integration/bitbucket_server.md#sourcegraph-bitbucket-server-plugin) enables the Bitbucket Server instance to send webhooks to Sourcegraph.

Using webhooks is highly recommended when using [Campaigns](../../user/campaigns.md), since they speed up the syncing of pull request data between Bitbucket Server and Sourcegraph and make it more efficient.

To set up webhooks:

1. Connect Bitbucket Server to Sourcegraph (_see instructions above_).
1. Install the [Sourcegraph Bitbucket Server plugin](../../integration/bitbucket_server.md#sourcegraph-bitbucket-server-plugin) on your Bitbucket Server instance.
1. In Sourcegraph, go to **Site admin > Manage repositories** and edit the Bitbucket Server configuration.
1. Add the `"plugin.webhooks"` property (you can generate a secret with `openssl rand -hex 32`):<br /> `"plugin.webhooks": [ {"secret": "verylongrandomsecret"} ]`
1. Click **Update repositories**.
1. Sourcegraph now automatically creates a webhook on your Bitbucket Server instance with the name `sourcegraph-campaigns` and the `pr`.
1. On your Bitbucket Server instance, go to **Administration > Add-ons > Sourcegraph** and make sure that the new `sourcegraph-campaigns` webhook is listed under **All webhooks** with a timestamp in the **Last successful** column.

Done! Sourcegraph will now receive webhook events from Bitbucket Server and use them to sync pull request events, used by [Campaigns](../../user/campaigns.md), fast and more efficiently.

## Repository permissions

By default, all Sourcegraph users can view all repositories. To configure Sourcegraph to use Bitbucket Server's repository permissions, see [Repository permissions](../repo/permissions.md#bitbucket_server).

### Experimental: faster ACL permissions fetching

With the [Sourcegraph Bitbucket Server plugin](../../integration/bitbucket_server.md#sourcegraph-bitbucket-server-plugin) you can enable the **experimental** and faster permissions fetching:

1. Connect Bitbucket Server to Sourcegraph (_see instructions above_).
1. Install the [Sourcegraph Bitbucket Server plugin](../../integration/bitbucket_server.md#sourcegraph-bitbucket-server-plugin) on your Bitbucket Server instance.
1. In Sourcegraph, go to **Site admin > Manage repositories** and edit the Bitbucket Server configuration.
1. Add the `"plugin.permissions"` property:<br /> `"plugin.permissions": "enabled"`

### Authentication for older Bitbucket Server versions

Bitbucket Server versions older than v5.5 require specifying a less secure username and password combination, as those versions of Bitbucket Server do not support [personal access tokens](https://confluence.atlassian.com/bitbucketserver/personal-access-tokens-939515499.html).

### HTTPS cloning

Sourcegraph by default clones repositories from your Bitbucket Server via HTTP(S), using the access token or account credentials you provide in the configuration. The [`username`](bitbucket_server.md#configuration) field is always used when cloning, so it is required.

## Repository labels

Sourcegraph will mark repositories as archived if they have the `archived` label on Bitbucket Server. You can exclude these repositories in search with `archived:no` [search syntax](../../user/search/queries.md).

## Configuration

Bitbucket Server connections support the following configuration options, which are specified in the JSON editor in the site admin "Manage repositories" area.

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/bitbucket_server.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/bitbucket_server) to see rendered content.</div>
