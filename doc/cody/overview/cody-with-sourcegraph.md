<style>

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

.limg a {
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

</style>

# Enabling Cody with Sourcegraph.com

<p class="subtitle">Learn how to connect and use Cody with the Sourcegraph instance.</p>

Cody uses Sourcegraph to provide contextually relevant answers and code. To install and use Cody with your [code editor](./../overview/install-vscode.md), you must connect it to your Sourcegraph.com account. For more advanced and customized usage, it's recommended to [enable Cody for Enterprise](enable-cody-enterprise.md).

<ul class="limg">
  <li>
    <a class="card text-left" target="_blank" href="https://sourcegraph.com/cody">
      <h3><img alt="JetBrains" src="https://sourcegraph.com/.assets/img/sourcegraph-mark.svg" />Cody with Sourcegraph.com</h3>
      <p>Connect Cody with your Sourcegraph account.</p>
    </a>
  </li>
  </ul>

## Initial setup

- [Create a Sourcegraph.com account](https://sourcegraph.com/sign-up)
- Install [the Cody VS Code extension](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai)
- Open the Cody extension
- Click on **Other Sign In Options...** and select Sign in to Sourcegraph.com
- Follow the prompts to authorize Cody to access your Sourcegraph.com account

You've now successfully connected Cody to your Sourcegraph.com instance.

## Configure code graph context for code-aware answers

After connecting Cody to Sourcegraph.com, you can optionally use [code graph context](./../core-concepts/code-graph.md) to improve Cody's context of existing code. Note that code graph context is only available for public repositories on sourcegraph.com, which have embeddings.

You can view the [list of repositories with embeddings here](../embedded-repos.md). To add any of these to your dev environment, contact a Sourcegraph team member via [Discord](https://discord.gg/8wJF5EdAyA) to get help with the access and setup.

To use Cody with code graph on private code, it's recommended to [enable Cody for Enterprise](enable-cody-enterprise.md).

### Enable code graph context

The `Cody: Codebase` setting in VS Code enables codebase-aware answers for the Cody extension. Enter the repository's name with embeddings, and Cody can provide more accurate and relevant answers to your coding questions based on that repository's content. To configure this setting in VS Code:

- Open the **Cody Extension Settings**
- Search for the `Cody: Codebase`
- Enter the repository name
- For example: `github.com/sourcegraph/sourcegraph` without the `https` protocol

## More benefits

Read more about [Cody capabilities](./../capabilities.md) to learn about all the features it provides to boost your development productivity.

## More resources

For more information on what to do next, we recommend the following resources:

<div class="cards">
  <a class="card text-left" href="./../quickstart"><b>Cody Quickstart</b><p>This guide recommends how to use Cody once you have installed the extension in your VS Code editor.</p></a>
  <a class="card text-left" href="./../use-cases"><b>Cody Use Cases</b><p>Explore some of the most common use cases of Cody that helps you with your development workflow.</p></a>
</div>
