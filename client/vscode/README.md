# Sourcegraph for Visual Studio Code

[![vs marketplace](https://img.shields.io/vscode-marketplace/v/sourcegraph.sourcegraph.svg?label=vs%20marketplace)](https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph) [![downloads](https://img.shields.io/vscode-marketplace/d/sourcegraph.sourcegraph.svg)](https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph)

![Search Gif](https://storage.googleapis.com/sourcegraph-assets/VS%20Marketplace/tableContainer2.gif)

Sourcegraph’s code search allows you to find & fix things fast across all your code.

Sourcegraph for VS Code allows you to search millions of open source repositories right from your VS Code IDE—for free. You can learn from helpful code examples, search best practices, and re-use code from millions of repositories across the open source universe.

Plus, with a free Sourcegraph Cloud account, you can sync your own private and public repositories and search all of your code in a single view in VS Code. Sourcegraph’s Code Intelligence feature provides fast, cross-repository navigation with “Go to definition” and “Find references” features, allowing you to understand new code quickly and find answers in your code across codebases of any size.

You can read more about Sourcegraph on our [website](https://about.sourcegraph.com/).

## Installation

### From the Visual Studio Marketplace:

1. Install Sourcegraph from the [Visual Studio Marketplace](https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph).
2. Launch VS Code, and click on the Sourcegraph (Wildcard) icon in the VS Code Activity Bar to open the Sourcegraph extension. Alternatively, you can launch the extension by pressing <kbd>Cmd</kbd>+<kbd>Shift</kbd>+<kbd>P</kbd> or <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>P</kbd> and searching for “Sourcegraph: Open search tab.”

### From within VS Code:

1. Open the extensions tab on the left side of VS Code (<kbd>Cmd</kbd>+<kbd>Shift</kbd>+<kbd>X</kbd> or <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>X</kbd>).
2. Search for `Sourcegraph` -> `Install` and `Reload`.

## Using the Sourcegraph extension

To get started and open the Sourcegraph extension, simply click the Sourcegraph (Wildcard) icon in the VS Code Activity Bar.

Sourcegraph functions like any search engine; simply type in your search query, and Sourcegraph will populate search results.

Sourcegraph offers 3 different ways to search:

1. [Literal search](https://learn.sourcegraph.com/how-to-search-code-with-sourcegraph-using-literal-patterns)
2. [Structural search](https://learn.sourcegraph.com/how-to-search-with-sourcegraph-using-structural-patterns)
3. [Regular expressions](https://learn.sourcegraph.com/how-to-search-with-sourcegraph-using-regular-expression-patterns)

Sourcegraph also accepts filters to narrow down search results, such as `repo`, `file`, and `lang`. Check out our search [cheat sheet](https://learn.sourcegraph.com/how-to-search-code-with-sourcegraph-a-cheat-sheet).

For example, you can search for "auth provider" in a Go repository with a search like this one:

```
repo:sourcegraph/sourcegraph lang:go auth provider
```

![Lang search gif](https://storage.googleapis.com/sourcegraph-assets/VS%20Marketplace/sourcegraph_search.gif)

## Adding and searching your own code

### Creating an account

In addition to searching open source code, you can create a Sourcegraph Cloud account to search your own private and public repositories. You can create an account and sync your repositories with the following steps:

1. Click the `Create an account` button in the sidebar of the Sourcegraph extension. You will be directed to sourcegraph.com in your browser.
2. Create an account using your email or connect directly to your code host.
3. Once you have created an account, navigate to Sourcegraph Cloud. Click on your profile icon in the navigation bar to go to `Your repositories`.
4. Click `Manage repositories`. From here, you can add your repositories to be synced to Sourcegraph.

### Connecting Sourcegraph Cloud account
Once you have repositories synced to Sourcegraph, you can generate an access token to connect your VS Code extension back to your Sourcegraph Cloud account.

1. Back in Sourcegraph Cloud, in your account settings, navigate to `Access tokens`, then click `Generate new token`.
2. Once you have generated a token, navigate back to the Sourcegraph extension. In the sidebar, under `Create an account`, click `Have an account?`.
3. Copy and paste the generated token from step 4 into the input field in the sidebar.
4. Alternatively, you can copy and paste the generated token from step 4 in this format: `“sourcegraph.accessToken": "e4234234123112312”` into your VS Code Setting by going to `Code` > `Preference` > `Settings` > Search for "Sourcegraph" > `Edit in settings.json`.
5. The Editor will be reloaded automatically to use the newly added token.

### Connecting to a private Sourcegraph instance 
1.  In Sourcegraph, in your account settings, navigate to `Access tokens`, then click `Generate new token`.
2. Once you have generated a token, navigate to your VS Code Settings, then navigate to "Extension settings".
3. Navigate to `Code preferences`, then click `Settings`.
4. Search for `Sourcegraph`, and enter the newly generated access token and Sourcegraph instance URL. 


## Keyboard Shortcuts:

| Description                                  | Mac                                          | Linux / Windows                               |
| -------------------------------------------- | -------------------------------------------- | --------------------------------------------- |
| Open Sourcegraph Search Tab/Search Selection | <kbd>Cmd</kbd>+<kbd>Shift</kbd>+<kbd>8</kbd> | <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>8</kbd> |
| Open File in Sourcegraph Cloud               | <kbd>Option</kbd>+<kbd>A</kbd>               | <kbd>Alt</kbd>+<kbd>A</kbd>                   |
| Search Selected Text in Sourcegraph Cloud    | <kbd>Option</kbd>+<kbd>S</kbd>               | <kbd>Alt</kbd>+<kbd>S</kbd>                   |

## Extension Settings

This extension contributes the following settings:

- `sourcegraph.url`: Specify your on-premises Sourcegraph instance here, if applicable. For example: `"sourcegraph.url": "https://sourcegraph.com"`
- `sourcegraph.accessToken`: The access token to query the Sourcegraph API. Required to use this extension with private instances.
- `sourcegraph.remoteUrlReplacements`: Object, where each `key` is replaced by `value` in the remote url.
- `sourcegraph.defaultBranch`: String to set the name of the default branch. Always open files in the default branch.

## Questions & Feedback

Please take a look at our [troubleshooting docs](https://docs.sourcegraph.com/admin/how-to/troubleshoot-sg-extension#vs-code-extension) for common issues in the VS Code extension. New issues and feature requests can be submitted at https://github.com/sourcegraph/sourcegraph-vscode/issues/new.

## Uninstallation

1.  Open the extensions tab on the left side of VS Code (<kbd>Cmd</kbd>+<kbd>Shift</kbd>+<kbd>X</kbd> or <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>X</kbd>).
2.  Search for `Sourcegraph` -> Gear icon -> `Uninstall` and `Reload`.

## Development

To develop the extension:

- `git clone` the sourcegraph repository
- Run `yarn generate`
- Open the repo in VS Code with `code .`
- Make your changes to the files within the `vscode` directory
- Click on `Run and Debug` from the activity bar on the left to open the sidebar
- Select `Launch & Watch VS Code Extension` from the dropdown menu to see your changes
- Select `Launch & Watch VS Code Web Extension` from the dropdown menu to see your changes with VS Code Web
