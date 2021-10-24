# GitHub

Site admins can sync Git repositories hosted on [GitHub.com](https://github.com) and [GitHub Enterprise](https://enterprise.github.com) with Sourcegraph so that users can search and navigate the repositories.

To connect GitHub to Sourcegraph:

1. Depending on whether you are a site admin or user:
    1. *Site admin*: Go to **Site admin > Manage repositories > Add repositories**
    1. *User*: Go to **Settings > Manage repositories**.
1. Select **GitHub**.
1. Configure the connection to GitHub using the action buttons above the text field, and additional fields can be added using <kbd>Cmd/Ctrl+Space</kbd> for auto-completion. See the [configuration documentation below](#configuration).
1. Press **Add repositories**.

> NOTE: Adding code hosts as a user is currently in private beta.

## Supported versions

- GitHub.com
- GitHub Enterprise v2.10 and newer

## Selecting repositories for code search

There are four fields for configuring which repositories are mirrored/synchronized:

- [`repos`](github.md#configuration)<br>A list of repositories in `owner/name` format.
- [`orgs`](github.md#configuration)<br>A list of organizations (every repository belonging to the organization will be cloned).
- [`repositoryQuery`](github.md#configuration)<br>A list of strings with three pre-defined options (`public`, `affiliated`, `none`, none of which are subject to result limitations), and/or a [GitHub advanced search query](https://github.com/search/advanced). Note: There is an existing limitation that requires the latter, GitHub advanced search queries, to return [less than 1000 results](#repositoryquery-returns-first-1000-results-only). See [this issue](https://github.com/sourcegraph/sourcegraph/issues/2562) for ongoing work to address this limitation.
- [`exclude`](github.md#configuration)<br>A list of repositories to exclude which takes precedence over the `repos`, `orgs`, and `repositoryQuery` fields.

## GitHub API token and access

The GitHub service requires a `token` in order to access their API. There are two different types of tokens you can supply:

- **[Personal access token](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line)**:<br>This gives Sourcegraph the same level of access to repositories as the account that created the token. If you're not wanting to mix your personal repositories with your organizations repositories, you could add an entry to the `exclude` array, or you can use a machine user token.
- **[Machine user token](https://developer.github.com/v3/guides/managing-deploy-keys/#machine-users)**:<br>Generates a token for a machine user that is affiliated with an organization instead of a user account.

No token scopes are required if you only want to sync public repositories and don't want to use any of the following features. Otherwise, the following token scopes are required for specific features:

| Feature                                               | Required token scopes                                                                              |
| ----------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| [Sync private repositories](#github)                  | `read:repo`                                                                                        |
| [Sync repository permissions][permissions]            | `write:repo`                                                                                       |
| [Repository permissions caching][permissions-caching] | `write:org`                                                                                        |
| [Batch changes][batch-changes]                        | `repo`, `read:org`, `user:email`, and `read:discussion` ([learn more][batch-changes-interactions]) |

[permissions]: ../repo/permissions.md#github
[permissions-caching]: ../repo/permissions.md#teams-and-organizations-permissions-caching
[batch-changes]: ../../batch_changes/index.md
[batch-changes-interactions]: ../../batch_changes/explanations/permissions_in_batch_changes.md#code-host-interactions-in-batch-changes

## GitHub.com rate limits

You should always include a token in a configuration for a GitHub.com URL to avoid being denied service by GitHub's [unauthenticated rate limits](https://developer.github.com/v3/#rate-limiting). If you don't want to automatically synchronize repositories from the account associated with your personal access token, you can create a token without a [`repo` scope](https://developer.github.com/apps/building-oauth-apps/scopes-for-oauth-apps/#available-scopes) for the purposes of bypassing rate limit restrictions only.

### Internal rate limits

Internal rate limiting can be configured to limit the rate at which requests are made from Sourcegraph to GitHub. 

If enabled, the default rate is set at 5000 per hour which can be configured via the `requestsPerHour` field (see below). If rate limiting is configured more than once for the same code host instance, the most restrictive limit will be used.

**NOTE** Internal rate limiting is only currently applied when synchronising changesets in [batch changes](../../batch_changes/index.md), repository permissions and repository metadata from code hosts.

## Repository permissions

By default, all Sourcegraph users can view all repositories. To configure Sourcegraph to use
GitHub's per-user repository permissions, see "[Repository permissions](../repo/permissions.md#github)".

## User authentication

To configure GitHub as an authentication provider (which will enable sign-in via GitHub), see the
[authentication documentation](../auth/index.md#github).

## Webhooks

The `webhooks` setting allows specifying the organization webhook secrets necessary to authenticate incoming webhook requests to `/.api/github-webhooks`.

```json
"webhooks": [
  {"org": "your_org", "secret": "verylongrandomsecret"}
]
```

Using webhooks is highly recommended when using [batch changes](../../batch_changes/index.md), since they speed up the syncing of pull request data between GitHub and Sourcegraph and make it more efficient.

To set up webhooks:

1. In Sourcegraph, go to **Site admin > Manage repositories** and edit the GitHub configuration.
1. Add the `"webhooks"` property to the configuration (you can generate a secret with `openssl rand -hex 32`):<br /> `"webhooks": [{"org": "your_org", "secret": "verylongrandomsecret"}]`
1. Click **Update repositories**.
1. Copy the webhook URL displayed below the **Update repositories** button.
1. On GitHub, go to the settings page of your organization. From there, click **Settings**, then **Webhooks**, then **Add webhook**.
1. Fill in the webhook form:
   * **Payload URL**: the URL you copied above from Sourcegraph.
   * **Content type**: this must be set to `application/json`.
   * **Secret**: the secret token you configured Sourcegraph to use above.
   * **Which events**: select **Let me select individual events**, and then enable:
     - Issue comments
     - Pull requests
     - Pull request reviews
     - Pull request review comments
     - Check runs
     - Check suites
     - Statuses
   * **Active**: ensure this is enabled.
1. Click **Add webhook**.
1. Confirm that the new webhook is listed.

Done! Sourcegraph will now receive webhook events from GitHub and use them to sync pull request events, used by [batch changes](../../batch_changes/index.md), faster and more efficiently.

## Configuration

GitHub connections support the following configuration options, which are specified in the JSON editor in the site admin "Manage repositories" area.

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/github.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/github) to see rendered content.</div>

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
### repositoryQuery returns first 1000 results only

GitHub's [Search API](https://developer.github.com/v3/search/) only returns the first 1000 results. Therefore a `repositoryQuery` (other than the three pre-defined options) needs to return a 1000 results or less otherwise Sourcegraph will not synchronize some repositories. To workaround this limitation you can split your query into multiple queries, each returning less than a 1000 results. For example if your query is `org:Microsoft fork:no` you can adjust your query to:

```jsonx
{
  // ...
  "repositoryQuery": [
    "org:Microsoft fork:no created:>=2019",
    "org:Microsoft fork:no created:2018",
    "org:Microsoft fork:no created:2016..2017",
    "org:Microsoft fork:no created:<2016"
  ]
}
```

If splitting by creation date does not work, try another field. See [GitHub advanced search query](https://github.com/search/advanced) for other fields you can try.

See [Handle GitHub repositoryQuery that has more than 1000 results](https://github.com/sourcegraph/sourcegraph/issues/2562) for ongoing work to address this limitation.
