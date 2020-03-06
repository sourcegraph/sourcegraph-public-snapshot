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

The plugin also has the optional functionality to enable **faster ACL permission syncing between Sourcegraph and Bitbucket Server** and can add **webhooks with configurable scope to Bitbucket Server**, which are used by and highly recommended for [Campaigns](../user/campaigns.md).

Additionally, activated [Sourcegraph extensions](../extensions) will be able to add information to Bitbucket Server code views and pull requests, such as test coverage data or trace/log information.

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

To configure a webhook on the Sourcegraph side, set the [`"plugin.webhooks"` property in the Bitbucket Server configuration](../admin/external_service/bitbucket_server.md#webhooks). Once that is configured Sourcegraph automatically makes sure in the background that a global webhook for usage with [Campaigns](../user/campaigns.md) is created on the Bitbucket Server instance.

### Experimental: faster ACL permissions fetching

The plugin also supports an optional, experimental method of faster ACL permissions fetching that aims to improve search speed.

You can enable this feature by setting the [`"plugin.permissions"` property in the Bitbucket Server configuration](../admin/external_service/bitbucket_server.md#repository-permissions) to `"enabled"`.

The speed improvements are subtle and more noticeable for larger instances with thousands of repositories. This may remove the occasional slow search that has incurred the overhead of refreshing expired permissions information.

### Technical Details

This section provides some technical insight into the Bitbucket Server plugin to make it easier to users to decide whether or not to install it on their Bitbucket Server instance.

You can find the full source code for the plugin at [github.com/sourcegraph/bitbucket-server-plugin](https://github.com/sourcegraph/bitbucket-server-plugin/).

#### Native Code Intelligence

The Bitbucket Server plugin provides **native code intelligence** (e.g.: **go-to-definition** and **find-references** functionality while browsing code or reviewing pull requests) without users having to install the [Sourcegraph browser extension](browser_extension.md).

It does that by fetching the required JavaScript code from the configured Sourcegraph instance and injecting it into the HTML that the Bitbucket Server instance serves. See the [`sourcegraph-bitbucket.js`](https://github.com/sourcegraph/bitbucket-server-plugin/blob/master/src/main/resources/js/sourcegraph-bitbucket.js) file for how it does that.

The code that's injected is the code of the [Sourcegraph browser extension](#browser-extension), hosted by your Sourcegraph instance, and adds the same code intelligence functionality to files and pull requests viewed on Bitbucket Server.

The code talks directly to the Sourcegraph instance that's configured in the Bitbucket Server plugin configuration. It doesn't add any more load to the Bitbucker Server instance.

If it failed to load or talk to the Sourcegraph instance messages are logged to the browser console.

When the Sourcegraph instance is updated to a newer version, the code that's loaded and injected might also be updated.

#### Webhooks

Bitbucket Server natively only [provides **per-repository** webhooks](https://confluence.atlassian.com/bitbucketserver/managing-webhooks-in-bitbucket-server-938025878.html).

Sourcegraph's Bitbucket Server adds support for webhooks with a **configurable scope**. Each webhook can be configured to listen to specific events **globally**, per **project** or per **repository**.

The motivation behind this added functionality is to more efficiently react to updates to Bitbucket Server pull requests when using [Campaigns](../user/campaigns.md) by requiring only a single webhook to receive events for hundreds or thousands of pull requests across projects and repositories.

The plugin adds a `/webhook` endpoint that accepts `GET`, `POST` and `DELETE` HTTP request to list, create and delete webhooks respectively. The full URL for this endpoint would be something like `https://your-bbs-instance.example.com/rest/sourcegraph-admin/1.0/webhook`. See the [webhooks README](https://github.com/sourcegraph/bitbucket-server-plugin/blob/master/src/main/java/com/sourcegraph/webhook/README.md) for detailed information on which payloads this endpoint accepts.

Once the plugin is installed it registers an asynchronous listener (see [`WebhookListener.java`](https://github.com/sourcegraph/bitbucket-server-plugin/blob/master/src/main/java/com/sourcegraph/webhook/WebhookListener.java)) that listens to `PullRequestEvent`s and `BuildStatusEvent`s. When an event is dispatched to the listener it checks whether a webhook has been registered for the scope and type of the event and if so, it enqueues the sending of a request to the webhook's endpoint in a thread pool. (See [`WebhookListener.handle`](https://github.com/sourcegraph/bitbucket-server-plugin/blob/master/src/main/java/com/sourcegraph/webhook/WebhookListener.java#L62-L76) and [`Dispatcher.java`](https://github.com/sourcegraph/bitbucket-server-plugin/blob/master/src/main/java/com/sourcegraph/webhook/Dispatcher.java).)

In order to persist the configured webhooks across restarts of the Bitbucket Server instance the plugin uses the [Active Objects ORM](https://developer.atlassian.com/server/framework/atlassian-sdk/active-objects/) of the Atlassian SDK. It registers two Active Objects: [`WebhookEntity` and `EventEntity`](https://github.com/sourcegraph/bitbucket-server-plugin/blob/94e4be96b57286429cc543205164586af03e9b9b/src/main/resources/atlassian-plugin.xml#L10-L14).

If Sourcegraph is configured to make use of the Bitbucket Server plugin webhooks (which is done by setting the [`"plugin.webhooks"` property in the Bitbucket Server configuration](../admin/external_service/bitbucket_server.md#webhooks)), it sends a request to the Bitbucket Server instance, every 30 seconds, to make sure that a webhook on the Bitbucket Server instance exists and points to the Sourcegraph instance.

#### Fast permissions fetching

When Sourcegraph is configured to use [Bitbucket Server's repository permissions](../../admin/repo/permissions.md#bitbucket_server) to control access to repositories on Sourcegraph, it needs to fetch permissions for each user.

The Bitbucket Server REST API only provides **paginated** endpoints to fetch either the list of repositories a given user has access to, or the list of users that have access to a given repository. Both endpoints return the **full representation of the entities**.

Since Sourcegraph is only interested in the IDs of either repositories or users (those are already synced to its database) the Bitbucket Server plugins adds two REST endpoints to provide more efficient endpoints:

- `/permissions/repositories?user=<USERNAME>&permission=<PERMISSION_LEVEL>`<br /> Returns **a list of repository IDs** the given `user` has access to on the given `permission` level.
- `/permissions/users?repository=<REPO>&permission=<PERMISSION_LEVEL>`<br /> Returns **a list of user IDs** that have access to the given `repository` on the given `permission` level.

The lists returned by both endpoints are encoded as [Roaring Bitmaps](https://roaringbitmap.org/).

Since only a single request is required to fetch the complete list of desired IDs and the response contains only IDs, encoded in an efficient binary format, these two endpoints make the fetching of permissions roughly **eight times faster** (measured on an instance with 10000 repositories) than using Bitbucket Server's REST API. (Bitbucket Server admins can further increase the performance of these endpoints by increasing the [`page.max.repositories` property in the Bitbucket Server configuration](https://confluence.atlassian.com/bitbucketserver/bitbucket-server-config-properties-776640155.html#BitbucketServerconfigproperties-Paging).)

The plugin uses `RepositoryService`, `UserManager`, `UserService` and `SecuritySevice` provided by the Atlassian SDK to fetch users or repositories from Bitbucket Server's database. You can see the full code for these two endpoints in [`PermissionRouter.java`](https://github.com/sourcegraph/bitbucket-server-plugin/blob/master/src/main/java/com/sourcegraph/permission/PermissionRouter.java)

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
