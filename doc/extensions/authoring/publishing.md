# Publishing a Sourcegraph extension

Publishing a Sourcegraph extension is fast and easy. It involves building (compiling and bundling) one or more TypeScript files into a single JavaScript file.

When [setting up your development environment](development_environment.md), you'll already have:

1. Installed the [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli#installation)
1. [Configured `src` with an access token](https://github.com/sourcegraph/src-cli#authentication)

Now publish your extension by running:

```shell
src extensions publish
```

At this point, your extension has been built and sent to Sourcegraph. The output will include a link to a detail page where you can enable your extension and start using it.

## Private extensions

Any user can publish to the Sourcegraph.com extension registry, all Sourcegraph instances can use extensions from Sourcegraph.com, and all Sourcegraph.com extensions are visible to everyone. If you need to publish an extension privately, use a private extension registry on your own self-hosted Sourcegraph instance.

## Testing your extension

Your extension will need to be published to Sourcegraph.com or an Enterprise instance in order for it to be tested. While we are working on [publishing to a local instance for testing](https://github.com/sourcegraph/sourcegraph/issues/489), flagging your extension as a work-in-progress (WIP) is the best solution for now.

### WIP extensions

An extension with no published releases, or whose `package.json` extension manifest has a `"wip": true` property, is considered a work-in-progress (WIP) extension. WIP extensions:

- are sorted last on the list of extensions (unless the user has previously enabled the WIP extension);
- have a red "WIP" badge on the extension card in the list; and
- have a red "WIP" badge on the extension's page.

You can use WIP extensions for testing in-development extensions, as well as new versions of an existing extension.

## Refreshing extension code without republishing

When iterating on your extension, each code change requires republishing. You can avoid this by using the Parcel bundler's development server to override the URL for the extension file when publishing. This lets you see the latest changes in your browser by reloading the page, without republishing.

To set this up:

1. If the extension is in use, add a `wip-` prefix to the current name (so that you don't publish your work-in-progress changes to users that rely on the extension) and set `"wip": true"` in the extension manifest.

1. In a terminal window, run `npm run serve` in your extension's directory to run the Parcel dev server. Wait until it reports that it's listening on http://localhost:1234 (or another port number).

In another terminal window, run `src extensions publish -url http://localhost:1234/my-extension.js` (my-extension.js being the bundled JavaScript file in your dist directory). Sourcegraph will now fetch the extension code from the value of the `-url` argument.

1. Make a change inside `src`, then save. Your code will be re-bundled and a reload of the browser window will cause your changes to be loaded.

### When you are ready to publish

You've written the code, you've tested your extension, and now you're almost ready to publish. Lastly, you'll need to remove the WIP extension:

1. Open the WIP extension detail page
- Click the **Manage** tab
- Click the **Delete extension** button

Now publish the extension:

```
run src extensions publish
```
