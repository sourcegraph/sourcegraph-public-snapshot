# Repository permissions

Sourcegraph can be configured to enforce repository permissions from code hosts.

Currently, GitLab permissions are supported. Check the [roadmap](../../dev/roadmap.md) for plans to
support other code hosts. If your desired code host is not yet on the roadmap, please [open a
feature request](https://github.com/sourcegraph/sourcegraph/issues/new?template=feature_request.md).

## GitHub

To enable GitHub permissions,

1. [Create a GitHub OAuth application.](https://developer.github.com/apps/building-oauth-apps/creating-an-oauth-app/)

1. [Add a `"github"` entry to `auth.providers`](../auth.md#github) in the critical site configuration (via the management console).

1. [Add or edit your GitHub external service
   configuration](../../integration/github.md#syncing-github-repositories) to contain the
   `authorization` field.

    ```json
    {
       "url": "https://github.com",
       "token": "$PERSONAL_ACCESS_TOKEN",
       "authorization": {
         "ttl": "3h"
       }
    }
    ```

    See the [GitHub connection documentation](../../admin/site_config/all.md#githubconnection-object) for the meaning of specific fields.


## GitLab

Enabling GitLab repository permissions on Sourcegraph requires the following:

* A GitLab access token with `api` and `sudo` scope.
* Single sign-on (SSO) configured for both Sourcegraph and GitLab. The same SSO provider should be
  used to authenticate to both.

Sourcegraph uses the above to associated a GitLab user account with each Sourcegraph user. It then
uses the GitLab API to determine the set of repositories that are accessible to each user. Note that
Sourcegraph admin users will have access to **all** repositories on Sourcegraph regardless of what
permissions are associated with their GitLab user.

To enable GitLab permissions,

1. [Add or edit your GitLab external service
   configuration](../../integration/gitlab.md#syncing-gitlab-repositories) to contain the
   `authorization` field.

    ```json
    {
      "url": "$GITLAB_URL",
      "token": "$GITLAB_TOKEN",  // Token should have both `sudo` and `api` scope
      "authorization": {
        "authnProvider": {
          "configID": "$USER_SPECIFIED_AUTHENTICATION_ID",
          "type": "$AUTHENTICATION_TYPE",
          "gitlabProvider": "$GITLAB_PROVIDER"
        },
        "ttl": "1h"
      }
    }
    ```

    See the [GitLab configuration documentation](../../admin/site_config/all.md#gitlabconnection-object) for the meaning of specific fields.

1. Add an entry to `auth.providers` in the critical site configuration (via the management console).
    ```json
    {
      "auth.providers": [
        {
          "type": "$AUTHENTICATION_TYPE",
          "configID": "$USER_SPECIFIED_AUTHENTICATION_ID"
          ...
        }
      ],
    }
    ```

    Note that the `configID` and `type` fields in the GitLab `authorization.authnProvider` object
    must match the `configID` and `type` of exactly one element of `auth.providers`.

    See the [auth providers configuration documentation](../../admin/site_config/all.md#auth-providers-array) for the meaning of specific fields.
