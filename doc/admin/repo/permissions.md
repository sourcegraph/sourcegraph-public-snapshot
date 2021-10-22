# Repository permissions

Sourcegraph can be configured to enforce repository permissions from code hosts.

Currently, GitHub, GitHub Enterprise, GitLab and Bitbucket Server permissions are supported. Check our [product direction](https://about.sourcegraph.com/direction) for plans to support other code hosts. If your desired code host is not yet on the roadmap, please [open a feature request](https://github.com/sourcegraph/sourcegraph/issues/new?template=feature_request.md).

If the Sourcegraph instance is configured to sync repositories from multiple code hosts (regardless of whether they are the same code host, e.g. `GitHub + GitHub` or `GitHub + GitLab`), setting up permissions for each code host will make repository permissions apply holistically on Sourcegraph. 

Setting up a unified SSO for code hosts and Sourcegraph is also possible: how to [Set up Sourcegraph with two GitLab and Keycloak using SAML](https://unknwon.io/posts/200915_setup-sourcegraph-gitlab-keycloak/).

> NOTE: Site admin users bypass all permission checks and have access to every repository on Sourcegraph.

<span class="virtual-br"></span>

> WARNING: It can take some time to complete mirroring repository permissions from a code host. [Learn more](#permissions-sync-times).

<br />

## GitHub

Prerequisite: [Add GitHub as an authentication provider](../auth/index.md#github).

Then, [add or edit a GitHub connection](../external_service/github.md#repository-syncing) and include the `authorization` field:

```json
{
  // The GitHub URL used to set up the GitHub authentication provider must match this URL.
  "url": "https://github.com",
  "token": "$PERSONAL_ACCESS_TOKEN",
  "authorization": {}
}
```

> WARNING: It can take some time to complete mirroring repository permissions from a code host. [Learn more](#permissions-sync-times).

### Faster permissions syncing via GitHub webhooks

<span class="badge badge-note">Sourcegraph 3.22+</span>

Sourcegraph can speed up permissions syncing by receiving webhooks from GitHub for events related to user and repo permissions. To set up webhooks, follow the guide in the [GitHub Code Host Docs](../external_service/github.md#webhooks). These events will enqueue permissions syncs for the repositories or users mentioned, meaning things like publicising / privatising repos, or adding collaborators will be reflected in your Sourcegraph searches more quickly. For this to work the user must have logged in via the [GitHub OAuth provider](../auth.md#github).

The events we consume are:

* [public](https://developer.github.com/webhooks/event-payloads/#public)
* [repository](https://developer.github.com/webhooks/event-payloads/#repository)
* [member](https://developer.github.com/webhooks/event-payloads/#member)
* [membership](https://developer.github.com/webhooks/event-payloads/#membership)
* [team_add](https://developer.github.com/webhooks/event-payloads/#team_add)
* [organization](https://developer.github.com/webhooks/event-payloads/#organization)

### Teams and organizations permissions caching

<span class="badge badge-experimental">Experimental</span> <span class="badge badge-note">Sourcegraph 3.31+</span>

For GitHub providers, Sourcegraph can leverage caching of GitHub [team](https://docs.github.com/en/organizations/managing-access-to-your-organizations-repositories/managing-team-access-to-an-organization-repository) and [organization](https://docs.github.com/en/organizations/managing-access-to-your-organizations-repositories/repository-permission-levels-for-an-organization) permissions - [learn more about permissions caching](#permissions-caching).

> NOTE: You should only try this if your GitHub setup makes extensive use of GitHub teams and organizations to distribute access to repositories and your number of `users * repos` is greater than 500,000 (which roughly corresponds to the scale at which [GitHub rate limits might become an issue](#permissions-sync-times)).
<!-- 5,000 requests an hour * 100 items per page = approx. 500,000 items before hitting a limit -->

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

When enabling this feature, we currently recommend a default of `72` (hours, or 3 days) for `groupsCacheTTL`. A lower value can be set if your teams and organizations change frequently, though the chosen value must be at least several hours for the cache to be leveraged in the event of being rate-limited (which takes [an hour to recover from](https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting)).

Caches can also be [manually invalidated](#permissions-caching) if necessary.
Cache invaldiation also happens automatically on certain [webhook events](#faster-permissions-syncing-via-github-webhooks).

> NOTE: The token associated with the external service must have `repo` and `write:org` scope in order to read the repo, orgs, and teams permissions and cache them - [learn more](../external_service/github.md#github-api-token-and-access).

<br />

## GitLab

GitLab permissions can be configured in three ways:

1. Set up GitLab as an OAuth sign-on provider for Sourcegraph (recommended)
2. Use a GitLab administrator (sudo-level) personal access token in conjunction with another SSO provider
   (recommended only if the first option is not possible)
3. Assume username equivalency between Sourcegraph and GitLab (warning: this is generally unsafe and
   should only be used if you are using strictly `http-header` authentication).

> WARNING: It can take some time to complete mirroring repository permissions from a code host. [Learn more](#permissions-sync-times).

### OAuth application

Prerequisite: [Add GitLab as an authentication provider.](../auth/index.md#gitlab)

Then, [add or edit a GitLab connection](../external_service/gitlab.md#repository-syncing) and include the `authorization` field:

```json
{
  "url": "https://gitlab.com",
  "token": "$PERSONAL_ACCESS_TOKEN",
  "authorization": {
    "identityProvider": {
      "type": "oauth"
    }
  }
}
```

### Administrator (sudo-level) access token

This method requires administrator access to GitLab so that Sourcegraph can access the [admin GitLab Users API endpoint](https://docs.gitlab.com/ee/api/users.html#for-admins). For each GitLab user, this endpoint provides the user ID that comes from the authentication provider, so Sourcegraph can associate a user in its system to a user in GitLab.

Prerequisite: Add the [SAML](../auth/index.md#saml) or [OpenID Connect](../auth/index.md#openid-connect)
authentication provider you use to sign into GitLab.

Then, [add or edit a GitLab connection](../external_service/gitlab.md#repository-syncing) using an administrator (sudo-level) personal access token, and include the `authorization` field:

```json
{
  "url": "https://gitlab.com",
  "token": "$PERSONAL_ACCESS_TOKEN",
  "authorization": {
    "identityProvider": {
      "type": "external",
      "authProviderID": "$AUTH_PROVIDER_ID",
      "authProviderType": "$AUTH_PROVIDER_TYPE",
      "gitlabProvider": "$AUTH_PROVIDER_GITLAB_ID"
    }
  }
}
```

`$AUTH_PROVIDER_ID` and `$AUTH_PROVIDER_TYPE` identify the authentication provider to use and should
match the fields specified in the authentication provider config
(`auth.providers`). The authProviderID can be found in the `configID` field of the auth provider config.

`$AUTH_PROVIDER_GITLAB_ID` should match the `identities.provider` returned by
[the admin GitLab Users API endpoint](https://docs.gitlab.com/ee/api/users.html#for-admins).

### Username

Prerequisite: Ensure that `http-header` is the *only* authentication provider type configured for
Sourcegraph. If this is not the case, then it will be possible for users to escalate privileges,
because Sourcegraph usernames are mutable.

[Add or edit a GitLab connection](../external_service/gitlab.md#repository-syncing) and include the `authorization` field:

```json
{
  "url": "https://gitlab.com",
  "token": "$PERSONAL_ACCESS_TOKEN",
  "authorization": {
    "identityProvider": {
      "type": "username"
    }
  }
}
```

<br />

## Bitbucket Server

Enforcing Bitbucket Server permissions can be configured via the `authorization` setting in its configuration.

> WARNING: It can take some time to complete mirroring repository permissions from a code host. [Learn more](#permissions-sync-times).

### Prerequisites

1. You have the exact same user accounts, **with matching usernames**, in Sourcegraph and Bitbucket Server. This can be accomplished by configuring an [external authentication provider](../auth/index.md) that mirrors user accounts from a central directory like LDAP or Active Directory. The same should be done on Bitbucket Server with [external user directories](https://confluence.atlassian.com/bitbucketserver/external-user-directories-776640394.html).
1. Ensure you have set `auth.enableUsernameChanges` to **`false`** in the [site config](../config/site_config.md) to prevent users from changing their usernames and **escalating their privileges**.


### Setup

This section walks you through the process of setting up an *Application Link between Sourcegraph and Bitbucket Server* and configuring the Sourcegraph Bitbucket Server configuration with `authorization` settings. It assumes the above prerequisites are met.

As an admin user, go to the "Application Links" page. You can use the sidebar navigation in the admin dashboard, or go directly to [https://bitbucketserver.example.com/plugins/servlet/applinks/listApplicationLinks](https://bitbucketserver.example.com/plugins/servlet/applinks/listApplicationLinks).

<img src="https://imgur.com/Hg4bzOf.png" width="800">

---

Write Sourcegraph's external URL in the text area (e.g. `https://sourcegraph.example.com`) and click **Create new link**. Click **Continue** even if Bitbucket Server warns you about the given URL not responding.

<img src="https://imgur.com/x6vFKIL.png" width="800">

---

Write `Sourcegraph` as the *Application Name* and select `Generic Application` as the *Application Type*. Leave everything else unset and click **Continue**.

<img src="https://imgur.com/161rbB9.png" width="800">

---


Now click the edit button in the `Sourcegraph` Application Link that you just created and select the `Incoming Authentication` panel.

<img src="https://imgur.com/sMGmzhH.png" width="800">

---


Generate a *Consumer Key* in your terminal with `echo sourcegraph$(openssl rand -hex 16)`. Copy this command's output and paste it in the *Consumer Key* field. Write `Sourcegraph` in the *Consumer Name* field.

<img src="https://imgur.com/1kK2Y5x.png" width="800">

---

Generate an RSA key pair in your terminal with `openssl genrsa -out sourcegraph.pem 4096 && openssl rsa -in sourcegraph.pem -pubout > sourcegraph.pub`. Copy the contents of `sourcegraph.pub` and paste them in the *Public Key* field.

<img src="https://imgur.com/YHm1uSr.png" width="800">

---

Scroll to the bottom and check the *Allow 2-Legged OAuth* checkbox, then write your admin account's username in the *Execute as* field and, lastly, check the *Allow user impersonation through 2-Legged OAuth* checkbox. Press **Save**.

<img src="https://imgur.com/1qxEAye.png" width="800">

---

Go to your Sourcegraph's *Manage repositories* page (i.e. `https://sourcegraph.example.com/site-admin/external-services`) and either edit or create a new *Bitbucket Server* connection. Add the following settings:

```
{
	// Other config goes here
	"authorization": {
		"identityProvider": {
			"type": "username"
		},
		"oauth": {
			"consumerKey": "<KEY GOES HERE>",
			"signingKey": "<KEY GOES HERE>"
		}
	}
}
```

Copy the *Consumer Key* you generated before to the `oauth.consumerKey` field and the output of the command `base64 sourcegraph.pem | tr -d '\n'` to the `oauth.signingKey` field. Save your changes.


Finally, **save the configuration**. You're done!

### Fast permission sync with Bitbucket Server plugin

By installing the [Bitbucket Server plugin](../../../integration/bitbucket_server.md), you can make use of the fast permission sync feature that allows using Bitbucket Server permissions on larger instances.

<br />

## Permissions sync times

When syncing permissions from code hosts with large numbers of users and repositories, it can take some time to complete mirroring repository permissions from a code host, typically due to rate limits on a code host that limits how quickly Sourcegraph can query for repository permissions.

To mitigate this, Sourcegraph can leverage:

* [Background permissions syncing](#background-permissions-syncing)
* [Provider-specific optimizations](#provider-specific-optimizations)

### Background permissions syncing

<span class="badge badge-note">Sourcegraph 3.17+</span>

Sourcegraph supports syncing permissions in the background by default to better handle repository permissions at scale for GitHub, GitLab, and Bitbucket Server code hosts, and has been the only permissions mirror option since Sourcegraph 3.19. Rather than syncing a user's permissions when they log in and potentially blocking them from seeing search results, Sourcegraph syncs these permissions asynchronously in the background, opportunistically refreshing them in a timely manner.

For older versions (Sourcegraph 3.14, 3.15, and 3.16), background permissions syncing is behind a feature flag in the [site configuration](../config/site_config.md):

```json
"permissions.backgroundSync": {
	"enabled": true
}
```

Benefits of background syncing:

1. More predictable load on the code host API due to maintaining a schedule of permission updates.
1. Permissions are quickly synced for new repositories added to the Sourcegraph instance.
1. Users who sign up on the Sourcegraph instance can immediately get search results from some repositories they have access to on the code host as we begin to incrementally sync their permissions.

Considerations when enabling for the first time:

1. While the initial sync for all repositories and users is happening, users can gradually see more and more search results from repositories they have access to.
1. It takes time to complete the first sync. Depending on how many private repositories and users you have on the Sourcegraph instance, it can take from a few minutes to several hours. This is generally not a problem for fresh installations, since admins should only make the instance available after it's ready, but for existing installations, active users may not see the repositories they expect in search results because the initial permissions syncing hasn't finished yet.
1. More requests to the code host API need to be done during the first sync, but their pace is controlled with rate limiting.

Please contact [support@sourcegraph.com](mailto:support@sourcegraph.com) if you have any concerns/questions about enabling this feature for your Sourcegraph instance.

#### Complete sync vs incremental sync

A complete sync means a repository or user has done a repository-centric or user-centric syncing respectively, which presists the most accurate permissions from code hosts to Sourcegraph.

An incremental sync is in fact a side effect of a complete sync because a user may grant or lose access to repositories and we react to such changes as soon as we know to improve permissions accuracy.

### Provider-specific optimizations

Each provider can implement optimizations to improve sync performance - please refer to the relevant provider documentation on this page for more details. For example, [the GitHub provider has support for using webhooks to improve sync speed](#faster-permissions-syncing-via-github-webhooks).

#### Permissions caching

<span class="badge badge-experimental">Experimental</span> <span class="badge badge-note">Sourcegraph 3.31+</span>

Some permissions providers in Sourcegraph can leverage caching mechanisms to reduce the number of API calls used when syncing permissions. This can significantly reduce the amount of time it takes to perform a full permissions sync due to reduced instances of being rate limited by the code host, and is useful for code hosts with very large numbers of users and repositories.

To see if your provider supports permissions caching, please refer to the relevant provider documentation on this page. For example, [the GitHub provider supports teams and organizations permissions caching](#teams-and-organizations-permissions-caching).

Note that this can mean that permissions can be out of date. To configure caching behaviour, please refer to the relevant provider documentation on this page. To force a bypass of caches during a sync, you can manually queue users or repositories for sync with the `invalidateCaches` options via the Sourcegraph GraphQL API:

```gql
mutation {
  scheduleUserPermissionsSync(user: "userid", options: {invalidateCaches: true}) {
    alwaysNil
  }
}
```

<br />

## Explicit permissions API

Sourcegraph exposes a GraphQL API to explicitly set repository permissions as an alternative to the code-host-specific repository permissions sync mechanisms.

To enable the permissions API, add the following to the [site configuration](../config/site_config.md):

```json
"permissions.userMapping": {
    "enabled": true,
    "bindID": "email"
}
```

The `bindID` value specifies how to uniquely identify users when setting permissions:

- `email`: You can [set permissions](#settings-repository-permissions-for-users) for users by specifying their email addresses (which must be verified emails associated with their Sourcegraph user account).
- `username`: You can [set permissions](#settings-repository-permissions-for-users) for users by specifying their Sourcegraph usernames.

If the permissions API is enabled, all other repository permissions mechanisms are disabled.

After you enable the permissions API, you must [set permissions](#settings-repository-permissions-for-users) to allow users to view repositories. (Site admins bypass all permissions checks and can always view all repositories.) 

> If you were previously using [background permissions syncing](#background-permissions-syncing), then those permissions are used as the initial state. Otherwise, the initial state is for all repositories to have an empty set of authorized users, so users will not be able to view any repositories.

### Setting repository permissions for users

Setting the permissions for a repository can be accomplished with 2 [GraphQL API](../../api/graphql.md) calls.

First, obtain the ID of the repository from its name:

```graphql
query {
  repository(name: "github.com/owner/repo") {
    id
  }
}
```

Next, set the list of users allowed to view the repository:

```graphql
mutation {
  setRepositoryPermissionsForUsers(
    repository: "<repo ID>", 
    userPermissions: [
      { bindID: "user@example.com" }
    ]) {
    alwaysNil
  }
}
```

Now, only the users specified in the `userPermissions` parameter will be allowed to view the repository. Sourcegraph automatically enforces these permissions for all operations. (Site admins bypass all permissions checks and can always view all repositories.)

You can call `setRepositoryPermissionsForUsers` repeatedly to set permissions for each repository, and whenever you want to change the list of authorized users.

### Listing a user's authorized repositories

You may query the set of repositories visible to a particular user with the `authorizedUserRepositories` [GraphQL API](../../api/graphql.md) mutation, which accepts a `username` or `email` parameter to specify the user:

```graphql
query {
  authorizedUserRepositories(email: "user@example.com", first: 100) {
    nodes {
      name
    }
    totalCount
  }
}
```

## Permissions for multiple code hosts

When integrating multiple code hosts with Sourcegraph, repository permissions typically need to be inherited and enforced across those respective code hosts and repositories. The steps below will walk you through configuring and enforcing repository permissions on a per-user basis across all of the code hosts and repos connected to Sourcegraph.

### Using the explicit permissions API

The recommended approach for inheriting permissions across multiple code hosts is via the [Explicit Permissions API](#explicit-permissions-api). The workaround provided in below is recommended only if using the Explicit Permissions API is not feasible.

### Using GitHub Enterprise and GitHub.com

> NOTE: This workaround is currently only verified to work when connecting both GitHub Enterprise and Github.com OAuth applications. For other code hosts and configuration options, please reach out to us.

Setup and add GitHub Enterprise (GHE) and GitHub.com (GHC) using our standard [GitHub integration](../external_service/github.md).

**Configure GitHub Enterprise SSO:**

1. Add GHE repos.
2. Configure auth for GHE using [OAuth](../auth/index.md#github).

    > NOTE: Ensure that the `allowSignup` field is set to `true`. This will ensure that users signing in via GHE will have a new user account created on Sourcegraph.

3. Add the [authorization field](#github) to the GHE code host connection (this is what enforces repository permissions).
4. Test that the GitHub Enterprise OAuth is working correctly (users should be able to sign into Sourcegraph using their GitHub Enterprise credentials).

**Configure Github.com SSO:**

1. Add GHC repos.
2. Configure auth for GHC using [OAuth](../auth/index.md#github).
  
    > NOTE: Ensure that the `allowSignup` field is set to `false`. This will ensure that users signing in via GHC will not have a new user account created on Sourcegraph.

3. Add the [authorization field](#github) to the GHC code host connection (this is what enforces repository permissions).
4. Test that the GitHub.com OAuth is working correctly (users should be able to sign into Sourcegraph using their GitHub.com credentials).

#### User sign in flow

When multiple code hosts/authentication providers are connected to Sourcegraph, a specific sign-in flow needs to be utilized when users are creating an account and signing into Sourcegraph for the first time.

1. Sign in to Sourcegraph using the GitHub Enterprise button
2. Once signed in, sign out and return to the sign in page
3. On the sign in page, sign in again using the Github.com button
4. Once signed in via Github.com, users should now have access to repositories on both code hosts and have all repository permissions enforced.

> NOTE: These steps are not required at every sign in - only during the initial account creation.
