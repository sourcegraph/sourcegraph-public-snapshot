# Contributing to the Sourcegraph Cody VS Code Extension

1. Update your VS Code user setting to turn on debugging mode:

   ```json
   "cody.debug": true,
   ```

2. Run `pnpm install` from the **root** of this repository
3. Select `Launch Cody Extension` from the dropdown menu in the `RUN AND DEBUG` sidebar
   1. Remove `node_modules` from `root` and `client/cody` before re-running `pnpm install` if the start up failed
4. Refresh the extension to see updated changes

## File structure

- `src`: source code of the components for the extension
  host
- `webviews`: source code of the extension UI (webviews),
  build with Vite and rollup.js using the `vite.config.ts` file at directory
  root
- `dist`: build outputs from both webpack and vite
- `resources`: everything in this directory will be move to
  the ./dist directory automatically during build time for easy packaging
- `index.html`: the entry file that Vite looks for to build
  the webviews. The extension host reads this file at run time and replace
  the variables inside the file with webview specific uri and info

## Testing

1. Unit tests:

   ```shell
   $ cd client/cody
   $ pnpm test:unit
   ```

2. Integration tests:

   ```shell
   $ cd client/cody
   $ pnpm test:integration
   ```

## Release Process

Follow the steps below to package and publish the VS Code extension.

> NOTE: Since the extension has already been bundled during build, we will need to add the `--no-dependencies` flag to the `vsce` step during the packaging step to exclude the npm dependencies ([source](https://github.com/microsoft/vscode-vsce/issues/421#issuecomment-1038911725))

### Prerequisite

- Install the [VSCE CLI tool](https://code.visualstudio.com/api/working-with-extensions/publishing-extension#vsce)

### Release Steps

1. Increment the `version` in [`package.json`](package.json) and then run:

   ```shell
   $ cd client/cody
   $ pnpm run vsce:package
   ```

2. To try the packaged extension locally, disable any other installations of it and then run:

   ```sh
   $ cd client/cody
   $ code --install-extension dist/cody.vsix
   ```

3. When the changes look good, create and merge a pull request with the changes into `main` and push an update to `cody/release` branch to trigger an automated release:

   ```shell
   $ git push origin main:cody/release
   ```

   - This will trigger the build pipeline for publishing the extension using the `pnpm release` command
   - Publish release to [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai)
   - Publish release to [Open VSX Registry](https://open-vsx.org/extension/sourcegraph/cody-ai)

   4. Visit the [buildkite page for the vsce/release pipeline](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=cody%2Frelease) to watch the build process
