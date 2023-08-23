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

<p class="subtitle">Learn and understand what features Cody offers to ensure a streamlined code AI.</p>

## Code chatbot

Cody is your AI-powered coding assistant that understands your entire codebase inside out. It goes beyond other AI code chatbots, aligning perfectly with your project's code conventions and architecture. You can chat with Cody right within your code editor or through the Sourcegraph sidebar.

Cody tells you which code files it reads to generate its response. In case of a wrong answer, please share feedback so we can improve it.

Examples of questions Cody can handle:

- How is our app's secret storage implemented on Linux?
- Where is the CI config for the web integration tests?
- Write a new GraphQL resolver for the AuditLog.
- Why is the UserConnectionResolver giving an error `unknown user`, and how do I fix it

<div class="getting-started">
  <a class="demo text-center" target="_blank" href="https://twitter.com/beyang/status/1647744307045228544">View Demo</a>
</div>

## Code Autocomplete

Cody provides real-time code auto-completion as you type, based on the context around your open files and file history. This predictive feature tells what you are trying to implement for a smoother coding experience.

![Example of Cody autocomplete. You see a code snippet starting with async function getWeather(city: string) { and Cody response with a multi-line suggestion using a public weather API to return the current weather ](https://storage.googleapis.com/sourcegraph-assets/website/Product%20Animations/GIFS/cody-completions-may2023-optim.gif)

### Configure autocomplete on Sourcegraph enterprise

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

## Commands

Cody supports **Commands** with its [VS Code extension](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai). These allow you to create and run reusable prompts without submitting pull requests or opening the chat sidebar.

You can run a command in 3 ways:

1. Type `/` in the chat bar. Cody will then suggest a list of commands
2. Right click > Cody > Select a command
3. Press the command hotkey (`⌥` + `C` / `Alt` + `C`)

![Cody Commands in VS Code](https://storage.googleapis.com/sourcegraph-assets/Docs/cody-commands.png)

### Custom Commands

The **Custom Commands** feature lets you customize Cody's abilities for your projects without writing code. This provides a flexible way to create reusable "commands" that give Cody targeted instructions and context from your project to enhance productivity.

You can define customized prompts using `JSON` files that can be invoked simply by typing a `/` in the chat box or using keyboard shortcuts `⌥` + `C` / `Alt` + `C` without opening the chat sidebar.

Right-click on the selected code to open the Cody context menu and select the **Custom Commands (Experimental)** option.

![Cody Custom Commands in VS Code](https://storage.googleapis.com/sourcegraph-assets/Docs/create-custom-commands.png)

### `cody.json` file

You can define custom commands for Cody in the `cody.json` file. To make commands accessible only in a specific project, create the `cody.json` file in the `.vscode` directory of your project. Cody will execute these workspace-specific custom commands when working in that project.

To make commands accessible across multiple projects, create a new `cody.json` file in the `.vscode` directory of your home directory. Cody will execute these global custom commands in any workspace.

## Fix code inline

Cody can help you make interactive edits and refactor code by following natural-language instructions. To do so, select the relevant code snippet, and ask Cody a question or request inline fix with `/fix` or `/touch` commands.

Cody will take it from there and figure out what edits to make.

![Example of Cody inline code fix ](https://storage.googleapis.com/sourcegraph-assets/website/Product%20Animations/GIFS/cody_inline_June23-sm.gif)

Examples of fix-up instructions Cody can handle:

- Factor out any common helper functions (when multiple functions are selected)
- Use the imported CSS module's class `n`
- Extract the list item to a separate React component
- Handle errors better
- Add helpful debug log statements
- Make this work (and yes, it often does work—give it a try!)

<div class="getting-started">
  <a class="demo text-center" target="_blank" href="https://twitter.com/sqs/status/1647673013343780864">View Demo</a>
</div>
