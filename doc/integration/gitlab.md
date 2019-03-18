# GitLab integration with Sourcegraph

Sourcegraph integrates with GitLab.com, GitLab CE, and GitLab EE.

## GitLab configuration

Sourcegraph supports syncing repositories from GitLab.com, GitLab CE, and GitLab EE (version 10.0 and newer). To add repositories associated with a GitLab user:

1.  Go to the [site configuration editor](../admin/site_config/index.md).
2.  Press **Add GitLab projects**.
3.  Fill in the fields in the generated `gitlab` configuration option.

By default, it adds every GitLab project where the token's user is a member. To see other optional GitLab configuration settings, view [all settings](../admin/site_config/index.md) or press Ctrl+Space or Cmd+Space in the site configuration editor.

## Browser extension

The [Sourcegraph browser extension](browser_extension.md) supports GitLab. When installed in your web browser, it adds hover tooltips, go-to-definition, find-references, and code search to files and merge requests viewed on GitLab.

1.  Install the [Sourcegraph browser extension](browser_extension.md).
1.  [Configure the browser extension](browser_extension.md#configuring-the-sourcegraph-instance-to-use) to use your Sourcegraph instance (where you've added the `gitlab` site config property).

- You can also use [`https://sourcegraph.com`](https://sourcegraph.com) for public code from GitLab.com only.

1.  Click the Sourcegraph icon in the browser toolbar to open the settings page. If a permissions notice is displayed, click **Grant permissions** to allow the browser extension to work on your GitLab instance.
1.  Visit any file or merge request on GitLab. Hover over code or click the "View file" and "View repository" buttons.

![Sourcegraph for GitLab](https://cl.ly/7916fe1453a4/download/sourcegraph-for-gitLab.gif)
