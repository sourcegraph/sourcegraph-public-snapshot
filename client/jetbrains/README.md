<!-- Plugin description -->

# Sourcegraph Cody + Code Search

Use Sourcegraph’s AI assistant Cody and Sourcegraph Code Search directly from your JetBrains IDE.

- [Cody](https://about.sourcegraph.com/cody) is a free and [open-source](https://github.com/sourcegraph/cody) AI coding assistant that can write, understand and fix your code. Cody is powered by Sourcegraph’s code graph, and has knowledge of your entire codebase.
- With [Code Search](https://about.sourcegraph.com/code-search), you can search code across all your repositories and code hosts—even the code you don’t have locally.

**Cody for JetBrains IDEs is experimental right now. We’d love your [feedback](https://github.com/sourcegraph/sourcegraph/discussions/new?category=product-feedback&labels=cody,cody/jetbrains)**!

## Cody Features

### Autocomplete: Cody writes code for you

Cody autocompletes single lines, or whole functions, in any programming language, configuration file, or documentation. It’s powered by latest instant LLM models for accuracy and performance.

![Example of using code autocomplete](https://storage.googleapis.com/sourcegraph-assets/website/Product%20Animations/AutoCompletion_IntelliJ_SS.jpg)

### Chat: Ask Cody about anything in your codebase

Cody understands your entire codebase — not just your open files. Ask Cody any question about your code, and it will use Sourcegraph's code graph to answer using knowledge of your codebase.

For example, you can ask Cody:

- "How is our app's secret storage implemented on Linux?"
- "Where is the CI config for the web integration tests?"
- "Write a new GraphQL resolver for the AuditLog"
- "Why is the UserConnectionResolver giving an "unknown user" error, and how do I fix it?"
- "Add helpful debug log statements"
- "Make this work" _(seriously, it often works—try it!)_

![Example of chatting with Cody](https://storage.googleapis.com/sourcegraph-assets/website/Product%20Animations/Chat_IntelliJ_SS.jpg)

### Built-in commands

Cody has quick commands for common actions. Select the commands tab or right-click on a selection of code and choose one of the `Ask Cody > ...` commands, such as:

- Explain code
- Generate unit test
- Generate docstring
- Improve variable names
- Smell code

_We also welcome also pull request contributions for new, useful commands!_

### Swappable LLMs

On entreprise plans, Cody supports Anthropic Claude, Claude 2, and OpenAI GPT-4/3.5 models, with more coming soon.

### Free usage

Cody is currently in beta, and includes free LLM usage for individual users on both personal and work code. Fair use limits apply.

### Programming language support

Cody works for any programming language because it uses LLMs trained on broad data. Cody works great with Python, Go, JavaScript, and TypeScript code.

### Code graph

Cody is powered by Sourcegraph’s code graph and uses context of your codebase to extend its capabilities. By using context from entire repositories, Cody is able to give more accurate answers and generate idiomatic code.

For example:

- Ask Cody to generate an API call. Cody can gather context on your API schema to inform the code it writes.
- Ask Cody to find where in your codebase a specific component is defined. Cody can retrieve and describe the exact files where that component is written.
- Ask Cody questions that require an understanding of multiple files. For example, ask Cody how frontend data is populated in a React app; Cody can find the React component definitions to understand what data is being passed and where it originates.

### Embeddings

Cody indexes your entire repository and generates embeddings, which are a vector representation of your entire codebase. Cody queries this embeddings database on-demand, and passes that data to the LLM as context. Embeddings make up one part of Sourcegraph’s code graph.

For users who use Cody for free, embeddings are available for Open Source projects through the Sourcegraph.com instance. Support for embeddings on private codes for individual users will be available in the near future.

For those with a Cody Enterprise subscription, your Sourcegraph Enterprise system will generate the embeddings specifically for your codebase.

### Cody Enterprise

Cody Enterprise requires the use of a Sourcegraph Enterprise instance, and gives you access to AI coding tools across your entire organization. [Contact us](https://about.sourcegraph.com/contact/request-info) to set up a trial of Cody Enterprise. If you’re an existing Sourcegraph Enterprise customer, contact your technical advisor.

## Feedback

- [Issue tracker](https://github.com/sourcegraph/sourcegraph/issues)
- [Discord chat](https://discord.gg/s2qDtYGnAE)
- [Twitter (@sourcegraph)](https://twitter.com/sourcegraph)

## License

[Cody's code](https://github.com/sourcegraph/sourcegraph/tree/main/client/jetbrains) is open source (Apache License 2.0).

## Code Search features

- Search with Sourcegraph directly from inside the IDE
- Instantly search in all open source repos and your private code
- Peek into any remote repo in the IDE, without checking it out locally

## URL sharing features

- Create URLs to specific code blocks to share them with your teammates
- Open your files on Sourcegraph

<!-- Plugin description end -->

## Supported IDEs [![JetBrains Plugin](https://img.shields.io/badge/JetBrains-Sourcegraph-green.svg)](https://plugins.jetbrains.com/plugin/9682-sourcegraph)

The plugin works with all JetBrains IDEs, including:

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

**Versions 2022+ Recommended**

**Exception:** Due to a Java bug, search doesn't work with IDE versions **2021.1** and **2021.2** for users with **Apple Silicone** CPUs.

## Installation

- Open settings
  - Windows: Go to `File | Settings` (or use <kbd>Ctrl+Alt+S</kbd>)
  - Mac: Go to `IntelliJ IDEA | Preferences` (or use <kbd>⌘,</kbd>)
- Click `Plugins` in the left-hand pane, then the `Marketplace` tab at the top
- Search for `Sourcegraph`, then click the `Install` button
- Make sure that the `git` command is available in your PATH. We’re going
  to [get rid of this dependency](https://github.com/sourcegraph/sourcegraph/issues/40452), but for now, the plugin
  relies on it.
- Restart your IDE if needed
- To search with Sourcegraph, press <kbd>Alt+S</kbd> (<kbd>⌥S</kbd> on Mac).
- To share a link to your code or search through the website, right-click in the editor, and choose an action under
  the `Sourcegraph` context menu item.
- To use your private Sourcegraph instance, open `Settings | Tools | Sourcegraph` and enter your URL and access token.

## Settings

### List of in-app settings and how to use them

- **Authorization**: List of accounts that can be used to interact with the plugin. Each account can be configured with:
  - **Server**: The URL of your Sourcegraph instance. It can be configured with your private instance if you're adding an enterprise account.
  - **Token**: See our [user docs](https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token) for a video guide on how to
    create an access token.
  - **Custom request headers**: Any custom headers to send with every request to Sourcegraph.
    - Use any number of pairs: `header1, value1, header2, value2, ...`.
    - Example: `Authorization, Bearer 1234567890, X-My-Header, My-Value`.
    - Whitespace around commas doesn't matter.
- **Default branch name**: The branch to use if the current branch is not yet pushed to the remote.
  - Usually "main" or "master", but can be any name
- **Remote URL replacements**: You can replace specific strings in your repo's remote URL.
  - Use any number of pairs: `search1, replacement1, search2, replacement2, ...`.
  - Pairs are replaced from left to right. Whitespace around commas doesn't matter.
  - **Important:** The replacements are done on the Git remote-formatted URL, not the URL you see in the browser!
    - Example replacement subject for Git: `git@github.com:sourcegraph/sourcegraph.git`
    - Example replacement subject for Perforce: `perforce@perforce.company.com:depot-name.perforce`
    - Anatomy of the replacement subjects:
      - The username is not used.
      - Between the `@` and the `:` is the hostname
      - After the `:` is the organization/repo name (for Git) or the depot name (for Perforce)
      - The `.git` / `.perforce` extension is not used.
    - When you do the replacements, make sure you **keep the colon**.
    - In the case of a custom `repositoryPathPattern` being set for your code host in your private Sourcegraph instance,
      you may try to set up a pattern that uses the `@`, `:`, and `.git`/`.perforce` boundaries, _or_ specify a
      replacement
      pair for _each repo_ or _each depot_ you may have. If none of these solutions work for you, please raise this
      at [support@sourcegraph.com](mailto:support@sourcegraph.com), and we'll prioritize making this more convenient.
- **Globbing**: Determines whether you can specify sets of filenames with wildcard characters.
- **Cody completions**: Enables/disables Cody completions in the editor.
  - The completions are disabled by default.

### Git remote setting

By default, the plugin will use the git remote called `origin` to determine which repository on Sourcegraph corresponds
to your local repository. If your `origin` remote doesn't match Sourcegraph, you may instead configure a Git remote by
the name of `sourcegraph`. It will take priority when creating Sourcegraph links.

### Setting levels

You can configure the plugin on three levels:

1. **Project-level** On the project level you are able to configure your default account, default branch name and remote url replacements
2. **Application-level** All other settings are stored here

## Managing Custom Keymaps

![A screenshot of the JetBrains preferences panel inside the Keymap tab](docs/keymaps.png)

You can configure JetBrains to set custom keymaps for Sourcegraph actions:

1. Open the JetBrains preferences panel and go to the Keymap page.
2. Filter by "sourcegraph" to see actions supplied by this plugin.
3. Now select an option to overwrite the keymap information and supply the new bindings.

## Questions & Feedback

If you have any questions, feedback, or bug report, we appreciate if you [open an issue on GitHub](https://github.com/sourcegraph/sourcegraph/issues/new?title=JetBrains:+&labels=jetbrains-ide).

## Uninstallation

- Open settings
  - Windows: Go to `File | Settings` (or use <kbd>Ctrl+Alt+S</kbd>)
  - Mac: Go to `IntelliJ IDEA | Preferences` (or use <kbd>⌘,</kbd>)
- Click `Plugins` in the left-hand pane, then the `Installed` tab at the top
- Find `Sourcegraph` → Right click → `Uninstall` (or uncheck to disable)

## Version History

See [`CHANGELOG.md`](https://github.com/sourcegraph/sourcegraph/blob/main/client/jetbrains/CHANGELOG.md).
