# Quickstart for Cody in VS Code

Get started with the Cody VS Code extension and start using Cody recipes in under 10 minutes.

If you haven't yet enabled Cody for your Sourcegraph instance, go here first:

- [Enabling Cody for Sourcegraph Enterprise customers](explanations/enabling_cody_enterprise.md)
- [Enabling Cody for Sourcegraph.com users](explanations/enabling_cody.md)

## Introduction

In this guide, you will:

- Install the VS Code extension
- Connect the extension to your Sourcegraph Enterprise instance or Sourcegraph.com account
- Generate a unit test for your code
- Have Cody summarize recent changes to your code

## Requirements

- A Sourcegraph instance with Cody enabled on it OR a Sourcegraph.com account with Cody enabled on it.

If you haven't yet done this, see Step 1 on the following pages:

- [Enabling Cody for Sourcegraph Enterprise](explanations/enabling_cody_enterprise.md)
- [Enabling Cody for Sourcegraph.com](explanations/enabling_cody.md)

## Install the VS Code extension

The first thing you need to do is install the VS Code extension. You can do this in 2 ways:

- Click the Extensions icon on the VS Code activity bar
- Search for "Sourcegraph Cody"
- Install the extension directly to VS Code

Or:

- [Download and install the extension from the VS Code marketplace](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai)

## Generate an access token from your Sourcegraph instance

Next, you must generate an access token so that Cody can talk to your Sourcegraph instance. You'll add this to your extension in the next step.

**For Sourcegraph Enterprise users:**

Log in to your Sourcegraph instance and go to `settings` / `access token` (`https://<your-instance>.sourcegraph.com/users/<your-instance>/settings/tokens`). From here, generate a new access token.

**For Sourcegraph.com users:**

Log in to Sourcegraph.com and go to the [Access tokens page](https://sourcegraph.com/user/settings/tokens). Generate a new access token.

## Configure your extension settings

When you first install the Cody extension, you will see this screen:

  <img width="553" alt="image" src="https://user-images.githubusercontent.com/25070988/227510233-5ce37649-6ae3-4470-91d0-71ed6c68b7ef.png">

In the `Access Token` field, paste in your access token from the previous step.

In the `Sourcegraph Instance` field, paste in the URL of your Sourcegraph instance.

- For Sourcegraph Enterprise, this is your own instance's URL.
- For Sourcegraph.com, this is `https://sourcegraph.com`

## Generate a unit test

The Cody icon should now appear in the activity bar in VS Code. Clicking the icon will open the Cody side panel. The `Chat` tab can be used to ask Cody questions and paste in snippets of code, and the `Recipes` tab can be used to run premade functions over whatever code you currently have highlighted.

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

## Try other recipes and Cody chat

Cody can do many other things, including:

1. Explain code snippets
2. Refactor code and variable names
3. Translate code to different languages
4. Answer general questions about your code
5. Suggest bugfixes and changes to code snippets

To see this in action, try asking Codying a question in the chat sidebar such as "Are there any bugs in this code snippet?"

## Congratulations!

**You're now up-and-running with your very own AI code asisstant!** 🎉
