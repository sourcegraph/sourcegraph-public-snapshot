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

## Using extensions in local development (sideloading)

When developing an extension, you can sideload it from your local development machine's Parcel dev server (instead of re-publishing it after each code change). This speeds up the development cycle and avoids breaking the published version of your extension. Your extension does not need to be published for you to be able to sideload it.

To set this up:

1. In your extension's directory, run `npm run serve` to run the Parcel dev server. Wait until it reports that it's listening
2. Reveal the **Ext ▲** debug menu by running the following JavaScript code in your browser's devtools console on a Sourcegraph page: `localStorage.debug=true;location.reload()`
3. In the **Ext ▲** debug menu, click **Sideload Extension -> Load Extension**
3. Enter the URL the Parcel dev server is listening on
4. Your extension should appear in the debug menu's "active extensions" list. If it doesn't, there may have been an error when activating your extension - check the debug console for error messages.

After doing this, the development cycle is as follows:

1. Make a change to your extension's code, then save the file.
2. Reload your browser window. (It will fetch the package.json and the newly compiled JavaScript bundle for your extension.)

When you're done, clear the sideload URL from the extensions debug menu.

*NOTE:* this workflow assumes that, when running the Parcel dev server, a symlink exists in the `dist/` directory pointing to your `package.json`. If you created the boilerplate for your extension using `npm init sourcegraph-extension`, this was automatically taken care of for you. Otherwise:
1. add the following npm script to your `package.json`:

    ```
    "symlink-package": "[ ! -f dist/package.json ] && (mkdir -p dist && pushd dist && ln -s ../package.json package.json && popd) || echo \"dist/package.json already symlinked\""
    ```
2. edit the `serve` npm script to run `symlink-package`:

    ```
    "serve": "npm run symlink package && parcel serve --no-hmr --out-file dist/your-extension.js src/your-extension.ts"
    ```

## WIP extensions

An extension with no published releases, or whose `package.json` extension manifest has a `"wip": true` property, is considered a work-in-progress (WIP) extension. WIP extensions:

- are sorted last on the list of extensions (unless the user has previously enabled the WIP extension);
- have a red "WIP" badge on the extension card in the list; and
- have a red "WIP" badge on the extension's page.

You can use WIP extensions for testing in-development extensions, as well as new versions of an existing extension.

Don't forget to delete your WIP extension when it's no longer needed (in the **Manage** tab on the extension's registry page).
