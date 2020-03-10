# Bitbucket Server integration with Sourcegraph

You can use Sourcegraph with Git repositories hosted on [Bitbucket Server](https://www.atlassian.com/software/bitbucket/server) (and the [Bitbucket Data Center](https://www.atlassian.com/enterprise/data-center/bitbucket) deployment option).

Feature | Supported?
------- | ----------
[Repository syncing](../admin/external_service/bitbucket_server.md) | ✅
[Webhooks](../admin/external_service/bitbucket_server.md#webhooks) | ✅
[Repository permissions](../admin/external_service/bitbucket_server.md#repository-permissions) | ✅
[Sourcegraph Bitbucket Server plugin](#sourcegraph-bitbucket-server-plugin) | ✅
[Browser extension](#browser-extension) | ✅

## Repository syncing

Site admins can [add Bitbucket Server repositories to Sourcegraph](../admin/external_service/bitbucket_server.md).

## Repository permissions

Site admins can [configure Sourcegraph to respect Bitbucket Server's repository access permissions](../admin/external_service/bitbucket_server.md#repository-permissions).

## Sourcegraph Bitbucket Server Plugin

We recommend installing the [Sourcegraph Bitbucket Server plugin](https://github.com/sourcegraph/bitbucket-server-plugin/tree/master) so users don't need to install and configure the browser extension to get code intelligence when browsing code or reviewing pull requests on Bitbucket Server.

The plugin also enables **faster ACL permission syncing between Sourcegraph and Bitbucket Server** and adds **webhooks to Bitbucket Server**, which are used by and highly recommended for [Campaigns](../user/campaigns.md).

![Bitbucket Server code intelligence](https://storage.googleapis.com/sourcegraph-assets/bitbucket-code-intel-pr-short.gif)

### Installation and requirements

See the [bitbucket-server-plugin](https://github.com/sourcegraph/bitbucket-server-plugin) repository for instructions on how to install the plugin on your Bitbucket Server instance.

For the Bitbucket Server plugin to then communicate with the Sourcegraph instance, the Sourcegraph site configuration must be updated to include the [`corsOrigin` property](../admin/config/site_config.md) with the Bitbucket Server URL

```json
{
  // ...
  "corsOrigin": "https://my-bitbucket.example.com"
  // ...
}
```

### Webhooks

Once the plugin is installed, go to **Administration > Add-ons > Sourcegraph** to see a list of all configured webhooks and to create a new one.

Sourcegraph automatically creates a webhook for usage with [Campaigns](../user/campaigns.md) once the [`"plugin.webhooks"` property in the Bitbucket Server configuration](../admin/external_service/bitbucket_server.md) is configured in Sourcegraph.

### Experimental: faster ACL permissions fetching

The plugin also supports an experimental method of faster ACL permissions fetching that aims to improve search speed. You can enable this in the experimental section of the [site configuration](../admin/config/site_config.md):

```json
{
  // ...
  "experimentalFeatures": {
    "bitbucketServerFastPerm": "enabled"
  }
  // ...
}
```
The speed improvements are subtle and more noticeable for larger instances with thousands of repositories. This may remove the occasional slow search that has incurred the overhead of refreshing expired permissions information.

## Browser extension

The [Sourcegraph browser extension](browser_extension.md) supports Bitbucket Server. When installed in your web browser, it adds hover tooltips, go-to-definition, find-references, and code search to files and pull requests viewed on Bitbucket Server.

1. Install the [Sourcegraph browser extension](browser_extension.md).
1. [Configure the browser extension](browser_extension.md#configuring-the-sourcegraph-instance-to-use) to use your Sourcegraph instance.
1. To allow the browser extension to work on your Bitbucket Server instance:
    - Navigate to any page on Bitbucket Server.
    - Right-click the Sourcegraph icon in the browser extension toolbar.
    - Click "Enable Sourcegraph on this domain".
    - Click "Allow" in the permissions request popup.
1. Visit any file or pull request on Bitbucket Server. Hover over code or click the "View file" and "View repository" buttons.
