# Sourcegraph documentation

[Sourcegraph](https://sourcegraph.com) Sourcegraph is a web-based, open-source, self-hosted code search and navigation tool for developers, used by Uber, Lyft, Yelp, and more.

## Quickstart guide

> NOTE: If you get stuck or need help, [file an issue](https://github.com/sourcegraph/sourcegraph/issues/new?&title=Improve+Sourcegraph+quickstart+guide), [tweet (@srcgraph)](https://twitter.com/srcgraph) or [email](mailto:support@sourcegraph.com?subject=Sourcegraph%20quickstart%20guide).


It takes less than 5 minutes to install Sourcegraph using Docker. If you've got [Docker installed](https://docs.docker.com/engine/installation/), you're ready to start the server which listens on port `7080` by default.

<!--
  DO NOT CHANGE THIS TO A CODEBLOCK.
  We want line breaks for readability, but backslashes to escape them do not work cross-platform.
  This uses line breaks that are rendered but not copy-pasted to the clipboard.
-->

<pre class="pre-wrap"><code>docker run<span class="virtual-br"></span> --publish 7080:7080 --publish 2633:2633 --rm<span class="virtual-br"></span> --volume ~/.sourcegraph/config:/etc/sourcegraph<span class="virtual-br"></span> --volume ~/.sourcegraph/data:/var/opt/sourcegraph<span class="virtual-br"></span> sourcegraph/server:3.1.2</code></pre>

Access the server on port `7080`, then the below screencast will show you how to configure Sourcegraph to search public and private repositories, and enable code intelligence on Sourcegraph and GitHub.com.

<p class="container">
  <div style="padding:56.25% 0 0 0;position:relative;">
    <iframe src="https://player.vimeo.com/video/314926561?color=0CB6F4&title=0&byline=0&portrait=0" style="position:absolute;top:0;left:0;width:100%;height:100%;" frameborder="0" webkitallowfullscreen mozallowfullscreen allowfullscreen></iframe>
  </div>
</p>

Once Sourcegraph has been configured, head to the [site administration documentation](admin/index.md) for next steps.

## Documentation

Sourcegraph development is open source at [github.com/sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph). Need help? Use the [issue tracker](https://github.com/sourcegraph/sourcegraph/issues).

### Core documentation

- [**User documentation**](user/index.md)
- [**Administrator documentation**](admin/index.md)
- [Install Sourcegraph](admin/install/index.md) or [update Sourcegraph](admin/updates.md)
- [Sourcegraph extensions](extensions/index.md)
- [Roadmap](dev/roadmap.md)

### Features and tutorials

- [Overview](user/index.md): What is Sourcegraph?
- [Tour](user/tour.md): A walkthrough of Sourcegraph's features, with real-world example use cases.
- [How to run a Sourcegraph trial](adopt/trial/index.md) at your company
- [Integrations](integration/index.md) with GitHub, GitLab, Bitbucket, etc.
- [Chrome and Firefox browser extensions](integration/browser_extension.md)
- [Query syntax reference](user/search/queries.md)
- [GraphQL API](api/graphql.md)
- [Sourcegraph Enterprise](admin/subscriptions/index.md)

<!-- TODO(sqs): Add link to ./graphbook when it has more content. -->

## Sourcegraph subscriptions

You can use Sourcegraph in 2 ways:

- [Self-hosted Sourcegraph](admin/install/index.md): Deploy and manage your own Sourcegraph instance.
- [Sourcegraph.com](https://sourcegraph.com): For public code only. No signup or installation required.

For self-hosted Sourcegraph instances, you run a Docker image or Kubernetes cluster on-premises or on your preferred cloud provider. There are [2 tiers](https://about.sourcegraph.com/pricing): Core (free) and Enterprise. Enterprise features require a [Sourcegraph subscription](https://sourcegraph.com/user/subscriptions).

## Other links

- [Sourcegraph open-source repository](https://github.com/sourcegraph/sourcegraph)
- [Contributing to Sourcegraph](dev/index.md)
- [Sourcegraph blog](https://about.sourcegraph.com/blog/)
- [Issue tracker](https://github.com/sourcegraph/sourcegraph/issues)
- [about.sourcegraph.com](https://about.sourcegraph.com) (general information about Sourcegraph)
- [@srcgraph on Twitter](https://twitter.com/srcgraph)
