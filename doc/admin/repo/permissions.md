# Repository permissions

Sourcegraph can be configured to enforce repository permissions from code hosts.

Currently, GitHub, GitHub Enterprise, and GitLab permissions are supported. Check the [roadmap](../../dev/roadmap.md) for plans to
support other code hosts. If your desired code host is not yet on the roadmap, please [open a
feature request](https://github.com/sourcegraph/sourcegraph/issues/new?template=feature_request.md).

## GitHub

Prerequisite: [Add GitHub as an authentication provider.](../auth.md#github)

Then, [add or edit a GitHub external service](../external_service/github.md#repository-syncing) and include the `authorization` field:

```json
{
  "url": "https://github.com",
  "token": "$PERSONAL_ACCESS_TOKEN",
  "authorization": {
    "ttl": "3h"
  }
}
```

## GitLab

GitLab permissions can be configured in three ways:

1. Set up GitLab as an OAuth sign-on provider for Sourcegraph (recommended)
2. Use a GitLab sudo-level personal access token in conjunction with another SSO provider
   (recommended only if the first option is not possible)
3. Assume username equivalency between Sourcegraph and GitLab (warning: this is generally unsafe and
   should only be used if you are using strictly `http-header` authentication).

### OAuth application

Prerequisite: [Add GitLab as an authentication provider.](../auth.md#gitlab)

Then, [add or edit a GitLab external service](../external_service/gitlab.md#repository-syncing) and include the `authorization` field:

```json
{
  "url": "https://gitlab.com",
  "token": "$PERSONAL_ACCESS_TOKEN",
  "authorization": {
    "identityProvider": {
      "type": "oauth"
    },
    "ttl": "3h"
  }
}
```

### Sudo access token

Prerequisite: Add the [SAML](../auth.md#saml) or [OpenID Connect](../auth.md#openid-connect)
authentication provider you use to sign into GitLab.

Then, [add or edit a GitLab external service](../external_service/gitlab.md#repository-syncing) and include the `authorization` field:

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
    },
    "ttl": "3h"
  }
}
```

`$AUTH_PROVIDER_ID` and `$AUTH_PROVIDER_TYPE` identify the authentication provider to use and should
match the fields specified in the authentication provider config
(`auth.providers`). `$AUTH_PROVIDER_GITLAB_ID` should match the `identities.provider` returned by
[the admin GitLab Users API endpoint](https://docs.gitlab.com/ee/api/users.html#for-admins).

### Username

Prerequisite: Ensure that `http-header` is the _only_ authentication provider type configured for
Sourcegraph. If this is not the case, then it will be possible for users to escalate privileges,
because Sourcegraph usernames are mutable.

[Add or edit a GitLab external service](../external_service/gitlab.md#repository-syncing) and include the `authorization` field:

```json
{
  "url": "https://gitlab.com",
  "token": "$PERSONAL_ACCESS_TOKEN",
  "authorization": {
    "identityProvider": {
      "type": "username"
    },
    "ttl": "3h"
  }
}
```
