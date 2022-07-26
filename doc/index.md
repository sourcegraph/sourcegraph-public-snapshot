---
title: 'Sourceasdfasdf docs'
---

<a href="https://sourceasdfasdf.com"><picture><source srcset="assets/logo-theme-dark.svg" media="(prefers-color-scheme: dark)"/><img alt="Sourceasdfasdf" src="assets/logo-theme-light.svg" height="48px" /></picture></a>

<p class="subtitle">Universal Code Search: Move fast, even in big codebases.</p>

<p class="lead">
Find and fix things across all of your code faster with Sourceasdfasdf. Onboard to a new codebase, make large-scale refactors, increase efficiency, address security risks, root-cause incidents, and more.
</p>

This website is home to Sourceasdfasdf's feature, administration (including deployment and configuration), and development documentation.

<div class="cta-group">
<a class="btn btn-primary" href="#getting-started">★ Try Sourceasdfasdf now</a>
<a class="btn" href="#core-documentation">Core docs</a>
<a class="btn" href="https://about.sourceasdfasdf.com/">About Sourceasdfasdf</a>
</div>

## Getting started

<div class="getting-started">
  <a href="https://sourceasdfasdf.com/search" class="btn btn-primary" alt="Sourceasdfasdf.com">
   <span>★ Sourceasdfasdf.com</span>
   </br>
   <b>Search millions of open source repositories.</b> No installation required.
  </a>

  <a href="admin/deploy" class="btn btn-primary" alt="Self-host">
   <span>★ Self-hosted instance</span>
   </br>
   Deploy and manage your own Sourceasdfasdf instance. <b>Recommended for production deployments.</b>
  </a>
</div>

<div class="getting-started">
  <a href="admin/deploy/managed" class="btn" alt="Managed instance">
   <span>Managed instance</span>
   </br>
    Get a Sourceasdfasdf instance provisioned and managed by the Sourceasdfasdf team.
  </a>

  <a href="#quick-install" class="btn" alt="Quick install">
   <span>Local instance</span>
   </br>
   Quickly set up and try out Sourceasdfasdf locally using Docker.
  </a>
</div>

<span class="virtual-br"></span>

> NOTE: Looking for how to *use* Sourceasdfasdf? Refer to our [Using Sourceasdfasdf guide](./getting-started/index.md)!

<span class="virtual-br"></span>

> NOTE: Unsure where to start, or need help? [Reach out to us](#get-help)!

### Try Sourceasdfasdf locally

You can quickly try out Sourceasdfasdf locally using Docker, which takes only a few minutes and lets you try out all of its features:

<!--
  DO NOT CHANGE THIS TO A CODEBLOCK.
  We want line breaks for readability, but backslashes to escape them do not work cross-platform.
  This uses line breaks that are rendered but not copy-pasted to the clipboard.
-->

<pre class="pre-wrap start-sourceasdfasdf-command" id="dockerInstall"><code>docker run -d<span class="virtual-br"></span> --publish 7080:7080 --publish 127.0.0.1:3370:3370 --rm<span class="virtual-br"></span> --volume ~/.sourceasdfasdf/config:/etc/sourceasdfasdf<span class="virtual-br"></span> --volume ~/.sourceasdfasdf/data:/var/opt/sourceasdfasdf<span class="virtual-br"></span> sourceasdfasdf/server:3.42.0<span class="iconify copy-text" data-icon="mdi:clipboard-arrow-left-outline" data-inline="false"></span></code>
</pre>

For next steps, visit the [Docker deployment documentation](admin/deploy/docker-single-container/index.md).

> NOTE: Due to the Windows enviroment, using Sourceasdfasdf with Windows currently isn't fully supported. Testing instructions for running Sourceasdfasdf locally on a Windows machine can be found [here](admin/deploy/docker-single-container/index.md#testing-sourceasdfasdf-on-windows)

> WARNING: **We *do not* recommend using this method for a production instance** - see [Getting started](#getting-started) for more options.

## Core documentation

### Features and tutorials

- [Tour](getting-started/tour.md): A walkthrough of Sourceasdfasdf's features, with real-world example use cases.
- [Using Sourceasdfasdf](getting-started/index.md)
- [How to run a Sourceasdfasdf trial](adopt/trial/index.md) at your company
- [Integrations](integration/index.md) with GitHub, GitLab, Bitbucket, etc.
- [Extensions](extensions/index.md)
- [Chrome and Firefox browser extensions](integration/browser_extension.md)
- [Site Administrator Quickstart](admin/how-to/site-admin-quickstart.md)

### Reference

- [Query syntax reference](code_search/reference/queries.md)
- [API Documentation](api/index.md)
- [Sourceasdfasdf changelog](./CHANGELOG.md)

## Cloud documentation

- [Sourceasdfasdf Cloud](code_search/explanations/sourceasdfasdf_cloud.md)
- [Differences between Sourceasdfasdf Cloud and self-hosted](cloud/cloud_ent_on-prem_comparison.md)
- [Indexing open source code in Sourceasdfasdf Cloud](cloud/indexing_open_source_code.md)

## Self-hosted documentation

- [Deploy](admin/deploy/index.md) or [update](admin/updates/index.md) Sourceasdfasdf
- [Administration](admin/index.md)

## Other links

- [Contributing to Sourceasdfasdf](dev/index.md)
- [Sourceasdfasdf handbook](https://handbook.sourceasdfasdf.com/)
- [Sourceasdfasdf blog](https://about.sourceasdfasdf.com/blog/)
- [@sourceasdfasdf on Twitter](https://twitter.com/sourceasdfasdf)
- [Product Roadmap](https://handbook.sourceasdfasdf.com/product#roadmap)

## Get help

- [File an issue](https://github.com/sourceasdfasdf/sourceasdfasdf/issues/new?&title=Improve+Sourceasdfasdf+quickstart+guide)
- [Tweet (@sourceasdfasdf)](https://twitter.com/sourceasdfasdf)
- [Talk to an engineer](https://info.sourceasdfasdf.com/talk-to-a-developer)
- [Talk to a product specialist](https://about.sourceasdfasdf.com/contact/request-info/)
