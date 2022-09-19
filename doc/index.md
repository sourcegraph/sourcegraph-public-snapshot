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
    border: 1px solid;
    background-color: white;
    text-align: center;
    font-weight: 100;
  }
  .app-btn:hover {
    box-shadow: 0 0 10px #00cbec;
  }
  .app-btn > img {
    height: 4em;
  }
  .app-btn > h3 {
    font-size: 1.5em;
    font-weight: lighter;
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

### Self-hosted

<form class="grid">
  <!-- AWS -->
  <button class="app-btn btn" formaction="/admin/deploy/docker-compose/aws">
    <img alt="aws-logo" src="https://user-images.githubusercontent.com/1646931/187976316-727d2b75-ff90-43ee-acfb-b63dc4b615f2.png"/>
    <h3>AWS</h3>
    <p>Deploy onto AWS EC2</p>
  </button>
  <!-- AWS One Click-->
  <button class="app-btn btn" formaction="/admin/deploy/docker-compose/aws-oneclick">
    <img alt="aws-logo" src="https://user-images.githubusercontent.com/1646931/187976316-727d2b75-ff90-43ee-acfb-b63dc4b615f2.png"/>
    <h3>AWS One-Click</h3>
    <span class="badge badge-warning">Coming soon</span> 
    <!-- <p>Deploy onto AWS in one click</p> -->
  </button>
  <!-- Azure -->
  <button class="app-btn btn" formaction="/admin/deploy/docker-compose/azure">
    <img alt="azure-logo" src="https://user-images.githubusercontent.com/1646931/187978161-771cfb91-6cb3-4f00-befd-657502b95ed4.png"/>
    <h3>Azure</h3>
    <p>Deploy onto Microsoft Azure</p>
  </button>
  <!-- digital ocean -->
  <button class="app-btn btn" formaction="/admin/deploy/docker-compose/digitalocean">
    <img alt="digital-ocean-logo" src="https://res.cloudinary.com/crunchbase-production/image/upload/c_lpad,h_170,w_170,f_auto,b_white,q_auto:eco,dpr_1/v1478792253/gnlwek2zwhq369yryrzv.jpg"/>
    <h3>DigitalOcean</h3>
    <p>Deploy onto Digital Ocean</p>
  </button>
  <!-- Docker Compose -->
  <button class="app-btn btn" formaction="/admin/install/docker-compose">
    <img alt="docker-compose-logo" src="https://user-images.githubusercontent.com/1646931/187978472-1219f3a0-8c89-433c-8a72-223228952814.png"/>
    <h3>Docker Compose</h3>
    <p>Deploy with Docker Compose</p>
  </button>
  <!-- GCP -->
  <button class="app-btn btn" formaction="/admin/deploy/docker-compose/google_cloud">
    <img alt="gcp-logo" src="https://user-images.githubusercontent.com/1646931/187977350-3618e506-6fab-47c5-9a7c-286484cbd5a8.png"/>
    <h3>Google Cloud</h3>
    <p>Deploy onto Google Cloud (GCP)</p>
  </button>
  <!-- Others -->
  <button class="app-btn btn" formaction="/admin/deploy">
    <img alt="private-cloud-logo" src="https://user-images.githubusercontent.com/1646931/187978634-6c4b2d06-2808-497d-8069-7adbee5bc703.png"/>
    <h3>Private cloud</h3>
    <p>Deploy into a generic cloud environment</p>
  </button>
  <!-- Kubernetes -->
  <button class="app-btn btn" formaction="/admin/deploy/kubernetes">
    <img alt="kubernetes-logo" src="https://user-images.githubusercontent.com/1646931/187978853-ee9efe0b-a18c-45a1-8375-c6c29647342a.png"/>
    <h3>Kubernetes</h3>
    <p><strong>Enterprise-only</strong></p>
	<p>Deploy a multi-node cluster</p>
  </button>
</form>

### Local machine

<form class="grid">
  <button class="app-btn btn" formaction="/admin/deploy/docker-single-container">
    <img alt="docker-logo" src="https://user-images.githubusercontent.com/1646931/187978472-1219f3a0-8c89-433c-8a72-223228952814.png"/>
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
