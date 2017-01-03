# vscode

We do NOT want to maintain a fork of vscode, for obvious reasons.

But we will need to modify some of vscode's behavior. Here are our rules:

* It is OK for us to make limited (in scope or complexity) changes to vscode's plumbing. E.g., how it loads web workers, how it imports CSS modules, etc. These changes MUST be applied in `ui/scripts/update-vscode.sh`, not as manual edits to the `ui/vendor/node_modules/vscode` files.
* It is NOT OK for us to make changes to vscode's UI/feature code, because that code will change more frequently and it'll be difficult to maintain our changes.
* It is OK for us to copy UI/feature code files into our own repository and modify them when we need to customize them.
* All overrides of VSCode modules belong in the workbench/overrides folder. They must include a note justifying the change. They are provided as overrides via webpack aliases. To reference the original module, use an import like `vscode/src/vs/*` instead of `vs/*`.

# Vocabulary
* `workbench` - The interface of VSCode. It controls all of the `part`s (editor, file tree).
* `part` - A major UI element. We use the editor part, the title bar part, and the file explorer part.
* `service` - VSCode has a dependency injection system to provide each class its dependencies. We override many of these services to provide Sourcegraph-specific functionality.
* `common`, `browser`, `electron`, `electron-browser` - These are various
  layers that VSCode source is separated into. We can use anything in the
  `browser` or `common` directory, and must replace functionality provided by
  functionality that depends on `electron`, `electron-browser` or `node`. Read
  https://github.com/Microsoft/vscode/wiki/Code-Organization for details.
