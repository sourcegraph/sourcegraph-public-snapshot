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

body.theme-dark img.toggle {
    filter: invert(100%);
}

img.toggle {
    width: 20px;
    height: 20px;
}

.toggle-container {
  border: 1px solid;
  border-radius: 3px;
  display: inline-flex;
  vertical-align: bottom;
}

</style>

<!-- # <picture title="Cody"><img class="theme-dark-only" alt="Cody" src="https://storage.googleapis.com/sourcegraph-assets/cody/20230417/logomark-default-text-white.png" width="200"><img class="theme-light-only" alt="Cody" src="https://storage.googleapis.com/sourcegraph-assets/cody/20230417/logomark-default-text-black.png" width="200"><div style="display:none">Cody</div></picture> -->

# Cody

<p class="subtitle">Learn how Cody understands your entire codebase and enhances your development process with features like autocomplete and commands.</p>

<aside class="beta">
<p>
<span class="badge badge-beta">Beta</span>
Cody is currently available in beta for all users.
</p>
</aside>

## What is Cody?

Cody is a free and open-source AI coding assistant that writes, fixes, and maintains your code. Cody understands your entire codebase by leveraging the power of [Code Graph](./../core-concepts/code-graph) to gather context, which assists you in writing accurate code.

<video width="1920" height="1080" loop playsinline controls style="width: 100%; height: auto; max-width: 50rem">
  <source src="https://storage.googleapis.com/sourcegraph-assets/Docs/Media/1.mp4" type="video/mp4">
</video>

Cody connects seamlessly with codehosts like <a target="blank" href="https://sourcegraph.com/get-cody">GitHub</a>, <a target="blank" href="https://gitlab.com/users/sign_in">GitLab</a> and IDEs like <a target="blank" href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai">VS Code</a> and <a target="blank" href="https://plugins.jetbrains.com/plugin/9682-sourcegraph-cody--code-search">JetBrains</a>. Once connected, Cody acts as your personal AI coding assistant, equipped with a comprehensive understanding of the following three crucial elements:

1. Your entire codebase
2. Vast knowledge of open source code
3. Extensive training data for code understanding and problem-solving

## Getting started

You can start using Cody with one of the following clients:

<ul class="limg">
  <li>
    <a class="card text-left" target="_blank" href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai">
    <h3><img alt="VS Code" src="https://storage.googleapis.com/sourcegraph-assets/docs/images/cody/vscode.svg"/> Cody for VS Code</h3>
    <p>Install Cody's free and open source extension for VS Code.</p>
    </a>
  </li>
  <li>
    <a class="card text-left" target="_blank" href="https://plugins.jetbrains.com/plugin/9682-cody-ai-by-sourcegraph">
      <h3><img alt="JetBrains" src="https://storage.googleapis.com/sourcegraph-assets/docs/images/cody/jb_beam.svg" />Cody for JetBrains (beta)</h3>
      <p>Install Cody's free and open source extension for JetBrains.</p>
    </a>
  </li>
    <li>
    <a class="card text-left" target="_blank" href="https://github.com/sourcegraph/sg.nvim">
      <h3><img alt="Neovim" src="https://storage.googleapis.com/sourcegraph-assets/Docs/neovim-logo.png" />Cody for Neovim (experimental)</h3>
      <p>Install Cody's free and open source extension for Neovim.</p>
    </a>
  </li>
  <li>
     <a class="card text-left" target="_blank" href="https://sourcegraph.com/get-cody">
      <h3><img alt="Cody App" src="https://storage.googleapis.com/sourcegraph-assets/docs/images/cody/cody-logomark-default.svg"/>Cody for Desktop</h3>
      <p>Free desktop app to try Cody with your local codebase.</p>
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
| [Commands](./../capabilities.md#commands)    | Cody offers quick ready-to-use commands for common actions, such as adding code documentation, generating unit tests, and detecting code smells. |

## What data is collected and how it's used?

Sourcegraph collects data to improve user experience in various deployments including self-hosted, cloud, Cody app, and Sourcegraph.com. This includes usage data and user feedback.

In the Cloud deployment, data is collected for providing the service, including user prompts and responses. In the Self-Hosted deployment, Sourcegraph doesn't access data unless shared through support channels. For Sourcegraph.com users, data is collected to enhance the user experience but not used to train models.

<a target="_blank" href="https://about.sourcegraph.com/terms/cody-notice">Read more about Cody Usage and Privacy policy here →</a>

## Compatible with Sourcegraph products

Cody is compatible to use with the other Sourcegraph products, like [Code Search](./../../code_search/index.md). You can use Cody's chat to ask questions about your codebase. When you run any search query, you'll find an **Ask Cody** button that takes you to Cody's default chat interface that you can use to ask questions about the codebase.

On a free tier, you can use Cody chat with Code Search on ten public and one private repository. For enterprise users, Cody Chat extends to repositories indexed by your site administrator.

[Read more in the Cody FAQs to learn more about such queries →](./../faq.md)

## Join our community

If you have any questions regarding Cody, you can always ask our community on [GitHub discussions](https://github.com/sourcegraph/cody/discussions), [Discord](https://discord.com/invite/s2qDtYGnAE), or [create a post on X](https://twitter.com/sourcegraph).

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
