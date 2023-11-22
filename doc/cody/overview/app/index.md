<style>
  .markdown-body .cards {
  display: flex;
  align-items: stretch;
}

.markdown-body .cards .card {
  flex: 1;
  margin: 0.5em;
  color: var(--text-color);
  border-radius: 4px;
  border: 1px solid var(--sidebar-nav-active-bg);
  padding: 1.5rem;
  padding-top: 1.25rem;
}

.markdown-body .cards .card:hover {
  color: var(--link-color);
}

.markdown-body .cards .card span {
  color: var(--link-color);
  font-weight: bold;
}

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

.limg {
  list-style: none;
  margin: 3rem 0 !important;
  padding: 0 !important;
}
.limg li {
  margin-bottom: 1rem;
  padding: 0 !important;
}

.limg li:last {
  margin-bottom: 0;
}

.limg a {
    display: flex;
    flex-direction: column;
    transition-property: all;
   transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
     transition-duration: 350ms;
     border-radius: 0.75rem;
  padding-top: 1rem;
  padding-bottom: 1rem;

}

.limg a:hover {
  padding-left: 1rem;
  padding-right: 1rem;
  background: rgb(113 220 232 / 19%);
}

.limg p {
  margin: 0rem;
}
.limg a img {
  width: 1.5rem;
}

.limg h3 {
  display:flex;
  gap: 0.6rem;
  margin-top: 0;
  margin-bottom: .25rem

}
</style>

# Cody App

<p class="subtitle">Learn how to use Cody and its features with the native Cody app.</p>

The Cody app is a free, lightweight, native desktop application that connects your local code to our AI coding assistant, Cody. You can ask Cody questions about your code using the app's interface. However, Cody will be available directly inside your code editor if connected to the VS Code extension.

## Installation

You can download and install the Cody desktop app for the following:

<ul class="limg">
  <li>
    <a class="card text-left" target="_blank" href="https://sourcegraph.com/.api/app/latest?arch=aarch64&target=darwin">
    <h3><img alt="macOS (Apple Silicon)" src="https://storage.googleapis.com/sourcegraph-assets/Docs/mac-logo.png"/> macOS (Apple Silicon)</h3>
    <p>Install Cody app for your Apple Silicon computers.</p>
    </a>
  </li>
  <li>
    <a class="card text-left" target="_blank" href="https://sourcegraph.com/.api/app/latest?arch=x86_64&target=darwin">
      <h3><img alt="macOS(Intel)" src="https://storage.googleapis.com/sourcegraph-assets/Docs/mac-logo.png" />macOS (Intel)</h3>
      <p>Install Cody app for your Apple Intel-based computers.</p>
    </a>
  </li>
  <li>
     <a class="card text-left" target="_blank" href="https://sourcegraph.com/.api/app/latest?arch=x86_64&target=linux">
      <h3><img alt="Linux" src="https://storage.googleapis.com/sourcegraph-assets/Docs/linux-icon.png"/>Linux</h3>
      <p>Install Cody app for your Linux-based computers.</p>
      </a>
  </li>
</ul>

> NOTE: The Cody app is not yet available for Windows. However, you can use the [Cody extension for VS Code](./../install-vscode.md) on Windows.

## Setup

After a successful installation, follow these steps to complete the app setup:

- Sign in and connect the app with your Sourcegraph.com account, or create a new account if you haven't got one
- Next, select up to 10 local repositories to add to your code graph
- This code is sent to OpenAI to create [embeddings](./../../core-concepts/embeddings.md), which helps Cody build the code graph and generate more accurate answers about your code
- If you use [VS Code](./../install-vscode.md) or [JetBrains IntelliJ](./../install-jetbrains.md) IDEs, it's recommended to install their extensions and Ask Cody questions right within your editor

> NOTE: The [JetBrains IntelliJ](./../install-jetbrains.md) extension is in the `Experimental` stage.

## Rate limiting

Several forms of rate limiting help us manage costs for free versions of Cody. We're working actively to relax these limits and facilitate users with [Cody Gateway](./../../core-concepts/cody-gateway.md).

If you hit these limits, you can request an increased limit by visiting our [Discord](https://discord.com/servers/sourcegraph-969688426372825169) channel for both chats and completions. If you'd like to use your third-party LLM provider instead of Cody Gateway, create your key with Anthropic or OpenAI and [update your app configuration](app_configuration.md).

### Embeddings

The default Cody app setup allows you to select up to 10 repos for embeddings. However, you can add more repos and "Schedule Embeddings". Go to **Cody Settings > Advanced settings > Embedding jobs**, and you can add and schedule repositories for embeddings.

However, these additional number of supported repos will vary depending on their size.

## Updating the app

You'll get prompts to upgrade the app automatically. Let us know if you have any issues, and we'll be happy to help.

If you're on a version that's `2023.6.13` or older, it's recommended to [uninstall](#uninstalling-the-app) the app and download the most recent version to add your projects to your code graph during the setup.

> NOTE: The old app versions were called "Sourcegraph" and included Sourcegraph [Code Search](./../../../code_search.md). However, the recent versions provide a Cody-only experience.

## Uninstalling the app

To uninstall the app, navigate to **Troubleshooting > Clear All Data** from the system tray and delete the app from your applications folder.

If you're on an older version of the app and don't see a "Clear All Data" option, run the following script in your terminal:

```shell
rm -rf ~/.sourcegraph-psql ~/Library/Application\ Support/com.sourcegraph.cody ~/Library/Caches/com.sourcegraph.cody ~/Library/WebKit/com.sourcegraph.cody
```

## Get help give feedback

If you run into any trouble or have ideas/feedback, you can reach out via the following:

<div class="socials">
  <a href="https://discord.com/invite/s2qDtYGnAE"><img alt="Discord" src="discord.svg"></img></a>
  <a href="https://twitter.com/sourcegraph"><img alt="Twitter" src="twitter.svg"></img></a>
  <a href="https://github.com/sourcegraph/app"><img alt="GitHub" src="github.svg"></img></a>
</div>

## More benefits

Read more about [App API integrations](integrations.md) to learn about how extensions and other clients can integrate with Cody app.

## More resources

For more information on what to do next, we recommend the following resources:

<div class="cards">
  <a class="card text-left" href="./troubleshooting"><b>Troubleshooting</b><p>See app troubleshooting guide if you run into any issues.</p></a>
  <a class="card text-left" href="./release-pipeline"><b>Release pipeline</b><p>Read more about our app release pipeline to get all updates.</p></a>
</div>
