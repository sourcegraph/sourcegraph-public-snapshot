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

# Installing Cody in VS Code

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
- Click the **Extensions** icon in the Activity Bar on the side of VS Code, or use the keyboard shortcut `Cmd+Shift+X` (macOS) or `Ctrl+Shift+X` (Windows/Linux)
- Type **Cody AI** in the search bar and click the **Install** button
- After installing, you may be prompted to reload VS Code to activate the extension

![install-vscode-extension](https://storage.googleapis.com/sourcegraph-assets/Docs/install-cody-vscode.png)

Alternatively, you can also [download and install the extension from the VS Code Marketplace][cody-vscode-marketplace] directly.

## Connect the extension to Sourcegraph

After a successful installation, the Cody icon appears in the [Activity sidebar](https://code.visualstudio.com/api/ux-guidelines/activity-bar). Clicking it prompts you to start with codehosts like GitHub, GitLab, and your Google login. This allows Cody to access your Sourcegraph.com account.

![cody-sign-flow](https://storage.googleapis.com/sourcegraph-assets/Docs/cody-signin-vscode.png)

You can use Sourcegraph Enterprise with the Cody VS Code extension. Click the **Sign in to Enterprise Instance** at the bottom of the Cody panel, and it connects to your enterprise environment.

## Verifying the installation

Once connected, click the Cody icon from the sidebar again, and a panel will open. To verify that the Cody extension has been successfully installed and is working as expected, let's create an autocomplete suggestion.

Cody provides intelligent code suggestions and context-aware autocompletions for numerous programming languages like JavaScript, Python, TypeScript, Go, etc.

- Create a new file in VS Code for example, `code.js`
- Next, type the following algorithm function to sort an array of numbers

```js
function bubbleSort(array){
```

- As you start typing, Cody will automatically provide suggestions and context-aware completions based on your coding patterns and the code context
- These autocomplete suggestions appears as grayed text. To accept the suggestion, press the `Tab` key

<video width="1920" height="1080" loop playsinline controls style="width: 100%; height: auto; max-width: 50rem">
  <source src="https://storage.googleapis.com/sourcegraph-assets/Docs/Media/cody-in-action.mp4" type="video/mp4">
</video>

## Commands

Cody offers quick, ready-to-use [Commands](./../capabilities.md#commands) for common actions to write, describe, fix, and smell code. These allow you to run predefined actions with smart context-fetching anywhere in the editor, like:

- `/ask`: Asks a question
- `/edit[instruction]`: Edits code
- `/doc`: Generates code documentation
- `/explain`: Explains code
- `/test`: Generates unit tests
- `/smell`: Find code smells
- `/reset`: Clears the Cody chat

Let's understand how the `/doc` command generates code documentation for a function.

<video width="1920" height="1080" loop playsinline controls style="width: 100%; height: auto; max-width: 50rem">
  <source src="https://storage.googleapis.com/sourcegraph-assets/Docs/Media/vscode-doc-command.mp4" type="video/mp4">
</video>

In addition, to support customization and advanced use cases, you can create Custom Commands tailored to your requirements. Custom Commands are currently supported by Cody for the VS Code extension version 0.8 and above.

[Learn more about Custom Commands here →](./../custom-commands.md)

## Enable code graph context for context-aware answers (Optional)

After connecting Cody extension to Sourcegraph.com, you can optionally use [Code Graph Context](./../core-concepts/code-graph.md) to improve Cody's context of existing code. Note that Code Graph Context is only available for public repositories on Sourcegraph.com, which have embeddings.

You can view the [list of repositories with embeddings here](../embedded-repos.md). To add any of these to your dev environment, contact a Sourcegraph team member via [Discord](https://discord.gg/8wJF5EdAyA) to get help with the access and setup.

To use Cody with code graph on private code, it's recommended to [enable Cody for Enterprise](enable-cody-enterprise.md).

### Configure Code Graph Context

The `Cody: Codebase` setting in VS Code enables codebase-aware answers for the Cody extension. Enter the repository's name with embeddings, and Cody can provide more accurate and relevant answers to your coding questions based on that repository's content. To configure this setting in VS Code:

- Open the **Cody Extension Settings**
- Search for the `Cody: Codebase`
- Enter the repository name
- For example: `github.com/sourcegraph/sourcegraph` without the `https` protocol

Learn more about how to:

- [Configure code graph context for Sourcegraph.com][cody-with-sourcegraph-config-graph]
- [Configure code graph context for Sourcegraph Enterprise][enable-cody-enterprise-config-graph]

## Updating the extension

VS Code will typically notify you when updates are available for installed extensions. Follow the prompts to update the Cody AI extension to the latest version.

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
