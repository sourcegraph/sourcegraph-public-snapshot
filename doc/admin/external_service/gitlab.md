# GitLab

Site admins can sync Git repositories hosted on [GitLab](https://gitlab.com) (GitLab.com and GitLab CE/EE) with Sourcegraph so that users can search and navigate the repositories.

To set this up, add GitLab as an external service to Sourcegraph:

1. Go to **User menu > Site admin**.
1. Open the **External services** page.
1. Press **+ Add external service**.
1. Enter a **Display name** (using "GitLab" is OK if you only have one GitLab instance).
1. In the **Kind** menu, select **GitLab**.
1. Configure the connection to GitLab in the JSON editor. Use Cmd/Ctrl+Space for completion, and [see configuration documentation below](#configuration).
1. Press **Add external service**.

## Supported versions

- GitLab.com
- GitLab CE/EE (v10.0 and newer)

## Repository syncing

By default, Sourcegraph adds every GitLab project where the token's user is a member. If you wish to limit the set of repositories that is indexed by Sourcegraph, the recommended way is to create a Sourcegraph "bot" user, which is just a normal user account with the desired access scope. For instance, if you wanted to add all internal GitLab projects to Sourcegraph, you could create a user "sourcegraph-bot" and give it no explicit access to any GitLab repositories.

### Troubleshooting

You can test your access token's permissions by running a cURL command against the GitLab API. This is the same API and the same project list used by Sourcegraph. 

Replace `$ACCESS_TOKEN` with the access token you are providing to Sourcegraph, and `$GITLAB_HOSTNAME` with your GitLab hostname:

```
curl -H 'Private-Token: $ACCESS_TOKEN' -XGET 'https://$GITLAB_HOSTNAME/api/v4/projects'
```

## Repository permissions

By default, all Sourcegraph users can view all repositories. To configure Sourcegraph to use
GitLab's per-user repository permissions, see "[Repository
permissions](../repo/permissions.md#gitlab)".

## User authentication

To configure GitLab as an authentication provider (which will enable sign-in via GitLab), see the
[authentication documentation](../auth.md#gitlab).

## Configuration

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/gitlab.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/gitlab) to see rendered content.</div>
