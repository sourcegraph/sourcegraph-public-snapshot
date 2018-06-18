# Sourcegraph for VS Code <a href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph"><img src="https://storage.googleapis.com/sourcegraph-assets/vscode_badge.png" width="145" height="20"></img></a>

The Sourcegraph extension for VS Code enables you to connect your Sourcegraph extensions to VS Code.

## Installation

1. Run `vsce package` in this directory.
1. Run `code --install-extension sourcegraph-platform-VERSION.vsix`

## Extension Settings

This extension contributes the following settings:

* `sourcegraph.URL`: The Sourcegraph instance to use. Specify your on-premises Sourcegraph instance here, if applicable (e.g. `sourcegraph.example.com`).
* `sourcegraph.token`: Your Sourcegraph authentication token.


## Questions & Feedback

Please file an issue: https://github.com/sourcegraph/issues/issues/new


## Uninstallation

1. Open the extensions tab on the left side of VS Code (<kbd>Cmd+Shift+X</kbd> or <kbd>Ctrl+Shift+X</kbd>).
2. Search for `Sourcegraph` -> Gear icon -> `Uninstall` and `Reload`.


## Development

To develop the extension:

- Run `npm install` in the directory
- Open the repo with `code .`
- Press <kbd>F5</kbd> to open a new VS Code window with the extension loaded.
- After making changes to `src/extension.ts`, reload the window by clicking the reload icon in the debug toolbar or with <kbd>F5</kbd>.
- To release a new version:
  1. Update `README.md` (describe ALL changes)
  2. Update `CHANGELOG.md` (copy from README.md change above)
  3. Update `src/extension.ts` (`VERSION` constant)
  4. Publish on the VS Code store by following https://code.visualstudio.com/docs/extensions/publish-extension (contact @slimsag or @lindaxie for access)
    - `vsce login sourcegraph` (see also https://marketplace.visualstudio.com/manage/publishers/sourcegraph)
    - `cd sourcegraph-vscode` and `vsce publish <major|minor|patch>`
  7. `git add . && git commit -m "all: release v<THE VERSION>" && git push`
  8. `git tag v<THE VERSION> && git push --tags`
