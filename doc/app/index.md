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

Check the [latest release](https://github.com/sourcegraph/sourcegraph/releases/tag/app-v2023.6.21%2B1321.8c3a4999f2) to find the right download link for your operating system. The app is currently supported on MacOS (use _aarch64.dmg for Apple Silicon or _x64.dmg for Mac Intel) and Linux (_amd64.deb), and we're working on Windows support.

## Setup

Follow the setup instructions to connect the app to your Sourcegraph.com account (or create one for free if you don't have an account yet), add your local projects, and select up to 10 of those projects to build your code graph. The code graph helps Cody generate more accurate answers about your code by sending your code to Sourcegraph to create [embeddings](../cody/explanations/code_graph_context.md#embeddings). (This may take a few minutes, depending on the size of your repos.) We are working on making it so that *all* the projects you connect to your app get embeddings and bumping the cap up from 10.   

If you use VS Code, we recommend you follow the steps to download the VS Code extension, which enables Cody within your editor. If you already have the extension, use the settings gear in the Cody chat window in the editor to log out and log back in through the app. Cody in VS Code will then talk to your Sourcegraph app to answer questions.

Note: We're only supporting VS Code right now, but a Jetbrains extension is coming soon!

## Get help & give feedback

Cody app is early stages. If you run into any trouble or have ideas/feedback, we'd love to hear from you!

* [Join our community Discord](https://discord.com/invite/s2qDtYGnAE) for live help/discussion
* [Create a GitHub issue](https://github.com/sourcegraph/app/issues/new)

## Upgrading

We're shipping new releases of the app quickly, and you should get prompts to upgrade automatically. Let us know if you have any issues and we'll be happy to help.

If you're on a version that's 2023.6.13 or older, we recommend you uninstall the app (see below) and download the most recent version in order to add your projects to your code graph during setup. Also note that these older versions of the app were called "Sourcegraph" and included Sourcegraph code search. Going forward, the app will be a Cody-only experience (and branded accordingly) so that we can focus on making Cody as useful and intuitive as possible.

## Rate limiting

There are several forms of rate limiting that help us control costs for free versions of Cody. We expect to relax these limits as we continue development on [Cody Gateway](../cody/explanations/cody_gateway.md).

### Cody Chat
Interactions with Cody Chat (whether in the app UI or in the editor extension) are capped at 100 requests per day. 

### Completions
Cody completions (autocomplete powered by Cody) are capped at 100 requests per day. Learn more about [Cody Completions](../cody/completions.md).

### Embeddings
The setup experience allows users to select up to 10 repos for embeddings. Additional repos can be added, and embeddings can be scheduled, under Settings > Advanced settings > Embedding jobs, but the number of additional repos supported will vary depending on size.  

## Uninstallation

We're working on a better way to clear all data including webkit storage, but for now you can run the following command to uninstall the app:
```bash
rm -rf ~/.sourcegraph-psql ~/Library/Application\ Support/com.sourcegraph.cody ~/Library/Caches/com.sourcegraph.cody ~/Library/WebKit/com.sourcegraph.cody
```

## Troubleshooting

Known issues: 
- The app is slow to load on initial install, due to Postgres issues we're working to tighten up. 
- In the app's Cody Chat UI, Cody will read N files to provide an answer. Those files are listed as hyperlinks, but they should not be. 
- In the settings dropdown in the app, there's an option for "Teams" that should not be there as the app is a single-player experience. 
- The cloud icon in the app UI always says "indexing" and refers to Sourcegraph code search, not projects that are being added to the app's code graph. 

See [App troubleshooting](troubleshooting.md)

## Release pipeline

See [App release pipeline](release-pipeline.md)

