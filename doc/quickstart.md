# Sourcegraph quickstart guide

> NOTE: If you get stuck or need help, [file an issue](https://github.com/sourcegraph/sourcegraph/issues/new?&title=Improve+Sourcegraph+quickstart+guide), [tweet (@srcgraph)](https://twitter.com/srcgraph) or [email](mailto:support@sourcegraph.com?subject=Sourcegraph%20quickstart%20guide).


It takes less than 5 minutes to install Sourcegraph using Docker. If you've got [Docker installed](https://docs.docker.com/engine/installation/), you're ready to start the server which listens on port `7080` by default.

<!--
  DO NOT CHANGE THIS TO A CODEBLOCK.
  We want line breaks for readability, but backslashes to escape them do not work cross-platform.
  This uses line breaks that are rendered but not copy-pasted to the clipboard.
-->

<pre class="pre-wrap"><code>docker run<span class="virtual-br"></span> --publish 7080:7080 --publish 2633:2633 --rm<span class="virtual-br"></span> --volume ~/.sourcegraph/config:/etc/sourcegraph<span class="virtual-br"></span> --volume ~/.sourcegraph/data:/var/opt/sourcegraph<span class="virtual-br"></span> sourcegraph/server:3.1.1</code></pre>

Access the server on port `7080`, then the below screencast will show you how to configure Sourcegraph to search public and private repositories, and enable code intelligence on Sourcegraph and GitHub.com.

<p class="container">
  <div style="padding:56.25% 0 0 0;position:relative;">
    <iframe src="https://player.vimeo.com/video/314926561?color=0CB6F4&title=0&byline=0&portrait=0" style="position:absolute;top:0;left:0;width:100%;height:100%;" frameborder="0" webkitallowfullscreen mozallowfullscreen allowfullscreen></iframe>
  </div>
</p>

Once Sourcegraph has been configured, head to the [site administration documentation](admin/index.md) for next steps.

## Learn about Sourcegraph

Learn more about Sourcegraph at [Sourcegraph.com](https://sourcegraph.com/start) where you can also use it for [searching open source repositories](https://sourcegraph.com/search).

Sourcegraph development is open source at [github.com/sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph). Need help? Use the [issue tracker](https://github.com/sourcegraph/sourcegraph/issues).

## Next steps

- [Core documentation](index.md#core-documentation)
- [Features and tutorials](index.md#features-and-tutorials)
- [Extensions management](extensions/index.md#usage)
- [Sourcegraph subscriptions](index.md#sourcegraph-subscriptions)
- [Other links](index.md#other-links)
