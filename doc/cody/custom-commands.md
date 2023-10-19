# Cody Custom Commands

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This feature is experimental and might change or be removed in the future. We've released it as an experimental feature for testing and feedback.
</p>
</aside>

Custom Commands allow you to define reusable prompts tailored for your own needs. They are currently supported by [Cody for VS Code](overview/install-vscode.md) version 0.8 and above.

### Use cases

* Integrate with your build system to suggest fixes for errors/warnings in the latest build
* Read in test command failures to explain and suggest fixes
* Analyze software dependencies output to explain compatibility or suggest upgrades
* Explain code quality output like linter warnings

## Creating a Command

You can can create a custom command by editing the [configuration JSON file](#file-paths), or by using command builder within VS Code.

To access the command builder within VS Code, open the Cody commands menu (⌥C on Mac or Alt-C on Windows/Linux) → "Configure Custom Commands..." → "New Custom Command..."

![Cody Custom Command Setup in VS Code](https://storage.googleapis.com/sourcegraph-assets/docs/images/cody/custom-command-setup-1.png)

## Configuration File Paths

Custom Commands can be defined in the following two file paths:

* User Settings (`~/.vscode/code.json`) — Stored locally. Use this for personal commands to use across all your projects.
* Workspace Settings (`.vscode/cody.json`) — Stored in project’s repository. Use this to share commands with others working on the same codebase, and for project-specific commands.

See the [examples](#examples) and [configuration properties](#configuration-properties) below for more details on how to configure custom commands.

## Examples

### `/commit-message`

```json
{
  "commit-message": {
    "description": "Commit message for current changes",
    "prompt": "Suggest an informative commit message by summarizing code changes from the shared command output. The commit message should follow the conventional commit format and provide meaningful context for future readers.",
    "context": {
      "selection": false,
      "command": "git diff --cached"
    }
  }
}
```

### `/compare-tabs`

```json
{
  "compare-tabs": {
    "description": "Compare files in opened tabs",
    "prompt": "Examine the shared code snippets from my opened tabs. Then explain the relationship between the code samples by answering these questions: 1. What are the main tasks or functions the code is performing? 2. Are there any similarities in functions or logic across the samples? 3. Does one code snippet call or import another? If so, how do they interact? 4. Are there any notable differences in how they approach similar problems? 5. Overall, how are the code snippets related - do they work together as part of a larger program, solve variants of the same problem, or serve entirely different purposes?",
    "context": {
      "openTabs": true,
      "selection": false
    }
  }
}
```

### `/current-dir`

```json
{
  "current-dir": {
    "description": "Explain current directory",
    "prompt": "At a high level, explain what this directory is used for.",
    "context": {
      "currentDir": true,
      "seleection": false
    }
  }
}
```

### `/latest-cody-release`

```json
{
  "latest-cody-release": {
    "description": "Summarize latest Cody release",
    "prompt": "What is the latest stable version of Cody? Briefly summarize the changes that were included in that release based on this CHANGELOG excerpt.",
    "context": {
      "selection": false,
      "command": "curl https://raw.githubusercontent.com/sourcegraph/cody/main/vscode/CHANGELOG.md | head -n 50"
    }
  }
}
```

### `/readme`

```json
{
  "readme": {
    "description": "Readme for current dir",
    "prompt": "Write a detailed README.md file to document the code located in the same directory as my current selection. Summarize what the code in this directory is meant to accomplish. Explain the key files, functions, classes, and features. Use Markdown formatting for headings, code blocks, lists, etc. to make the it organized and readable. Aim for a beginner-friendly explanation that gives a developer unfamiliar with the code a good starting point to understand it. Make sure to include: - Overview of directory purpose - Functionality explanations - Relevant diagrams or visuals if helpful. Write the README content clearly and concisely using complete sentences and paragraphs based on the shared context. Use proper spelling, grammar, and punctuation throughout. Surround your full README text with triple backticks so it renders properly as a code block. Do not make assumptions or fabricating additional details.",
    "context": {
      "currentDir": true,
      "selection": true
    }
  }
}
```

### `/recent-git-changes`

```json
{
  "recent-git-changes": {
    "description": "Summarize recent changes",
    "prompt": "Summarize the given git changes in 3-5 sentences",
    "context": {
      "command": "git log -10 --pretty=format:'%h - %an: %s' --stat",
      "selection": false
    }
  }
}
```

## Configuration Properties

#### `commands` (required)

An object containing the commands.

Type: `object`

Example: `{ "commands": {} }`

#### `commands.<id>` (required)

The slash name used for the command.

Type: `string`

Example: `"summarize-git-changes"`

#### `commands.<id>.description` (required)

A short, sentence-case string description of what the command does.

Type: `string`

Example: `"Summarize recent git changes"`

#### `commands.<id>.prompt`

The set of instructions for Cody.

Type: `string`

Example: `"Summarize the given git changes in 3-5 sentences"`

#### `commands.<id>.mode`

The interaction mode to use.

Type: `string`

Default: `"ask"`

Values:
- `"ask"` — Focus the chat view and add to the current chat
- `"inline"` — Start a new inline chat
- `"insert"` — Insert the response above the code
- `"replace"` — Replace the code with the response

#### `commands.<id>.context`

Optional context data to generate and pass to Cody.

Type: `object`

Default: `{ "codebase": true }`

#### `commands.<id>.context.codebase`

Include embeddings and/or keyword code search (depending on availability).

Type: `boolean`

Default: `true`

#### `commands.<id>.context.command`

Terminal command to run and include the output of.

Type: `string`

Default: `""`

Example: ` "git log --since=$(git describe --tags --abbrev=0) --pretty='Commit author: %an%nCommit message: %s%nChange description:%b%n'"`

#### `commands.<id>.context.currentDir`

Include snippets from first 10 files in the current directory.

Type: `boolean`

Default: `false`

#### `commands.<id>.context.currentFile`

Include snippets from the current file. If the file is too long, only the content surrounding the current selection will be included.

Type: `boolean`

Default: `false`

#### `commands.<id>.context.directoryPath`

Include snippets from the first five files within the given relative path of the directory. Content will be limited and truncated according to the token limit.

Type: `string`

Default: `""`

Example: `"lib/common"`

#### `commands.<id>.context.filePath`

Include snippets from the given file path relative to the codebase. If the file is too long, content will be truncated.

Type: `string`

Default: `""`

Example: `"CHANGELOG.md", "test/unit/example.test.ts"`

#### `commands.<id>.context.none`

Provide only the prompt, and no additional context. If `true`, overrides all other context settings.

Type: `boolean`

Default: `false`

#### `commands.<id>.context.openTabs`

Include the text content of opened editor tabs.

Type: `boolean`

Default: `false`

#### `commands.<id>.context.selection`

Include currently selected code. When not specified, Cody will try to use visible content from the current file instead.

Type: `boolean`

Default: `false`

