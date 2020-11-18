# Sourcegraph extensions
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

<div class="getting-started">
   <a href="usage" class="btn">
   <span>Using extensions</span>
   </br>
   How to enable extensions on Sourcegraph.
  </a>

  <a href="https://sourcegraph.com/extensions" class="btn">
   <span>Extension registry</span>
   </br>
   Browse popular extensions like git extras, link previews, open-in-editor, Codecov, or Sonarqube.
  </a>

  <a href="security" class="btn">
   <span>Security and privacy</span>
   </br>
    All the ways we protect the privacy of your code. 
  </a>
</div>

## Administering extensions

Site admins can control instance-wide extensions settings and additional permissions features. See [Administering extensions on your Sourcegraph instance](../admin/extensions/index.md). 

## Authoring extensions

Have an improvement or an idea for a new Sourcegraph extension? Want to publish your extension publicly? See [**Authoring extensions**](authoring/index.md).
