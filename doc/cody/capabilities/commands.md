# Commands

<p class="subtitle">Learn how Commands help you kick-start with reusuable prompts.</p>

Commands allow you to run common actions quickly. Commands are predefined, reusable prompts accessible by hotkey from within the [VS Code extension](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai). Like autocomplete and chat, commands will search for context in your codebase to provide more contextually aware and informed answers (or to generate more idiomatic code snippets).

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
