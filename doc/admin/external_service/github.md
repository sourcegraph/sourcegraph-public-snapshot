# GitHub

Site admins can sync Git repositories hosted on [GitHub.com](https://github.com) and [GitHub Enterprise](https://enterprise.github.com) with Sourcegraph so that users can search and navigate the repositories.

There are 2 ways to connect with GitHub:
1. [Using a GitHub App (recommended)](#using-a-github-app)
2. [Using an access token](#using-an-access-token)

## Supported versions

- GitHub.com
- GitHub Enterprise v2.10 and newer

## Using a GitHub App

<span class="badge badge-note">Sourcegraph 5.1+</span>

To create a GitHub app and connect it to Sourcegraph:  

1. Go to **Site admin > Github Apps** in the Sourcegraph app.  
    ![Github Apps page screenshot](https://storage.googleapis.com/sourcegraph-assets/docs/images/administration/config/github-apps/empty-github-apps-list.png)
2. Click **Create GitHub App**.  
3. Enter a name for your app, the URL of your GitHub instance, and optionally an organization that will own the app. If no organization is specified, the app will be owned by the user account that creates it in GitHub.  
    ![Create Github App form screenshot](https://storage.googleapis.com/sourcegraph-assets/docs/images/administration/config/github-apps/create-app-form.png)
4. Click **Create GitHub App**. You will be redirected to GitHub to confirm the app creation.  
    ![Create Github App screenshot](https://storage.googleapis.com/sourcegraph-assets/docs/images/administration/config/github-apps/create-app-page.png)
5. After creating the app, you will be redirected to the app installation screen on GitHub. Review the app permissions and select which repositories the app can access. The default is **All repositories**. You can change this later.  
    ![Install Github App screenshot](https://storage.googleapis.com/sourcegraph-assets/docs/images/administration/config/github-apps/installation-page.png)
6. Click **Install**. You will be redirected back to Sourcegraph, where the GitHub app setup screen will display.  
7. Sourcegraph needs to map Sourcegraph users to GitHub users. Click **Reveal secret** to get the JSON configuration for the auth provider and copy/paste it into the `"auth.providers"` section of your site configuration.  
    ![Auth provider JSON screenshot](https://storage.googleapis.com/sourcegraph-assets/docs/images/administration/config/github-apps/auth-provider-json.png)
8. After setting up the auth provider, go back to the GitHub App setup screen in Sourcegraph. Click **Add connection** under your new installation to create a code host connection to GitHub with this app installation. By default, it will sync all repositories the app can access within the organization it was installed on. Repository permission enforcement will also be turned on by default.
    ![Add connection button screenshot](https://storage.googleapis.com/sourcegraph-assets/docs/images/administration/config/github-apps/add-connection-button.png)

You can now [select repositories to sync](#selecting-repositories-to-sync) or see more configuration options in the [configuration section](#configuration).

### Mutliple installations

In some cases, having multiple installations is beneficial, especially if you want to sync repositories from multiple organizations.  

By default, Sourcegraph creates a private GitHub app, which allows only one installation on the owner account. To allow multiple installations, the app needs to be made public. For security considerations, see [GitHub's documentation on private vs public apps](https://docs.github.com/en/apps/creating-github-apps/setting-up-a-github-app/making-a-github-app-public-or-private). To make the app public, [follow GitHub's documentation](https://docs.github.com/en/apps/maintaining-github-apps/modifying-a-github-app#changing-the-visibility-of-a-github-app).  

After making the app public, go back to Sourcegraph **Site admin > GitHub Apps > [YOUR APP]** and click **Add installation**. You will be redirected to GitHub to pick which other organization to install the app on and finish the installation process.

To sync repositories from this installation, click **Add connection** in Sourcegraph.  

> NOTE: Each organization that the app needs to access requires its own installation.
## Using an access token

To connect GitHub to Sourcegraph with an access token:

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

## GitHub API access

GitHub requires a `token` in order to access their API. There are different types of tokens that can be supplied. When using GitHub apps, this is handled automatically by Sourcegraph.

- **[GitHub app installation access token](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-an-installation-access-token-for-a-github-app)**:<br>An installation access token is created automatically when you install a GitHub app. Do not set this token in the code host connection configuration. This token gives Sourcegraph the same level of access to repositories as the GitHub app installation.
- **[Personal access token](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line)**:<br>This gives Sourcegraph the same level of access to repositories as the account that created the token. If you don't want to mix your personal repositories with your organizations repositories, you could add an entry to the `exclude` array, or you can use a machine user token or a fine-grained access token.
- **[Fine-grained access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token#creating-a-fine-grained-personal-access-token)**:<br>Allows scoping access tokens to specific repositories with specific permissions. Consult the [table below](#fine-grained-access-token-permissions) for the required permissions.
- **[Machine user token](https://developer.github.com/v3/guides/managing-deploy-keys/#machine-users)**:<br>Generates a token for a machine user that is affiliated with an organization instead of a user account.

### Personal access token scopes

No [token scopes](https://docs.github.com/en/developers/apps/building-oauth-apps/scopes-for-oauth-apps#available-scopes) are required if you only want to sync public repositories and don't want to use any of the following features. Otherwise, the following token scopes are required for specific features:

| Feature                                               | Required token scopes                                                                                          |
| ----------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| [Sync private repositories](#private-repositories)    | `repo`                                                                                                         |
| [Sync repository permissions][permissions]            | `repo`                                                                                                         |
| [Batch changes][batch-changes]                        | `repo`, `read:org`, `user:email`, `read:discussion`, and `workflow` ([learn more][batch-changes-interactions]) |

[permissions]: #repository-permissions
[permissions-caching]: #teams-and-organizations-permissions-caching
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

### Fine-grained access token permissions

Fine-grained tokens can access public repositories, but can only access the private repositories of the account they are scoped to.

When creating your fine-grained access token, select the following permissions depending on the purpose of the token:

| Feature                                               | Required token permissions                                  |
| ----------------------------------------------------- | ------------------------------------------------------ |
| [Sync private repositories](#private-repositories)    | `Repository permissions: Contents - Access: Read-only` |
| [Sync repository permissions][permissions]            | `Repository permissions: Contents - Access: Read-only` |
| [Batch changes][batch-changes]                        | `Unsupported`                                          |

<br>

> WARNING: Fine-grained tokens don't support the `repositoryQuery` code host connection option or batch changes. Both of these features rely on GitHub's GraphQL API, which is [unsupported by fine-grained access tokens](https://docs.github.com/en/graphql/guides/forming-calls-with-graphql#authenticating-with-graphql).

### Private repositories

To clone and search private repositories, we need [a GitHub access token](#github-api-access) with the required scopes and at least read access to the relevant private repositories.

For more details, see [GitHub API access](#github-api-access).

## Selecting repositories to sync

There are four fields for configuring which repositories are mirrored/synchronized:

- [`repos`](github.md#repos)<br>A list of repositories in `owner/name` format. The order determines the order in which we sync repository metadata and is safe to change.
- [`orgs`](github.md#orgs)<br>A list of organizations (every repository belonging to the organization will be cloned).
- [`repositoryQuery`](github.md#repositoryQuery)<br>A list of strings with three pre-defined options (`public`, `affiliated`, `none`, none of which are subject to result limitations), and/or a [GitHub advanced search query](https://github.com/search/advanced). Note: There is an existing limitation that requires the latter, GitHub advanced search queries, to return [less than 1000 results](#repositoryquery-returns-first-1000-results-only). See [this issue](https://github.com/sourcegraph/sourcegraph/issues/2562) for ongoing work to address this limitation.
- [`exclude`](github.md#exclude)<br>A list of repositories to exclude which takes precedence over the `repos`, `orgs`, and `repositoryQuery` fields.

## Rate limits

Always include a token in a configuration for a GitHub.com URL to avoid being denied service by GitHub's [unauthenticated rate limits](https://docs.github.com/en/rest/overview/resources-in-the-rest-api?apiVersion=2022-11-28#rate-limiting). If you don't want to automatically synchronize repositories from the account associated with your personal access token, you can create a token without a [`repo` scope](https://developer.github.com/apps/building-oauth-apps/scopes-for-oauth-apps/#available-scopes) for the purposes of bypassing rate limit restrictions only.

When Sourcegraph hits a rate limit imposed by GitHub, Sourcegraph waits the appropriate amount of time specified by GitHub before retrying the request. This can be several minutes in extreme cases.

### GitHub Enterprise Server rate limits

Rate limiting may not be enabled by default. To check and verify the current rate limit settings, you may make a request to the `/rate_limit` endpoint like this:

```
$ curl -s https://<github-enterprise-url>/api/v3/rate_limit -H "Authorization: Bearer <token>"
{
  "message": "Rate limiting is not enabled.",
  "documentation_url": "https://docs.github.com/enterprise/3.3/rest/reference/rate-limit#get-rate-limit-status-for-the-authenticated-user"
}
```

### Internal rate limits

See [Internal rate limits](./rate_limits.md#internal-rate-limits).

## Repository permissions

Prerequisite for configuring repository permission syncing: [Add GitHub as an authentication provider](../auth/index.md#github).

Then, add or edit the GitHub connection as described above and include the `authorization` field:

```json
{
  // ...
  "authorization": {}
}
```

This needs to be done for every github code host connection if there is more than one configured.

Repo-centric permission syncing is done by calling the [list repository collaborators GitHub API endpoint](https://docs.github.com/en/rest/collaborators/collaborators#list-repository-collaborators). To call this API endpoint correctly, we need a [GitHub access token](#github-api-access) with the required scopes and read and write access to all relevant repositories.

> IMPORTANT: We strongly recommend configuring both read and write access to associated repositories for permission syncing due to GitHub's token scope requirements. Without write access, there will be a conflict between [user-centric sync](../permissions/syncing.md#troubleshooting) and repo-centric sync. In that case, [disable repo-centric permission sync](../permissions/syncing.md#disable-repo-centric-permission-sync) (supported in <span class="badge badge-note">Sourcegraph 5.0.4+</span>).

<span class="virtual-br"></span>

> IMPORTANT: Optional, but strongly recommended - [continue with configuring webhoooks for permissions](../config/webhooks/incoming.md#user-permissions).

<span class="virtual-br"></span>

> NOTE: It can take some time to complete full cycle of repository permissions sync if you have a large number of users or repositories. [See sync duration time](../permissions/syncing.md#sync-duration) for more information.

### Trigger permissions sync from GitHub webhooks

Follow the link to [configure webhooks for permissions for Github](../config/webhooks/incoming.md#user-permissions)

### Teams and organizations permissions caching

<span class="badge badge-experimental">Experimental</span>

> WARNING: The following section is experimental and might not work properly anymore on new Sourcegraph versions (post 4.0+). Please prefer [configuring webhooks for permissions instead](../config/webhooks/incoming.md#user-permissions)

Github code host can leverage caching mechanisms to reduce the number of API calls used when syncing permissions. This can significantly reduce the amount of time it takes to perform a full cycle of permissions sync due to reduced instances of being rate limited by the code host, and is useful for code hosts with very large numbers of users and repositories.

Sourcegraph can leverage caching of GitHub [team](https://docs.github.com/en/organizations/managing-access-to-your-organizations-repositories/managing-team-access-to-an-organization-repository) and [organization](https://docs.github.com/en/organizations/managing-access-to-your-organizations-repositories/repository-permission-levels-for-an-organization) permissions.

> NOTE: You should only try this if your GitHub setup makes extensive use of GitHub teams and organizations to distribute access to repositories and your number of `users * avg_repositories` is greater than 250,000 (which roughly corresponds to the scale at which [GitHub rate limits might become an issue](../permissions/syncing.md#sync-duration)).
<!-- 5,000 requests an hour * 100 items per page / 2-way sync = approx. 250,000 items before hitting a limit -->

This caching behaviour can be enabled via the `authorization.groupsCacheTTL` field:

```json
{
   "url": "https://github.example.com",
   "token": "$PERSONAL_ACCESS_TOKEN",
   "authorization": {
     "groupsCacheTTL": 72, // hours
   }
}
```

In the corresponding [authorization provider](../auth/index.md#github) in [site configuration](./../config/site_config.md), the `allowGroupsPermissionsSync` field must be set as well for the correct auth scopes to be requested from users:

```json
{
  // ...
  "auth.providers": [
    {
      "type": "github",
      "url": "https://github.example.com",
      "allowGroupsPermissionsSync": true,
    }
  ]
}
```

A [token that has the required scopes](#github-api-access) and both read and write access to all relevant repositories and organizations is needed to fetch repository permissions and team memberships. 
Read-only access will *not* work with cached permissions sync, but will work with careful configuration for [regular GitHub permissions sync](#repository-permissions).

When enabling this feature, we currently recommend a default `groupsCacheTTL` of `72` (hours, or 3 days). A lower value can be set if your teams and organizations change frequently, though the chosen value must be at least several hours for the cache to be leveraged in the event of being rate-limited (which takes [an hour to recover from](https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting)).

Cache invaldiation happens automatically on certain webhook events, so it is recommended to configure webhook support when using cached permissions sync.
Caches can also be [manually invalidated](#manually-invalidate-caches) if necessary.

#### Manually invalidate caches

To force a bypass of caches during a sync, you can manually queue users or repositories for sync with the `invalidateCaches` options via the Sourcegraph GraphQL API:

```gql
mutation {
  scheduleUserPermissionsSync(user: "userid", options: {invalidateCaches: true}) {
    alwaysNil
  }
}
```

## User authentication

To configure GitHub as an authentication provider (which will enable sign-in via GitHub), see the
[authentication documentation](../auth/index.md#github).

## Webhooks

Using the `webhooks` property on the external service has been deprecated.

Please consult [this page](../config/webhooks/incoming.md) in order to configure webhooks.

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

If you would like to sync all public repositories while omitting archived repos, consider generating a GitHub token with access to only public repositories, then use `repositoryQuery` with option `affiliated` and an `exclude` argument with option `public` as seen in the example below:
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
