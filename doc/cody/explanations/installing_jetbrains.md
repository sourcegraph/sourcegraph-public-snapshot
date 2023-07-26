# Installing Cody in Jetbrains (Experimental)

## Introduction

In this guide, you will:

- Install the Jetbrains Code extension
- Connect the extension to your Sourcegraph Enterprise instance or Sourcegraph.com account

## Requirements

- A Sourcegraph instance with Cody enabled on it OR a Sourcegraph.com account.

If you haven't yet done this, see Step 1 on the following pages:

- [Enabling Cody for Sourcegraph Enterprise](enabling_cody_enterprise.md)
- [Enabling Cody for Sourcegraph.com](enabling_cody.md)

## Install the Jetbrains IntelliJ Cody extension

You can install Cody in Jetbrains in 2 ways:

- Open Settings (Windows: Ctrl+Alt+S, Mac: ⌘,) and select Plugins.
- Search for "Sourcegraph Cody"
- Install the extension directly to IntelliJ

Or:

- [Download and install the extension from the Jetbrains marketplace](https://plugins.jetbrains.com/plugin/9682-sourcegraph)

## Connect the extension to Sourcegraph

Next, you'll open the Cody extension and configure it to connect to a Sourcegraph instance (either an enterprise instance or Sourcegraph.com).

**For Sourcegraph Enterprise users:**

Log in to your Sourcegraph instance and go to `settings` / `access token` (`https://<your-instance>.sourcegraph.com/users/<your-instance>/settings/tokens`). From here, generate a new access token.

Then, you select the option to `Use an exnterprise instance` and you will paste your access token and instance URL address in to the Cody extension.

**For Sourcegraph.com users:**

Click `Continue with Sourcegraph.com` in the Cody extension. From there, you'll be taken to Sourcegraph.com, which will authenticate your extension.

## (Optional) Enable code graph context for context-aware answers

You can optional configure code graph content, which gives Cody the ability to provide context-aware answers. For example, Cody can write example API calls if has context of a codebase's API schema.

- [Configure code graph context for Sourcegraph.com](enabling_cody.md#configure-code-graph-context-for-code-aware-answers)
- [Configure code graph context for Sourcegraph Enterprise](enabling_cody_enterprise.md#enabling-codebase-aware-answers)

## Get started with Cody

You're now ready to use Cody! See our recommendations for getting started with using Cody.

## Congratulations!

**You're now up-and-running with your very own AI code asisstant!** 🎉

