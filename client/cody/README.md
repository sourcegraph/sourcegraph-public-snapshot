<div align="center">
    <p>Cody: An AI-Powered Programming Assistant</p>
    <a href="https://docs.sourcegraph.com/cody">Docs</a> •
    <a href="https://discord.gg/s2qDtYGnAE">Discord</a> •
    <a href="https://twitter.com/sourcegraph">Twitter</a>
    <br /><br />
    <a href="https://srcgr.ph/discord">
        <img src="https://img.shields.io/discord/969688426372825169?color=5765F2" alt="Discord" />
    </a>
    <a href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai">
        <img src="https://img.shields.io/vscode-marketplace/v/sourcegraph.cody-ai.svg?label=vs%20marketplace" alt="VS Marketplace" />
    </a>
</div>

## Cody (experimental)

Cody is a coding assistant that answers code questions and writes code for you by reading your entire codebase and the code graph.

**Status:** experimental ([request access](https://about.sourcegraph.com/cody))

## Main features

- Answer questions about your codebase instantly
- Generate documentation and unit tests on demand
- Write code snippets and prototypes for you
- Translate comments and functions in your code between languages

## Installation

Here are the ways to install Cody in Visual Studio Code:

### In Visual Studio Code

1. Open the Extensions tab on the left side of VS Code (<kbd>Cmd</kbd>+<kbd>Shift</kbd>+<kbd>X</kbd> or <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>X</kbd>).
2. Search for `Sourcegraph Cody` and click Install.
3. Once installed, **reload** VS Code.
4. After reloading, click the Cody icon in the VS Code Activity Bar to open the extension.
   - Alternatively, you can launch the extension by pressing <kbd>Cmd</kbd>+<kbd>Shift</kbd>+<kbd>P</kbd> or <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>P</kbd> and searching for "Cody: Focus on chat view" and searching for "Cody: Focus on chat view".

### Through the Visual Studio Marketplace

1. Install Cody from the [Visual Studio Marketplace](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai).
2. After installing the plugin, click the Cody icon in the VS Code Activity Bar to open the extension.

## Setting up the Cody extension

To set up the Cody extension, you will need to enter the URL of your Sourcegraph instance and an access token used for authentication.

1. Open the Cody chat view by clicking on the Cody icon in the sidebar.
2. If you are setting up Cody for the first time, you should see the terms of service.
3. To proceed, read the terms and click "I accept", if you accept the terms of service.
4. Aftewards, you should see the login screen, where you have to enter the URL of your Sourcegraph instance and an access token used for authentication.
5. Once you have filled out the login form, click the Login button to login to Cody.

### Generating a Sourcegraph access token

1. Go to your Sourcegraph instance.
2. In your account settings, navigate to `Access tokens`.
3. Click `Generate new token`.
4. Copy the token.

### Codebase

To enable codebase-aware answers, you have to set the codebase setting to let Cody know which repository you are working on in the current workspace. You can do that as follows:

1. Open the VS Code workspace settings by pressing <kbd>Cmd</kbd>, (or File > Preferences (Settings) on Windows & Linux).
2. Search for the "Cody: Codebase" setting.
3. Enter the repository name as listed on your Sourcegraph instance.
   1. For example: `github.com/sourcegraph/sourcegraph` without the `https` protocol

Setting the codebase will edit the `.vscode/settings.json` file in your repository, which you can then commit and save for future usage.

## Extension Settings

This extension contributes the following settings:

| Setting             | Description                                                                                                                  | Example                                                      |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------ |
| cody.enabled        | Enable or disable Cody.                                                                                                      | true/false                                                   |
| cody.serverEndpoint | URL of the Sourcegraph instance.                                                                                             | "https://example.sourcegraph.com"                            |
| cody.codebase       | Name of the repository opened in the current workspace. Use the same repository name as listed on your Sourcegraph instance. | "github.com/sourcegraph/sourcegraph"                         |
| cody.useContext     | Context source for Cody. One of: "embeddings", "keyword", "blended", or "none".                                              | "embeddings"                                                 |
| cody.customHeaders  | Takes object, where each value pair will be added to the request headers made to your Sourcegraph instance.                  | {"Cache-Control": "no-cache", "Proxy-Authenticate": "Basic"} |

## Development

Please see the [CONTRIBUTING](./CONTRIBUTING.md) document if you are interested in contributing to our code base.
