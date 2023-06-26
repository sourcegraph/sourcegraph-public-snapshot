# Cody AI by Sourcegraph

Cody for VS Code is an AI code assistant that can write code and answers questions across your entire codebase. It combines the power of large language models with Sourcegraph’s Code Graph API, generating deep knowledge of all of your code (and not just your open files). Large monorepos, multiple languages, and complex codebases are no problem for Cody.

For example, you can ask Cody:

- Where is the CI config for the web integration tests?
- Write a new GraphQL resolver for the AuditLog
- Why is the UserConnectionResolver giving an error "unknown user", and how do I fix it?
- Add helpful debug log statements
- Make this work _(seriously, it often works—try it!)_

  **Cody AI is in beta, and we’d love your [feedback](https://github.com/sourcegraph/sourcegraph/discussions/new?category=product-feedback&labels=cody,cody/vscode)**!

## Features

<!-- NOTE: These should stay roughly in sync with doc/cody/index.md, although that page needs to be not specific to VS Code. -->

### 🤖 Ask Cody about anything in your codebase

Cody understands your entire codebase — not just your open files. Ask questions, insert code, and use the built-in recipes such as "Summarize recent code changes" and "Improve variable names".

![Example of chatting with Cody](https://storage.googleapis.com/sourcegraph-assets/website/Product%20Animations/GIFS/cody-chat-may2023-optim.gif)

### ✨ Tools for fixing and explaining code

Cody can perform complex inline fixups, or answer questions inline. Interact with Cody within your code through Inline Assist, or use the "Fixup code from inline instructions" recipe for more involved fixups.

![Example of using a fixup](https://storage.googleapis.com/sourcegraph-assets/website/Product%20Animations/GIFS/cody-fixup-may2023-optim.gif)

### 🔨 Let Cody write code for you

Cody can provide real-time code completions as you're typing. As you start coding, or after you type a comment, Cody will look at the context around your open files and file history to predict what you're trying to implement and provide completions.

![Example of using code completions](https://storage.googleapis.com/sourcegraph-assets/website/Product%20Animations/GIFS/cody-completions-may2023-optim.gif)

## 🍿 See it in action

- https://twitter.com/beyang/status/1647744307045228544
- https://twitter.com/sqs/status/1647673013343780864

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
