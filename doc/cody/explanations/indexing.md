# Generate Index to Enable Codebase-Aware Answers for Cody

This docs provides instructions on how to generate an index that enables codebase-aware answers for Cody. Codebase-aware answers leverage code graph context to enhance Cody's understanding of code from selected codebases. 

By following the steps outlined in this guide, you can configure the necessary prerequisites and enable Cody to provide more accurate and contextually relevant answers to your coding questions based on the codebase you are working in.

## General Index

You can enhance Cody's understanding of existing code by generating index for your codebases to enable [code graph context](https://docs.sourcegraph.com/cody/explanations/code_graph_context).

### Sourcegraph Enterprise

To generate an index for your codebase and enable codebase-aware answers for Cody, your Site Admins must:

- Configure code graph context on your Sourcegraph instance
- Enable Cody for your Sourcegraph account

> If you are a Site Admin, please refer to our documentation on [Code Graph Context](https://docs.sourcegraph.com/cody/explanations/code_graph_context) for detailed instructions.

### Sourcegraph.com

Code graph context is available only for public repositories on Sourcegraph.com that are already embedded.

Please refer to the [list of repositories with embeddings](https://docs.sourcegraph.com/cody/embedded-repos) for instant access.

If the codebase you want to connect to is not on the list, it will not provide code graph context to Cody. However, you can request assistance in the #cody-embeddings channel on the Sourcegraph team's [Discord](https://discord.gg/8wJF5EdAyA).

Please note that Sourcegraph.com currently does not support private repositories.

## Enable Codebase-Aware Answers

To enable codebase-aware answers for the Cody extension, you need to set the `Cody: Codebase` (`cody.codebase`) configuration option in VS Code. Set this option to the repository name on your Sourcegraph instance. By doing so, Cody will provide more accurate and relevant answers to your coding questions, referring to the context of the codebase you are currently working in.

### Extension Settings

Here are the steps to configure the `codebase` setting for Cody via the [Extension Settings](https://code.visualstudio.com/docs/getstarted/settings#_extension-settings) in VS Code:

1. Open the VS Code workspace settings by clicking: 
   - Mac: `Code` > `Settings` > `Settings`
   - Windows & Linux: `File` > `Preferences (Settings)`
2. Enter `Cody: Codebase` in the search bar
3. Enter the repository name as listed on your Sourcegraph instance.
  - For example, the name for the [Sourcegraph repository on Sourcegraph.com](https://sourcegraph.com/github.com/sourcegraph/sourcegraph) is `github.com/sourcegraph/sourcegraph`, so we will enter it to the setting field without the https protocol as `github.com/sourcegraph/sourcegraph`

### Settings.json

Alternatively, if you can configure via [settings.json](https://code.visualstudio.com/docs/getstarted/settings#_settingsjson) using the `cody.codebase` configuration contribution point:

```json
{
  "cody.serverEndpoint": "https://sourcegraph.com",
  "cody.codebase": "github.com/sourcegraph/sourcegraph"
}
```
