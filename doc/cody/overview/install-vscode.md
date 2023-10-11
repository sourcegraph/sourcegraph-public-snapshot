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

# Installing Cody in VS Code <span class="badge badge-experimental" style="margin-left: 0.5rem; vertical-align:middle;">Beta</span>

<p class="subtitle">Learn how to use Cody and its features with the VS Code editor.</p>

The Cody extension by Sourcegraph enhances your coding experience in VS Code by providing intelligent code suggestions, context-aware autocomplete, and advanced code analysis. This guide will walk you through the steps to install and set up the Cody within your VS Code environment.

<ul class="limg">
  <li>
    <a class="card text-left" target="_blank" href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai">
    <h3><img alt="VS Code" src="https://storage.googleapis.com/sourcegraph-assets/docs/images/cody/vscode.svg"/> Cody: VS Code Extension</h3>
    <p>Install Cody's free and open source extension for VS Code.</p>
    </a>
  </li>
</ul>

## Prerequisites

- You have the latest version of [VS Code](https://code.visualstudio.com/) installed
- You have enabled an instance for [Cody from your Sourcegraph.com](cody-with-sourcegraph.md) account

## Install the VS Code extension

Follow these steps to install the Cody AI extension for VS Code:

- Open VS Code editor on your local machine
- Click the "Extensions" icon in the Activity Bar on the side of VS Code, or use the keyboard shortcut `Cmd+Shift+X` (macOS) or `Ctrl+Shift+X` (Windows/Linux)
- Type "Cody AI" in the search bar and press "Enter"
- Click on the "Install" button next to the "Cody AI" by Sourcegraph
- After installing the extension, you may be prompted to restart VS Code to activate the extension

Alternatively, you can also [download and install the extension from the VS Code Marketplace][cody-vscode-marketplace] directly.

## Connect the extension to Sourcegraph

After a successful installation, the Cody icon appears in the [Activity sidebar](https://code.visualstudio.com/api/ux-guidelines/activity-bar). Clicking it prompts you to start with codehosts like GitHub, GitLab, and your Google login. This allows Cody to access your Sourcegraph.com account.

![cody-sign-flow](https://storage.googleapis.com/sourcegraph-assets/Docs/cody-sign-in.png)

You can also connect with your Sourcegraph Enterprise Instance.

## Verifying the installation

Once connected, click the Cody icon from the sidebar again, and a panel will open. To verify that the Cody extension has been successfully installed and is working as expected:

- Open a file in a supported programming language like JavaScript, Python, Go, etc.
- As you start typing, Cody should begin providing intelligent suggestions and context-aware completions based on your coding patterns and the context of your code

## Commands

Cody also supports executing reusable prompts known as **Commands** from within the [VS Code extension](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai). They allow you to run predefined actions with smart context-fetching anywhere in the editor, like:

- `/ask`: Asks a question
- `/edit[instruction]`: Edits code
- `/doc`: Generates code documentation
- `/explain`: Explains code
- `/test`: Generates unit tests
- `/smell`: Find code smells
- `/reset`: Clears the Cody chat

[Learn more about Commands here →](./../capabilities.md#commands)

## Enable code graph context for context-aware answers (Optional)

You can optionally configure code graph content, which gives Cody the ability to provide context-aware answers. For example, Cody can write example API calls if has context of a codebase's API schema.

Learn more about how to:

- [Configure code graph context for Sourcegraph.com][cody-with-sourcegraph-config-graph]
- [Configure code graph context for Sourcegraph Enterprise][enable-cody-enterprise-config-graph]

## Updating the extension

VS Code will typically notify you when updates are available for installed extensions. Follow the prompts to update the Cody AI extension to the latest version.

## More benefits

Read more about [Cody Capabilities](./../capabilities.md) to learn about all the features it provides to boost your development productivity.

## More resources

For more information on what to do next, we recommend the following resources:

<div class="cards">
  <a class="card text-left" href="./../quickstart"><b>Cody Quickstart</b><p>This guide recommends how to use Cody once you have installed the extension in your VS Code editor.</p></a>
  <a class="card text-left" href="https://docs.sourcegraph.com/cody/capabilities#commands"><b>Commands in VS Code</b><p>Explore how Cody supports reusable prompts called Commands from within the VS Code extension.</p></a>
</div>

[cody-with-sourcegraph]: cody-with-sourcegraph.md
[cody-with-sourcegraph-config-graph]: cody-with-sourcegraph.md#configure-code-graph-context-for-code-aware-answers
[enable-cody-enterprise]: enable-cody-enterprise.md
[enable-cody-enterprise-config-graph]: enable-cody-enterprise.md#enabling-codebase-aware-answers
[cody-vscode-marketplace]: https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai
