# Sourcegraph Cody for Visual Studio Code

[![vs marketplace](https://img.shields.io/vscode-marketplace/v/sourcegraph.cody.svg?label=vs%20marketplace)](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody)

## Installation

### From the Visual Studio Marketplace:

1. Install Sourcegraph from the [Visual Studio Marketplace](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody).
2. Launch VS Code, and click on the Cody icon in the VS Code Activity Bar to open the extension. Alternatively, you can launch the extension by pressing <kbd>Cmd</kbd>+<kbd>Shift</kbd>+<kbd>P</kbd> or <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>P</kbd> and searching for "Cody: Focus on chat view".

### From within VS Code:

1. Open the extensions tab on the left side of VS Code (<kbd>Cmd</kbd>+<kbd>Shift</kbd>+<kbd>X</kbd> or <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>X</kbd>).
2. Search for `Sourcegraph Cody` -> `Install` and `Reload`.

## Setting up the Cody extension

To set up the Cody extension, you will need to enter the URL of your Sourcegraph instance and an access token used for authentication.

1. Open the Cody chat view by clicking on the Cody icon in the sidebar.
2. If you are setting up Cody for the first time, you should see the terms of service.
3. To proceed, read the terms and click "I accept", if you accept the terms of service.
4. Aftewards, you should see the login screen, where you have to enter the URL of your Sourcegraph instance and an access token used for authentication.

> To generate an access token: go to your Sourcegraph instance, then in your account settings, navigate to `Access tokens`, click `Generate new token`, and copy the token.

5. Once you have filled out the form, click the Login button to login into Cody.

### Codebase

To enable codebase-aware answers, you have to set the codebase setting to let Cody know which repository you are working on in the current workspace. You can do that by opening the VSCode workspace settings, search for the "Cody: Codebase" setting, and enter the repository name as listed on your Sourcegraph instance. Setting the codebase will edit the `.vscode/settings.json` file in your repository, which you can then commit and save for future usage.

## Extension Settings

This extension contributes the following settings:

| Setting             | Description                                                                                                                  | Example                              |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ------------------------------------ |
| cody.enabled        | Enable or disable Cody.                                                                                                      | true/false                           |
| cody.serverEndpoint | URL of the Sourcegraph instance.                                                                                             | "https://<instance>.sourcegraph.com" |
| cody.codebase       | Name of the repository opened in the current workspace. Use the same repository name as listed on your Sourcegraph instance. | "github.com/sourcegraph/sourcegraph" |
| cody.useContext     | Context source for Cody. One of: "embeddings", "keyword", "blended", or "none".                                              | "embeddings"                         |

## Development

Please see the [CONTRIBUTING](./CONTRIBUTING.md) document if you are interested in contributing to our code base.
