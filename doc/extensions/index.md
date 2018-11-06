# Sourcegraph extensions

Sourcegraph's extension API makes easy to add new features and information to Sourcegraph, GitHub, and other code hosts and review tools that our browser extensions integrate with. The Sourcegraph extension API allows extensions to provide hovers, definitions, references, line decorations, buttons, menu items, and other similar contributions. For more information, see [sourcegraph-extension-api](https://github.com/sourcegraph/sourcegraph/blob/master/packages/sourcegraph-extension-api/README.md).

Sourcegraph extensions are available in alpha on Sourcegraph.com, in Sourcegraph 2.11.2+, [Sourcegraph for Chrome](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack), and [Sourcegraph for Firefox](https://addons.mozilla.org/en-US/firefox/addon/sourcegraph/).

## Usage

To view all available extensions, click **User menu > Extensions** in the top navigation bar.

To enable an extension for yourself, visit its page and toggle the slider to on.

To enable an extension for all users (site admins only) or for all organization members, add to the `extensions` object in global or organization settings (as shown below).

```json
{
  ...,
  "extensions": {
    ...,
    "alice/myextension": true,
    ...
  },
  ...
}
```

## Publishing a local copy of an extension

If your Sourcegraph instance is unable to connect to Sourcegraph.com (due to a firewall), or if you want to customize an extension, you need to publish a local copy to your Sourcegraph instance. To do so, follow these steps:

1.  Download and install the latest [src](https://github.com/sourcegraph/src-cli) (Sourcegraph CLI).
1.  [Configure and authenticate `src`](https://github.com/sourcegraph/src-cli#authentication) with the URL and an access token for your Sourcegraph instance.
1.  Clone the repository of the extension you want to publish: [sourcegraph-codecov](https://github.com/sourcegraph/sourcegraph-codecov) or [sourcegraph-basic-code-intel](https://github.com/sourcegraph/sourcegraph-basic-code-intel).
1.  Run `npm install` in the clone directory to install dependencies.
1.  Run `src extensions publish -extension-id $USER/$NAME` in the clone directory to publish the extension locally to your Sourcegraph instance. Replace `$USER` with your Sourcegraph username and `$NAME` with with `codecov` or `basic-code-intel`.
1.  Enable the extension for your Sourcegraph user account by clicking on **User menu > Extensions** in the top navigation bar and then toggling the slider to on.

## Private extension registry

On Sourcegraph Enterprise, you can publish Sourcegraph extensions on your Sourcegraph instance to a private extension registry and control which extensions are available. Sourcegraph extensions that are published to the private extension registry on your instance are only visible to other users on your instance.

### Inheritance of Sourcegraph extensions from Sourcegraph.com

Sourcegraph Core, Enterprise Starter, and Enterprise instances inherit extensions from Sourcegraph.com with [`extensions.remoteRegistry`](../admin/site_config/all.md#remoteregistry) set to `"https://sourcegraph.com/.api/registry"`. The OSS version of Sourcegraph has no dependencies on external services, and its `extensions.remoteRegistry` defaults to `false`.

You can disable inheritance by setting [`extensions.remoteRegistry`](../admin/site_config/all.md#remoteregistry) to `false` in your site configuration:

```json
{
  "extensions": { "remoteRegistry": false }
}
```

### Allowing specific extensions to be inherited from Sourcegraph.com

On Sourcegraph Enterprise, you can also set [`extensions.allowRemoteExtensions`](../admin/site_config/all.md#alloweemoteextensions) so that only extensions in that list will be inherited from Sourcegraph.com:

```json
{
  "extensions": { "allowRemoteExtensions": ["chris/token-highlights"] }
}
```

## Next steps

- [Sourcegraph extension authoring documentation](https://github.com/sourcegraph/sourcegraph-extension-docs)
- [sourcegraph-extension-api](https://github.com/sourcegraph/sourcegraph/blob/master/packages/sourcegraph-extension-api/README.md)
- Sourcegraph extensions also work on Sourcegraph.com for public code, and on GitHub via [Sourcegraph for Chrome](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack) or [Sourcegraph for Firefox](https://addons.mozilla.org/en-US/firefox/addon/sourcegraph/). (Support for more code hosts is coming soon.) See the [sourcegraph-codecov README](https://github.com/sourcegraph/sourcegraph-codecov) for usage instructions.
