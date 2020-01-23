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

There are three fields for configuring which projects are mirrored/synchronized:

- [`projects`](gitlab.md#configuration)<br>A list of projects in `{"name": "group/name"}` or `{"id": id}` format.
- [`projectQuery`](gitlab.md#configuration)<br>A list of strings with one pre-defined option (`none`), and/or an URL path and query that targets a GitLab API endpoint returning a list of projects.
- [`exclude`](gitlab.md#configuration)<br>A list of projects to exclude which takes precedence over the `projects`, and `projectQuery` fields. It has the same format as `projects`.

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
[authentication documentation](../auth/index.md#gitlab).

## Configuration

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/gitlab.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/gitlab) to see rendered content.</div>

## Native integration

To provide out-of-the-box code intelligence and navigation features to your users on GitLab, you will need to [configure your GitLab instance](https://docs.gitlab.com/ee/integration/sourcegraph.html).

The Sourcegraph instance's site admin must [update the `corsOrigin` site config property](../config/site_config.md) to allow the GitLab instance to communicate with the Sourcegraph instance. For example:

```json
{
  // ...
  "corsOrigin":
    "https://my-gitlab.example.com"
  // ...
}
```

## Access token scopes

Sourcegraph requires an access token with `api` permissions (and `sudo`, if you are using an `external` identity provider type). Below is an explanation for why we require this scope instead of e.g. just `read_user`, and what we may also use the access token for in the future.

We are actively collaborating with GitLab on multiple fronts to improve our integration with them (e.g. through our [GitLab native integration](https://docs.gitlab.com/ee/integration/sourcegraph.html) and [working torwards better APIs Sourcegraph could use for querying repository permissions](https://gitlab.com/gitlab-org/gitlab/issues/20532)), so if you have any feedback please let us know!

| Request Type | Required GitLab scope | Sourcegraph usage |
|--------------|-----------------------|-------------------|
| [`GET /projects`](https://docs.gitlab.com/ee/api/projects.html#list-all-projects) | `api` | (1) For repository discovery when specifying `projectQuery` in external service configuration; (2) If using an `external` identity provider type, also used as a test query to ensure token is `sudo` (`sudo` not required otherwise). |
| [`GET /users`](https://docs.gitlab.com/ee/api/users.html#list-users) | `read_user` or `api` | If you are using an `external` identity provider type, used to discover user accounts. |
| [`GET /users/:id`](https://docs.gitlab.com/ee/api/users.html#single-user) | `read_user` or `api` | If using GitLab OAuth, used to fetch user metadata during the OAuth sign in process. |
| [`GET /projects/:id`](https://docs.gitlab.com/ee/api/projects.html#get-single-project) | `api` | (1) If using GitLab OAuth and repository permissions, used to determine if a user has access to a given _project_; (2) Used to query repository metadata (e.g. description) for display on Sourcegraph. |
| [`GET /projects/:id/repository/tree`](https://docs.gitlab.com/ee/api/repositories.html#list-repository-tree) | `api` | If using GitLab OAuth and repository permissions, used to verify a given user has access to the file contents of a repository within a project (i.e. does not merely have `Guest` permissions). |

Sourcegraph in the future may do more with the provided access token as well, including:

- Enabling Sourcegraph site-admins to perform large-scale code refactors, with Sourcegraph issuing and managing the merge requests on GitLab repositories, company-wide.
- Using more efficient APIs to get repository and user permissions from GitLab more efficiently. We are actively working with GitLab to make this more efficient and use fewer requests, see https://gitlab.com/gitlab-org/gitlab/issues/20532
- Improving the GitLab native integration and Sourcegraph browser extension integration: https://docs.gitlab.com/ee/integration/sourcegraph.html
