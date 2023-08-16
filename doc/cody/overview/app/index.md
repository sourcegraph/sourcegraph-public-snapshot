<style>
.socials {
  display: flex;
  flex-direction: row;
}
.socials a {
  padding: 0.25rem;
  margin: 1rem;
  background: #dddddd;
  border-radius: 0.25rem;
  width: 3.5rem;
  height: 3.5rem;
  display: flex;
  align-items: center;
}
.socials a:hover {
  filter: brightness(0.75);
}
.socials a img {
  width: 100%;
  height: 100%;
}
</style>

# Cody App

The Cody app is a free, lightweight, native desktop application that connects your local code to our AI coding assistant, Cody. You can ask Cody questions about all the code you connect to your app in both the app interface and, if you connect the VS code extension, in your editor.

<div class="socials">
  <a href="https://discord.com/invite/s2qDtYGnAE"><img alt="Discord" src="discord.svg"></img></a>
  <a href="https://twitter.com/sourcegraph"><img alt="Twitter" src="twitter.svg"></img></a>
  <a href="https://github.com/sourcegraph/app"><img alt="GitHub" src="github.svg"></img></a>
</div>

## Installation

Download the app:
- [MacOS (Apple Silicon)](https://sourcegraph.com/.api/app/latest?arch=aarch64&target=darwin)
- [MacOS (Intel)](https://sourcegraph.com/.api/app/latest?arch=x86_64&target=darwin)
- [Linux](https://sourcegraph.com/.api/app/latest?arch=x86_64&target=linux)
- Windows coming in October 2023

## Setup

Follow the setup instructions to connect the app to your Sourcegraph.com account (or create one for free if you don't have an account yet), add your local projects, and select up to 10 of those projects to build your code graph. The code graph helps Cody generate more accurate answers about your code by sending your code to Sourcegraph to create [embeddings](./../../explanations/code_graph_context.md#embeddings). (This may take a few minutes, depending on the size of your repos.) We are working on making it so that *all* the projects you connect to your app get embeddings and bumping the cap up from 10.

If you use VS Code or a JetBrains IDE, we recommend you follow the steps to download the extension, which enables Cody within your editor. (If you installed the extension before you downloaded the app, you'll see a prompt in your editor to download the app.) Cody in your editor will then talk to your Sourcegraph app to answer questions.

Note: The JetBrains extension is still `Experimental`.

## Upgrading

You'll get prompts to upgrade automatically. Let us know if you have any issues and we'll be happy to help.

If you're on a version that's 2023.6.13 or older, we recommend you uninstall the app (see [Uninstallation](#uninstallation) below) and download the most recent version in order to add your projects to your code graph during setup. Also note that these older versions of the app were called "Sourcegraph" and included Sourcegraph code search. Going forward, the app will be a Cody-only experience (and branded accordingly) so that we can focus on making Cody as useful and intuitive as possible.

## Rate limiting

There are several forms of rate limiting that help us control costs for free versions of Cody. We expect to relax these limits as we continue development on [Cody Gateway](./../../explanations/cody_gateway.md). If you hit these limits, you can can request an increase by visiting our [discord](https://discord.com/servers/sourcegraph-969688426372825169) channel and requesting a higher limit for both chats and completions. If you'd like to use your own third-party LLM provider instead of Cody Gateway, you must create your own key with Anthropic or OpenAI and [update your app configuration](app_configuration.md).

### Embeddings

The setup experience allows users to select up to 10 repos for embeddings. Additional repos can be added, and embeddings can be scheduled, under Settings > Advanced settings > Embedding jobs, but the number of additional repos supported will vary depending on size.

## Uninstallation

Select "Troubleshooting > Clear All Data" from the system tray and delete the app from your applications folder. If you're on an older version of the app and don't see a "Clear All Data" option, run:

```bash
rm -rf ~/.sourcegraph-psql ~/Library/Application\ Support/com.sourcegraph.cody ~/Library/Caches/com.sourcegraph.cody ~/Library/WebKit/com.sourcegraph.cody
```

## Troubleshooting

See [App troubleshooting](troubleshooting.md)

## Release pipeline

See [App release pipeline](release-pipeline.md)

## API and integrations

See [App API integrations](integrations.md)

## Get help & give feedback

Cody app is new and we're iterating on it quickly. If you run into any trouble or have ideas/feedback, we'd love to hear from you!

* [Join our community Discord](https://discord.com/invite/s2qDtYGnAE) for live help/discussion
* [Create a GitHub issue](https://github.com/sourcegraph/app/issues/new)
