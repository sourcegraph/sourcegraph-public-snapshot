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

The Cody extension by Sourcegraph enhances your coding experience in VS Code by providing intelligent code suggestions, context-aware autocomplete, and advanced code analysis. This guide will walk you through the steps to install and set up Cody within your VS Code environment.

<ul class="limg">
  <li>
    <a class="card text-left" target="_blank" href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai">
    <h3><img alt="VS Code" src="https://storage.googleapis.com/sourcegraph-assets/docs/images/cody/vscode.svg"/> Cody: VS Code Extension</h3>
    <p>Install Cody's free extension for VS Code.</p>
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

![install-vscode-extension](https://storage.googleapis.com/sourcegraph-assets/Docs/cody-new-ui.png)

Alternatively, you can also [download and install the extension from the VS Code Marketplace][cody-vscode-marketplace] directly.

## Connect the extension to Sourcegraph

After a successful installation, the Cody icon appears in the [Activity sidebar](https://code.visualstudio.com/api/ux-guidelines/activity-bar). Users on free Cody tier can sign in to their Sourcegraph.com accounts through GitHub, GitLab, or Google.

![cody-sign-flow](https://storage.googleapis.com/sourcegraph-assets/Docs/cody-signin-vscode.png)

You can use Sourcegraph Enterprise with the Cody VS Code extension. Click the **Sign In to Your Enterprise Instance**, and it connects to your enterprise environment.

## Verifying the installation

Once connected, click the Cody icon from the sidebar again. The Cody extension will open in a configurable side panel.

![code-panel](https://storage.googleapis.com/sourcegraph-assets/Docs/cody-nw-panel-.png)

Let's create an autocomplete suggestion to verify that the Cody extension has been successfully installed and is working as expected.

Cody provides intelligent code suggestions and context-aware autocompletions for numerous programming languages like JavaScript, Python, TypeScript, Go, etc.

- Create a new file in VS Code, for example, `code.js`
- Next, type the following algorithm function to sort an array of numbers

```js
function bubbleSort(array){
```

- As you start typing, Cody will automatically provide suggestions and context-aware completions based on your coding patterns and the code context
- These autocomplete suggestions appear as grayed text. To accept the suggestion, press the `Tab` key

<video width="1920" height="1080" loop playsinline controls style="width: 100%; height: auto; max-width: 50rem">
  <source src="https://storage.googleapis.com/sourcegraph-assets/Docs/Media/cody-in-action.mp4" type="video/mp4">
</video>

## Chat

Cody chat in VS Code is available in a tab next to your code. Once connected to Sourcegraph, a **New Chat** button opens the chat window on the right. You can have multiple Cody Chats going simultaneously in separate tabs.

All previous and existing chats are stored under the chats panel on the left. You can download these to share or use later in a `.json` file, or delete them altogether.

### Enhanced Context Selector

Cody's Enhanced Context enables Cody to leverage search and embeddings-based context. It's enabled by default, though embeddings for Enterprise users are controlled by administrators and community users will need to generate embeddings for their projects by clicking the icon next to the chat input. The icon is also where you can disable Enhanced Context if you'd like more granular control of Cody's context by tagging `@-file` or `@#-symbol` in the chat input.

![](https://storage.googleapis.com/sourcegraph-assets/Docs/enhanced-context.png)

The following tables shows what happens when Enhanced Context Selection is enabled or disabled.

|                          | Opened Files                 | Highlighted Code            | Embeddings (If available)  | Symf (as backup)            | Graph Context (BFG)         |
|--------------------------|------------------------------|-----------------------------|-----------------------------|-----------------------------|-----------------------------|
| Enhanced Context Enabled  | ✅                           | ✅                          | ✅                          | ✅                          | ❌                          |
| Enhanced Context Disabled | ❌                           | ❌                          | ❌                          | ❌                          | ❌                          |


### LLM Selection

Cody Community users can choose the LLM they'd like Cody to use for chats right within the chat panel. The default LLM is set to Claude 2.0 by Anthropic, but the drop-down allows you to experiment with different LLMs and choose the one that's best for you. For Enterprise users, the LLM is determined by the administrator and cannot be changed within the editor.

![](https://storage.googleapis.com/sourcegraph-assets/Docs/llm-select.png)

## Commands

Cody offers quick, ready-to-use [Commands](./../capabilities.md#commands) for common actions to write, describe, fix, and smell code. These allow you to run predefined actions with smart context-fetching anywhere in the editor, like:

- Ask Cody a question
- Add code documentation
- Edit code with instructions
- Explain code
- Identify code smells
- Generate unit tests
- Custom commands

Let's understand how the `/doc` command generates code documentation for a function.

<video width="1920" height="1080" loop playsinline controls style="width: 100%; height: auto; max-width: 50rem">
  <source src="https://storage.googleapis.com/sourcegraph-assets/Docs/Media/code-comments-cody.mp4" type="video/mp4">
</video>

In addition, you can also select which files and symbols to add as additional context. Type `@` to include files of your choice. Currently, only local files are available for this feature to work.

The file paths are relative to your workspace, and you can start with the root folder and type out the rest of the path, for example `src/util/<YOUR-FILE>`.

### Custom Commands

In addition, to support customization and advanced use cases, you can create **Custom Commands** tailored to your requirements. Custom Commands are currently supported by Cody for the VS Code extension version 0.8 and above.

[Learn more about Custom Commands here →](./../custom-commands.md)

## Cody VS Code Actions

Cody VS Code extension users can also use the **Code Actions** feature to `fix`, `explain`, and `edit` code. These Code Actions are triggered by the following:

- Ask Cody to Fix
- Ask Cody to Explain
- Ask Cody to Edit

These Code Actions can be initiated by clicking the **lightbulb** icon in your code file. For example, whenever there's an error in code syntax, and you make a mistake while writing code, Cody's Code Actions come into play, and a red warning triggers. Along this appears the lightbulb icon. Click this lightbulb icon in the project file.

- Select **Ask Cody to fix** option
- **Cody is working** notice will appear and provide a quick fix with options for **Edits Applied**, **Retry**, **Undo**, and **Done**
- If you are satisfied with the fix, click **Edits Applied**
- To verify the applied changes, you can see a diff view of the fix in a new tab
- If you are not satisfied with the fix, you can **Retry** or **Undo** the changes

Here's a demo that shows how Code Actions work to fix an error:

<video width="1920" height="1080" loop playsinline controls style="width: 100%; height: auto; max-width: 50rem">
  <source src="https://storage.googleapis.com/sourcegraph-assets/Docs/Media/code-actions.mp4" type="video/mp4">
</video>

A similar process applies to explain and edit Code Actions.

## Enable code graph context for context-aware answers (Optional)

After connecting Cody's extension to Sourcegraph.com, you can optionally use [Code Graph Context](./../core-concepts/code-graph.md) to improve Cody's context of existing code. Code Graph Context is only available for public repositories on Sourcegraph.com, which have embeddings.

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
