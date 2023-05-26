# Cody completions

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This feature is experimental and might change or be removed in the future. We've released it as an experimental feature to provide a preview of functionality we're working on.
</p>
</aside>

![Example of Cody completions. You see a code snippet starting with async function getWeather(city: string) { and Cody response with a multi-line suggestion using a public weather API to return the current weather ](https://storage.googleapis.com/sourcegraph-assets/cody_completions.png)

## What are Cody completions?

Cody provides real-time code completions as you're typing. As you start coding, or after you type a comment, Cody will look at the context around your open files and file history to predict what you're trying to implement and provide completions. It's autocomplete powered by Cody!

## Enabling Cody completions

While in experimental state, Cody completions need to be enabled manually. To do that:

1. Make sure your Cody AI by Sourcegraph extension is on the latest version
   - <kbd>shift</kbd>+<kbd>cmd</kbd>+<kbd>x</kbd> to see all extensions, select Cody, confirm the version
1. Next, go to the Cody Extension Settings and enable completions
   - Click to check the box for: `Cody > Experimental Suggestions`
1. Finally, restart VS Code and test it out!

### Configuring on Sourcegraph Enterprise

Please follow the steps in [Enabling Cody on Sourcegraph Enterprise](.explanations/enabling_cody_enterprise#enabling-cody-on-sourcegraph-enterprise) to enable Cody on Sourcegraph enterprise.

You have to configure the model used for Cody completions via the `completionModel` option inside the `completions` site config.

> NOTE: Self-hosted customers need to update to a minimum of version 5.0.4 to use completions.

> NOTE: Cody completions currently only work with Claude Instant or our LLMProxy configured with Claude Instant. Support for other models will be coming later.
