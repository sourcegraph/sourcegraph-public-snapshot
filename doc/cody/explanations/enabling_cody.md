# Enabling Cody with Sourcegraph.com

Cody uses Sourcegraph to fetch relevant context to generate answers and code. These instructions walk through installing Cody and connecting it to Sourcegraph.com. For private instances of Sourcegraph, see [this page about enabling Cody for Enterprise](enabling_cody_enterprise.md).

## Initial setup

1. Sign into [Sourcegraph.com](https://sourcegraph.com). {You can create an account for free](https://sourcegraph.com/sign-up) if you don't already have one.
2. [Join the open beta](https://forms.gle/cffMa8mrr8YuHv8o8). We'll email you when you're added, usually within a day.
3. [Create a Sourcegraph access token](https://sourcegraph.com/user/settings/tokens)
4. Install [the Cody VS Code extension](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai)
5. Set the Sourcegraph URL to be `https://sourcegraph.com`
6. Set the access token to be the token you just created

  <img width="553" alt="image" src="https://user-images.githubusercontent.com/25070988/227510233-5ce37649-6ae3-4470-91d0-71ed6c68b7ef.png">

7. 

## Use embeddings for code-aware answers

After installing, you can use embeddings to improve Cody's context of existing code. Embeddings significantly improve the accuracy and quality of Cody's responses. Note that embeddings are only available for public repositories on sourcegraph.com. [See the list](embedded-repos.md) of embedded repositories and request any that you'd like to add by pinging a Sourcegraph team member in [Discord](https://discord.gg/8wJF5EdAyA).

If you want to use Cody with embeddings on private code, consider moving to a Sourcegraph Enterprise instance.

### Enable embeddings

The `Cody: Codebase` setting in VS Code enables codebase-aware answers for the Cody extension. By setting this configuration option to the name of a repository with embeddings, Cody will be able to provide more accurate and relevant answers to your coding questions based on that repository's content.

1. Open the VS Code workspace settings by pressing <kbd>Cmd/Ctrl+,</kbd>, (or File > Preferences (Settings) on Windows & Linux).
2. Search for the `Cody: Codebase` setting.
3. Enter the repository name.
   1. For example: `github.com/sourcegraph/sourcegraph` without the `https` protocol

## Provide feedback

Please spread the word online and send us your feedback in Discord! Cody is open source and we'd love to hear from you if you have bug reports or feature requests.
