# Sourcegraph extensions - setup and usage instructions (alpha)

Sourcegraph is getting an extension API that makes easy to add new features and information to Sourcegraph, GitHub, and other code hosts and review tools that our browser extensions integrate with. The [Sourcegraph extension API](https://github.com/sourcegraph/sourcegraph-extension-api) allows extensions to provide hovers, definitions, references, line decorations, buttons, menu items, and other similar contributions. For more information, see the open-source [sourcegraph-extension-api repository](https://github.com/sourcegraph/sourcegraph-extension-api).

Sourcegraph extensions are available in alpha on Sourcegraph.com, in Sourcegraph 2.11.2+, [Sourcegraph for Chrome](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack), and [Sourcegraph for Firefox](https://addons.mozilla.org/en-US/firefox/addon/sourcegraph/).

### Using Sourcegraph extensions on Sourcegraph

After following the steps above, Sourcegraph extensions are enabled, and a new **Extensions** link is visible in the top navigation bar. On this new extensions page, you can choose which extensions to add. To add an extension, click **Add** and then select your username.

In this alpha release, we recommend using the following 2 extensions.

> Note: These recommended extensions are fetched from Sourcegraph.com. See the steps below for each extension if you need to publish a local copy of an extension (because of a firewall preventing network access to Sourcegraph.com, or to run a version of the extension with custom modifications).

#### Codecov: Code coverage overlays from Codecov

Add the **Codecov** extension as described above. After adding it, visit a repository that contains Codecov code coverage data. If the repository is private, run the "Codecov: Set API token" command in the menu in the top navigation bar.

For more information, see the [sourcegraph-codecov repository](https://github.com/sourcegraph/sourcegraph-codecov).

#### Basic code intelligence: Definitions and references from text search

Add the **Basic code intelligence** extension as described above. After adding it, visit any file. Toggle between precise code intelligence and fuzzy (basic) code intelligence by pressing the **Precise** and **Fuzzy** buttons in the file header. The toggle affects the behavior of **Go to definition** and the reference results shown for **Find references**.

For more information, see the [sourcegraph-basic-code-intel repository](https://github.com/sourcegraph/sourcegraph-basic-code-intel).

#### Publishing a local copy of an extension

If your Sourcegraph instance is unable to connect to Sourcegraph.com (due to a firewall), or if you want to customize an extension, you need to publish a local copy to your Sourcegraph instance. To do so, follow these steps:

1.  Download and install the latest [src](https://github.com/sourcegraph/src-cli) (Sourcegraph CLI).
1.  [Configure and authenticate `src`](https://github.com/sourcegraph/src-cli#authentication) with the URL and an access token for your Sourcegraph instance.
1.  Clone the repository of the extension you want to publish: [sourcegraph-codecov](https://github.com/sourcegraph/sourcegraph-codecov) or [sourcegraph-basic-code-intel](https://github.com/sourcegraph/sourcegraph-basic-code-intel).
1.  Run `npm install` in the clone directory to install dependencies.
1.  Run `src extensions publish -extension-id $USER/$NAME` in the clone directory to publish the extension locally to your Sourcegraph instance. Replace `$USER` with your Sourcegraph username and `$NAME` with with `codecov` or `basic-code-intel`.
1.  Enable the extension for your Sourcegraph user account by clicking the **Extensions** link in the top navigation, clicking **Add**, and selecting your username.

### Private extension registry

On Sourcegraph Enterprise, you can publish Sourcegraph extensions on your Sourcegraph instance to a private extension registry and control which extensions are available. Sourcegraph extensions that are published to the private extension registry on your instance are only visible to other users on your instance.

#### Inheritance of Sourcegraph extensions from Sourcegraph.com

Sourcegraph Core, Enterprise Starter, and Enterprise instances inherit extensions from Sourcegraph.com with [`extensions.remoteRegistry`](../admin/site_config/index.md#remoteregistry) set to `"https://sourcegraph.com/.api/registry"`. The OSS version of Sourcegraph has no dependencies on external services, and its `extensions.remoteRegistry` defaults to `false`.

You can disable inheritance by setting [`extensions.remoteRegistry`](../admin/site_config/index.md#remoteregistry) to `false` in your site configuration:

```json
{
  "extensions": { "remoteRegistry": false }
}
```

#### Allowing specific extensions to be inherited from Sourcegraph.com

On Sourcegraph Enterprise, you can also set [`extensions.allowRemoteExtensions`](../admin/site_config/index.md#allowRemoteExtensions) so that only extensions in that list will be inherited from Sourcegraph.com:

```json
{
  "extensions": { "allowRemoteExtensions": ["chris/token-highlights"] }
}
```

### Next steps

- Sourcegraph extensions also work on Sourcegraph.com for public code, and on GitHub via [Sourcegraph for Chrome](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack) or [Sourcegraph for Firefox](https://addons.mozilla.org/en-US/firefox/addon/sourcegraph/). (Support for more code hosts is coming soon.) See the [sourcegraph-codecov README](https://github.com/sourcegraph/sourcegraph-codecov) for usage instructions.
- You'll be able to create and distribute your own extensions in the upcoming beta release. See the [sourcegraph-extension-api README](https://github.com/sourcegraph/sourcegraph-extension-api) for an preview of the capabilities and design for extensions.
