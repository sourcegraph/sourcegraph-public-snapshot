# GitHub

Site admins can sync Git repositories hosted on [GitHub.com](https://github.com) and [GitHub Enterprise](https://enterprise.github.com) with Sourcegraph so that users can search and navigate the repositories.

To connect GitHub to Sourcegraph:

1. Go to **Site admin > Manage code hosts**
2. Select **GitHub**.
3. Configure the connection to GitHub using the action buttons above the text field, and additional fields can be added using <kbd>Cmd/Ctrl+Space</kbd> for auto-completion. See the [configuration documentation below](#configuration).
4. Press **Add repositories**.

In this example, the kubernetes public repository on GitHub is added by selecting **Add a singe repository** and replacing `<owner>/<repository>` with `kubernetes/kubernetes`:

```
{
  "url": "https://github.com",
  "token": "<access token>",
  "orgs": [],
  "repos": [
    "kubernetes/kubernetes"
  ]
}
```

## Supported versions

- GitHub.com
- GitHub Enterprise v2.10 and newer

## Selecting repositories for code search

There are four fields for configuring which repositories are mirrored/synchronized:

- [`repos`](github.md#repos)<br>A list of repositories in `owner/name` format. The order determines the order in which we sync repository metadata and is safe to change.
- [`orgs`](github.md#orgs)<br>A list of organizations (every repository belonging to the organization will be cloned).
- [`repositoryQuery`](github.md#repositoryQuery)<br>A list of strings with three pre-defined options (`public`, `affiliated`, `none`, none of which are subject to result limitations), and/or a [GitHub advanced search query](https://github.com/search/advanced). Note: There is an existing limitation that requires the latter, GitHub advanced search queries, to return [less than 1000 results](#repositoryquery-returns-first-1000-results-only). See [this issue](https://github.com/sourcegraph/sourcegraph/issues/2562) for ongoing work to address this limitation.
- [`exclude`](github.md#exclude)<br>A list of repositories to exclude which takes precedence over the `repos`, `orgs`, and `repositoryQuery` fields.

### Private repositories

A [token that has the prerequisite scopes](#github-api-token-and-access) is required in order to clone private repositories for search, as well as at least read access to the relevant private repositories.

See [GitHub API token and access](#github-api-token-and-access) for more details.

## GitHub API token and access

The GitHub service requires a `token` in order to access their API. There are two different types of tokens you can supply:

- **[Personal access token](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line)**:<br>This gives Sourcegraph the same level of access to repositories as the account that created the token. If you're not wanting to mix your personal repositories with your organizations repositories, you could add an entry to the `exclude` array, or you can use a machine user token.
- **[Machine user token](https://developer.github.com/v3/guides/managing-deploy-keys/#machine-users)**:<br>Generates a token for a machine user that is affiliated with an organization instead of a user account.

No [token scopes](https://docs.github.com/en/developers/apps/building-oauth-apps/scopes-for-oauth-apps#available-scopes) are required if you only want to sync public repositories and don't want to use any of the following features. Otherwise, the following token scopes are required for specific features:

| Feature                                               | Required token scopes                                                                                          |
| ----------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| [Sync private repositories](#private-repositories)    | `repo`                                                                                                         |
| [Sync repository permissions][permissions]            | `repo`                                                                                                         |
| [Repository permissions caching][permissions-caching] | `repo`, `write:org`                                                                                            |
| [Batch changes][batch-changes]                        | `repo`, `read:org`, `user:email`, `read:discussion`, and `workflow` ([learn more][batch-changes-interactions]) |

[permissions]: ../repo/permissions.md#github
[permissions-caching]: ../repo/permissions.md#teams-and-organizations-permissions-caching
[batch-changes]: ../../batch_changes/index.md
[batch-changes-interactions]: ../../batch_changes/explanations/permissions_in_batch_changes.md#code-host-interactions-in-batch-changes

<span class="virtual-br"></span>

> WARNING: In addition to the prerequisite token scopes, the account attached to the token must actually have the same level of access to the relevant resources that you are trying to grant. For example:
>
> - If read access to repositories is required, the token must have `repo` scope *and* the token's account must have read access to the relevant repositories. This can happen by being directly granted read access to repositories, being on a team with read access to the repository, and so on.
> - If write access to repositories is required, the token must have `repo` scope *and* the token's account must have write access to all repositories. This can happen by being added as a direct contributor, being on a team with write access to the repository, being an admin for the repository's organization, and so on.
> - If write access to organizations is required, the token must have `write:org` scope *and* the token's account must have write access for all organizations. This can happen by being an admin in all relevant organizations.
>
> Learn more about how the GitHub API is used and what level of access is required in the corresponding feature documentation.

### Fine-grained personal access tokens

GitHub's fine-grained personal access tokens are not yet supported.

## GitHub.com rate limits

You should always include a token in a configuration for a GitHub.com URL to avoid being denied service by GitHub's [unauthenticated rate limits](https://developer.github.com/v3/#rate-limiting). If you don't want to automatically synchronize repositories from the account associated with your personal access token, you can create a token without a [`repo` scope](https://developer.github.com/apps/building-oauth-apps/scopes-for-oauth-apps/#available-scopes) for the purposes of bypassing rate limit restrictions only.

## GitHub Enterprise Server rate limits

Rate limiting may not be enabled by default. To check and verify the current rate limit settings, you may make a request to the `/rate_limit` endpoint like this:

```
$ curl -s https://<github-enterprise-url>/api/v3/rate_limit -H "Authorization: Bearer <token>"
{
  "message": "Rate limiting is not enabled.",
  "documentation_url": "https://docs.github.com/enterprise/3.3/rest/reference/rate-limit#get-rate-limit-status-for-the-authenticated-user"
}
```

### Internal rate limits

Internal rate limiting can be configured to limit the rate at which requests are made from Sourcegraph to GitHub. 

If enabled, the default rate is set at 5000 per hour which can be configured via the `requestsPerHour` field (see below):

- For Sourcegraph <=3.38, if rate limiting is configured more than once for the same code host instance, the most restrictive limit will be used.
- For Sourcegraph >=3.39, rate limiting should be enabled and configured for each individual code host connection.

**NOTE** Internal rate limiting is only currently applied when synchronising changesets in [batch changes](../../batch_changes/index.md), repository permissions and repository metadata from code hosts.

## Repository permissions

By default, all Sourcegraph users can view all repositories. To configure Sourcegraph to use
GitHub's per-user repository permissions, see "[Repository permissions](../repo/permissions.md#github)".

## User authentication

To configure GitHub as an authentication provider (which will enable sign-in via GitHub), see the
[authentication documentation](../auth/index.md#github).

## Webhooks

Using the `webhooks` property on the external service has been deprecated.

Please consult [this page](../config/webhooks.md) in order to configure webhooks.

## Configuration

GitHub connections support the following configuration options, which are specified in the JSON editor in the site admin "Manage code hosts" area.

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/github.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/github) to see rendered content.</div>

## Default branch

Sourcegraph displays search results from the default branch of a repository when no `revision:` [parameter](https://docs.sourcegraph.com/code_search/reference/queries#repository-revisions) is specified. If you'd like the search results to be displayed from another branch by default, you may [change a repo's default branch on the github repo settings page](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-branches-in-your-repository/changing-the-default-branch). If this is not an option, consider using [search contexts](https://docs.sourcegraph.com/code_search/how-to/search_contexts) instead. 

## Troubleshooting

### Hitting GitHub Search API rate limit with repositoryQuery
When Sourcegraph syncs repositories configured  via `repositoryQuery`, it consumes GitHub API search rate limit, which is lower than the normal rate limit. The `affiliated`, `public` and `none` special values, however, trigger normal API requests instead of search API requests.

When the search rate limit quota is exhausted, an error like `failed to list GitHub repositories for search: page=..., searchString=\"...\"` can be found in logs. To work around this try reducing the frequency with which repository syncing happens by setting a higher value (in minutes) of `repoListUpdateInterval` in your Sourcegraph [site config] (https://docs.sourcegraph.com/admin/config/site_config).

`repositoryQuery` is the only repo syncing method that consumes GitHub search API quota, so if setting `repoListUpdateInterval` doesn't work consider switching your syncing method to use another option, like `orgs`, or using one of the special values described above.

### "repositoryQuery": ["public"] does not return archived status of a repo

The  `repositoryQuery` option `"public"` is valuable in that it allows sourcegraph to sync all public repositories, however, it does not return whether or not a repo is archived. This can result in archived repos appearing in normal search. You can see an example of what is returned by the GitHub API for a query to "public" [here](https://docs.github.com/en/rest/reference/repos#list-public-repositories).

If you would like to sync all public repositories while omitting archived repos, consider generating a GitHub token with access to only public repositories, then use `respositoryQuery` with option `affiliated` and an `exclude` argument with option `public` as seen in the example below:
```
{
    "url": "https://github.example.com",
    "gitURLType": "http",
    "repositoryPathPattern": "devs/{nameWithOwner}",
    "repositoryQuery": [
        "affiliated"
    ],
    "token": "TOKEN_WITH_PUBLIC_ACCESS",
    "exclude": [
        {
            "archived": true
        }
    ]
}
```
