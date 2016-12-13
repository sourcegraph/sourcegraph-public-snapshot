# vscode

We do NOT want to maintain a fork of vscode, for obvious reasons.

But we will need to modify some of vscode's behavior. Here are our rules:

* It is OK for us to make limited (in scope or complexity) changes to vscode's plumbing. E.g., how it loads web workers, how it imports CSS modules, etc. These changes MUST be applied in `ui/scripts/update-vscode.sh`, not as manual edits to the `ui/node_modules/vscode` files.
* It is NOT OK for us to make changes to vscode's UI/feature code, because that code will change more frequently and it'll be difficult to maintain our changes.
* It is OK for us to copy UI/feature code files into our own repository and modify them when we need to customize them.
* All overrides of VSCode modules belong in the workbench/overrides folder. They must include a note justifying the change.
