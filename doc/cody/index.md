# <picture title="Cody"><img class="theme-dark-only" src="https://storage.googleapis.com/sourcegraph-assets/cody/20230417/logomark-default-text-white.png" width="200"><img class="theme-light-only" src="https://storage.googleapis.com/sourcegraph-assets/cody/20230417/logomark-default-text-black.png" width="200"><div style="display:none">Cody</div></picture>

<span class="badge badge-experimental">Experimental</span>

Cody is an AI code assistant that writes code and answers questions for you by reading your entire codebase and the code graph.

Cody uses a combination of Sourcegraph's code graph and Large Language Models (LLMs) to eliminate toil and keep human devs in flow. You can think of Cody as your coding assistant who has read through all the code in open source, all the questions on StackOverflow, and your own entire codebase, and is always there to answer questions you might have or suggest ways of doing something based on prior knowledge.

## Get Cody

- **Sourcegraph Enterprise customers:** Contact your Sourcegraph techical advisor or [request enterprise access](https://sourcegraph.typeform.com/to/pIXTgwrd) to use Cody on your existing Sourcegraph instance.
- **Everyone:** [Join the open beta.](https://forms.gle/cffMa8mrr8YuHv8o8) We'll email you when you're added, usually within a day.

Cody is available as a [VS Code extension](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai) and in Sourcegraph itself.

<div class="cta-group">
<a class="btn btn-primary" href="quickstart">★ VS Code Extension Quickstart</a>
<a class="btn" href="explanations/use_cases">Cody Use Cases</a>
<a class="btn" href="faq">FAQ</a>
<a class="btn" href="https://discord.com/servers/sourcegraph-969688426372825169">Join our Discord</a>
</div>

## Features

<!-- NOTE: These should stay roughly in sync with client/cody/README.md, although these need to be not specific to VS Code. -->

- **🤖 Chatbot that knows _your_ code:** Writes code and answers questions with knowledge of your entire codebase, following your project's code conventions and architecture better than other AI code chatbots.
- **✨ Fixup code:** Interactively writes and refactors code for you, based on quick natural-language instructions.
- **🧪 Recipes:** Generates unit tests, documentation, and more, with full codebase awareness.

### 🤖 Chatbot that knows _your_ code

[**📽️ Demo**](https://twitter.com/beyang/status/1647744307045228544)

You can chat with Cody in VS Code or in the Sourcegraph sidebar.

Examples of the kinds of questions Cody can handle:

- How is our app's secret storage implemented on Linux?
- Where is the CI config for the web integration tests?
- Write a new GraphQL resolver for the AuditLog.
- Why is the UserConnectionResolver giving an error `unknown user`, and how do I fix it?

Cody tells you which code files it read to generate its response. (If Cody gives a wrong answer, please share feedback so we can improve it.)

### ✨ Fixup code

[**📽️ Demo**](https://twitter.com/sqs/status/1647673013343780864)

In VS Code, just sprinkle your code with instructions in natural language, select the code, and run `Cody: Fixup` (<kbd>Ctrl+Alt+/</kbd>/<kbd>Ctrl+Opt+/</kbd>). Cody will figure out what edits to make.

Examples of the kinds of fixup instructions Cody can handle:

- "Factor out any common helper functions" (when multiple functions are selected)
- "Use the imported CSS module's class names"
- "Extract the list item to a separate React component"
- "Handle errors better"
- "Add helpful debug log statements"
- "Make this work" (seriously, it often works--try it!)

### 🧪 Recipes

In VS Code, right-click on a selection of code and choose one of the `Ask Cody > ...` recipes, such as:

- Explain Code
- Generate Unit Test
- Generate Docstring
- Improve Variable Names

## Explanations

- [Enabling Cody for Sourcegraph Enterprise customers](explanations/enabling_cody_enterprise.md)
- [Enabling Cody for Sourcegraph.com users](explanations/enabling_cody.md)
- [Configuring code graph context](explanations/code_graph_context.md)
