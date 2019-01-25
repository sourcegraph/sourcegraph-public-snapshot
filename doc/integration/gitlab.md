# GitLab integration with Sourcegraph

Sourcegraph integrates with GitLab.com, GitLab CE, and GitLab EE.

## Syncing GitLab repositories

Sourcegraph supports syncing repositories from GitLab.com, GitLab CE, and GitLab EE (version 10.0 and newer).

- Add GitLab as an external service (in **Site admin > External services**, or in the site config JSON editor in Sourcegraph 2.x)

- Read the [GitLab configuration documentation](../admin/site_config/all.md#gitlabconnection-object) or press Ctrl+Space or Cmd+Space in the configuration editor.

By default, it adds every GitLab project where the token's user is a member. If you wish to limit the set of repositories that is indexed by Sourcegraph, the recommended way is to create a Sourcegraph "bot" user, which is just a normal user account with the desired access scope. For instance, if you wanted to add all internal GitLab projects to Sourcegraph, you could create a user "sourcegraph-bot" and give it no explicit access to any GitLab repositories.

### Debugging

You can test your access token's permissions by running a cURL command against the GitLab API. This is the same API and the same project list used by Sourcegraph. 

Replace `$ACCESS_TOKEN` with the access token you are providing to Sourcegraph, and `$GITLAB_HOSTNAME` with your GitLab hostname:

```
curl -H 'Private-Token: $ACCESS_TOKEN' -XGET 'https://$GITLAB_HOSTNAME/api/v4/projects'
```

## Authentication

To configure GitLab as an authentication provider (which will enable sign-in via GitLab), see the
[authentication documentation](../admin/auth.md#gitlab).


## Repository permissions

By default, all Sourcegraph users can view all repositories. To configure Sourcegraph to use
GitLab's per-user repository permissions, see "[Repository
permissions](../admin/repo/permissions.md#gitlab)".

## Browser extension

The [Sourcegraph browser extension](browser_extension.md) supports GitLab. When installed in your web browser, it adds hover tooltips, go-to-definition, find-references, and code search to files and merge requests viewed on GitLab.

1.  Install the [Sourcegraph browser extension](browser_extension.md).
1.  [Configure the browser extension](browser_extension.md#configuring-the-sourcegraph-instance-to-use) to use your Sourcegraph instance (where you've added the `gitlab` site config property).

- You can also use [`https://sourcegraph.com`](https://sourcegraph.com) for public code from GitLab.com only.

1.  Click the Sourcegraph icon in the browser toolbar to open the settings page. If a permissions notice is displayed, click **Grant permissions** to allow the browser extension to work on your GitLab instance.
1.  Visit any file or merge request on GitLab. Hover over code or click the "View file" and "View repository" buttons.

![Sourcegraph for GitLab](https://cl.ly/7916fe1453a4/download/sourcegraph-for-gitLab.gif)
