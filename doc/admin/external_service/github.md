# GitHub

Site admins can sync Git repositories hosted on [GitHub.com](https://github.com) and [GitHub Enterprise](https://enterprise.github.com) with Sourcegraph so that users can search and navigate the repositories.

To set this up, add GitHub as an external service to Sourcegraph:

1. Go to **User menu > Site admin**.
1. Open the **External services** page.
1. Press **+ Add external service**.
1. Enter a **Display name** (using "GitHub" is OK if you only have one GitHub instance).
1. In the **Kind** menu, select **GitHub**.
1. Configure the connection to GitHub in the JSON editor. Use Cmd/Ctrl+Space for completion, and [see configuration documentation below](#configuration).
1. Press **Add external service**.

## Supported versions

- GitHub.com
- GitHub Enterprise v2.10 and newer

## Repository syncing

By default, Sourcegraph syncs `none` of the repositories that are affiliated with the user whose token you provide, as defined by default [`repositoryQuery`](github.md#configuration) setting.

You can explicitly configure which repositories to mirror via the [`repositoryQuery`](github.md#configuration) setting.

If you want to synchronize public repositories from GitHub.com, the repositories need to be explicitly enumerated in the [`repos`](github.md#configuration) field of the GitHub external service configuration.

If you don't want to use an access token from your personal GitHub user account, generate a token for a [machine user](https://developer.github.com/v3/guides/managing-deploy-keys/#machine-users) affiliated with the organizations whose repositories you wish to make available.

### GitHub.com rate limits

You should always include a token in a configuration for a GitHub.com URL to avoid being denied service by GitHub's [unauthenticated rate limits](https://developer.github.com/v3/#rate-limiting). If you don't want to automatically synchronize repositories from the account associated with your personal access token, you can create a token without a [`repo` scope](https://developer.github.com/apps/building-oauth-apps/scopes-for-oauth-apps/#available-scopes) for the purposes of bypassing rate limit restrictions only.

## Repository permissions

By default, all Sourcegraph users can view all repositories. To configure Sourcegraph to use
GitHub's per-user repository permissions, see "[Repository permissions](../repo/permissions.md#github)".

## User authentication

To configure GitHub as an authentication provider (which will enable sign-in via GitHub), see the
[authentication documentation](../auth.md#github).

## Configuration

GitHub external service connections support the following configuration options, which are specified in the JSON editor in the site admin external services area.

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/github.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/github) to see rendered content.</div>
