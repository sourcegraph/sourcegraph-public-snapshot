---
title: 'Sourcegraph docs'
---

<a href="https://sourcegraph.com"><picture><source srcset="assets/logo-theme-dark.svg" media="(prefers-color-scheme: dark)"/><img alt="Sourcegraph" src="assets/logo-theme-light.svg" height="48px" /></picture></a>

<p class="subtitle">Universal Code Search: Move fast, even in big codebases.</p>

<p class="lead">
Find and fix things across all of your code faster with Sourcegraph. Onboard to a new codebase, make large-scale refactors, increase efficiency, address security risks, root-cause incidents, and more.
</p>

This website is home to Sourcegraph's feature, administration (including deployment and configuration), and development documentation.

<div class="cta-group">
<a class="btn btn-primary" href="#getting-started">★ Try Sourcegraph now</a>
<a class="btn" href="#core-documentation">Core docs</a>
<a class="btn" href="#cloud-documentation">Cloud docs</a>
<a class="btn" href="https://about.sourcegraph.com/">About Sourcegraph</a>
</div>

## Getting started

<div class="getting-started">
  <a href="https://sourcegraph.com/search" class="btn btn-primary" alt="Sourcegraph Cloud">
   <span>★ Sourcegraph Cloud</span>
   </br>
   <b>Search 1 million+ open source repositories or your own repositories.</b> No installation required.
  </a>

  <a href="admin/deploy" class="btn btn-primary" alt="Self-host">
   <span>★ Self-hosted instance</span>
   </br>
   Deploy and manage your own Sourcegraph instance. <b>Recommended for production deployments.</b>
  </a>
</div>

<div class="getting-started">
  <a href="admin/deploy/managed" class="btn" alt="Managed instance">
   <span>Managed instance</span>
   </br>
    Get a Sourcegraph instance provisioned and managed by the Sourcegraph team.
  </a>

  <a href="#quick-install" class="btn" alt="Quick install">
   <span>Local instance</span>
   </br>
   Quickly set up and try out Sourcegraph locally using Docker.
  </a>
</div>

<span class="virtual-br"></span>

> NOTE: Looking for how to *use* Sourcegraph? Refer to our [Using Sourcegraph guide](./getting-started/index.md)!

<span class="virtual-br"></span>

> NOTE: Unsure where to start, or need help? [Reach out to us](#get-help)!

### Try Sourcegraph locally

You can quickly try out Sourcegraph locally using Docker, which takes only a few minutes and lets you try out all of its features:

<!--
  DO NOT CHANGE THIS TO A CODEBLOCK.
  We want line breaks for readability, but backslashes to escape them do not work cross-platform.
  This uses line breaks that are rendered but not copy-pasted to the clipboard.
-->

<pre class="pre-wrap start-sourcegraph-command" id="dockerInstall"><code>docker run -d<span class="virtual-br"></span> --publish 7080:7080 --publish 127.0.0.1:3370:3370 --rm<span class="virtual-br"></span> --volume ~/.sourcegraph/config:/etc/sourcegraph<span class="virtual-br"></span> --volume ~/.sourcegraph/data:/var/opt/sourcegraph<span class="virtual-br"></span> sourcegraph/server:3.39.1<span class="iconify copy-text" data-icon="mdi:clipboard-arrow-left-outline" data-inline="false"></span></code>
</pre>

For next steps, visit the [Docker deployment documentation](admin/deploy/docker-single-container/index.md).

> NOTE: Due to the Windows enviroment, using Sourcegraph with Windows currently isn't fully supported. Testing instructions for running Sourcegraph locally on a Windows machine can be found [here](admin/deploy/docker-single-container/index.md#testing-sourcegraph-on-windows)

> WARNING: **We *do not* recommend using this method for a production instance** - see [Getting started](#getting-started) for more options.

## Core documentation

### Features and tutorials

- [Tour](getting-started/tour.md): A walkthrough of Sourcegraph's features, with real-world example use cases.
- [Using Sourcegraph](getting-started/index.md)
- [How to run a Sourcegraph trial](adopt/trial/index.md) at your company
- [Integrations](integration/index.md) with GitHub, GitLab, Bitbucket, etc.
- [Extensions](extensions/index.md)
- [Chrome and Firefox browser extensions](integration/browser_extension.md)
- [Site Administrator Quickstart](admin/how-to/site-admin-quickstart.md)

### Reference

- [Query syntax reference](code_search/reference/queries.md)
- [API Documentation](api/index.md)
- [Sourcegraph changelog](./CHANGELOG.md)

## Cloud documentation

- [Sourcegraph Cloud](code_search/explanations/sourcegraph_cloud.md)
- [Differences between Sourcegraph Cloud and self-hosted](cloud/cloud_ent_on-prem_comparison.md)
- [Indexing open source code in Sourcegraph Cloud](cloud/indexing_open_source_code.md)

## Self-hosted documentation

- [Deploy](admin/deploy/index.md) or [update](admin/updates/index.md) Sourcegraph
- [Administration](admin/index.md)

## Other links

- [Contributing to Sourcegraph](dev/index.md)
- [Sourcegraph handbook](https://handbook.sourcegraph.com/)
- [Sourcegraph blog](https://about.sourcegraph.com/blog/)
- [@sourcegraph on Twitter](https://twitter.com/sourcegraph)
- [Product Roadmap](https://handbook.sourcegraph.com/product#roadmap)

## Get help

- [File an issue](https://github.com/sourcegraph/sourcegraph/issues/new?&title=Improve+Sourcegraph+quickstart+guide)
- [Tweet (@sourcegraph)](https://twitter.com/sourcegraph)
- [Talk to an engineer](https://info.sourcegraph.com/talk-to-a-developer)
- [Talk to a product specialist](https://about.sourcegraph.com/contact/request-info/)
