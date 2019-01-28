# Sourcegraph quickstart guide

Sourcegraph is used by companies like Uber and Lyft to help developers search, navigate and review code at enterprise scale.

It takes less than 5 minutes to install and configure a self-hosted instance with GitHub integration and code intelligence enabled.

> NOTE: If you get stuck or need help, [tweet (@srcgraph)](https://twitter.com/srcgraph), [email](mailto:support@sourcegraph.com?subject=Sourcegraph%20quickstart%20guide) or [file an issue](https://github.com/sourcegraph/sourcegraph/issues/new?assignees=ryan-blunden&labels=docs&template=bug_report.md&title=Improve+Sourcegraph+quickstart+guide).

### 1. Install Sourcegraph

If you've got [Docker installed](https://docs.docker.com/engine/installation/), you're ready to start the server which listens on port `7080` by default.

<!--
  DO NOT CHANGE THIS TO A CODEBLOCK.
  We want line breaks for readability, but backslashes to escape them do not work cross-platform.
  This uses line breaks that are rendered but not copy-pasted to the clipboard.
-->
<pre class="pre-wrap"><code>docker run<span class="virtual-br"></span> --publish 7080:7080 --publish 2633:2633 --rm<span class="virtual-br"></span> --volume ~/.sourcegraph/config:/etc/sourcegraph<span class="virtual-br"></span> --volume ~/.sourcegraph/data:/var/opt/sourcegraph<span class="virtual-br"></span> sourcegraph/server:3.0.0-beta</code></pre>

<div class="container media-youtube">
  <iframe class="mx-auto" width="560" height="315" src="https://www.youtube.com/embed/HZAGneTpsyA" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
</div>

### 2. Sync repositories from GitHub

Select Git repositories for searching by integrating Sourcegraph with GitHub.

<div class="container media-youtube">
  <iframe class="mx-auto" width="560" height="315" src="https://www.youtube.com/embed/NvPeBtI9gkA" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
</div>

### 3. Enable code intelligence

Enable basic code intelligence (go-to-definition, hover tooltips, and find-references) for all languages.

<div class="container media-youtube">
  <iframe class="mx-auto" width="560" height="315" src="https://www.youtube.com/embed/Pfy2CjeJ2N4" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
</div>

## Configuration complete

Now you should have a fully functioning Sourcegraph instance. If something didn't work or you've got a suggestion for making this guide better, [tweet](https://twitter.com/srcgraph), [email](mailto:ryan@sourcegraph.com?subject=Sourcegraph%20quickstart%20guide) or [file an issue](https://github.com/sourcegraph/sourcegraph/issues/new?assignees=ryan-blunden&labels=docs&template=bug_report.md&title=Improve+Sourcegraph+quickstart+guide).

## Learn about Sourcegraph

Learn more about Sourcegraph at [Sourcegraph.com](https://sourcegraph.com/start) or use it for public repositories [Sourcegraph.com](https://sourcegraph.com/search).

Sourcegraph development is open source at [github.com/sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph). Need help? Use the [issue tracker](https://github.com/sourcegraph/sourcegraph/issues).

## Next steps

- [Core documentation](index.md#core-documentation)
- [Features and tutorials](index.md#features-and-tutorials)
- [Sourcegraph subscriptions](index.md#sourcegraph-subscriptions)
- [Other links](index.md#other-links)
