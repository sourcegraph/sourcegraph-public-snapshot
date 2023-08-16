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
</style>

# Installing Cody in VS Code

<p class="subtitle">Learn how to install and configure Cody VS Code extension.</p>

When working with Visual Studio Code, you have the ability to seamlessly integrate and implement suggestions from Cody by Sourcegraph right within your editor. This guide explains how to install Cody in VS Code across macOS, Windows, or Linux platforms.

<ul class="limg">
  <li>
    <a class="card text-left" target="_blank" href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai">
    <h3><img alt="VS Code" src="https://storage.googleapis.com/sourcegraph-assets/docs/images/cody/vscode.svg"/> Cody: VS Code Extension →</h3>
    <p>Install Cody's free and open source extension for VS Code.</p>
    </a>
  </li>
</ul>

## Prerequisites

- To use Cody in VS Code, make sure you have it installed. Read more information [here →](https://code.visualstudio.com/Download).
- A Sourcegraph instance with Cody enabled on it or a Sourcegraph.com account.

Learn more with the following resources:
If you haven't yet done this, read the following resources:

- [Enabling Cody for Sourcegraph Enterprise][enable-cody-enterprise]
- [Enabling Cody for Sourcegraph.com][cody-with-sourcegraph]

## Install the VS Code extension

You can install Cody in VS Code in 2 ways:

- Click the Extensions icon on the VS Code activity bar
- Search for "Cody AI"
- Install the extension directly to VS Code

Or:

- [Download and install the extension from the VS Code marketplace][cody-vscode-marketplace]

## Connect the extension to Sourcegraph

Next, you'll open the VS Code extension and configure it to connect to a Sourcegraph instance (either an enterprise instance or Sourcegraph.com).

**For Sourcegraph Enterprise users:**

Log in to your Sourcegraph instance and go to `settings` / `access token` (`https://<your-instance>.sourcegraph.com/users/<your-instance>/settings/tokens`). From here, generate a new access token.

Then, you will paste your access token and instance address in to the Cody extension.

**For Sourcegraph.com users:**

Click `Continue with Sourcegraph.com` in the Cody extension. From there, you'll be taken to Sourcegraph.com, which will authenticate your extension.

## (Optional) Enable code graph context for context-aware answers

You can optional configure code graph content, which gives Cody the ability to provide context-aware answers. For example, Cody can write example API calls if has context of a codebase's API schema.

- [Configure code graph context for Sourcegraph.com][cody-with-sourcegraph-config-graph]
- [Configure code graph context for Sourcegraph Enterprise][enable-cody-enterprise-config-graph]

## Get started with Cody

You're now ready to use Cody! See our recommendations for getting started with using Cody.

## Congratulations!

**You're now up-and-running with your very own AI code asisstant!** 🎉

[cody-with-sourcegraph]: cody-with-sourcegraph.md
[cody-with-sourcegraph-config-graph]: cody-with-sourcegraph.md#configure-code-graph-context-for-code-aware-answers
[enable-cody-enterprise]: enable-cody-enterprise.md
[enable-cody-enterprise-config-graph]: enable-cody-enterprise.md#enabling-codebase-aware-answers
[cody-vscode-marketplace]: https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai
