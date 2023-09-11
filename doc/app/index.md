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

The app is a free, lightweight, native desktop version of Sourcegraph that connects your local code to our AI coding assistant, Cody. You can ask Cody questions about all the code you connect to your app in both the app interface and, if you connect the VS code extension, in your editor. 

<div class="socials">
  <a href="https://discord.com/invite/s2qDtYGnAE"><img alt="Discord" src="discord.svg"></img></a>
  <a href="https://twitter.com/sourcegraph"><img alt="Twitter" src="twitter.svg"></img></a>
  <a href="https://github.com/sourcegraph/app"><img alt="GitHub" src="github.svg"></img></a>
</div>

## Installation

Check the [latest release](https://github.com/sourcegraph/sourcegraph/releases/tag/app-v2023.6.13%2B1311.1af08ae774) to find the right download link for your operating system. The app is currently supported on MacOS and Linux, and we're working on Windows support.

## Setup

Follow the setup instructions to connect the app to your Sourcegraph.com account (or create one for free if you don't have an account yet) and add your local projects. 

If you use VS Code or a JetBrains IDE, we recommend you follow the steps to download the extension, which enables Cody within your editor. (If you installed the extension before you downloaded the app, you'll see a prompt in your editor to download the app.) Cody in your editor will then talk to your Sourcegraph app to answer questions.

Note: The JetBrains extension is still `Experimental`. 

### (Optional) batch changes & precise code intel

Batch changes and precise code intel require the following optional dependencies be installed and on your PATH:

- The `src` CLI ([installation](https://sourcegraph.com/github.com/sourcegraph/src-cli))
- `docker`

## Tips

- Use the app icon in your system tray to open a Cody chat window. This is especially helpful when arranged alongside your IDE if you prefer an editor other than VS Code. 

## Get help & give feedback

Cody app is early-stages, if you run into any trouble or have ideas/feedback, we'd love to hear from you!

* [Join our community Discord](https://discord.com/invite/s2qDtYGnAE) for live help/discussion
* [Create a GitHub issue](https://github.com/sourcegraph/app/issues/new)

## Upgrading

We're shipping new releases of the app quickly, and you should get prompts to upgrade automatically. Let us know if you have any issues!

## Uninstallation

We're working on a better way to clear all data including webkit storage, but for now you can run `rm -rf ~/.sourcegraph-psql ~/Library/Application\ Support/com.sourcegraph.cody ~/Library/Caches/com.sourcegraph.cody ~/Library/WebKit/com.sourcegraph.cody` to uninstall the app.

## Troubleshooting

See [App troubleshooting](troubleshooting.md)

## Release pipeline

See [App release pipeline](release-pipeline.md)

