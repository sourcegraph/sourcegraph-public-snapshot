---
title: 'Sourcegraph docs'
---

<a href="https://sourcegraph.com"><picture><source srcset="assets/logo-theme-dark.svg" media="(prefers-color-scheme: dark)"/><img alt="Sourcegraph" src="assets/logo-theme-light.svg" height="48px" /></picture></a>

<p class="subtitle">Code search and intelligence</p>

<p class="lead">
Understand, fix, and automate across your codebase with Sourcegraph.
</p>

### Try Sourcegraph

You can quickly try Sourcegraph locally using Docker:

<!--
  DO NOT CHANGE THIS TO A CODEBLOCK.
  We want line breaks for readability, but backslashes to escape them do not work cross-platform.
  This uses line breaks that are rendered but not copy-pasted to the clipboard.
-->

<pre class="pre-wrap start-sourcegraph-command" id="dockerInstall"><code>docker run<span class="virtual-br"></span> --publish 7080:7080 --publish 127.0.0.1:3370:3370 --rm<span class="virtual-br"></span> --volume ~/.sourcegraph/config:/etc/sourcegraph<span class="virtual-br"></span> --volume ~/.sourcegraph/data:/var/opt/sourcegraph<span class="virtual-br"></span> sourcegraph/server:3.43.0<span class="iconify copy-text" data-icon="mdi:clipboard-arrow-left-outline" data-inline="false"></span></code>
</pre>

For more information, see [single-container Docker deployment documentation](admin/deploy/docker-single-container/index.md).

To use Sourcegraph on 2 million open-source repositories, visit [Sourcegraph.com](https://sourcegraph.com/search).

**For production deployments:** [use Sourcegraph Cloud](cloud/index.md) or [deploy self-hosted Sourcegraph](admin/deploy/index.md).

## Popular documentation

- [Tour](getting-started/tour.md): A walkthrough of Sourcegraph's features, with real-world example use cases.
- [Using Sourcegraph](getting-started/index.md)
- [How to run a Sourcegraph trial](adopt/trial/index.md) at your company
- [Integrations](integration/index.md) with GitHub, GitLab, Bitbucket, etc.
- [Chrome and Firefox browser extensions](integration/browser_extension.md)
- Reference:
  - [Query syntax reference](code_search/reference/queries.md)
  - [API documentation](api/index.md)
  
### Other links

- [Sourcegraph changelog](./CHANGELOG.md)
- [Sourcegraph handbook](https://handbook.sourcegraph.com/)
- [Sourcegraph blog](https://about.sourcegraph.com/blog/)
- [@sourcegraph on Twitter](https://twitter.com/sourcegraph)
- [File an issue](https://github.com/sourcegraph/sourcegraph/issues/new?&title=Improve+Sourcegraph+quickstart+guide)
- [Request a demo](https://about.sourcegraph.com/demo)
- [Talk to a product specialist](https://about.sourcegraph.com/contact/request-info/)
