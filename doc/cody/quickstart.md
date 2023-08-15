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

- Trying the `Generate a unit test` recipe
- Trying the `Summarize recent code changes` recipe
- Asking Cody to pull information from documentation
- Asking Cody to write context-aware code

## Getting started with the Cody extension and recipes

The Cody icon should now appear in the activity bar in VS Code. Clicking the icon will open the Cody side panel. The `Chat` tab can be used to ask Cody questions and paste in snippets of code, and the `Recipes` tab can be used to run premade functions over whatever code you currently have highlighted.

## Generate a unit test

As one recipe example, let's generate a unit test:

1. Open a code file in VS Code
2. Open the `Recipes` tab in the Cody sidebar
3. Highlight a code snippet in your code that you'd like to test
4. Click the `Generate a unit test` recipe button OR right click -> `Ask Cody` -> `Ask Cody: Generate Unit Test`
5. Cody will provide a unit test as a code snippet in the sidebar

## Get up-to-date on recent changes to your code (Sourcegraph Enterprise users only)

If you're using Cody with Sourcegraph Enterprise, Cody can utilize Sourcegraph's search and the code graph to understand your own codebase and provide context-aware answers. Using this, Cody can tell you about recent changes to your code and quickly provide a summary.

Imagine you've stepped away from a project for the last week. Let's see what's changed in that time:

1. Open the `Recipes` tab in the Cody sidebar
2. Click the `Summarize recent changes` recipe button
3. A dropdown will appear. Click `Last week`
4. Cody will provide a summary of changes to your codebase that occurred in that time period

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

## Try other recipes and Cody chat

Cody can do many other things, including:

1. Explain code snippets
2. Refactor code and variable names
3. Translate code to different languages
4. Answer general questions about your code
5. Suggest bugfixes and changes to code snippets

[cody-with-sourcegraph]: overview//cody-with-sourcegraph.md
