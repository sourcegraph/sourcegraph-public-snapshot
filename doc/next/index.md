---
title: Deploy Sourcegraph
---

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1em;
  }
  .app-btn {
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

# Deploy Sourcegraph

Get started with one of the available deployment methods below, from creating a managed instance hosted by Sourcegraph Cloud to setting up a self-hosted instance with [AWS One-Click](./aws-oneclick.md) --we've got you covered!

## Recommended

<form class="grid">
  <button class="app-btn btn" formaction="#TODO">
			<img alt="sourcegraph-logo" src="https://handbook.sourcegraph.com/departments/engineering/design/brand_guidelines/logo/versions/Sourcegraph_Logomark_Color.svg"/>
			<h3>Sourcegraph Cloud</h3>
		  <p>Create a single-tenant instance managed by Sourcegraph</p>
  </button>
</form>

## Self-Hosted

<form class="grid">
  <!-- AWS -->
  <button class="app-btn btn" formaction="/admin/deploy/docker-compose/aws">
    <img alt="aws-logo" src="https://user-images.githubusercontent.com/1646931/187976316-727d2b75-ff90-43ee-acfb-b63dc4b615f2.png"/>
    <h3>AWS</h3>
    <p>Deploy manually onto AWS EC2</p>
  </button>
  <!-- AWS One Click-->
  <button class="app-btn btn" formaction="/next/aws-oneclick">
    <img alt="aws-logo" src="https://user-images.githubusercontent.com/1646931/187976316-727d2b75-ff90-43ee-acfb-b63dc4b615f2.png"/>
    <h3>AWS One-Click</h3>
    <p>Deploy onto AWS in one click</p>
  </button>
  <!-- Azure -->
  <button class="app-btn btn" formaction="/admin/deploy/kubernetes/azure">
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
  <!-- Kubernetes -->
  <button class="app-btn btn" formaction="/admin/deploy/kubernetes">
    <img alt="kubernetes-logo" src="https://user-images.githubusercontent.com/1646931/187978853-ee9efe0b-a18c-45a1-8375-c6c29647342a.png"/>
    <h3>Kubernetes</h3>
    <p><strong>Enterprise-only</strong>. Helm Chart and Kustomize Overlays supported</p>
  </button>
  <!-- Others -->
  <button class="app-btn btn" formaction="/admin/deploy">
    <img alt="private-cloud-logo" src="https://user-images.githubusercontent.com/1646931/187978634-6c4b2d06-2808-497d-8069-7adbee5bc703.png"/>
    <h3>Private cloud</h3>
    <p>Deploy into a generic cloud environment</p>
  </button>
</form>

## Local Machine

<form class="grid">
  <button class="app-btn btn" formaction="/admin/deploy/docker-single-container">
    <img alt="docker-logo" src="https://user-images.githubusercontent.com/1646931/187978472-1219f3a0-8c89-433c-8a72-223228952814.png"/>
    <h3>Docker Container</h3>
    <p>Spin up a local Sourcegraph instance</p>
  </button>
</form>
