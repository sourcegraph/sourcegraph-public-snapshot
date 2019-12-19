# Bitbucket Server

Site admins can sync Git repositories hosted on [Bitbucket Server](https://www.atlassian.com/software/bitbucket/server) (and the [Bitbucket Data Center](https://www.atlassian.com/enterprise/data-center/bitbucket) deployment option) with Sourcegraph so that users can search and navigate the repositories.

To set this up, add Bitbucket Server as an external service to Sourcegraph:

1. Go to **User menu > Site admin**.
1. Open the **External services** page.
1. Press **+ Add external service**.
1. Enter a **Display name** (using "Bitbucket Server" is OK if you only have one Bitbucket Server instance).
1. In the **Kind** menu, select **Bitbucket Server**.
1. Configure the connection to Bitbucket Server in the JSON editor. Use Cmd/Ctrl+Space for completion, and [see configuration documentation below](#configuration).
1. Press **Add external service**.

## Repository syncing

There are four fields for configuring which repositories are mirrored:

- [`repos`](bitbucket_server.md#configuration)<br>A list of repositories in `projectKey/repositorySlug` format.
- [`repositoryQuery`](bitbucket_server.md#configuration)<br>A list of strings with some pre-defined options (`none`, `all`), and/or a [Bitbucket Server Repo Search Request Query Parameters](https://docs.atlassian.com/bitbucket-server/rest/6.1.2/bitbucket-rest.html#idp355).
- [`exclude`](bitbucket_server.md#configuration)<br>A list of repositories to exclude which takes precedence over the `repos`, and `repositoryQuery` fields.
- ['excludePersonalRepositories'](bitbucket_server.md#configuration)<br>With this enabled, Sourcegraph will exclude any personal repositories from being imported, even if it has access to them.

## Repository permissions

By default, all Sourcegraph users can view all repositories. To configure Sourcegraph to use
Bitbucket Server's repository permissions, see [Repository permissions](../repo/permissions.md#bitbucket_server).

### Authentication for older Bitbucket Server versions

Bitbucket Server versions older than v5.5 require specifying a less secure username and password combination, as those versions of Bitbucket Server do not support [personal access tokens](https://confluence.atlassian.com/bitbucketserver/personal-access-tokens-939515499.html).

### HTTPS cloning

Sourcegraph by default clones repositories from your Bitbucket Server via HTTP(S), using the access token or account credentials you provide in the configuration. The [`username`](bitbucket_server.md#configuration) field is always used when cloning, so it is required.

## Repository labels

Sourcegraph will mark repositories as archived if they have the `archived` label on Bitbucket Server. You can exclude these repositories in search with `archived:no` [search syntax](../../user/search/queries.md).

## Configuration

Bitbucket Server external service connections support the following configuration options, which are specified in the JSON editor in the site admin external services area.

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/bitbucket_server.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/bitbucket_server) to see rendered content.</div>

## Sourcegraph native code intelligence plugin

Learn more about the [Sourcegraph Bitbucket Server plugin](../../integration/bitbucket_server.md#sourcegraph-native-code-intelligence-plugin) for enabling native code intelligence for every Bitbucket user when browsing code and reviewing pull requests.
