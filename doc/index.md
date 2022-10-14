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
    font-size: 0.85rem;
    cursor: pointer;
    text-decoration: none;
    padding-top: 1.5rem !important;
    width: 100%;
    border-radius: 1em;
    border: 2px solid var(--input-focus-border);
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
  .app-btn > p {
    margin-bottom: 0 !important;
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

- For most customers, we recommend Sourcegraph Cloud. A Sourcegraph Cloud instance is a single-tenant instance that is managed entirely by Sourcegraph.
- For customers that want to self-host, we recommend one of the single-node [deployment options](admin/deploy/index.md).
- For enterprise customers that require a multi-node, self-hosted deployment, we offer a Kubernetes option. We strongly encourage you to get in touch by email (sales@sourcegraph.com) if you pursue this option.

### Recommended

<div class="grid">
  <!-- Sourcegraph Cloud -->
  <a class="app-btn btn" href="/cloud">
			<img alt="sourcegraph-logo" src="https://handbook.sourcegraph.com/departments/engineering/design/brand_guidelines/logo/versions/Sourcegraph_Logomark_Color.svg"/>
			<h3>Sourcegraph Cloud</h3>
		  <p>Create a single-tenant instance managed by Sourcegraph</p>
  </a>
</div>

### Self-hosted

<div class="grid">
  <!-- AWS AMI-->
  <a class="app-btn btn" href="/admin/deploy/machine-images/aws-ami">
    <img alt="aws-logo" src="/assets/other-logos/aws-light.svg" class="theme-light-only" />
    <img alt="aws-logo" src="/assets/other-logos/aws-dark.svg" class="theme-dark-only" />
    <h3>AWS</h3>
    <p>Launch a pre-configured Sourcegraph instance from an AWS AMI</p>
  </a>
</div>

<div class="grid">
  <!-- Azure -->
  <a class="app-btn btn" href="/admin/deploy/docker-compose/azure">
    <img alt="azure-logo" src="https://storage.googleapis.com/sourcegraph-resource-estimator/assets/azure.png"/>
    <h3>Azure</h3>
    <p>Deploy onto Microsoft Azure</p>
  </a>
  <!-- AWS One Click-->
  <a class="app-btn btn" href="/admin/deploy/docker-compose/aws-oneclick">
    <img alt="aws-logo" src="/assets/other-logos/aws-light.svg" class="theme-light-only" />
    <img alt="aws-logo" src="/assets/other-logos/aws-dark.svg" class="theme-dark-only" />
    <h3>AWS One-Click</h3>
    <span class="badge badge-warning">Coming soon</span> 
    <!-- <p>Deploy onto AWS in one click</p> -->
  </a>
  <!-- Digital Ocean -->
  <a class="app-btn btn" href="/admin/deploy/docker-compose/digitalocean">
    <img alt="digital-ocean-logo" src="https://storage.googleapis.com/sourcegraph-resource-estimator/assets/DigitalOcean.png"/>
    <h3>DigitalOcean</h3>
    <p>Deploy onto DigitalOcean</p>
  </a>
  <!-- Docker Compose -->
  <a class="app-btn btn" href="/admin/deploy/docker-compose">
    <img alt="docker-compose-logo" src="https://storage.googleapis.com/sourcegraph-resource-estimator/assets/Docker.png"/>
    <h3>Docker Compose</h3>
    <p>Deploy with Docker Compose</p>
  </a>
  <!-- GCP -->
  <a class="app-btn btn" href="/admin/deploy/docker-compose/google_cloud">
    <img alt="gcp-logo" src="https://storage.googleapis.com/sourcegraph-resource-estimator/assets/googlecloud.png"/>
    <h3>Google Cloud</h3>
    <p>Deploy onto Google Cloud (GCP)</p>
  </a>
  <!-- Others -->
  <a class="app-btn btn" href="/admin/deploy">
    <img alt="private-cloud-logo" src="https://storage.googleapis.com/sourcegraph-resource-estimator/assets/cloud.png"/>
    <h3>Private cloud</h3>
    <p>Deploy into a generic cloud environment</p>
  </a>
  <!-- Kubernetes -->
  <a class="app-btn btn" href="/admin/deploy/kubernetes">
    <img alt="kubernetes-logo" src="https://storage.googleapis.com/sourcegraph-resource-estimator/assets/kubernetes.png"/>
    <h3>Kubernetes</h3>
	  <p>Deploy a multi-node cluster</p>
    <p><strong>Enterprise-only</strong></p>
  </a>
</div>

### Local machine

<div class="grid">
  <a class="app-btn btn" href="/admin/deploy/docker-single-container">
    <img alt="docker-logo" src="https://storage.googleapis.com/sourcegraph-resource-estimator/assets/Docker.png"/>
    <h3>Docker Container</h3>
    <p>Spin up a local Sourcegraph instance</p>
  </a>
  <div></div><div></div>
</div>

---

## Quickstart

- [Learn Sourcegraph](getting-started/index.md)
  - Sourcegraph 101: how to use Sourcegraph
- [Tour Sourcegraph](getting-started/tour.md)
  - Take a tour of Sourcegraph’s features using real-world examples and use cases
- [Trial Sourcegraph](adopt/trial/index.md)
  - Start a Sourcegraph trail at your company

## Community

- [Blog](https://about.sourcegraph.com/blog/)
- [Discord](https://discord.gg/s2qDtYGnAE)
- [Twitter](https://twitter.com/sourcegraph)
- [Handbook](https://handbook.sourcegraph.com/)

## Support

- [File an issue](https://github.com/sourcegraph/sourcegraph/issues/new?&title=Improve+Sourcegraph+quickstart+guide)
- [Request a demo](https://about.sourcegraph.com/demo)
- [Talk to a product specialist](https://about.sourcegraph.com/contact/request-info/)
