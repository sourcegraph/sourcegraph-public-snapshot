# Cody AI by Sourcegraph

Cody for VS Code is an AI code assistant that can write code and answers questions across your entire codebase. It combines the power of large language models with Sourcegraph’s Code Graph API, generating deep knowledge of all of your code (and not just your open files). Large monorepos, multiple languages, and complex codebases are no problem for Cody.

For example, you can ask Cody:
- Where is the CI config for the web integration tests?
- Write a new GraphQL resolver for the AuditLog
- Why is the UserConnectionResolver giving an error "unknown user", and how do I fix it?
- Add helpful debug log statements
- Make this work _(seriously, it often works—try it!)_

 **Cody AI is in beta, and we’d love your [feedback](#feedback)**!

## Features

### 🤖 Ask Cody about anything in your codebase

Cody understands your entire codebase — not just your open files. Ask questions, insert code, and use the built-in recipes such as "Summarize recent code changes" and "Improve variable names".

<img src="walkthroughs/media/cody-ask.gif" width="800">

### ✨ Tools for writing, fixing and explaining code

Cody can autocomplete the line you’re on, perform complex inline fixups, or answer questions inline. TODO: complete.

<img src="walkthroughs/media/cody-fixup.gif" width="800">

## 🍿 See it in action

- https://twitter.com/beyang/status/1647744307045228544
- https://twitter.com/sqs/status/1647673013343780864

## 🍳 Built-in recipes

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

* [Sourcegraph Search Extension](https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph)

## License

[Cody's code](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/client/cody) is open source (Apache License 2.0).
