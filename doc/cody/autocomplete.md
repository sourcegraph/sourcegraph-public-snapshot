# Code autocomplete


<aside class="beta">
<p>
<span class="badge badge-beta">Beta</span> <strong>This feature is currently in beta.</strong>
</p>
</aside>

![Example of Cody autocomplete. You see a code snippet starting with async function getWeather(city: string) { and Cody response with a multi-line suggestion using a public weather API to return the current weather ](https://storage.googleapis.com/sourcegraph-assets/cody_completions.png)

## What is Cody code autocomplete?

Cody provides real-time code autocompletion as you're typing. As you start coding, or after you type a comment, Cody will look at the context around your open files and file history to predict what you're trying to implement and provide autocomplete. It's autocomplete powered by Cody!

## Enabling autocomplete

While in beta state, autocomplete need to be enabled manually. To do that:

1. Make sure your [Cody AI by Sourcegraph](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai) extension is on the latest version
   - <kbd>shift</kbd>+<kbd>cmd</kbd>+<kbd>x</kbd> to see all extensions, select Cody, confirm the version
1. Go to the Cody Extension Settings and enable autocomplete
   - Click to check the box for: `Cody > Autocomplete: Enabled`
1. That's it!

### Configuring on Sourcegraph Enterprise

Please follow the steps in [Enabling Cody on Sourcegraph Enterprise](explanations/enabling_cody_enterprise.md) to enable Cody on Sourcegraph Enterprise.

By default, a fully configured Sourcegraph instance picks a default LLM to generate code autocomplete. Custom models can be used for Cody autocomplete via the `completionModel` option inside the `completions` site config.

> NOTE: Self-hosted customers need to update to a minimum of version 5.0.4 to use autocomplete.

<br />

> NOTE: Cody autocomplete currently only work with Anthropic's Claude Instant model. Support for other models will be coming later.

## Accessing autocomplete logs

VS Code logs can be accessed in the _Outputs_ view. To do this:

1. Make sure to have [autocomplete enabled](#enabling-autocomplete)
1. Enable Cody logs in verbose mode.
  - Go to the Cody Extension Settings and enable: `Cody › Debug: Enable` and `Cody › Debug: Verbose`
1. Restart or reload VS Code
1. You can now see the logs in the outputs view
  - Open the view via the menu bar: `View > Output`
  - Select Cody AI by Sourcegraph in the dropdown list
