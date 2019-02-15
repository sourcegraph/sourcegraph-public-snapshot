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

Prerequisite: [Add GitLab as an authentication provider.](../auth.md#gitlab)

Then, [add or edit a GitLab external service](../external_service/gitlab.md#repository-syncing) and include the `authorization` field:

```json
{
  "url": "https://gitlab.com",
  "token": "$PERSONAL_ACCESS_TOKEN",
  "authorization": {
    "ttl": "3h"
  }
}
```
