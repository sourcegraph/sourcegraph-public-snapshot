# GitLab

Site admins can sync Git repositories hosted on [GitLab](https://gitlab.com) (GitLab.com and GitLab CE/EE) with Sourcegraph so that users can search and navigate the repositories.

To connect GitLab to Sourcegraph:

1. Go to **Site admin > Manage repositories > Add repositories**
1. Select **GitLab**.
1. Configure the connection to GitLab using the action buttons above the text field, and additional fields can be added using <kbd>Cmd/Ctrl+Space</kbd> for auto-completion. See the [configuration documentation below](#configuration).
1. Press **Add repositories**.

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

## Internal rate limits

Internal rate limiting can be configured to limit the rate at which requests are made from Sourcegraph to GitLab. 

If enabled, the default rate is set at 36,000 per hour (10 per second) which can be configured via the `requestsPerHour` field (see below). If rate limiting is configured more than once for the same code host instance, the most restrictive limit will be used.

**NOTE** Internal rate limiting is only currently applied when synchronising [campaign](../../user/campaigns/index.md) changesets.

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

Sourcegraph requires an access token with `api` permissions (and `sudo`, if you are using an `external` identity provider type). These permissions are required for the following reasons:

We are actively collaborating with GitLab to improve our integration (e.g. the [Sourcegraph GitLab native integration](https://docs.gitlab.com/ee/integration/sourcegraph.html) and [better APIs for querying repository permissions](https://gitlab.com/gitlab-org/gitlab/issues/20532)).

| Request Type | Required GitLab scope | Sourcegraph usage |
|--------------|-----------------------|-------------------|
| [`GET /projects`](https://docs.gitlab.com/ee/api/projects.html#list-all-projects) | `api` | (1) For repository discovery when specifying `projectQuery` in code host configuration; (2) If using an `external` identity provider type, also used as a test query to ensure token is `sudo` (`sudo` not required otherwise). |
| [`GET /users`](https://docs.gitlab.com/ee/api/users.html#list-users) | `read_user` or `api` | If you are using an `external` identity provider type, used to discover user accounts. |
| [`GET /users/:id`](https://docs.gitlab.com/ee/api/users.html#single-user) | `read_user` or `api` | If using GitLab OAuth, used to fetch user metadata during the OAuth sign in process. |
| [`GET /projects/:id`](https://docs.gitlab.com/ee/api/projects.html#get-single-project) | `api` | (1) If using GitLab OAuth and repository permissions, used to determine if a user has access to a given _project_; (2) Used to query repository metadata (e.g. description) for display on Sourcegraph. |
| [`GET /projects/:id/repository/tree`](https://docs.gitlab.com/ee/api/repositories.html#list-repository-tree) | `api` | If using GitLab OAuth and repository permissions, used to verify a given user has access to the file contents of a repository within a project (i.e. does not merely have `Guest` permissions). |
| (future) write access | `api` | graph site-admins (only) to perform large-scale code refactors, with Sourcegraph issuing and managing the merge requests on GitLab repositories, company-wide. |

## Webhooks

The `webhooks` setting allows specifying the webhook secrets necessary to authenticate incoming webhook requests to `/.api/gitlab-webhooks`.

```json
"webhooks": [
  {"secret": "verylongrandomsecret"}
]
```

Using webhooks is optional, but if configured on GitLab, they allow faster campaign updates than the normal background syncing (i.e. polling) which `repo-updater` provides.

The following [webhook events](https://docs.gitlab.com/ee/user/project/integrations/webhooks.html) are currently used:

- Merge request events
- Pipeline events

Webhooks must be configured on each project that you wish to monitor on GitLab. To set up a project webhook on GitLab, go to the Webhook settings page of the project.

Fill in the **URL** displayed after saving the `webhooks` Sourcegraph setting mentioned above and make sure it is publicly available.

Add the **Secret Token** that was configured within Sourcegraph: in the example above, that's `verylongrandomsecret`.

Select the event types you want to monitor. At a minimum, these should include **merge request events** and **pipeline events**.

Click on **Enable SSL verification** if you have configured SSL with a valid certificate in your Sourcegraph instance.

Finally, select **Add webhook** to create the webhook.
