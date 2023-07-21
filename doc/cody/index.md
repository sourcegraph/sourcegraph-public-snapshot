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

## Get started

<div class="getting-started">
  <a class="btn btn-primary text-left" href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai"><b>VS Code Extension</b><p>Install Cody's free and open source IDE extension for VS Code.</p></a>
    <a class="btn btn-primary text-left" href="https://plugins.jetbrains.com/plugin/9682-cody-ai-by-sourcegraph"><b>JetBrains Extension (experimental)</b><p>Install Cody's free and open source IDE extension for JetBrains.</p></a>
</div>

<div class="getting-started">
  <a class="btn btn-primary text-left" href="https://sourcegraph.com/get-cody"><b>Cody App!</b><p>Install the free desktop app to try Cody with your local codebase.</p></a>
  <a class="btn btn-primary text-left" href="https://about.sourcegraph.com/cody/pricing"><b>Cody Enterprise!</b><p>Get in touch with our team to try Cody for Sourcegraph Enterprise.</p></a>
</div>

## Cody features

<!-- NOTE: These should stay roughly in sync with client/cody/README.md, although these need to be not specific to VS Code. -->

### Your code-wise chatbot

Cody is your AI-powered coding assistant that understands your entire codebase inside out. It goes beyond other AI code chatbots, aligning perfectly with your project's code conventions and architecture. You can chat with Cody right within your code editor or through the Sourcegraph sidebar.

Examples of questions Cody can handle:

- How is our app's secret storage implemented on Linux?
- Where is the CI config for the web integration tests?
- Write a new GraphQL resolver for the AuditLog.
- Why is the UserConnectionResolver giving an error `unknown user`, and how do I fix it

Cody tells you which code files it reads to generate its response. In case of a wrong answer, please share feedback so we can improve it.

<div class="getting-started">
  <a class="btn text-center" href="https://twitter.com/beyang/status/1647744307045228544">View Demo</a>
</div>

### Fix code inline

Cody can help you make interactive edits and refactor code by following natural-language instructions. To do so, add natural-language instructions to your code, select the relevant code, and run:

```bash
 Cody: Fixup(Ctrl+Opt+/) — for MacOS

 Cody: Fixup(Ctrl+Alt+/) — for Windows
```

Cody will take it from there and figure out what edits to make.

Examples of fix-up instructions Cody can handle:

- Factor out any common helper functions (when multiple functions are selected)
- Use the imported CSS module's class `n`
- Extract the list item to a separate React component
- Handle errors better
- Add helpful debug log statements
- Make this work (and yes, it often does work—give it a try!)

<div class="getting-started">
  <a class="btn text-center" href="https://twitter.com/sqs/status/1647673013343780864">View Demo</a>
</div>

### Recipes

Cody can generate unit tests, documentation, and more, leveraging its full awareness of your codebase. You can access various helpful recipes to streamline your development process with just a few clicks.

Select the recipes tab in your editor or right click on a code section, then choose one of the `Ask Cody > ...` recipes. You'll find options such as:

- Explain code
- Generate unit test
- Generate `docstring`
- Detect code smells
- Improve variable names
- Generate release notes
- Summarize recent code changes
- Translate to different language

### Autocomplete

While typing, Cody provides real-time code auto-completion based on the context around your open files and file history. This predictive feature ensures a smoother coding experience.

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
  <a class="btn text-left" href="quickstart"><b>Cody Quickstart</b><p>This guide recommends first things to try once Cody is up and running.</p></a>
</div>
<div class="getting-started">
  <a class="btn text-left" href="explanations/use_cases"><b>Cody Use Cases</b><p>Explore some of the handy use cases with Cody.</p></a>
</div>
<div class="getting-started">
  <a class="btn text-left" href="faq"><b>FAQs</b><p>Learn about some of the frequently asked questions about Cody.</p></a>
</div>
<div class="getting-started">
  <a class="btn text-left" href="troubleshooting"><b>Cody Troubleshooting Guide</b><p>Having trouble with Cody? Review our troubleshooting guide for help.</p></a>
</div>
