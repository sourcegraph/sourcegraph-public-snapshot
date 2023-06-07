# Cody AI by Sourcegraph

Cody for VS Code is an AI code assistant that can write code and answers questions across your entire codebase. It combines the power of large language models with Sourcegraph’s Code Graph API, generating deep knowledge of all of your code (and not just your open files). Large monorepos, multiple languages, and complex codebases are no problem for Cody.

For example, you can ask Cody:

- Where is the CI config for the web integration tests?
- Write a new GraphQL resolver for the AuditLog
- Why is the UserConnectionResolver giving an error "unknown user", and how do I fix it?
- Add helpful debug log statements
- Make this work _(seriously, it often works—try it!)_

  **Cody AI is in beta, and we’d love your [feedback](https://github.com/sourcegraph/sourcegraph/discussions/new?category=product-feedback&labels=cody,cody/vscode)**!

## Usage

1. Install the [Cody VS Code extension](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai).
1. Open Cody (from the activity bar or by pressing <kbd>Alt+/</kbd>/<kbd>Opt+/</kbd>) and sign in.
1. Start using it! Read on to learn about the chatbot, fixups, and recipes.

## Features

<!-- NOTE: These should stay roughly in sync with doc/cody/index.md, although that page needs to be not specific to VS Code. -->

- **🤖 Chatbot that knows _your_ code:** Writes code and answers questions with knowledge of your entire codebase, following your project's code conventions and architecture better than other AI code chatbots.
- **✨ Fixup code:** Interactively writes and refactors code for you, based on quick natural-language instructions.
- **📖 Recipes:** Generates unit tests, documentation, and more, with full codebase awareness.

### 🤖 Chatbot that knows _your_ code

[**📽️ Demo**](https://twitter.com/beyang/status/1647744307045228544)

To start chatting with Cody, click on the Cody icon in the activity bar (or press <kbd>Alt+/</kbd>/<kbd>Opt+/</kbd>, or run the `Cody: Focus on Chat View` command).

Cody tells you which code files it read to generate its response. (If Cody gives a wrong answer, please share feedback so we can improve it.)

### ✨ Fixup code

[**📽️ Demo**](https://twitter.com/sqs/status/1647673013343780864)

Just sprinkle your code with instructions in natural language, select the code, and run `Cody: Fixup` (<kbd>Ctrl+Alt+/</kbd>/<kbd>Ctrl+Opt+/</kbd>). Cody will figure out what edits to make.

Examples of the kinds of fixup instructions Cody can handle:

- "Factor out any common helper functions" (when multiple functions are selected)
- "Use the imported CSS module's class names"
- "Extract the list item to a separate React component"
- "Handle errors better"
- "Add helpful debug log statements"
- "Make this work" (seriously, it often works--try it!)

## 🍳 Built-in recipes

Select the recipes tab or right-click on a selection of code and choose one of the `Ask Cody > ...` recipes, such as:

- Explain code
- Generate unit test
- Generate docstring
- Improve variable names
- Translate to different language
- Summarize recent code changes
- Detect code smells
- Generate release notes

_We also welcome also pull request contributions for new, useful recipes!_

## Feedback

- [Issue tracker](https://github.com/sourcegraph/sourcegraph/issues)
- [Discord chat](https://discord.gg/s2qDtYGnAE)
- [Twitter (@sourcegraph)](https://twitter.com/sourcegraph)

## Development

See [CONTRIBUTING.md](./CONTRIBUTING.md).

## Other Extensions by Sourcegraph

- [Sourcegraph Search Extension](https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph)

## License

[Cody's code](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/client/cody) is open source (Apache License 2.0).
