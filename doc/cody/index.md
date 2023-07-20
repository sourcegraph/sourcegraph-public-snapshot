# <picture title="Cody"><img class="theme-dark-only" src="https://storage.googleapis.com/sourcegraph-assets/cody/20230417/logomark-default-text-white.png" width="200"><img class="theme-light-only" src="https://storage.googleapis.com/sourcegraph-assets/cody/20230417/logomark-default-text-black.png" width="200"><div style="display:none">Cody</div></picture>

<aside class="beta">
<p>
<span class="badge badge-beta">Beta</span>
Cody is currently available in beta, and we value your feedback! You can <a href="https://about.sourcegraph.com/contact">contact us</a> directly, file an  <a href="https://github.com/sourcegraph/sourcegraph">issue</a>, or <a href="https://twitter.com/sourcegraph">tweet</a>.
</p>
</aside>

Cody is an AI coding assistant tool that utilizes Sourcegraph's <a href="https://docs.sourcegraph.com/cody/explanations/code_graph_context"> code graph</a> and Large Language Models (LLMs) to write code and provide answers based on your codebase and code graph.

Think of Cody as your personal coding assistant with a comprehensive understanding of your open source code, StackOverflow questions, and entire codebase.

Cody helps you answer questions, write code, and offer suggestions for code improvement.

## Get Started

<div class="getting-started">
  <a class="btn btn-primary text-left" href="https://about.sourcegraph.com/cody/pricing"><b>Cody Enterprise!</b><p>Get in touch with our team to try Cody for Sourcegraph Enterprise.</p></a>
  <a class="btn btn-primary text-left" href="https://sourcegraph.com/get-cody"><b>Cody App!</b><p>Install the free desktop app to try Cody with your local codebase.</p></a>
</div>

<div class="getting-started">
  <a class="btn btn-primary text-left" href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai"><b>VS Code Extension</b><p>Install Cody's free IDE extension for VS Code.</p></a>
    <a class="btn btn-primary text-left" href="https://plugins.jetbrains.com/plugin/9682-cody-ai-by-sourcegraph"><b>JetBrains Extension (experimental)</b><p>Install Cody's free IDE extension for JetBrains.</p></a>
</div>

[Read more about the Cody clients, extensions, and plugins](explanations/cody_clients.md), including a full breakdown of features available by client.

## Features

<!-- NOTE: These should stay roughly in sync with client/cody/README.md, although these need to be not specific to VS Code. -->

- **🤖 Chatbot that knows _your_ code:** Writes code and answers questions with knowledge of your entire codebase, following your project's code conventions and architecture better than other AI code chatbots.
- **✨ Fix code inline:** Interactively writes and refactors code for you, based on quick natural-language instructions.
- **📖 Recipes:** Generates unit tests, documentation, and more, with full codebase awareness.
- **Autocomplete:** Get suggestions from Cody as you're coding.

### 🤖 Chatbot that knows _your_ code

[**📽️ Demo**](https://twitter.com/beyang/status/1647744307045228544)

You can chat with Cody in the editor or in the Sourcegraph sidebar.

Examples of the kinds of questions Cody can handle:

- How is our app's secret storage implemented on Linux?
- Where is the CI config for the web integration tests?
- Write a new GraphQL resolver for the AuditLog.
- Why is the UserConnectionResolver giving an error `unknown user`, and how do I fix it?

Cody tells you which code files it read to generate its response. (If Cody gives a wrong answer, please share feedback so we can improve it.)

### ✨ Fix code inline

[**📽️ Demo**](https://twitter.com/sqs/status/1647673013343780864)

In your editor, just sprinkle your code with instructions in natural language, select the code, and run `Cody: Fixup` (<kbd>Ctrl+Alt+/</kbd>/<kbd>Ctrl+Opt+/</kbd>). Cody will figure out what edits to make.

Examples of the kinds of fixup instructions Cody can handle:

- "Factor out any common helper functions" (when multiple functions are selected)
- "Use the imported CSS module's class names"
- "Extract the list item to a separate React component"
- "Handle errors better"
- "Add helpful debug log statements"
- "Make this work" (seriously, it often works--try it!)

### 📖 Recipes

In your editor, select the recipes tab or  right-click on a selection of code and choose one of the `Ask Cody > ...` recipes, such as:

- Explain code
- Generate unit test
- Generate docstring
- Improve variable names
- Translate to different language
- Summarize recent code changes
- Detect code smells
- Generate release notes

### Autocomplete

Cody provides real-time code autocompletion as you're typing. As you start coding, or after you type a comment, Cody will look at the context around your open files and file history to predict what you're trying to implement and provide autocomplete.

## Troubleshooting

See [Cody troubleshooting guide](troubleshooting.md).

## Explanations

- [Cody clients, plugins, and extensions](explanations/cody_clients.md)
- [Enabling Cody for Sourcegraph Enterprise customers](explanations/enabling_cody_enterprise.md)
- [Enabling Cody for the Cody app](../app/index.md)
- [Enabling Cody for open source Sourcegraph.com users](explanations/enabling_cody.md)
- [Installing the Cody VS Code extension](explanations/installing_vs_code.md)
- [Installing the Jetbrains extension (experimental)](explanations/installing_jetbrains.md)
- [Configuring code graph context](explanations/code_graph_context.md)
- [Sourcegraph Cody Gateway](explanations/cody_gateway.md)

## More resources

For more information on what to do next, we recommend the following resources:

<div class="getting-started">
  <a class="btn btn-primary text-center" href="quickstart">Cody Quickstart</a>
  <a class="btn text-center" href="explanations/use_cases">Cody Use Cases</a>
</div>
<div class="getting-started">
  <a class="btn text-center" href="faq">FAQs</a>
  <a class="btn text-center" href="https://discord.com/servers/sourcegraph-969688426372825169">Join our Discord!</a>
</div>
