# Quickstart for Cody in VS Code

This guide provides recommendations for first things to try once you have Cody up and running.

If you haven't yet enabled Cody for your Sourcegraph instance or installed the VS Code extension, go here first:

- [Enabling Cody for Sourcegraph Enterprise customers](overview/enable-cody-enterprise.md)
- [Enabling Cody for Sourcegraph.com users](overview/cody-with-sourcegraph.md)
- [Installing Cody for VS Code](overview/install-vscode.md)
- [Installing Cody for Jetbrains](overview/install-jetbrains.md)
- [Installing the Cody app](overview/app/index.md)

## Introduction

Once you have access to Cody, we recommend:

- Trying the `Generate Unit Tests` command
- Asking Cody to pull information from documentation
- Asking Cody to write context-aware code

## Getting started with the Cody extension and commands

Once you've installed the Cody extension, the Cody icon should appear in the activity bar. Clicking the icon will open Cody's `Chat` panel. This can be used to ask Cody questions and paste in snippets of code.

You can also run `Commands` with Cody, quick actions that apply to any code you currently have highlighted. Once you've highlighted a snippet of code, you can run a command in 3 ways:

1. Type `/` in the chat bar. Cody will then suggest a list of commands.
2. Right click -> Cody -> Select a command.
3. Press the command hotkey (`⌥` + `c` / `alt` + `c`)

## Generate a unit test

Cody offers a number of commands. One command, `Generate Unit Tests`, quickly writes test code for any snippet that you have highlighted. To use this command:

1. Open a code file in VS Code
3. Highlight a code snippet that you'd like to test
3. Press the command hotkey (`⌥` + `c` / `alt` + `c`), then select `Generate Unit Tests`
5. Cody will provide a unit test as a code snippet in the sidebar

## Ask Cody to pull reference documentation

Cody can also directly reference documentation. If Cody has context of a codebase, and docs are committed within that codebase, Cody can search through the text of docs to understand documentation and quickly pull out information so that you don't have to search for it yourself.

In the `Chat` tab of the VS Code extension, you can ask Cody examples like:

- "What happens when Service X runs out of memory?"
- "Where can I find a list of all experimental features and their feature flags?"
- "How is the changelog generated?"

## Ask Cody to write context-aware code

One of Cody's strengths is writing code, especially boilerplate code or simple code that requires general awareness of the broader codebase.

One great example of this is writing an API call. Once Cody has context of a codebase, including existing API schemas within that codebase, you can ask Cody to write code for API calls.

For example, ask Cody:

- "Write an API call to retrieve user permissions data"

## Try other commands and Cody chat

Cody can do many other things, including:

1. Explain code snippets
2. Refactor code and variable names
3. Translate code to different languages
4. Answer general questions about your code
5. Suggest bugfixes and changes to code snippets

[cody-with-sourcegraph]: overview//cody-with-sourcegraph.md
