# Bitbucket Server integration with Sourcegraph

Sourcegraph integrates with Bitbucket Server and Bitbucket Data Center.

> NOTE: Support for [Bitbucket Cloud](https://bitbucket.org) is coming soon. In the meantime, you can add Bitbucket Cloud repositories using [`repos.list`](../admin/repo/add_from_git_repository.md).

## Syncing Bitbucket Server repositories

Sourcegraph supports automatically syncing repositories from [Bitbucket Server](https://www.atlassian.com/software/bitbucket/server). To add repositories associated with a Bitbucket Server user:

1.  Go to the [site configuration editor](../admin/site_config/index.md).
2.  Press **Add Bitbucket Server projects**.
3.  Fill in the fields in the generated `bitbucketServer` configuration option.

Note: Bitbucket Server versions older than v5.5 will require specifying a less secure username+password combination, as those versions of Bitbucket Server do not support [personal access tokens](https://confluence.atlassian.com/bitbucketserver/personal-access-tokens-939515499.html).

#### Excluding personal repositories

Sourcegraph will be able to view and clone the repositories that the account you provide has access to. For example, if you provide a personal access token or username/password of an administrator Bitbucket Server account, Sourcegraph will be able to view and clone all repositories -- even personal ones.

We recommend that you create a new Bitbucket user account specifically for Sourcegraph (e.g. a "Sourcegraph Bot" account) and only give that account access to the repositories you wish to be viewable on Sourcegraph.

If you don't wish to create a separate Bitbucket user account just for Sourcegraph, you can specify the `"excludePersonalRepositories": true` option in the site config in the `bitbucketServer` object. With this enabled, Sourcegraph will exclude any personal repositories from being imported -- even if it has access to them.

#### How cloning works

Sourcegraph by default clones repositories from your Bitbucket Server via HTTP(s), using the access token or account credentials you provide in the configuration. SSH cloning is not used by default and as such you do not need to configure SSH cloning.

## Browser extension

The [Sourcegraph browser extension](browser_extension.md) supports Bitbucket Server. When installed in your web browser, it adds hover tooltips, go-to-definition, find-references, and code search to files and pull requests viewed on Bitbucket Server.

1.  Install the [Sourcegraph browser extension](browser_extension.md).
1.  [Configure the browser extension](browser_extension.md#configuring-the-sourcegraph-instance-to-use) to use your Sourcegraph instance (where you've added the `bitbucketServer` site config property).
1.  Click the Sourcegraph icon in the browser toolbar to open the settings page. If a permissions notice is displayed, click **Grant permissions** to allow the browser extension to work on your Bitbucket Server instance.
1.  Visit any file or pull request on Bitbucket Server. Hover over code or click the "View file" and "View repository" buttons.
