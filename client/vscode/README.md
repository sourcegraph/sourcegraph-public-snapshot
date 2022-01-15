# Sourcegraph for Visual Studio Code

[![vs marketplace](https://img.shields.io/vscode-marketplace/v/sourcegraph.sourcegraph.svg?label=vs%20marketplace)](https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph) [![downloads](https://img.shields.io/vscode-marketplace/d/sourcegraph.sourcegraph.svg)](https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph) [![build](https://img.shields.io/github/workflow/status/sourcegraph/sourcegraph-vscode/build/master)](https://github.com/sourcegraph/sourcegraph-vscode/actions?query=branch%3Amaster+workflow%3Abuild) [![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release)
[![codecov](https://codecov.io/gh/sourcegraph/sourcegraph-vscode/branch/master/graph/badge.svg?token=8TLCsGxBeS)](https://codecov.io/gh/sourcegraph/sourcegraph-vscode)

The Sourcegraph extension for VS Code enables you to open and search code on Sourcegraph easily and efficiently without leaving your IDE!

## Installation

1.  Open the extensions tab on the left side of VS Code (<kbd>Cmd</kbd>+<kbd>Shift</kbd>+<kbd>X</kbd> or <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>X</kbd>).
2.  Search for `Sourcegraph` -> `Install` and `Reload`.

## Usage

In the command palette (<kbd>Cmd</kbd>+<kbd>Shift</kbd>+<kbd>P</kbd> or <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>P</kbd>), search for `Sourcegraph:` to see available actions.

## Keyboard Shortcuts:

| Description                             | Mac                                          | Linux / Windows                               |
| --------------------------------------- | -------------------------------------------- | --------------------------------------------- |
| Open Sourcegraph Search Tab             | <kbd>Cmd</kbd>+<kbd>Shift</kbd>+<kbd>8</kbd> | <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>8</kbd> |
| Open File in Sourcegraph Web            | <kbd>Option</kbd>+<kbd>A</kbd>               | <kbd>Alt</kbd>+<kbd>A</kbd>                   |
| Search Selected Text in Sourcegraph     | <kbd>Option</kbd>+<kbd>S</kbd>               | <kbd>Alt</kbd>+<kbd>S</kbd>                   |
| Search Selected Text in Sourcegraph Web | <kbd>Option</kbd>+<kbd>Q</kbd>               | <kbd>Alt</kbd>+<kbd>Q</kbd>                   |

## Extension Settings

This extension contributes the following settings:

- `sourcegraph.url`: Specify your on-premises Sourcegraph instance here, if applicable. For example: `"sourcegraph.url": "https://sourcegraph.com"`
- `sourcegraph.accessToken`: The access token to query the Sourcegraph API. Required to use this extension with private instances.
- `sourcegraph.corsURL`: CORS headers are necessary for the extension to fetch data when using VS Code Web with instances on version under 3.35.2.
- `sourcegraph.remoteUrlReplacements`: Object, where each `key` is replaced by `value` in the remote url.
- `sourcegraph.defaultBranch`: String to set the name of the default branch. Always open files in the default branch.

## Questions & Feedback

Please file an issue at https://github.com/sourcegraph/sourcegraph-vscode/issues/new

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
