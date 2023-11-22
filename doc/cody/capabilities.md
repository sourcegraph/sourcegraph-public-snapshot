<style>
  .demo{
    display: table;
    width: 35%;
    margin: 0.5em;
    padding: 1rem 1rem;
    color: var(--text-color);
    border-radius: 4px;
    border: 1px solid var(--sidebar-nav-active-bg);
    padding: 1rem;
    padding-top: 1rem;
    background-color: var(--sidebar-nav-active-bg);
  }
</style>

# Cody capabilities

<p class="subtitle">Learn and understand more about Cody's features and core AI functionality.</p>

## Autocomplete

Cody suggests completions as you type using context from your code, such as your open files and file history. It’s powered by the latest instant LLM models for accuracy and performance.

Autocomplete supports any programming language because it uses LLMs trained on broad data. We've found that it works exceptionally well with JavaScript, TypeScript, Python, and Go code.

![Example of Cody autocomplete. You see a code snippet starting with async function getWeather(city: string) { and Cody response with a multi-line suggestion using a public weather API to return the current weather ](https://storage.googleapis.com/sourcegraph-assets/website/Product%20Animations/GIFS/cody-completions-may2023-optim.gif)

### Configure autocomplete on an enterprise Sourcegraph instance

By default, a fully configured Sourcegraph instance picks a default LLM to generate code autocomplete. Custom models can be used for Cody autocomplete via the `completionModel` option inside the `completions` site config.

We also recommend reading the [Enabling Cody on Sourcegraph Enterprise](overview/enable-cody-enterprise.md) guide before you configure the autocomplete feature.

> NOTE: Self-hosted customers need to update to a minimum of version 5.0.4 to use autocomplete.

<br />

> NOTE: Cody autocomplete currently only work with Anthropic's Claude Instant model. Support for other models will be coming later.

### Access autocomplete logs

VS Code logs can be accessed via the **Outputs** view. To access autocomplete logs, you need to enable Cody logs in verbose mode. To do so:

- Go to the Cody Extension Settings and enable: `Cody › Debug: Enable` and `Cody › Debug: Verbose`
- Restart or reload your VS Code editor
- You can now see the logs in the Outputs view
- Open the view via the menu bar: `View > Output`
- Select **Cody by Sourcegraph** from the dropdown list

![View Cody's autocomplete logs from the Output View in VS Code](https://storage.googleapis.com/sourcegraph-assets/Docs/view-autocomplete-logs.png)

## Chat

Chat lets you ask Cody general programming questions or questions about your specific code. You can chat with Cody in the `Chat` panel of the editor extensions or with the `Ask Cody` button in the Sourcegraph UI.

Cody uses several search methods (including keyword and semantic search) to find files in your codebase that are relevant to your chat questions. It then uses context from those files to provide an informed response based on your codebase. Cody also tells you which code files it reads to generate its responses.

Context retrieval isn't perfect, and Cody occasionally uses incorrect context or hallucinates answers. When Cody returns an incorrect response, it is often worth asking the question again slightly differently to see if Cody can find better context the second time.

Cody's chat function can handle use cases like:

- Ask Cody to generate an API call. Cody can gather context on your API schema to inform the code it writes
- Ask Cody where a specific component is defined within your codebase. Cody can retrieve and describe the files where that component is written
- Ask Cody questions that require an understanding of multiple files, such as how data is populated in a React app. Cody can find the React component definitions to understand what data is being passed and where it originates

More specifically, Cody can answer questions like:

- How is our app's secret storage implemented on Linux?
- Where is the CI config for the web integration tests?
- Can you write a new GraphQL resolver for the AuditLog?
- Why is the UserConnectionResolver giving an error `unknown user`, and how do I fix it?

![Example of Cody chat. You see the user ask Cody to describe what a file does, and Cody returns an answers that explains how the file is working in the context of the project.](https://storage.googleapis.com/sourcegraph-assets/cody/Docs/Sept2023/Context_Chat_SM.gif)

<div class="getting-started">
  <a class="demo text-center" target="_blank" href="https://twitter.com/beyang/status/1647744307045228544">View Demo</a>
</div>

### Inline chat and code edits

You can also open Cody's chat inline in VS Code using the `+` icon. This opens a chat box that can be used for general chat questions, code edits, and refactors. Select a code snippet to ask Cody for an inline code edit, then type `/edit` plus your desired code change. Cody will generate edits, which you can accept or reject with the `Apply` button.

You can also use the or `/touch` command in the inline chat box if you'd like Cody to place its output in a new file.

![Example of Cody inline code fix ](https://storage.googleapis.com/sourcegraph-assets/website/Product%20Animations/GIFS/cody_inline_June23-sm.gif)

Examples of `/edit` instructions Cody can handle:

- Factor out any common helper functions (when multiple functions are selected)
- Use the imported CSS module's class `n`
- Extract the list item to a separate React component
- Handle errors in this code better
- Add helpful debug log statements
- Make this work (and yes, it often does work—give it a try!)

> NOTE: Inline chat functionality is currently only available in the VS Code extension. The `/edit` command was called `/fix` prior to version 0.10.0 of the VS Code extension.

<div class="getting-started">
  <a class="demo text-center" target="_blank" href="https://twitter.com/sqs/status/1647673013343780864">View Demo</a>
</div>

## Commands

**Commands** allow you to run common actions quickly. Commands are predefined, reusable prompts accessible by hotkey from within the [VS Code extension](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai). Like autocomplete and chat, commands will search for context in your codebase to provide more contextually aware and informed answers (or to generate more idiomatic code snippets).

The commands available in VS Code include:

- Document Code
- Explain Code
- Generate Unit Tests
- Code Smell

![Cody Commands in VS Code](https://storage.googleapis.com/sourcegraph-assets/Docs/cody-commands.png)

There are three ways to run a command in VS Code:

1. Type `/` in the chat bar. Cody will then suggest a list of available commands
2. Right click and select `"Cody"` > Choose a command from the list
3. Use the predefined command hotkey: `⌥` + `C` / `Alt` + `C`

> NOTE: This functionality is also available in the JetBrains extension under the name `Recipes`. To access it, navigate to the `Recipes` panel (next to the `Chat` panel), and you can find each available recipe as a button within the UI.

![Example of the Cody 'Explain Code' command. The user highlights a section of code, then uses the hotkey to open the commands menu, then selects the command.](https://storage.googleapis.com/sourcegraph-assets/cody/Docs/Sept2023/Explain_Code_SM.gif)

### Custom commands <span class="badge badge-experimental">Experimental</span>

**Custom commands** let you save your quick actions and prompts for Cody based on your common workflows. They are defined in JSON format and allow you to call CLI tools, write custom prompts, and select context to be sent to Cody. This provides a flexible way to tailor Cody to your needs.

![Cody Custom Commands in VS Code](https://storage.googleapis.com/sourcegraph-assets/Docs/create-custom-commands.png)

You can invoke custom commands with the same hotkey as predefined commands. Alternatively, you can right-click the selected code, open the Cody context menu, and select the `Custom Commands (Experimental)` option.

![Example of a user creating a custom "/Variables" command. The user creates a command that automatically makes the variable names in selected code more helpful.](https://storage.googleapis.com/sourcegraph-assets/cody/Docs/Sept2023/Custom_Command_SM.gif)

### Defining commands in the `cody.json` file

You can define custom commands for Cody in the `cody.json` file. To make commands only available for a specific project, create the `cody.json` file in that project's `.vscode` directory. When you work on that project, these workspace-specific custom commands will be available.

To make custom commands globally available across multiple projects, create a new `cody.json` file in your home directory's `.vscode` folder. These global custom commands will be available in Cody in any workspace.

## Language support

Cody uses LLMs trained on broad code, and we've found it to support all common programming languages effectively. However, the quality of autocompletion and other features may vary based on how well the underlying LLM model was trained in a given language.
