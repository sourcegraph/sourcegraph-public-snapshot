# Contributing to the Sourcegraph Cody VS Code Extension

## Getting started

1. Update your VS Code user setting to turn on debugging mode:

   ```json
   "cody.debug.enable": true,
   "cody.debug.verbose": true
   ```

2. Run `pnpm install` from the **root** of this repository
3. Run `scripts/download-rg.sh` from `client/cody` directory
4. Select `RUN AND DEBUG` from the sidebar > `Launch Cody Extension` in the top dropdown
   1. Remove `node_modules` from `root` and `client/cody` before re-running `pnpm install` if the start up failed
5. Refresh the extension to see updated changes

- To view dev tools: `cm + option + i` or `cm + shift + p > Toggle Developer Tools`
- To debug: `RUN AND DEBUG > ... elipsis > toggle 'Debug Console'`
- To view logs: `DEBUG CONSOLE`

## File structure

- `src`: source code of the components for the extension
  host
- `webviews`: source code of the extension UI (webviews),
  build with Vite and rollup.js using the `vite.config.ts` file at directory
  root
- `test/integration`: code for integration tests
- `test/e2e`: code for playwright UI tests
- `dist`: build outputs from both webpack and vite
- `resources`: everything in this directory will be move to
  the ./dist directory automatically during build time for easy packaging
- `index.html`: the entry file that Vite looks for to build
  the webviews. The extension host reads this file at run time and replace
  the variables inside the file with webview specific uri and info

## Testing

1. Unit tests:

   ```shell
   cd client/cody
   pnpm test:unit
   ```

2. Integration tests:

   ```shell
   cd client/cody
   pnpm test:integration
   ```

3. E2E tests:

   To run all the tests inside the `client/cody/test/e2e` directory:

   ```shell
   cd client/cody
   pnpm test:e2e
   ```

   To run test individually, pass in the name of the test by replacing $TEST_NAME below.

   ```sh
   pnpm test:e2e $TEST_NAME
   # Example: Run the inline-assist test only
   pnpm test:e2e inline-assist
   ```

   Example: Run the inline-assist test only

   ```sh
   pnpm test:e2e --debug
   # Example: Run the inline-assist test in debug mode
   pnpm test:e2e inline-assist --debug
   ```

## Release Process

Follow the steps below to package and publish the VS Code extension.

### Versioning

Starting from `0.2.0`, Cody is using:

- `major.EVEN_NUMBER.patch` for stable release versions
- `major.ODD_NUMBER.patch` for nightly pre-release versions

For example: 1.2._ for release and 1.3._ for pre-release.

### Prerequisite

- Install the [VSCE CLI tool](https://code.visualstudio.com/api/working-with-extensions/publishing-extension#vsce)

### Manual release steps

1. Increment the `version` in [`package.json`](package.json).
1. Update the contents of the change log in [`CHANGELOG.md](CHANGELOG.md).
1. Create PR to merge the changes into `main` ([here's an example](https://github.com/sourcegraph/sourcegraph/pull/54316)).
1. Once merged into `main`, push the changes onto the `cody/release` branch to start a release job:

- `git checkout main`
- `git pull`
- `git push origin main:cody/release`

The last part will trigger the build pipeline for publishing the extension using the `pnpm release` command

- Publish release to [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai)
- Publish release to [Open VSX Registry](https://open-vsx.org/extension/sourcegraph/cody-ai)
- Publish a pre-release version to [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai)
  - Create a [pre-release version with minor version bump](https://code.visualstudio.com/api/working-with-extensions/publishing-extension#prerelease-extensions) allow the Nightly build to patch the pre-release version at the correct version number through the [auto-incrementing the extension version feature from the VSCE API](https://code.visualstudio.com/api/working-with-extensions/publishing-extension#autoincrementing-the-extension-version)

## Testing the release locally

Before pushing a release, you can ensure that it meets the bar by building and testing it locally. To do that:

1. Create a release package:

   ```shell
   $ cd client/cody
   $ pnpm run vsce:package
   ```

> NOTE: Since the extension has already been bundled during build, we will need to add the `--no-dependencies` flag to the `vsce` step during the packaging step to exclude the npm dependencies ([source](https://github.com/microsoft/vscode-vsce/issues/421#issuecomment-1038911725))

2. To try the packaged extension locally, disable any other installations of it and then run:

   ```sh
   $ cd client/cody
   $ code --install-extension dist/cody.vsix
   ```

### Build Status

**For internal use only.**

Visit the following pages to follow the build status for:

- Stable: [Buildkite page for the cody/release pipeline](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=cody%2Frelease)
- Nightly: [Buildkite page for the cody nightly build](https://buildkite.com/sourcegraph/sourcegraph/settings/schedules/337676ef-c8a3-4977-a0d9-7990cb0916d0)

## Resources

- [VS Code Publishing Extensions](https://code.visualstudio.com/api/working-with-extensions/publishing-extension)
- [VS Code UX Guidelines](https://code.visualstudio.com/api/ux-guidelines/webviews)
- [VS Code Webview UI Toolkit](https://microsoft.github.io/vscode-webview-ui-toolkit)
- [VS Code Icons - Codicons](https://microsoft.github.io/vscode-codicons/dist/codicon.html)

## Debugging with dedicated Node DevTools

1. **Initialize the Build Watcher**: Run the following command from the monorepo root to start the build watcher:

```sh
pnpm --filter cody-ai run watch
```

2. **Launch the VSCode Extension Host**: Next, start the VSCode extension host by executing the command below from the monorepo root:

```sh
pnpm --filter cody-ai run start:debug
```

3. **Access the Chrome Inspector**: Open up your Google Chrome browser and navigate to `chrome://inspect/#devices`.
4. **Open Node DevTools**: Look for and click on the option that says "Open dedicated DevTools for Node".
5. **Specify the Debugging Endpoint**: At this point, DevTools aren't initialized yet. Therefore, you need to specify [the debugging endpoint](https://nodejs.org/en/docs/inspector/) `localhost:9333` (the port depends on the `--inspect-extensions` CLI flag used in the `start:debug` npm script)
6. **Start Debugging Like a PRO**: yay!

### More tips

1. To open the webviews developer tools: cmd+shift+p and select `Developer: Toggle Developer Tools`
2. To reload extension sources: cmd+shift+p and select `Developer: Reload Window`. If you have the watcher running it should be enough to get the latest changes to the extension host.
