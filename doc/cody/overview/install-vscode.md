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
</style>

# Installing Cody in VS Code

<p class="subtitle">Learn how to use Cody and its features with the VS Code editor.</p>

The Cody AI extension by Sourcegraph enhances your coding experience in VS Code by providing intelligent code sugsgestions, context-aware completions, and advanced code analysis. This guide will walk you through the steps to install and set up the Cody within your VS Code environment.

## Prerequisites

- You have the latest version of [VS Code](https://code.visualstudio.com/) installeds
- You have enabled an instance for [Cody from your Sourcegraph.com](overview/cody-with-sourcegraph.md) account

## Install the VS Code extension

Follow these steps to install the Cody AI extension for VS Code:

- Open VS Code editor on your local machine
- Click the "Extensions" icon in the Activity Bar on the side of VS Code, or use the keyboard shortcut `Cmd+Shift+X` (macOS) or `Ctrl+Shift+X` (Windows/Linux)
- Type "Cody AI" in the search bar and press "Enter"
- Click on the "Install" button next to the "Cody AI" by Sourcegraph
- After installing the extension, you may be prompted to restart VS Code to activate the extension

Alternatively, you can also [download and install the extension from the VS Code Marketplace][cody-vscode-marketplace] directly.

## Connect the extension to Sourcegraph

Next, open the VS Code extension and configure it to connect to a Sourcegraph instance (either an enterprise instance or Sourcegraph.com).

### For Sourcegraph enterprise users

Log in to your Sourcegraph instance and go to `settings` / `access token` (`https://<your-instance>.sourcegraph.com/users/<your-instance>/settings/tokens`). From here, generate a new access token.

Then, you will paste your access token and instance address in to the Cody extension.

### For Sourcegraph.com users

Click `Continue with Sourcegraph.com` in the Cody extension. From there, you'll be taken to Sourcegraph.com, which will authenticate your extension.

## Verifying the installation

Once connected, you should see the Cody icon on the left side of VS Code. Click it, and a panel will open up. To verify that the Cody AI extension has been successfully installed and is working as expected:

- Open a file in a supported programming language like JavaScript, Python, Go, etc.
- As you start typing, Cody AI should begin providing intelligent suggestions and context-aware completions based on your coding patterns and the context of your code

## Enable code graph context for context-aware answers (Optional)

You can optionally configure code graph content, which gives Cody the ability to provide context-aware answers. For example, Cody can write example API calls if has context of a codebase's API schema.

Learn more about how to:

- [Configure code graph context for Sourcegraph.com][cody-with-sourcegraph-config-graph]
- [Configure code graph context for Sourcegraph Enterprise][enable-cody-enterprise-config-graph]

## Updating the extension

VS Code will typically notify you when updates are available for installed extensions. Follow the prompts to update the Cody AI extension to the latest version.

## Congratulations!

Congratulations! You've successfully installed the Cody AI extension for VS Code. Enjoy the benefits of intutive code autocompletions, intelligent code suggestions and enhanced coding productivity in your development workflow.

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
