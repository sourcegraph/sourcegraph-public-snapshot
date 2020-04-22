# GitHub

Site admins can sync Git repositories hosted on [GitHub.com](https://github.com) and [GitHub Enterprise](https://enterprise.github.com) with Sourcegraph so that users can search and navigate the repositories.

To connect GitHub to Sourcegraph:

1. Go to **Site admin > Manage repositories > Add repositories**
1. Select **GitHub**.
1. Configure the connection to GitHub using the action buttons above the text field, and additional fields can be added using <kbd>Cmd/Ctrl+Space</kbd> for auto-completion. See the [configuration documentation below](#configuration).
1. Press **Add repositories**.

## Supported versions

- GitHub.com
- GitHub Enterprise v2.10 and newer

## Selecting repositories for code search

There are four fields for configuring which repositories are mirrored/synchronized:

- [`repos`](github.md#configuration)<br>A list of repositories in `owner/name` format.
- [`orgs`](github.md#configuration)<br>A list of organizations (every repository belonging to the organization will be cloned).
- [`repositoryQuery`](github.md#configuration)<br>A list of strings with three pre-defined options (`public`, `affiliated`, `none`), and/or a [GitHub advanced search query](https://github.com/search/advanced). Note: There is an existing limitation that requires GitHub advanced search queries to return [less than 1000 results](#repositoryquery-returns-first-1000-results-only). See [this issue](https://github.com/sourcegraph/sourcegraph/issues/2562) for ongoing work to address this limitation.
- [`exclude`](github.md#configuration)<br>A list of repositories to exclude which takes precedence over the `repos`, `orgs`, and `repositoryQuery` fields.

## GitHub API token and access

The GitHub service requires a `token` in order to access their API. There are two different types of tokens you can supply:

- **[Personal access token](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line)**:<br>This gives Sourcegraph the same level of acccess to repositories as the account that created the token. If you're not wanting to mix your personal repositories with your organizations repositories, you could add an entry to the `exclude` array, or you can use a machine user token.
- **[Machine user token](https://developer.github.com/v3/guides/managing-deploy-keys/#machine-users)**:<br>Generates a token for a machine user that is affiliated with an organization instead of a user account.

No token scopes are required if you only want to sync public repositories and don't want to use any of the following features. Otherwise, the following token scopes are required:

- `repo` to sync private repositories from GitHub to Sourcegraph.
- `read:org` to use the `"allowOrgs"` setting [with a GitHub authentication provider](../auth/index.md#github).
- `repo`, `read:org`, and `read:discussion` to use [Campaigns](../../user/campaigns/index.md) with GitHub repositories.

>NOTE: If you plan to use repository permissions with background syncing, an access token that has admin access to all private repositories is required. It is because only admin can list all collaborators of a repository.

## GitHub.com rate limits

You should always include a token in a configuration for a GitHub.com URL to avoid being denied service by GitHub's [unauthenticated rate limits](https://developer.github.com/v3/#rate-limiting). If you don't want to automatically synchronize repositories from the account associated with your personal access token, you can create a token without a [`repo` scope](https://developer.github.com/apps/building-oauth-apps/scopes-for-oauth-apps/#available-scopes) for the purposes of bypassing rate limit restrictions only.

## Repository permissions

By default, all Sourcegraph users can view all repositories. To configure Sourcegraph to use
GitHub's per-user repository permissions, see "[Repository permissions](../repo/permissions.md#github)".

## User authentication

To configure GitHub as an authentication provider (which will enable sign-in via GitHub), see the
[authentication documentation](../auth/index.md#github).

## Webhooks

The `webhooks` setting allows specifying the org webhook secrets necessary to authenticate incoming webhook requests to `/.api/github-webhooks`.

```json
"webhooks": [
  {"org": "your_org", "secret": "verylongrandomsecret"}
]
```

These organization webhooks are optional, but if configured on GitHub, they allow faster metadata updates than the background syncing (i.e. polling) with `repo-updater` permits.

The following [webhook events](https://developer.github.com/webhooks/) are currently used:

- Issue comments
- Pull requests
- Pull request reviews
- Pull request review comments
- Check runs
- Check suites
- Statuses

To set up a organization webhook on GitHub, go to the settings page of your organization. From there, click **Webhooks**, then **Add webhook**.

Fill in your Sourcegraph external URL with `/.api/github-webhooks` as the path and make sure it is publicly available.

The **Content Type** of the webhook should be `application/json`. Generate the secret with `openssl rand -hex 32` and paste it in the respective field. This value is what you need to specify in the GitHub config.

Click on **Enable SSL verification** if you have configured SSL with a valid certificate in your Sourcegraph instance.

Select **the events mentioned above** on the events section, ensure **Active** is checked and finally create the webhook.

## Configuration

GitHub connections support the following configuration options, which are specified in the JSON editor in the site admin "Manage repositories" area.

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/github.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/github) to see rendered content.</div>

## Troubleshooting

### RepositoryQuery returns first 1000 results only

GitHub's [Search API](https://developer.github.com/v3/search/) only returns the first 1000 results. Therefore a `repositoryQuery` needs to return a 1000 results or less otherwise Sourcegraph will not synchronize some repositories. To workaround this limitation you can split your query into multiple queries, each returning less than a 1000 results. For example if your query is `org:Microsoft fork:no` you can adjust your query to:

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
