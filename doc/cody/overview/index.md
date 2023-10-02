<style>
.limg {
  list-style: none;
  margin: 3rem 0 !important;
  padding: 0 !important;
}
.limg li {
  margin-bottom: 1rem;
  padding: 0 !important;
}

.limg li:last {
  margin-bottom: 0;
}

.limg a {
    display: flex;
    flex-direction: column;
    transition-property: all;
   transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
     transition-duration: 350ms;
     border-radius: 0.75rem;
  padding-top: 1rem;
  padding-bottom: 1rem;

}

.limg a:hover {
  padding-left: 1rem;
  padding-right: 1rem;
  background: rgb(113 220 232 / 19%);
}

.limg p {
  margin: 0rem;
}
.limg a img {
  width: 1rem;
}

.limg h3 {
  display:flex;
  gap: 0.6rem;
  margin-top: 0;
  margin-bottom: .25rem

}

th:first-child,
td:first-child {
   min-width: 200px;
}

.markdown-body table thead tr{
  border-top:0;
}

.markdown-body table th, .markdown-body table td {
    text-align: left;
    vertical-align: baseline;
    padding: 0.5714286em;
}

.markdown-body table tr:nth-child(2n) {
  background: unset;
}

.markdown-body table th, .markdown-body table td {
    border: none;
}

.markdown-body .cards {
  display: flex;
  align-items: stretch;
}

.markdown-body .cards .card {
  flex: 1;
  margin: 0.5em;
  color: var(--text-color);
  border-radius: 4px;
  border: 1px solid var(--sidebar-nav-active-bg);
  padding: 1.5rem;
  padding-top: 1.25rem;
}

.markdown-body .cards .card:hover {
  color: var(--link-color);
}

.markdown-body .cards .card span {
  color: var(--link-color);
  font-weight: bold;
}
</style>

# <picture title="Cody"><img class="theme-dark-only" alt="Cody" src="https://storage.googleapis.com/sourcegraph-assets/cody/20230417/logomark-default-text-white.png" width="200"><img class="theme-light-only" alt="Cody" src="https://storage.googleapis.com/sourcegraph-assets/cody/20230417/logomark-default-text-black.png" width="200"><div style="display:none">Cody</div></picture>

<aside class="beta">
<p>
<span class="badge badge-beta">Beta</span>
Cody is currently available in beta.

<br />
We value your feedback! You can <a href="https://about.sourcegraph.com/contact">contact us</a> directly, file an  <a href="https://github.com/sourcegraph/cody/issues">issue</a>, Join our <a href="https://discord.com/servers/sourcegraph-969688426372825169">Discord</a>, or <a href="https://twitter.com/sourcegraphcody">tweet</a> to share feedback.
</p>
</aside>

## What is Cody?

Cody is a free, open-source AI coding assistant that writes, understands, and fixes code. Cody is powered by Sourcegraph's <a href="https://docs.sourcegraph.com/cody/explanations/code_graph_context">code graph</a> and uses knowledge of your entire codebase to give more accurate answers and generate idiomatic code.

Think of Cody as your personal AI coding assistant, equipped with a comprehensive understanding of three crucial elements:

1. Your entire codebase
2. Vast knowledge of open source code
3. Extensive training data for code understanding and problem-solving

## Getting started

To start using Cody, pick one of the following:

<ul class="limg">
  <li>
    <a class="card text-left" target="_blank" href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai">
    <h3><img alt="VS Code" src="https://storage.googleapis.com/sourcegraph-assets/docs/images/cody/vscode.svg"/> Cody: VS Code Extension</h3>
    <p>Install Cody's free and open source extension for VS Code.</p>
    </a>
  </li>
  <li>
    <a class="card text-left" target="_blank" href="https://plugins.jetbrains.com/plugin/9682-cody-ai-by-sourcegraph">
      <h3><img alt="JetBrains" src="https://storage.googleapis.com/sourcegraph-assets/docs/images/cody/jb_beam.svg" />JetBrains Extension (experimental)</h3>
      <p>Install Cody's free and open source extension for JetBrains.</p>
    </a>
  </li>
  <li>
     <a class="card text-left" target="_blank" href="https://sourcegraph.com/get-cody">
      <h3><img alt="Cody App" src="https://storage.googleapis.com/sourcegraph-assets/docs/images/cody/cody-logomark-default.svg"/>Cody App</h3>
      <p>Free Cody desktop app to try Cody with your local codebase.</p>
      </a>
  </li>
  <li>
    <a class="card text-left" target="_blank" href="https://about.sourcegraph.com/cody/pricing">
      <h3><img alt="Cody Enterprise" src="https://sourcegraph.com/.assets/img/sourcegraph-mark.svg" />Cody Enterprise</h3>
      <p>Get in touch with our team to try Cody for Sourcegraph Enterprise.</p>
    </a>
  </li>
</ul>

## Main features

Cody's main features include:

<!-- NOTE: These should stay roughly in sync with client/cody/README.md, although these need to be not specific to VS Code. -->

|     Feature     |                                                                                         Description                                                                                         |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [Autocomplete](./../capabilities.md#autocomplete)    | Cody makes context-based code autocompletions. Cody can autocomplete single lines or whole functions in any programming language, configuration file, or documentation.
| [Chat](./../capabilities.md#chat)    | Ask Cody questions in the chat view or inline with code, and it will use Sourcegraph’s code graph to answer using knowledge of your codebase. |
| [Commands](./../capabilities.md#commands)    | Cody has quick commands for common actions, such as adding code documentation, generating unit tests, and detecting code smells. |

## Join our community

If you have any questions regarding Cody, you can always ask our community on [GitHub Discussions](https://github.com/sourcegraph/cody/discussions), [Discord](https://discord.com/invite/s2qDtYGnAE), or [Twitter](https://twitter.com/sourcegraph).

## Explanations

- [Cody clients, plugins, and extensions](./../explanations/cody_clients.md)
- [Enabling Cody for Sourcegraph Enterprise customers](./../overview/enable-cody-enterprise.md)
- [Enabling Cody for the Cody app](../overview/app/index.md)
- [Enabling Cody for open source Sourcegraph.com users](./../overview/cody-with-sourcegraph.md)
- [Installing the Cody VS Code extension](./../overview/install-vscode.md)
- [Installing the Jetbrains extension (experimental)](./../overview/install-jetbrains.md)
- [Configuring code graph context](./../explanations/code_graph_context.md)
- [Sourcegraph Cody Gateway](./../explanations/cody_gateway.md)

## More resources

For more information on what to do next, we recommend the following resources:

<div class="cards">
  <a class="card text-left" href="./../quickstart"><b>Cody Quickstart</b><p>This guide recommends first things to try once Cody is up and running.</p></a>
  <a class="card text-left" href="./../use-cases"><b>Cody Use Cases</b><p>Explore some of the handy use cases with Cody and try them yourself.</p></a>
</div>
<div class="cards">
   <a class="card text-left" href="./../troubleshooting"><b>Cody Troubleshooting Guide</b><p>Having trouble with Cody? Review our troubleshooting guide for help.</p></a>
  <a class="card text-left" href="./../faq"><b>FAQs</b><p>Learn about some of the frequently asked questions about Cody.</p></a>
</div>
