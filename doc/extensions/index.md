# Sourcegraph extensions

> NOTE: Sourcegraph extensions are being deprecated with the upcoming Sourcegraph September release. [Learn more](./deprecation.md).

<p class="lead">
If you've used Sourcegraph before, you've used a Sourcegraph extension: Sourcegraph extensions add features like git blame, code coverage, trace/log information, link previews, and third-party tool integrations.
</p>

For example, try hovering over tokens or toggling Git blame on [`tuf_store.go`](https://sourcegraph.com/github.com/theupdateframework/notary/-/blob/server/storage/tuf_store.go). Or see a [demo video of code coverage overlays](https://www.youtube.com/watch?v=j1eWBa3rWH8). The [Sourcegraph.com extension registry](https://sourcegraph.com/extensions) lists all publicly available extensions.

Sourcegraph extensions are like editor extensions, but run anywhere you view code in your web browser. To get extensions on your code host and review tools, you need the [Chrome/Firefox Sourcegraph browser extension](../integration/browser_extension.md), or you must set up a [native integration](../integration/index.md) with your code host.

<div style="text-align:center;margin:20px 0;display:flex">
<a href="https://github.com/sourcegraph/sourcegraph-codecov" target="_blank"><img src="https://user-images.githubusercontent.com/1976/45107396-53d56880-b0ee-11e8-96e9-ca83e991101c.png" style="padding:15px"></a>
<a href="https://github.com/sourcegraph/sourcegraph-git-extras" target="_blank"><img src="https://user-images.githubusercontent.com/1976/47624533-f3a1e800-dada-11e8-81d9-3d4bd67fc08a.png" style="padding:15px"></a>

</div>

## Getting started
- [Using extensions](usage.md) – How to enable extensions on Sourcegraph.
- [Principles](principles.md)
- [Security and privacy](security.md) – All the ways we protect the privacy of your code.

## Administering extensions

Site admins can control instance-wide extensions settings and additional permissions features. See [Administering extensions on your Sourcegraph instance](../admin/extensions/index.md).
