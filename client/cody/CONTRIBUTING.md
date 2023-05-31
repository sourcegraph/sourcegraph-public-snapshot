# Contributing to the Cody AI by Sourcegraph VS Code Extension

Thank you for taking the time to help improve our project. We are grateful for your interest in contributing to Cody.

## Code of Conduct

All interactions with the Sourcegraph open source project are governed by the [Sourcegraph Community Code of Conduct](https://handbook.sourcegraph.com/company-info-and-process/community/code_of_conduct/).

## Build and Run

At Sourcegraph we use [sg](https://docs.sourcegraph.com/dev/setup/quickstart), the Sourcegraph developer tool, to set up and manage our local development environment.

Prerequisites:

- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) (v2.18 or higher)
- [Node JS](https://nodejs.org/en/download) (see current recommended version in .nvmrc)
- [VS Code](https://code.visualstudio.com/download)

1. Update your VS Code user setting to turn on debugging mode:

   ```json
   "cody.debug.enable": true,
   "cody.debug.verbose": true
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

## Submitting a Pull Request

To facilitate a streamlined review process and enable us to properly evaluate your contributions, please follow these guidelines when submitting a Pull Request for your contributions to Cody:

1. Provide Sufficient Review Time

Please allow 5-7 business days for our team to review your PR. We value thorough and thoughtful reviews, and this time period enables us to provide meaningful feedback.

If you have not heard back from us after 7 days, please ping and remind the `sourcegraph/cody` team.

1. Include Relevant Tests

To expedite the review process and ensure the reliability of our codebase, we request that you include either integration test(s), unit test(s), or end-to-end test(s) relevant to your PR. Tests validate changes and maintain stability. They also help contributors and reviewers understand how new changes interact with existing features.

3. Provide Clear Setup and Reproduction Steps

Please provide clear and concise steps to set up or reproduce the changes in your PR. This information greatly helps reviewers and other contributors understand the context of your modifications.

4. Add the `cody/contributor` Label

To help us track and manage contributions from external contributors, please add the `cody/contributor` label to the related GitHub issue when submitting your PR. This label facilitates the identification and prioritization of external contributions.

Thank you for following these guidelines. Your cooperation enables us to maintain the quality and efficiency of our open-source project. We greatly appreciate your valuable contributions and look forward to reviewing your PR!

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

## Resources

- [VS Code UX Guidelines](https://code.visualstudio.com/api/ux-guidelines/webviews)
- [VS Code Webview UI Toolkit](https://microsoft.github.io/vscode-webview-ui-toolkit)
- [VS Code Icons - Codicons](https://microsoft.github.io/vscode-codicons/dist/codicon.html)
