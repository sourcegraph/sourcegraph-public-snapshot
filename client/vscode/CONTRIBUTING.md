# Contributing to Sourcegraph VS Code Extension

Thank you for your interest in contributing to Sourcegraph!
The goal of this document is to provide a high-level overview of how you can contribute to the Sourcegraph VS Code Extension.
Please refer to our [main CONTRIBUTING](https://github.com/sourcegraph/sourcegraph/blob/main/CONTRIBUTING.md) docs for general information regarding contributing to any Sourcegraph repository.

## Feedback

Your feedback is important to us and is greatly appreciated. Please do not hesitate to submit your ideas or suggestions about how we can improve the extension to our [VS Code Extension Feedback Discussion Thread](https://github.com/sourcegraph/sourcegraph/discussions/34821) on GitHub.

## Issues / Bugs

New issues and feature requests can be filed through our [issue tracker](https://github.com/sourcegraph/sourcegraph/issues/new?labels=team/integrations,vscode-extension&title=VSCode+Bug+report:+&projects=Integrations%20Project%20Board) using the `vscode-extension` & `team/integrations` label.

## Development

### Build and Run

#### Desktop and Web Version

1. `git clone` the [Sourcegraph repository](https://github.com/sourcegraph/sourcegraph)
1. Install dependencies via `yarn` for the Sourcegraph repository
1. Run `yarn generate` at the root directory to generate the required schemas
1. Make your changes to the files within the `client/vscode` directory with VS Code
1. Run `yarn build-vsce` to build or `yarn watch-vsce` to build and watch the tasks from the `root` directory
1. Select `Launch VS Code Extension` (`Launch VS Code Web Extension` for VS Code Web) from the dropdown menu in the `Run and Debug` sidebar view to see your changes

### Integration Tests

To perform integration tests:

1. In the Sourcegraph repository:
   1. `yarn`
   2. `yarn generate`
2. In the `client/vscode` directory:
   1. `yarn build:test` or `yarn watch:test`
   2. `yarn test-integration`

## GitPod

The Sourcegraph extension for VS Code also works on GitPod.

#### Desktop Version

To install this extension on GitPod Desktop:

1. Open the Extensions view by clicking on the Extensions icon in the Activity Bar on the side of your workspace
2. Search for `Sourcegraph`
3. Click `install` to install the Sourcegraph extension

#### Web Version

To run and test the web extension on GitPod Web (as well as VS Code and GitHub for the web), you must sideload the extension from your local machine as suggested in the following steps:

1. `git clone` the [Sourcegraph repository](https://github.com/sourcegraph/sourcegraph)
1. Run `yarn && yarn generate` at the root directory to install dependencies and generate the required schemas
1. Run `yarn build-vsce` at root to build the Sourcegraph VS Code extension for Web
1. Once the build has been completed, move to the extension’s directory: `cd client/vscode`
1. Start an HTTP server inside the extension’s path to host the extension locally: `npx serve --cors -l 8988`
1. In another terminal, generate a publicly-accessible URL from your locally running HTTP server using the localtunnel tool: `npx localtunnel -p 8988`
   1. A publicly-accessible URL will be generated for you in the output followed by “your url is:”
1. Copy and then open the newly generated URL in a browser and then select “Click to Continue”
1. Open the Command Palette in GitPod Web (a GitPod Workspace using the Open in Browser setting)
1. Select “Developer: Install Web Extension…”
1. Paste the newly generated URL in the input area and select Install
1. The extension is now installed

### Debugging

Please refer to the [How to Contribute](https://github.com/microsoft/vscode/wiki/How-to-Contribute#debugging) guide by VS Code for debugging tips.

## Questions

If you need guidance or have any questions regarding Sourcegraph or the extension in general, we invite you to connect with us on the [Sourcegraph Community Slack group](https://about.sourcegraph.com/community).

## Resources

- [Changelog](https://marketplace.visualstudio.com/items/sourcegraph.sourcegraph/changelog)
- [Code of Conduct](https://handbook.sourcegraph.com/company-info-and-process/community/code_of_conduct/)
- [Developing Sourcegraph guide](https://docs.sourcegraph.com/dev)
- [Developing the web clients](https://docs.sourcegraph.com/dev/background-information/web)
- [Feedback / Feature Request](https://github.com/sourcegraph/sourcegraph/discussions/34821)
- [Issue Tracker](https://github.com/sourcegraph/sourcegraph/labels/vscode-extension)
- [Report a bug](https://github.com/sourcegraph/sourcegraph/issues/new?labels=team/integrations,vscode-extension&title=VSCode+Bug+report:+&projects=Integrations%20Project%20Board)
- [Troubleshooting docs](https://docs.sourcegraph.com/admin/how-to/troubleshoot-sg-extension#vs-code-extension)

## License

Apache

## Release Process

The release process for the VS Code Extension for Sourcegraph is currently automated.

#### Release Steps

1. Make sure the main branch is up-to-date.
2. Make a commit in the following format: `$RELEASE_TYPE release vsce`
   - Replace $RELEASE_TYPE with one of the supporting types: `Major`, `minor`, and `patch`
3. Run `git push origin main:vsce/release` to trigger the build pipeline for releasing the extension.
   - The extension is built using the code from the release branch.
   - The package name and changelog will also be updated automatically.
   - The extension is published with the [auto-incremented](https://code.visualstudio.com/api/working-with-extensions/publishing-extension#autoincrementing-the-extension-version) version number by running the `vsce publish $RELEASE_TYPE` command provided by the [vsce CLI tool](https://code.visualstudio.com/api/working-with-extensions/publishing-extension#vsce)
4. Visit the [buildkite page for the vsce/release pipeline](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=vsce%2Frelease) to watch the build process
5. Once the build is completed with no error, you should see the new version being verified for the Sourcegraph extension in your [Marketplace Publisher Dashboard](https://marketplace.visualstudio.com/manage/publishers)
