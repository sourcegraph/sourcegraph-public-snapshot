# Cody (experimental)

Cody is an AI coding assistant that lives in your editor that can find, explain, and write code. Cody uses a combination of AI (specifically Large Language Models or LLMs), Sourcegraph search, and Sourcegraph code intelligence to provide answers that eliminate toil and keep human programmers in flow. You can think of Cody as your programmer buddy who has read through all the code on GitHub, all the questions on StackOverflow, and all your organization's private code, and is always there to answer questions you might have or suggest ways of doing something based on prior knowledge.

Cody is in private alpha (tagged as an [experimental](../doc/admin/beta_and_experimental_features.md) feature) at this stage. Contact your techical advisor or [signup here](https://t.co/4TMTW1b3lR) to get access.

In this initial release, Cody is only available as a VS Code extension.

To enable Cody:

- **Turn on Cody on your instance** (requires site-admin permissions):
  - TODO

- **Install the VS Code extension**
  - Open VSCode editor
  - Uninstall any previous Cody extension you manually installed. You have to perform the steps below even if you previously installed it manually.
  - Go to [this link](https://marketplace.visualstudio.com/items?itemName=hpargecruos.kodj) and install the Kodj extension. We use an obfuscated VS Code marketplace listing during the Experimental phase.
  - Restart/reload the VSCode editor.


Note: by enabling Cody, you agree to the [Cody Notice and Usage Policy](https://about.sourcegraph.com/cody-notice). In particular, some code snippets will be sent to a third-party language model provider when you use Cody questions.
