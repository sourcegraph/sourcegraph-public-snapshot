<!-- Plugin description -->

# Sourcegraph Cody + Code Search

Use Sourcegraph Code Search and Sourcegraph’s AI assistant Cody directly from your JetBrains IDE.

- With Code Search, you can search code across all your repositories and code hosts—even the code you don’t have locally.
- Cody can write code and answer questions across your entire codebase.

**Cody AI for JetBrains IDEs is experimental right now. We’d love your [feedback](https://github.com/sourcegraph/sourcegraph/discussions/new?category=product-feedback&labels=cody,cody/jetbrains)**!

## Features

### 🤖 Ask Cody about anything in your codebase

**Cody understands your entire codebase — not just your open files. Ask questions, insert code, and use the built-in commands such as "Generate unit test" and "Improve variable names".**

Cody combines the power of large language models (LLMs) with Sourcegraph’s Code Graph API, generating deep knowledge of all of your code—even the code you don’t have locally. Large monorepos, multiple languages, and complex codebases are no problem for Cody.

Example questions you can ask Cody:

- Where is the CI config for the web integration tests?
- Write a new GraphQL resolver for the AuditLog
- Why is the UserConnectionResolver giving an error "unknown user", and how do I fix it?
- Add helpful debug log statements
- Make this work _(seriously, it often works—try it!)_

![Example of chatting with Cody](https://storage.googleapis.com/sourcegraph-assets/website/Product%20Animations/Chat_IntelliJ_SS.jpg)

### 🔨 Let Cody write code for you

Cody can provide real-time code autocompletions as you're typing. As you start coding, or after you type a comment, Cody will look at the context around your open files, your file history, and your entire codebase to predict what you're trying to implement and provide suggestions.

![Example of using code autocomplete](https://storage.googleapis.com/sourcegraph-assets/website/Product%20Animations/AutoCompletion_IntelliJ_SS.jpg)

## 🍳 Built-in commands

Select the commands tab or right-click on a selection of code and choose one of the `Ask Cody > ...` commands, such as:

- Explain code
- Generate unit test
- Generate docstring
- Improve variable names
- Smell code

_We also welcome also pull request contributions for new, useful commands!_

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

- **Sourcegraph URL**: The URL of your Sourcegraph instance if you use a private instance.
  - To use Sourcegraph.com and search in public repos, just choose "Use sourcegraph.com".
- **Access token**:
  - If you want to use your private Sourcegraph instance, you'll need an access token to authorize
    yourself.
  - If you use Sourcegraph.com, using an access token is optional (and only necessary to use Cody).
  - The configuration for an access token to use with Sourcegraph.com & a private instance is separate,
    you can switch between them on the fly.
  - See our [user docs](https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token) for a video guide on how to
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

1. **Project-level** settings take the highest priority.
2. **Application-level** settings are second: For _each specific setting_, if the plugin finds no project-level value,
   then the app-level setting is used.
3. **User-level** (legacy) settings take the lowest priority. Also, note that only three of the settings are available
   on the user level.

Here is each level in detail.

#### Project level

These settings have the highest priority. You can set them in a less than intuitive way:

1. Create a new file at `{project root}/.idea/sourcegraph.xml` if it doesn't exist, with this content:

   ```xml
   <?xml version="1.0" encoding="UTF-8"?>
   <project version="4">
     <component name="Config">
       <option name="instanceType" value="DOTCOM" />
       <option name="url" value="https://company.sourcegraph.com/" />
       <option name="enterpriseAccessToken" value="" />
       <option name="defaultBranch" value="main" />
       <option name="remoteUrlReplacements" value="" />
       <option name="isGlobbingEnabled" value="false" />
     </component>
   </project>
   ```

   If the file already exists, then just add the option lines next to the original ones.

   **Replace `DOTCOM` with `ENTERPRISE` for private instanceType.**

2. Reopen your project to let the IDE catch up with the changes. Now you have custom settings enabled for this project. In the future, when you have this project open, and you edit your settings in the Settings UI, they will be saved to the **project-level** file.
3. To remove the project-level settings, open the XML again and remove the lines you want to set on the app level.

**Storage location:** `{project root}/.idea/sourcegraph.xml`

#### Application level

This is what you edit when you go to Settings and make changes in the UI. That is, unless you have project-specific settings for your current project.

**Storage location:** App-level settings are stored in a file called `sourcegraph.xml` together with the rest of the IDE settings. [This article](https://intellij-support.jetbrains.com/hc/en-us/articles/206544519-Directories-used-by-the-IDE-to-store-settings-caches-plugins-and-logs) will help you find it if you should need it for anything.

#### User level – ⚠️ DEPRECATED ⚠️

This type of settings take the lowest priority, and is something that's rarely used and is only kept for backwards compatibility, and might be removed in the future. So, the plugin is also configurable by removing all creating a file called `.sourcegraph-jetbrains.properties` in your home directory. Both the app-level and project-level XMLs override this, plus it only supports three settings:

```
url = https://sourcegraph.example.com
defaultBranch = example-branch
remoteUrlReplacements = git.example.com, git-web.example.com
```

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
