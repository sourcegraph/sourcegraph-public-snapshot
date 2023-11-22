# Cody troubleshooting guide

Here are common troubleshooting steps to run before filing Cody bugs on the [issue tracker](https://github.com/sourcegraph/cody/issues) or asking in our [Discord](https://discord.gg/s2qDtYGnAE). (We're always happy to help, though!)

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

### Codebase is `Not Indexed`

If you are logged into Sourcegraph.com, only public open source repositories on [this list](embedded-repos.md) are indexed. Please join the [Sourcegraph Discord](https://discord.gg/8wJF5EdAyA) and message the `#embeddings-indexing` channel to get an open source repository added to the public index.

If you’re connected to a Sourcegraph Enterprise instance, please ask your site admin to [Configure Code Graph Context](core-concepts/code-graph.md) for your Sourcegraph instance and then [Enable Cody](overview/enable-cody-enterprise.md) for your account.

If you're connected to the Cody app, you can trigger indexing for a repository by adding the repo to your app under Settings > Local repositories, navigating to Settings > Advanced settings > Embeddings jobs in the app, and scheduling embedding. If your repo has no git remote or still shows as `Not Indexed`, you'll need to follow the step below to set `Cody: Codebase` to the repository name as displayed at Settings > Local repositories in the Cody App.

If you've completed the above and still seeing your codebase showing up as `NOT INDEXED`, try updating the `Cody: Codebase` (`cody.codebase`) setting in VS Code to the repository name as listed on your Sourcegraph instance.

For more information, see [Generate Index to Enable Codebase-Aware Answers](core-concepts/embeddings/embedding-index.md).

### Signin fails on each VS Code restart

If you are automatically signed out of Cody upon every VS Code restart due to keychain authentication issues, please follow the suggested steps detailed in the official VS Code docs on [troubleshooting keychain issues](https://code.visualstudio.com/docs/editor/settings-sync#_troubleshooting-keychain-issues) to resolve this.

### Autocomplete rate limits

Cody supports 1,000 suggested autocompletions per day, per user. For Sourcegraph Enterprise instances, this rate limit is pooled across all users.

If you hit the rate limit, please wait a bit and try again later. You can also contact Sourcegraph Support to discuss increasing your rate limit for your use case.

### Error logging in VS Code on Linux

On Linux, if you encounter an issue where your are unable to login to Cody using your Sourcegraph instance URL, with a valid access token and you observe that during the sign-in process VS Code just hangs and Cody cannot start, it could be possible that you may be having possible underlying networking rules related to SSL certificates.

To resolve this via a workaround:

1. Quit VS Code.
2. Run `echo "export NODE_TLS_REJECT_UNAUTHORIZED=0" >> ~/.bashrc` in terminal.
3. Restart VS Code and sign in again.

### Error exceeding `localStorage` quota

When using Cody chat, you may come across this error:

```
Failed to execute 'setItem' on 'Storage': Setting the value of 'user-history:$user_id' exceeded the quota.
```

Experiencing this error means that the size of the chat history exceeded what the local storage in your browser can store. Cody saves all context data along with each chat message and this is what primarily causes this.
To fix this, navigate to https://sourcegraph.example.com/cody/chat and click on `Clear Chat History` if your instance is on v5.2.3+. If on an older version, please clear your browsing data/browser history.
