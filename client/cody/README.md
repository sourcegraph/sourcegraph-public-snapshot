# Cody: AI code assistant

Cody is an AI code assistant that writes code and answers questions for you by reading your entire codebase and the code graph.

**Status:** experimental ([join the open beta](https://docs.sourcegraph.com/cody))

[**Full documentation**](https://docs.sourcegraph.com/cody)

## Features

<!-- NOTE: These should stay roughly in sync with doc/cody/index.md, although that page needs to be not specific to VS Code. -->

- **🤖 Chatbot that knows _your_ code:** Writes code and answers questions with knowledge of your entire codebase, following your project's code conventions and architecture better than other AI code chatbots.
- **✨ Fixup code:** Interactively writes and refactors code for you, based on quick natural-language instructions.
- **🧪 Recipes:** Generates unit tests, documentation, and more, with full codebase awareness.

## Usage

1. Install the [Cody VS Code extension](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai).
1. Open Cody (from the activity bar or by pressing <kbd>Alt+/</kbd>/<kbd>Opt+/</kbd>) and sign in.
1. Start using it! Read on to learn about the chatbot, fixups, and recipes.

### 🤖 Chatbot that knows _your_ code

[**📽️ Demo**](https://twitter.com/beyang/status/1647744307045228544)

To start chatting with Cody, click on the Cody icon in the activity bar (or press <kbd>Alt+/</kbd>/<kbd>Opt+/</kbd>, or run the `Cody: Focus on Chat View` command).

Examples of the kinds of questions Cody can handle:

- How is our app's secret storage implemented on Linux?
- Where is the CI config for the web integration tests?
- Write a new GraphQL resolver for the AuditLog.
- Why is the UserConnectionResolver giving an error `unknown user`, and how do I fix it?

Cody tells you which code files it read to generate its response. (If Cody gives a wrong answer, please share feedback so we can improve it.)

**Note:** For full codebase awareness, you must set the `cody.codebase` setting to the name of the repository on the connected Sourcegraph instance.

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

### 🧪 Recipes

Right-click on a selection of code and choose one of the `Ask Cody > ...` recipes, such as:

- Explain Code
- Generate Unit Test
- Generate Docstring
- Improve Variable Names

We welcome PRs that contribute new, useful recipes.

## Feedback

- [Issue tracker](https://github.com/sourcegraph/sourcegraph/issues)
- [Discord chat](https://discord.gg/s2qDtYGnAE)
- [@sourcegraph (Twitter)](https://twitter.com/sourcegraph)

## Development

[Cody's code](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/client/cody) is open source (Apache 2). See [CONTRIBUTING.md](./CONTRIBUTING.md).
