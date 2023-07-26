<style>
.get-started {
  display: flex;
  flex-direction: row;
}
.get-started a {
  padding: 0.25rem;
  margin: 1rem;
  background: #dddddd;
  border-radius: 0.25rem;
  width: 3.5rem;
  height: 3.5rem;
  display: flex;
  align-items: center;
}
.get-started a:hover {
  filter: brightness(0.75);
}
.get-started a img {
  width: 100%;
  height: 100%;
}
</style>

# <picture title="Cody"><img class="theme-dark-only" src="https://storage.googleapis.com/sourcegraph-assets/cody/20230417/logomark-default-text-white.png" width="200"><img class="theme-light-only" src="https://storage.googleapis.com/sourcegraph-assets/cody/20230417/logomark-default-text-black.png" width="200"><div style="display:none">Cody</div></picture>

<aside class="beta">
<p>
<span class="badge badge-beta">Beta</span>
Cody is currently available in beta, and we value your feedback! You can <a href="https://about.sourcegraph.com/contact">contact us</a> directly, file an  <a href="https://github.com/sourcegraph/sourcegraph">issue</a>, or <a href="https://twitter.com/sourcegraph">tweet</a>.
</p>
</aside>

## What is Cody?

Cody is an AI coding assistant tool that utilizes Sourcegraph's <a href="https://docs.sourcegraph.com/cody/explanations/code_graph_context"> code graph</a> and Large Language Models (LLMs) to write code and provide answers based on your codebase and code graph.

Think of Cody as your personal coding assistant with a comprehensive understanding of your open source code, StackOverflow questions, and entire codebase.

Cody helps you answer questions, write code, and offer suggestions for code improvement.

## Getting Started

Start using Cody by one of the following

<div class="getting-started">
  <a class="btn btn-primary text-left" href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai"><div class="get-started"><img alt="VSCode" src="vscode.svg"></img></div><b>VS Code Extension</b><p>Install Cody's free and open source IDE extension for VS Code.</p></a>
    <a class="btn btn-primary text-left" href="https://plugins.jetbrains.com/plugin/9682-cody-ai-by-sourcegraph"><div class="get-started"><img alt="JetBrains" src="jb_beam.svg"></img></div><b>JetBrains Extension (experimental)</b><p>Install Cody's free and open source IDE extension for JetBrains.</p></a>
</div>

<div class="getting-started">
  <a class="btn btn-primary text-left" href="https://sourcegraph.com/get-cody"><div class="get-started"><img alt="Cody App" src="cody-logomark-default.svg"></img></div><b>Cody App</b><p>Install the free desktop app to try Cody with your local codebase.</p></a>
  <a class="btn btn-primary text-left" href="https://about.sourcegraph.com/cody/pricing"><div class="get-started"><img alt="Enterprise" src="enterprise.png"></img></div><b>Cody Enterprise</b><p>Get in touch with our team to try Cody for Sourcegraph Enterprise.</p></a>
</div>

## Main Features

Some of the main Cody features include:

<!-- NOTE: These should stay roughly in sync with client/cody/README.md, although these need to be not specific to VS Code. -->

| Feature | Description |
| ----------- | ----------- |
| Code chatbot | Your AI-powered code assistant. It fits your project's coding conventions and architecture, unlike other AI code chatbots. Chat with Cody from your code editor or the Sourcegraph sidebar.
| Fix code inline | Cody can help you make interactive edits and refactor code by following natural-language instructions.
| Recipes | Cody uses its codebase awareness to produce unit tests, documentation, and more. Use our pre-built recipes to simplify development with a few clicks.
| Autocomplete | Cody provides context-based code auto-completion while typing. Predictions make coding easier.

## Join our Community

If you have any questions regarding Cody, you can always ask our community on [GitHub Discussions](https://github.com/sourcegraph/sourcegraph/discussions), [Discord](https://discord.com/invite/s2qDtYGnAE), or [Twitter](https://twitter.com/sourcegraph).

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
  <a class="btn text-left" href="explanations/use_cases"><b>Cody Use Cases</b><p>Explore some of the handy use cases with Cody and try them yourself.</p></a>
</div>
<div class="getting-started">
   <a class="btn text-left" href="troubleshooting"><b>Cody Troubleshooting Guide</b><p>Having trouble with Cody? Review our troubleshooting guide for help.</p></a>
  <a class="btn text-left" href="faq"><b>FAQs</b><p>Learn about some of the frequently asked questions about Cody.</p></a>
</div>
