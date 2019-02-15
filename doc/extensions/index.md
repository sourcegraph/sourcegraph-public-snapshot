# Sourcegraph extensions

> Building an extension? See [**extension authoring documentation**](authoring/index.md).

Sourcegraph extensions enhance your code host, code reviews, and Sourcegraph itself by adding features such as:

- [Code intelligence](../user/code_intelligence/index.md) (go-to-definition, find references, hovers, etc.)
- Test coverage overlays
- Links to live traces, log output, and performance data for a line of code
- Git blame
- Usage examples for functions

When these features are built as Sourcegraph extensions, you get them everywhere you view and review code:

- Sourcegraph.com
- Your own Sourcegraph instance
- GitHub and GitHub Enterprise (with the [browser extension](../integration/browser_extension.md))
- GitLab (with the [browser extension](../integration/browser_extension.md))
- Bitbucket Server (with the [browser extension](../integration/browser_extension.md))
- Coming soon: editors, more code hosts, and more code review tools

Extensions can provide the following functionality:

- Go to definition
- Find references
- Hovers
- Line decorations and annotations
- Toolbar buttons
- Commands
- Search keywords

Screenshots of test coverage and Git blame extensions:

<div style="text-align:center;margin:20px 0;display:flex">
<a href="https://github.com/sourcegraph/sourcegraph-codecov" target="_blank"><img src="https://user-images.githubusercontent.com/1976/45107396-53d56880-b0ee-11e8-96e9-ca83e991101c.png" style="padding:15px"></a>
<a href="https://github.com/sourcegraph/sourcegraph-git-extras" target="_blank"><img src="https://user-images.githubusercontent.com/1976/47624533-f3a1e800-dada-11e8-81d9-3d4bd67fc08a.png" style="padding:15px"></a>
</div>

If you've used Sourcegraph before, you've used a Sourcegraph extension. For example, try hovering over tokens or toggling Git blame on [`tuf_store.go`](https://sourcegraph.com/github.com/theupdateframework/notary/-/blob/server/storage/tuf_store.go). Or see a [demo video of code coverage overlays](https://www.youtube.com/watch?v=j1eWBa3rWH8). These features (and many others) are provided by extensions. You could improve these features, or add new features, just by improving or creating extensions to Sourcegraph.

The [Sourcegraph.com extension registry](https://sourcegraph.com/extensions) lists all publicly available extensions.

## Usage

To view all available extensions on your Sourcegraph instance, click **User menu > Extensions** in the top navigation bar. (To see recommended extensions, click **Explore** in the top navigation bar.)

To enable/disable an extension for yourself, click **User menu > Extensions**, find the extension, and toggle the slider.

After enabling a Sourcegraph extension, it is immediately ready to use. Of course, some extensions only activate for certain files (e.g., the Python extension only adds code intelligence for `.py` files).

### On your code host

Install the [browser extension](../integration/browser_extension.md) and point it to your Sourcegraph instance (or Sourcegraph.com) in its options menu. It will consult your Sourcegraph user settings and activate the extensions you've enabled whenever you view code or diffs on your code host or review tool.

### For organizations

To enable/disable an extension for all organization members, add it to the `extensions` object in organization settings (as shown below).

```json
{
  ...,
  "extensions": {
    ...,
    "alice/myextension": true, // or false to disable
    ...
  },
  ...
}
```

### For all users

On a self-hosted Sourcegraph instance, add the same JSON above to global settings (in **Site admin > Global settings**).

## [Authoring](authoring/index.md)

Ready to create your own extension? See the [extension authoring documentation](authoring/index.md).

For inspiration:

- Check out the source code for existing extensions on the [Sourcegraph.com extension registry](https://sourcegraph.com/extensions). (The extension's repository is linked from most extensions' pages.)
- [Issues labeled `extension-request`](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+label%3Aextension-request)

## Administration

Site administrators can customize how Sourcegraph extensions are used on their instance, with options for:

- a private extension registry on their instance
- allowing only specific extensions to be enabled by users
- preventing users from enabling any extension from Sourcegraph.com

See "[Administration of Sourcegraph extensions and the extension registry](../admin/extensions/index.md)".

## Security and privacy

We designed Sourcegraph extensions with security and privacy in mind.

- Extensions do not send your code to Sourcegraph to operate. They run entirely on the client in your web browser.
- Extensions don't have direct access to private code. Extensions needing code access, such as to provide find-references in a project, must fetch code from the API of your code host or self-hosted Sourcegraph instance. This requires you to explicitly permit access (e.g., by creating a GitHub access token and configuring the extension to use it).
- Extensions run in isolation from your code host's web pages. They don't have direct DOM access (because they run in a Web Worker) and can only contribute actions and behavior allowed by the Sourcegraph extension API.
- Extensions are sandboxed by your web browser. Because *Sourcegraph extensions* run inside of the Sourcegraph for Chrome/Firefox *browser extension*, they are limited by the permissions you granted to the browser extension.
- Sourcegraph development is open source, so these claims are verifiable.

You can use Sourcegraph extensions without a Sourcegraph.com account or self-hosted Sourcegraph instance. To do so, just install the [browser extension](../integration/browser_extension.md). (Note: To use extensions other than the default set of language extensions, you currently do need an account or self-hosted instance. We plan to remove this limitation in Feb 2019.)

## More information

See "[Principles of extensibility for Sourcegraph](principles.md)" for or more information about why we built the Sourcegraph extension API---and how it's different from other attempts, such as LSP (Language Server Protocol).

See the [Sourcegraph roadmap](../dev/roadmap.md) for future plans related to extensions.

## Feedback

File bugs, feature requests, extension API questions, and all other issues on [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph/issues).
