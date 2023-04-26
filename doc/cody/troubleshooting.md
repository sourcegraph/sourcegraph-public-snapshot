# Cody troubleshooting guide

Here are common troubleshooting steps to run before filing Cody bugs on the [issue tracker](https://github.com/sourcegraph/sourcegraph/issues) or asking in our [Discord](https://discord.gg/s2qDtYGnAE). (We're always happy to help, though!)

## VS Code extension

### Cody is not responding in chat

1. Ensure you are on the latest version of the [Cody VS Code extension](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai). You can run the VS Code command `Extensions: Check for Extension Updates` to check.
1. Check the VS Code error console for relevant error messages. To open it, run the VS Code command `Developer: Toggle Developer Tools` and then look in the `Console` for relevant messages.

### Errors trying to install Cody on MacOS

If you are getting: 
```
Command 'Cody: Set Access Token' resulted in an error

command 'cody.set-access-token' not found
```
1. Close VS Code
2. Open Keychain Access.app 
3. Search for `cody`
4. Delete the `vscodesourcegraph.cody-ai` entry in the System keychain on the left. 
5. Try opening VS Code again

![Opening up Keychain Access](https://storage.googleapis.com/sourcegraph-assets/blog/cody-docs-troubleshooting-keychain-access.png)
