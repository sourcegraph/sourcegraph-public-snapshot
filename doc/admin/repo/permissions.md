# Repository permissions

Sourcegraph can be configured to enforce repository permissions from code hosts. The currently supported methods are:

- [GitHub / GitHub Enterprise](#github)
- [GitLab](#gitlab)
- [Bitbucket Server / Bitbucket Data Center](#bitbucket-server)
- [Unified SSO](https://unknwon.io/posts/200915_setup-sourcegraph-gitlab-keycloak/)
- [Explicit permissions API](#explicit-permissions-api)

For most supported repository permissions enforcement methods, Sourcegraph [syncs permissions in the background](#background-permissions-syncing).

If the Sourcegraph instance is configured to sync repositories from multiple code hosts, setting up permissions for each code host will make repository permissions apply holistically on Sourcegraph, so long as users log in from each code host - [learn more](#permissions-for-multiple-code-hosts).

> NOTE: Site admin users bypass all permission checks and have access to every repository on Sourcegraph.

<span class="virtual-br"></span>

> WARNING: It can take some time to complete [backgroung mirroring of repository permissions](#background-permissions-syncing) from a code host. [Learn more](#permissions-sync-duration).

<span class="virtual-br"></span>

> NOTE: If your desired code host is not yet supported, please [open a feature request](https://github.com/sourcegraph/sourcegraph/issues/new?template=feature_request.md).

<br />

## GitHub

Prerequisite: [Add GitHub as an authentication provider](../auth/index.md#github).

Then, [add or edit a GitHub connection](../external_service/github.md) and include the `authorization` field:

```json
{
  // The GitHub URL used to set up the GitHub authentication provider must match this URL.
  "url": "https://github.com",
  "token": "$PERSONAL_ACCESS_TOKEN",
  "authorization": {}
}
```

A [token that has the prerequisite scopes](../external_service/github.md#github-api-token-and-access) and both read and write access to all relevant repositories is required in order to list collaborators for each repository to perform a [complete sync](#complete-sync-vs-incremental-sync).

> NOTE: Both read and write access to the associated repos for permissions syncing are strongly suggested due to GitHub's token scope requirements. Without write permissions, sync will rely only on [user-centric sync](#background-permissions-syncing) and continue working as expected, though Sourcegraph may have out-of-date permissions more frequently.

<span class="virtual-br"></span>

> WARNING: It can take some time to complete [backgroung mirroring of repository permissions](#background-permissions-syncing) from a code host. [Learn more](#permissions-sync-duration).

### Trigger permissions sync from GitHub webhooks

<span class="badge badge-note">Sourcegraph 3.22+</span>

Sourcegraph can improve how up to date synchronized permissions stay by initiating syncs when receiving webhooks from GitHub for events related to user and repo permissions - [learn more about webhooks and permissions sync](#triggering-syncs-with-webhooks).

To set up webhooks, follow the guide in the [GitHub Code Host Docs](../external_service/github.md#webhooks). These events will enqueue permissions syncs for the repositories or users mentioned, meaning things like publicising / privatising repos, or adding collaborators will be reflected in your Sourcegraph searches more quickly. For this to work the user must have logged in via the [GitHub OAuth provider](../auth.md#github).

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

> NOTE: You should only try this if your GitHub setup makes extensive use of GitHub teams and organizations to distribute access to repositories and your number of `users * repos` is greater than 250,000 (which roughly corresponds to the scale at which [GitHub rate limits might become an issue](#permissions-sync-duration)).
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

A [token that has the prerequisite scopes](../external_service/github.md#github-api-token-and-access) and both read and write access to all relevant repositories and organizations is required to fetch repository and team permissions and team memberships is required and cache them across syncs.
Read-only access will *not* work with cached permissions sync, but will work with [regular GitHub permissions sync](#github) (with [some drawbacks](#github)).

When enabling this feature, we currently recommend a default `groupsCacheTTL` of `72` (hours, or 3 days). A lower value can be set if your teams and organizations change frequently, though the chosen value must be at least several hours for the cache to be leveraged in the event of being rate-limited (which takes [an hour to recover from](https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting)).

Cache invaldiation happens automatically on certain [webhook events](#trigger-permissions-sync-from-github-webhooks), so it is recommended that to configure webhook support when using cached permissions sync.
Caches can also be [manually invalidated](#permissions-caching) if necessary.

<br />

## GitLab

GitLab permissions can be configured in three ways:

1. Set up GitLab as an OAuth sign-on provider for Sourcegraph (recommended)
2. Use a GitLab administrator (sudo-level) personal access token in conjunction with another SSO provider
   (recommended only if the first option is not possible)
3. Assume username equivalency between Sourcegraph and GitLab (warning: this is generally unsafe and
   should only be used if you are using strictly `http-header` authentication).

> WARNING: It can take some time to complete [backgroung mirroring of repository permissions](#background-permissions-syncing) from a code host. [Learn more](#permissions-sync-duration).

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

## Bitbucket Server / Bitbucket Data Center

Enforcing Bitbucket Server / Bitbucket Data Center permissions can be configured via the `authorization` setting in its configuration.

> WARNING: It can take some time to complete [backgroung mirroring of repository permissions](#background-permissions-syncing) from a code host. [Learn more](#permissions-sync-duration).

### Prerequisites

1. You have the exact same user accounts, **with matching usernames**, in Sourcegraph and Bitbucket Server / Bitbucket Data Center. This can be accomplished by configuring an [external authentication provider](../auth/index.md) that mirrors user accounts from a central directory like LDAP or Active Directory. The same should be done on Bitbucket Server / Bitbucket Data Center with [external user directories](https://confluence.atlassian.com/bitbucketserver/external-user-directories-776640394.html).
1. Ensure you have set `auth.enableUsernameChanges` to **`false`** in the [site config](../config/site_config.md) to prevent users from changing their usernames and **escalating their privileges**.


### Setup

This section walks you through the process of setting up an *Application Link between Sourcegraph and Bitbucket Server / Bitbucket Data Center* and configuring the Sourcegraph Bitbucket Server / Bitbucket Data Center configuration with `authorization` settings. It assumes the above prerequisites are met.

As an admin user, go to the "Application Links" page. You can use the sidebar navigation in the admin dashboard, or go directly to [https://bitbucketserver.example.com/plugins/servlet/applinks/listApplicationLinks](https://bitbucketserver.example.com/plugins/servlet/applinks/listApplicationLinks).

<img src="https://imgur.com/Hg4bzOf.png" width="800">

Write Sourcegraph's external URL in the text area (e.g. `https://sourcegraph.example.com`) and click **Create new link**. Click **Continue** even if Bitbucket Server / Bitbucket Data Center warns you about the given URL not responding.

<img src="https://imgur.com/x6vFKIL.png" width="800">

Write `Sourcegraph` as the *Application Name* and select `Generic Application` as the *Application Type*. Leave everything else unset and click **Continue**.

<img src="https://imgur.com/161rbB9.png" width="800">

Now click the edit button in the `Sourcegraph` Application Link that you just created and select the `Incoming Authentication` panel.

<img src="https://imgur.com/sMGmzhH.png" width="800">

Generate a *Consumer Key* in your terminal with `echo sourcegraph$(openssl rand -hex 16)`. Copy this command's output and paste it in the *Consumer Key* field. Write `Sourcegraph` in the *Consumer Name* field.

<img src="https://imgur.com/1kK2Y5x.png" width="800">

Generate an RSA key pair in your terminal with `openssl genrsa -out sourcegraph.pem 4096 && openssl rsa -in sourcegraph.pem -pubout > sourcegraph.pub`. Copy the contents of `sourcegraph.pub` and paste them in the *Public Key* field.

<img src="https://imgur.com/YHm1uSr.png" width="800">

Scroll to the bottom and check the *Allow 2-Legged OAuth* checkbox, then write your admin account's username in the *Execute as* field and, lastly, check the *Allow user impersonation through 2-Legged OAuth* checkbox. Press **Save**.

<img src="https://imgur.com/1qxEAye.png" width="800">

Go to your Sourcegraph's *Manage repositories* page (i.e. `https://sourcegraph.example.com/site-admin/external-services`) and either edit or create a new *Bitbucket Server / Bitbucket Data Center* connection. Add the following settings:

```json
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

Copy the *Consumer Key* you generated before to the `oauth.consumerKey` field and the output of the command `base64 sourcegraph.pem | tr -d '\n'` to the `oauth.signingKey` field. Finally, **save the configuration**. You're done!

### Fast permission sync with Bitbucket Server plugin

By installing the [Bitbucket Server plugin](../../../integration/bitbucket_server.md), you can make use of the fast permission sync feature that allows using Bitbucket Server / Bitbucket Data Center permissions on larger instances.

<br />

## Background permissions syncing

<span class="badge badge-note">Sourcegraph 3.17+</span>

Sourcegraph syncs permissions in the background by default to better handle repository permissions at scale for [GitHub](#github), [GitLab](#gitlab), and [Bitbucket Server / Bitbucket Data Center](#bitbucket-server) code hosts. Rather than syncing a user's permissions when they log in and potentially blocking them from seeing search results, Sourcegraph syncs these permissions asynchronously in the background, opportunistically refreshing them in a timely manner.

Sourcegraph's background permissions syncing is a 2-way sync that combines data from both types of sync for each configured code host to populate the database tables Sourcegraph uses as its source-of-truth for what repositories a user has access to:

- **User-centric permissions syncs** update the complete list of repositories a user has access to, from the user's view. This typically uses authentication associated with the user where available.
- **Repository-centric permissions syncs** update the complete list of all users that have access to a repository, from the repository's view. This may require elevated permissions to request from a code host.

Both types of sync happen [repeatedly and continuously based on a variety of events and criteria](#permissions-sync-scheduling).

> NOTE: Failure cases for each type of sync is generally gracefully handled - unless the code host returns a non-error response, the result is not used to update permissions. This means that permissions may become outdated, but will usually not be deleted, if syncs fail.

Background permissions syncing enables:

1. More predictable load on the code host API due to maintaining a schedule of permission updates, though this can mean it [can take a long time for a sync to complete](#permissions-sync-duration).
2. Permissions are quickly synced for new repositories and users added to the Sourcegraph instance.
3. Users who sign up on the Sourcegraph instance can immediately get search results from some repositories they have access to on the code host as we begin to [incrementally sync](#complete-sync-vs-incremental-sync) their permissions.

> NOTE: Background permissions sync does not apply to the [explicit permissions API](#explicit-permissions-api).

### Complete sync vs incremental sync

The two types of sync, [user-centric and repository-centric](#background-permissions-syncing), means that **each user or repository** can be in one of two states:

- **Complete sync** means a user has completed user-centric permissions sync (or a repository has completed a repository-centric sync), which indicates the most accurate permissions from the code host has been presisted to Sourcegraph for the user (or vice versa for repositories).
- **Incremental sync** means a user has *not* yet completed a recent user-centric permissions sync, but has been recently granted some permissions from a repository-centric sync (or vice versa for repositories).
  - For example, if a user has *not* had a user-centric permissions sync, but has been granted permissions from one or more repository-centric syncs, the user will have only completed an incremental sync. In this state, a user might not have access to all repositories they should have access to, but will incrementally receive more access as repository-centric syncs complete.
  - It is possible to be in an incremental sync state where a user or repository has effectively completed a complete sync, and all access rules are aligned with what is in the code host - for example, if a user completed a complete sync and a single repository is added, the user will be granted access to that repository through incremental sync, so the user will have full access to everything the user should have access to despite being in an incremental sync state.

The state of permissions is [repeatedly and continuously updated in the background](#permissions-sync-scheduling).

#### Checking permissions sync state

The state of an user or repository's permissions can be checked in the UI by:

- For users: navigating to `/users/$USER/settings/permissions`
- For repositories: navigating to `/$CODEHOST/$REPO/-/settings/permissions`

The GraphQL API can also be used:

```gql
query {
  user(username: "user") {
    permissionsInfo {
      syncedAt
      updatedAt
    }
  }
  repository(name: "repository") {
    permissionsInfo {
      syncedAt
      updatedAt
    }
  }
}
```

In the GraphQL API, `syncedAt` indicates the last complete sync and `updatedAt` indicates the last incremental sync. If `syncedAt` is more recent than `updatedAt`, the user or repository is in a state of complete sync - [learn more](#complete-sync-vs-incremental-sync).

### Permissions sync scheduling

A variety of heuristics are used to determine when a user or a repository should be scheduled for a permissions sync (either [user-centric or repo-centric](#background-permissions-syncing) respectively) to ensure the permissions data Sourcegraph has is up to date. Scheduling of syncs happens repeatedly and continuously [in the background](#background-permissions-syncing) for both users and repositories.

For example, permissions syncs may be scheduled:

- When a user or repository is created
- When certain interactions happen, such as when a user logs in or a repository is visited
- When a user's or repository's permissions are deemed stale (i.e. some amount of time has passed since the last [complete sync](#complete-sync-vs-incremental-sync) for a user or repository)
- When a relevant [webhook is configured and received](#triggering-syncs-with-webhooks)
- When a [manual sync is scheduled](#manually-scheduling-a-sync)

When a sync is scheduled, it is added to a queue that is steadily processed to avoid overloading the code host - a sync [might not happen immediately](#permissions-sync-duration). Prioritization of permissions sync also happens to, for example, ensure users or repositories with no permissions get processed first.

#### Manually scheduling a sync

Permissions syncs are [typically scheduled automatically](#manually-scheduling-a-sync).
However, a sync can be manually scheduled through the UI in by site admins:

- For users: navigating to `/users/$USER/settings/permissions` and clicking "Schedule now"
- For repositories: navigating to `/$CODEHOST/$REPO/-/settings/permissions` and clicking "Schedule now"

The GraphQL API can also be used to schedule a sync:

```gql
mutation {
  scheduleUserPermissionsSync(user: "userid") {
    alwaysNil
  }
  scheduleRepositoryPermissionsSync(repository: "repositoryid") {
    alwaysNil
  }
}
```

### Permissions sync duration

When syncing permissions from code hosts with large numbers of users and repositories, it can take some time to complete mirroring repository permissions from a code host for every user and every repository, typically due to rate limits on a code host that limits how quickly Sourcegraph can query for repository permissions. This is generally not a problem for fresh installations, since admins should only make the instance available after it's ready, but for existing installations, active users may not see the repositories they expect in search results because the initial permissions syncing hasn't finished yet.

Since Sourcegraph [syncs permissions in the background](#background-permissions-syncing), while the initial sync for all repositories and users is happening, users will [gradually see more and more search results](#complete-sync-vs-incremental-sync) from repositories they have access to.

To further mitigate long sync times and API request load, Sourcegraph can also leverage [provider-specific optimizations](#provider-specific-optimizations).

### Provider-specific optimizations

Each provider can implement optimizations to improve [sync performance](#permissions-sync-duration) and [up-to-dateness of permissions](#triggering-syncs-with-webhooks) - please refer to the relevant provider documentation on this page for more details.

#### Triggering syncs with webhooks

Some permissions providers in Sourcegraph can leverage code host webhooks to help [trigger a permissions sync](#permissions-sync-scheduling) on relevant events, which helps ensure permissions data in Sourcegraph is up to date.

> NOTE: Webhook payloads is not used to populate permissions rules. All the prerequisite access for performing permissions sync for the relevant provider is still required.

To see if your provider supports triggering syncs with webhooks, please refer to the relevant provider documentation on this page. For example, [the GitHub provider supports webhook events](#trigger-permissions-sync-from-github-webhooks).

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

### Pending permissions

> NOTE: This section describes some very technical details behind background permissions sync. In most cases, you will not need to consult this section.

Pending permissions are created and stored when the repo permissions fetched from the code host contain users which are not yet having accounts on Sourcegraph. This information is stored for the purpose of immediate repo access for such users after joining Sourcegraph. During the process of user creation, `user_pending_permissions` is queried and if there are any permissions for the user being created, then these permissions are moved to `user_permissions` table and this user is ready to go in no time. Without pending permissions, new users will have to wait for their permissions sync to complete.

As soon as a new user is created on Sourcegraph, pending permissions (`repo_pending_permissions` and `user_pending_permissions`) are used to populate "ordinary" permissions (`repo_permissions` and `user_permissions` tables), after which the `user_pending_permissions` is cleared (however, `repo_pending_permissions` [is not](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@8e128dd3434b9e548176f8f1148ead3981458db9/-/blob/enterprise/internal/database/perms_store.go?L979-981) for performance concerns and user IDs are monotonically increasing and would never repeat).

#### External code host user to Sourcegraph user mapping

The [`user_pending_permissions` table](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/internal/database/schema.md#table-public-user-pending-permissions) has a `bind_id` column which is an ID of the user of the external code host, for example a username for Bitbucket Server, a GraphID for GitHub or a user ID for GitLab.

User pending permission is a composite entity comprising:
- `service_type` (e.g. `github`, `gitlab`, `bitbucketServer`)
- `service_id` (ID of the code host, e.g. `https://github.com/`, `https://gitlab.com/`)
- `permission` (access level, e.g. "read")
- `object_type` (type of what is enumerated in `object_ids_ints` column; for now it is `repos`)
- `bind_id`
 
All of which are included as a unique constraint. This entity is addressed in `user_ids_ints` column of [`repo_pending_permissions` table](#repo-pending-permissions) by `id`. Please see [this godoc](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@8e128dd3434b9e548176f8f1148ead3981458db9/-/blob/internal/authz/perms.go?L190-218=) for more information.

Overall, one entry of `user_pending_permissions` table means that _"There is a user with `bind_id` ID of this exact (`service_id`) external code host of this (`service_type`) type with such permissions for this (`object_ids_ints`) set of repos"_.

#### Repo pending permissions

[`repo_pending_permissions` table](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/internal/database/schema.md#table-public-repo-pending-permissions) maps `user_pending_permissions` entities to repo ID along with the permission type (currently only `read` is supported). Each row of the table maps a repo ID to an array of `user_pending_permissions` entries. It is designed as an inverted `user_pending_permissions` for more performant CRUD operations (see the DB migration description in [this commit](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/compare/0705aa790d31fcd51713f4432496cc6bbb49cce8...bc30ae1186cf7a491ef21a5c00cb2f565288dfbb#diff-660eca66a5fad95783448fa468b2ce2fR50)).

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

After you enable the permissions API, you must [set permissions](#settings-repository-permissions-for-users) to allow users to view repositories (site admins bypass all permissions checks and can always view all repositories).

> NOTE: If you were previously using [background permissions syncing](#background-permissions-syncing), e.g. using [GitHub permissions](#github), then those permissions are used as the initial state after enabling explicit permissions. Otherwise, the initial state is for all repositories to have an empty set of authorized users, so users will not be able to view any repositories.

> NOTE: If you're using Sourcegraph with multiple code hosts, it's not possible to use the explicit permissions API for some repositories and inherit code host permissions for others. (See [RFC 626: Permissions mechanisms in parallel](https://docs.google.com/document/d/1nWbmfM5clAH4pi_4tEt6zDtqN1-z1DuHlQ7A5KAijf8/edit#) for a design document about future support for this situation.)

### Setting a repository as unrestricted

Sometimes it can be useful to mark a repository as `unrestricted`, meaning that it is available to all Sourcegraph users. This can be done with the `setRepositoryPermissionsUnrestricted` mutation. Marking a repository as unrestricted will disregard any previously set explicit or synced permissions. Setting `unrestricted` back to `false` will restore the previous behaviour.

For example:

```graphql
mutation {
  setRepositoryPermissionsUnrestricted(repositories: ["A","B","C"], unrestricted: true)
}
```

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

<br />

## Permissions for multiple code hosts

If the Sourcegraph instance is configured to sync repositories from multiple code hosts (regardless of whether they are the same code host, e.g. `GitHub + GitHub` or `GitHub + GitLab`), Sourcegraph will enforce access to repositories from each code host with authorization enabled, so long as:

- users log in to Sourcegraph at least once from each code host's [authentication provider](../auth/index.md)
- users have the same primary email in Sourcegraph (under "User settings" > "Emails") as the code host at the time of the initial log in via that code host

To attach a user's Sourcegraph account to all relevant code host accounts, a specific sign-in flow needs to be utilized when users are creating an account and signing into Sourcegraph for the first time.

1. Sign in to Sourcegraph using the one of the code host's [authentication provider](../auth/index.md)
2. Once signed in, sign out and return to the sign in page
3. On the sign in page, sign in again using the next code host's [authentication provider](../auth/index.md)
4. Once repeated across all relevant code hosts, users should now have access to repositories on all code hosts and have all repository permissions enforced.

> NOTE: These steps are not required at every sign in - only during the initial account creation.
