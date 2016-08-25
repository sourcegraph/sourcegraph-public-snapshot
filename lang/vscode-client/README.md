# vscode-client

The vscode-client extension for Visual Studio Code helps you develop
and debug language servers. It lets you run multiple language servers
at once with minimal extra configuration per language.


## Using this extension

1. Follow the [Go language server installation instructions](../golang/README.md) and [sample language server installation instructions](../sample/README.md)
1. Run `npm install`.
1. Run `npm run vscode` to start a new VSCode instance. Use `npm run vscode -- /path/to/mydir` to open the editor to a specific directory.
1. Open a `.go` or `.txt` file and hover over text to start using the language servers. Refer to the language servers' installation instructions for detailed usage information.

To view a language server's stderr output in VSCode, select View → Output (temporary note: see the known issue section). To debug further, see the "Hacking on this extension" section below.

After updating the binary for a language server (during development or after an upgrade), just kill the process (e.g., `killall langserver-go`). VSCode will automatically restart and reconnect to the language server process.

> **Note for those who use VSCode as their primary editor:** Because this extension's functionality conflicts with other VSCode extensions (e.g., showing Go hover information), the `npm run vscode` script launches an separate instance of VSCode and stores its config in `../.vscode-dev`. It will still show your existing extensions in the panel (which seems to be a VSCode bug), but they won't be activated.


## Hacking on this extension

1. Run `npm install` in this directory (`vscode-client`).
1. Open this directory by itself in Visual Studio Code.
1. Hit F5 to open a new VSCode instance in a debugger running this extension. (This is equivalent to going to the Debug pane on the left and running the "Launch Extension" task.)

See the [Node.js example language server tutorial](https://code.visualstudio.com/docs/extensions/example-language-server) under "To test the language server" for more information.


## Known issues

* To view language server stderr output, we must apply the change from https://github.com/Microsoft/vscode-languageserver-node/pull/83 to the vscode-languageclient dependency. This patch is applied in a npm postinstall script automatically. When that PR is merged, we'll remove this hack.
