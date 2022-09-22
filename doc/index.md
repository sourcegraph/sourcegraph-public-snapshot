---
title: Get Started
---

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1em;
    margin-bottom: 1em;
  }
  .app-btn {
    cursor: pointer;
    text-decoration: none;
    height: 15em;
    width: 100%;
    border-radius: 1em;
    border: 2px solid var(--input-focus-bord);
    color: var(--text-color);
    background-color: var(--sidebar-bg);
    text-align: center;
    font-weight: 500;
    -moz-osx-font-smoothing: grayscale;
    -webkit-font-smoothing: antialiased;
  }
  .app-btn:hover {
    box-shadow: 0 0 10px var(--link-hover-color);
  }
  .app-btn > img {
    height: 4em;
  }
  .app-btn > h3 {
    font-size: 1.5em;
    font-weight: 400;
    margin-top: .2em;
    margin-bottom: 1em;
  }
</style>

# Get Started

## Deploy Sourcegraph

Sourcegraph is runnable in a variety of environments, from cloud to self-hosted to your local machine.

* For most customers, we recommend Sourcegraph Cloud, a single-tenant, auto-managed, and auto-upgrading option.
* For customers that desire to self-host, we recommend one of the single-node deployment options.
* For enterprise customers that require a multi-node, self-hosted deployment, we offer a Kubernetes option. We strongly encourage you to get in touch via Discord or email if you pursue this option.

### Recommended

<form class="grid">
  <!-- Sourcegraph Cloud -->
  <button class="app-btn btn" formaction="cloud">
			<img alt="sourcegraph-logo" src="https://handbook.sourcegraph.com/departments/engineering/design/brand_guidelines/logo/versions/Sourcegraph_Logomark_Color.svg"/>
			<h3>Sourcegraph Cloud</h3>
		  <p>Create a single-tenant instance managed by Sourcegraph</p>
  </button>
</form>
<form class="grid">
  <!-- AWS One Click-->
  <button class="app-btn btn" formaction="/admin/deploy/docker-compose/aws-oneclick">
    <img alt="aws-logo" src="https://user-images.githubusercontent.com/68532117/191821148-ebc139f4-3185-4995-823b-c9402eb3b009.png"/>
    <h3>AWS One-Click</h3>
    <span class="badge badge-warning">Coming soon</span> 
    <!-- <p>Deploy onto AWS in one click</p> -->
  </button>
</form>

### Self-hosted

<form class="grid">
  <!-- Azure -->
  <button class="app-btn btn" formaction="/admin/deploy/docker-compose/azure">
    <img alt="azure-logo" src="https://user-images.githubusercontent.com/68532117/191821152-01a6d6c4-1ed5-4486-92e6-29f09a528c01.png"/>
    <h3>Azure</h3>
    <p>Deploy onto Microsoft Azure</p>
  </button>
  <!-- Digital Ocean -->
  <button class="app-btn btn" formaction="/admin/deploy/docker-compose/digitalocean">
    <img alt="digital-ocean-logo" src="https://seeklogo.com/images/D/digital-ocean-logo-7B970FE624-seeklogo.com.png"/>
    <h3>DigitalOcean</h3>
    <p>Deploy onto Digital Ocean</p>
  </button>
  <!-- Docker Compose -->
  <button class="app-btn btn" formaction="/admin/install/docker-compose">
    <img alt="docker-compose-logo" src="https://user-images.githubusercontent.com/68532117/191821158-3c82043e-dc2c-4309-af15-2541d81d42c4.png"/>
    <h3>Docker Compose</h3>
    <p>Deploy with Docker Compose</p>
  </button>
  <!-- GCP -->
  <button class="app-btn btn" formaction="/admin/deploy/docker-compose/google_cloud">
    <img alt="gcp-logo" src="https://user-images.githubusercontent.com/68532117/191821160-bbfbe825-c75b-4d02-bd80-69e32aea7050.png"/>
    <h3>Google Cloud</h3>
    <p>Deploy onto Google Cloud (GCP)</p>
  </button>
  <!-- Others -->
  <button class="app-btn btn" formaction="/admin/deploy">
    <img alt="private-cloud-logo" src="https://user-images.githubusercontent.com/68532117/191821156-a9f2f88b-b4a0-4c60-8d46-bec9a72bf82f.png"/>
    <h3>Private cloud</h3>
    <p>Deploy into a generic cloud environment</p>
  </button>
  <!-- Kubernetes -->
  <button class="app-btn btn" formaction="/admin/deploy/kubernetes">
    <img alt="kubernetes-logo" src="https://user-images.githubusercontent.com/68532117/191821162-625f91fa-f5be-4fa7-a265-1393a7edac20.png"/>
    <h3>Kubernetes</h3>
    <p><strong>Enterprise-only</strong></p>
	<p>Deploy a multi-node cluster</p>
  </button>
</form>

### Local machine

<form class="grid">
  <button class="app-btn btn" formaction="/admin/deploy/docker-single-container">
    <img alt="docker-logo" src="https://user-images.githubusercontent.com/68532117/191821158-3c82043e-dc2c-4309-af15-2541d81d42c4.png"/>
    <h3>Docker Container</h3>
    <p>Spin up a local Sourcegraph instance</p>
  </button>
  <div></div><div></div>
</form>

---

## Quickstart

- [Learn Sourcegraph](getting-started/index.md) 
  - Sourcegraph 101: how to use Sourcegraph
- [Tour Sourcegraph](getting-started/tour.md)
  - Take a tour of Sourcegraph’s features using real-world examples and use cases
- [Trial Sourcegraph](adopt/trial/index.md) 
  - Start a Sourcegraph trail at your company
- [Sourcegraph AWS launch stack](admin/deploy/docker-compose/aws-oneclick.md) 
  - Launch a Sourcegraph instance in one-click

## Community

- [Blog](https://about.sourcegraph.com/blog/)
- [Discord](https://discord.gg/s2qDtYGnAE) 
- [Twitter](https://twitter.com/sourcegraph)
- [Handbook](https://handbook.sourcegraph.com/)

## Support

- [File an issue](https://github.com/sourcegraph/sourcegraph/issues/new?&title=Improve+Sourcegraph+quickstart+guide)
- [Request a demo](https://about.sourcegraph.com/demo)
- [Talk to a product specialist](https://about.sourcegraph.com/contact/request-info/)
