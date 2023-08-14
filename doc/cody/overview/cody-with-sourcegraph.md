# Enabling Cody with Sourcegraph.com

Cody uses Sourcegraph to fetch relevant context to generate answers and code. These instructions walk through installing Cody in your editor and connecting it to Sourcegraph.com and is the best option if you're interested in using Cody on public code. To use Cody on your local code, download the [Cody App](./../overview/app/index.md) or see [this page about enabling Cody for Enterprise](enable-cody-enterprise.md).

## Initial setup

1. [Create a Sourcegraph.com account](https://sourcegraph.com/sign-up)
2. Install [the Cody VS Code extension](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai)
3. Open the Cody extension
4. Click on **Other Sign In Options...** and select Sign in to Sourcegraph.com
5. Follow the prompts to authorize Cody to access your Sourcegraph.com account

You're now ready to use Cody in VS Code!

## Configure code graph context for code-aware answers

After installing, you can optionally use [code graph context](./../explanations/code_graph_context.md) to improve Cody's context of existing code. Note that code graph context is only available for public repositories on sourcegraph.com which have embeddings. [See the list](../embedded-repos.md) of repositories with embeddings and request any that you'd like to add by pinging a Sourcegraph team member in [Discord](https://discord.gg/8wJF5EdAyA).

If you want to use Cody with code graph context on private code, consider downloading the [Cody App](./../overview/app/index.md) or moving to a Sourcegraph Enterprise instance.

### Enable code graph context

The `Cody: Codebase` setting in VS Code enables codebase-aware answers for the Cody extension. By setting this configuration option to the name of a repository with embeddings, Cody will be able to provide more accurate and relevant answers to your coding questions based on that repository's content.

- Open the VS Code workspace settings by pressing <kbd>Cmd/Ctrl+,</kbd>, (or File > Preferences (Settings) on Windows & Linux).
- Search for the `Cody: Codebase` setting.
- Enter the repository name.
  - For example: `github.com/sourcegraph/sourcegraph` without the `https` protocol

## Provide feedback

Please spread the word online and send us your feedback in Discord! Cody is open source and we'd love to hear from you if you have bug reports or feature requests.
