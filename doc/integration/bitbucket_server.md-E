# Bitbucket Server integration with Sourcegraph

You can use Sourcegraph with Git repositories hosted on [Bitbucket Server](https://www.atlassian.com/software/bitbucket/server) (and the [Bitbucket Data Center](https://www.atlassian.com/enterprise/data-center/bitbucket) deployment option).

Feature | Supported?
------- | ----------
[Repository syncing](../admin/external_service/bitbucket_server.md) | ✅
[Repository permissions](../admin/external_service/bitbucket_server.md#repository-permissions) | ✅
[Sourcegraph native code intelligence plugin](#sourcegraph-native-code-intelligence-plugin) | ✅
[Browser extension](#browser-extension) | ✅

## Repository syncing

Site admins can [add Bitbucket Server repositories to Sourcegraph](../admin/external_service/bitbucket_server.md).

## Repository permissions

Site admins can [configure Sourcegraph to respect Bitbucket Server's repository access permissions](../admin/external_service/bitbucket_server.md#repository-permissions).

## Sourcegraph native code intelligence plugin

![Bitbucket Server code intelligence](https://storage.googleapis.com/sourcegraph-assets/bitbucket-code-intel-pr-short.gif)

We recommend installing the [Sourcegraph Bitbucket Server plugin](https://github.com/sourcegraph/bitbucket-server-plugin/tree/master) so each user doesn't need to install and configure the browser extension.

This involves adding a new add-on to your Bitbucket Server instance and see the [bitbucket-server-plugin](https://github.com/sourcegraph/bitbucket-server-plugin) repository for installation instructions and configuration settings.

> NOTE: For the Bitbucket Server plugin to communicate with the Sourcegraph instance, the Sourcegraph site configuration must be updated to add the Bitbucket Server URL to the [`corsOrigin` property](../admin/config/site_config.md)

```json
{
  // ...
  "corsOrigin":
    "https://my-bitbucket.example.com"
  // ...
}
```

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
