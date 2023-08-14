# Generate an Embeddings Index to Enable Codebase-Aware Answers

These docs provide instructions on how to generate an index that enables codebase-aware answers for Cody. Codebase-aware answers leverage code graph context to enhance Cody's understanding of code from selected codebases.

By following the steps outlined in this guide, you can configure the necessary prerequisites and enable Cody to provide more accurate and contextually relevant answers to your coding questions based on the codebase you are working in.

## Generate Embeddings Index

You can enhance Cody's understanding of existing code by embedding your code base.

### Sourcegraph Enterprise

To embed your codebase and enable codebase-aware answers for Cody, your site admin must complete the following:

- [Configure Code Graph Context](code_graph_context.md) for your Sourcegraph instance
- [Enable Cody for your Sourcegraph instance](../overview/enable-cody-enterprise.md#step-1-enable-cody-on-your-sourcegraph-instance)
- Embed your codebase by either
  - [scheduling a one-off embeddings job](schedule_one_off_embeddings_jobs.md), or
  - [creating an embeddings policy to automatically keep your index up-to-date](policies.md).

### Sourcegraph.com

 See the current [list of repositories with code graph context available](../embedded-repos.md), and request any that you'd like to add by pinging a Sourcegraph team member in the [Sourcegraph Discord](https://discord.gg/8wJF5EdAyA) under the `#cody-embeddings` channel.

> NOTE: Sourcegraph.com does not support connections to private repositories

## Enable Codebase-Aware Answers


When connected to the correct codebase, Cody provides accurate and relevant answers to your coding questions, taking into account the context of the codebase you are currently working in.

Cody attempts to connect to the appropriate codebase automatically using Git, eliminating the need for manual configuration of the `Cody: Codebase` (`cody.codebase`) option.

You only need to configure the `Cody: Codebase` (`cody.codebase`) setting if Cody displays `NOT INDEXED` below the chat window.


### Extension Settings


Cody tries to connect to the correct codebase using Git at start-up, so manual configuration of `cody.codebase` is unnecessary.

To enable Cody to set the codebase config using `git remote get-url origin`, ensure that the `Cody: Codebase` (`cody.codebase`) configuration field is unset in your VS Code User and Workspace settings, and remote workspace and folder-level settings if applicable.

If this attempt fails and you see `NOT INDEXED` below your Cody chatbox, you can manually specify the correct codebase for Cody using the `Cody: Codebase` (`cody.codebase`) configuration option.

#### Manual Configuration

Follow these steps to manually configure the `codebase` setting for Cody via the [Extension Settings](https://code.visualstudio.com/docs/getstarted/settings#_extension-settings) in VS Code:

1. Open the VS Code workspace settings by clicking:
   - Mac: `Code` > `Settings` > `Settings`
   - Windows & Linux: `File` > `Preferences (Settings)`
2. Switch to the Workspace settings tab
3. Enter `Cody: Codebase` in the search bar
4. Enter the repository name as listed on your Sourcegraph instance in the `Cody: Codebase` field

For example, the name for the [Sourcegraph repository on Sourcegraph.com](https://sourcegraph.com/github.com/sourcegraph/sourcegraph) is `github.com/sourcegraph/sourcegraph`, so we will enter it to the setting field without the https protocol as:

```
github.com/sourcegraph/sourcegraph
```

### Settings.json

Alternatively, you can configure `codebase` manually via [settings.json](https://code.visualstudio.com/docs/getstarted/settings#_settingsjson) using the `cody.codebase` configuration contribution point:

```json
{
  "cody.serverEndpoint": "https://sourcegraph.com",
  "cody.codebase": "github.com/sourcegraph/sourcegraph"
}
```
