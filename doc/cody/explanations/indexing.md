# Generate Index to Enable Codebase-Aware Answers

These docs provide instructions on how to generate an index that enables codebase-aware answers for Cody. Codebase-aware answers leverage code graph context to enhance Cody's understanding of code from selected codebases. 

By following the steps outlined in this guide, you can configure the necessary prerequisites and enable Cody to provide more accurate and contextually relevant answers to your coding questions based on the codebase you are working in.

## Generate Index

You can enhance Cody's understanding of existing code by generating index for your codebases to enable [code graph context](https://docs.sourcegraph.com/cody/explanations/code_graph_context).

### Sourcegraph Enterprise

To generate an index for your codebase and enable codebase-aware answers for Cody, your site admin must complete the following:

- [Configure Code Graph Context](https://docs.sourcegraph.com/cody/explanations/code_graph_context) for your Sourcegraph instance
- [Enable Cody for your Sourcegraph instance](../enabling_cody_enterprise.md#step-1-enable-cody-on-your-sourcegraph-instance)
- [Enable Cody for your Sourcegraph account](../enabling_cody_enterprise.md#turning-cody-off)

### Sourcegraph.com

Code graph context is available instantly for public repositories on Sourcegraph.com that are already embedded. Please refer to the [list of repositories with embeddings](https://docs.sourcegraph.com/cody/embedded-repos) for the latest information.

Code graph context is not available for codebases that are not included in the list. 

However, you can make requests and ask for assistance in the [Sourcegraph Discord](https://discord.gg/8wJF5EdAyA) under the `#cody-embeddings` channel.

> NOTE: Sourcegraph.com does not support connections to private repositories

## Enable Codebase-Aware Answers

To enable codebase-aware answers for the Cody extension, you must set the `Cody: Codebase` (`cody.codebase`) configuration option in VS Code to the repository name on your Sourcegraph instance. By doing so, Cody will provide more accurate and relevant answers to your coding questions, referencing to the context of the codebase you are currently working in.

### Extension Settings

Here are the steps to configure the `codebase` setting for Cody via the [Extension Settings](https://code.visualstudio.com/docs/getstarted/settings#_extension-settings) in VS Code:

1. Open the VS Code workspace settings by clicking: 
   - Mac: `Code` > `Settings` > `Settings`
   - Windows & Linux: `File` > `Preferences (Settings)`
2. Enter `Cody: Codebase` in the search bar
3. Enter the repository name as listed on your Sourcegraph instance in the `Cody: Codebase` field

For example, the name for the [Sourcegraph repository on Sourcegraph.com](https://sourcegraph.com/github.com/sourcegraph/sourcegraph) is `github.com/sourcegraph/sourcegraph`, so we will enter it to the setting field without the https protocol as:

```
github.com/sourcegraph/sourcegraph
```

### Settings.json

Alternatively, you can configure via [settings.json](https://code.visualstudio.com/docs/getstarted/settings#_settingsjson) using the `cody.codebase` configuration contribution point:

```json
{
  "cody.serverEndpoint": "https://sourcegraph.com",
  "cody.codebase": "github.com/sourcegraph/sourcegraph"
}
```
