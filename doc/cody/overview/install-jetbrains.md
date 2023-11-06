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

# Install Cody for JetBrains

<p class="subtitle">Learn how to use Cody and its features with the JetBrains IntelliJ editor.</p>

<aside class="beta">
<p>
<span style="margin-right:0.25rem;" class="badge badge-beta">Beta</span> Cody support for JetBrains is in the Beta stage.
<br />
For any feedback, you can <a href="https://about.sourcegraph.com/contact">contact us</a> directly, file an <a href="https://github.com/sourcegraph/cody/issues">issue</a>, join our <a href="https://discord.com/servers/sourcegraph-969688426372825169">Discord</a>, or <a href="https://twitter.com/sourcegraphcody">tweet</a>.
</p>
</aside>

The Cody extension by Sourcegraph enhances your coding experience in your IDE by providing intelligent code suggestions, context-aware completions, and advanced code analysis. This guide will walk you through the steps to install and set up the Cody within your JetBrains environment.

<ul class="limg">
  <li>
    <a class="card text-left" target="_blank" href="https://plugins.jetbrains.com/plugin/9682-cody-ai-by-sourcegraph">
      <h3><img alt="JetBrains" src="https://storage.googleapis.com/sourcegraph-assets/docs/images/cody/jb_beam.svg" />JetBrains Extension (Beta)</h3>
      <p>Install Cody's free and open source extension for JetBrains.</p>
    </a>
  </li>
  </ul>

## Prerequisites

- You have the latest version of <a href="https://www.jetbrains.com/idea/" target="_blank">JetBrains IDEs</a> installed
- You have enabled an instance for [Cody from your Sourcegraph.com](cody-with-sourcegraph.md) account
- Cody is compatible with the following JetBrains IDEs:
  - [Android Studio](https://developer.android.com/studio)
  - [AppCode](https://www.jetbrains.com/objc/)
  - [CLion](https://www.jetbrains.com/clion/)
  - [GoLand](https://www.jetbrains.com/go/)
  - [IntelliJ IDEA](https://www.jetbrains.com/idea/) (Community and Ultimate editions)
  - [PhpStorm](https://www.jetbrains.com/phpstorm/)
  - [PyCharm](https://www.jetbrains.com/pycharm/) (Community and Professional editions)
  - [Rider](https://www.jetbrains.com/rider/)
  - [RubyMine](https://www.jetbrains.com/ruby/)
  - [WebStorm](https://www.jetbrains.com/webstorm/)

## Install the JetBrains IntelliJ Cody extension

Follow these steps to install the Cody extension for JetBrains IntelliJ:

- Open JetBrains IntelliJ editor on your local machine
- Open **Settings** (Mac: `⌘+,` Windows: `Ctrl+Alt+S`) and select **Plugins**
- Type and search **Sourcegraph Cody + Code Search** extension and click **Install**

![cody-for-intellij](https://storage.googleapis.com/sourcegraph-assets/Docs/Media/cody-for-intellij.png)

Alternatively, you can also [Download and install the extension from the Jetbrains marketplace](https://plugins.jetbrains.com/plugin/9682-sourcegraph).

## Connect the extension to Sourcegraph

After a successful installation, Cody's icon appears in the sidebar. Clicking it prompts you to start with codehosts like GitHub, GitLab, and your Google login. This allows Cody to access your Sourcegraph.com account.

Alternatively, you can also click the **Sign in with and Enterprise Instance** to connect to your enterprise instance.

![cody-for-intellij-login](https://storage.googleapis.com/sourcegraph-assets/Docs/Media/sign-in-cody-jb.png)

### Context Selection

Cody automatically understands your codebase context based on the project opened in your workspace. However, at any point, you can override the automatic “codebase detection” by clicking on the repo name below the Cody chat and then adding the Git URL. This will allow Cody to start using the codebase context you’ve selected.



### For Sourcegraph enterprise users

Log in to your Sourcegraph instance and go to `settings` / `access token` (`https://<your-instance>.sourcegraph.com/users/<your-username>/settings/tokens`). From here, generate a new access token.

Then, you select the option to `Use an enterprise instance` and you will paste your access token and instance URL address in to the Cody extension.

### For Sourcegraph.com users

Click `Continue with Sourcegraph.com` in the Cody extension. From there, you'll be taken to Sourcegraph.com, which will authenticate your extension.

## Embeddings

For free users, embeddings are supported for the open source repos on Sourcegraph.com. Enterprise users are guided to contact their admin for a more customized embeddings experience.

## Verifying the installation

Once connected, click the Cody icon from the sidebar again, and a panel will open. To verify that the Cody extension has been successfully installed and is working as expected:

- Open a file in a supported programming language like JavaScript, Python, Go, etc.
- As you start typing, Cody should begin providing intelligent suggestions and context-aware completions based on your coding patterns and the context of your code

## Commands

The Cody JetBrains IntelliJ extension also supports pre-built reusable prompts called "Commands" that help you quickly get started with common programming tasks like:

- Explain code
- Generate unit test
- Generate docstring
- Improve variable names
- Translate to different language
- Summarize recent code changes
- Detect code smells
- Generate release notes

## Autocomplete

Autocomplete suggestions appear as inlay suggestions. Press `Option+\` (for macOS) or `Alt+\` (for Windows) to manually trigger the autocomplete. Press `tab` to accept a suggestion and `Escape` to reject it.

Autocomplete is triggered by default. You can disable the automatic trigger optionally through the Cody icon in the status bar or from the Cody Settings.

In addition, autocomplete suggestions use the same color as inline parameter hints according to your configured theme. You can optionally customize this color from the Cody settings.

## Supported JetBrains IDEs

The Cody plugin works with all JetBrains IDEs, including:

- IntelliJ IDEA
- IntelliJ IDEA Community Edition
- PhpStorm
- WebStorm
- PyCharm
- PyCharm Community Edition
- RubyMine
- AppCode
- CLion
- GoLand
- DataGrip
- Rider
- Android Studio

## Enable code graph context for context-aware answers (optional)

You can optionally configure code graph content, which gives Cody the ability to provide context-aware answers. For example, Cody can write example API calls if has context of a codebase's API schema.

Learn more about how to:

- [Configure code graph context for Sourcegraph.com](cody-with-sourcegraph.md#configure-code-graph-context-for-code-aware-answers)
- [Configure code graph context for Sourcegraph Enterprise](enable-cody-enterprise.md#enabling-codebase-aware-answers)

## Updating the extension

JetBrains IntelliJ will typically notify you when updates are available for installed extensions. Follow the prompts to update the Cody extension to the latest version.

## More benefits

Read more about [Cody capabilities](./../capabilities.md) to learn about all the features it provides to boost your development productivity.

## More resources

For more information on what to do next, we recommend the following resources:

<div class="cards">
  <a class="card text-left" href="./../quickstart"><b>Cody Quickstart</b><p>This guide recommends how to use Cody once you have installed the extension in your VS Code editor.</p></a>
  <a class="card text-left" href="./../use-cases"><b>Cody Use Cases</b><p>Explore some of the most common use cases of Cody that helps you with your development workflow.</p></a>
</div>
