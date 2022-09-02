---
title: Deploy Sourcegraph
---

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  }
  .app-button {
    flex-basis: 0;
    flex-grow: 1;
    text-decoration: none;
    height: 15em;
    max-width: 100%;
    border-radius: 1em;
    background-color: white;
    border: 1px solid lightgray;
    margin: 0.5em;
    text-align: center;
    font-weight: 100;
  }
  .app-button:hover {
    box-shadow: 0 0 10px #00cbec;
  }
  .app-button > img {
    height: 4em;
  }
  .app-button > h1 {
    font-size: 1.5em;
    font-weight: 150;
    margin-top: .2em;
    margin-bottom: 1em;
  }
</style>

# Deploy Sourcegraph

Click on a deployment method for detail instructions.

## Recommended

<form class="grid">
  <button class="app-button" formaction="#TODO">
			<img src="https://handbook.sourcegraph.com/departments/engineering/design/brand_guidelines/logo/versions/Sourcegraph_Logomark_Color.svg"/>
			<h1>Sourcegraph Cloud</h1>
		  <p>Create a single-tenant instance managed by Sourcegraph</p>
  </button>
</form>

## Cloud Providers

<form class="grid">
  <!-- AWS -->
  <button class="app-button" formaction="/admin/deploy/docker-compose/aws">
    <img src="https://user-images.githubusercontent.com/1646931/187976316-727d2b75-ff90-43ee-acfb-b63dc4b615f2.png"/>
    <h1>AWS</h1>
    <p>Deploy manually onto AWS EC2</p>
  </button>
  <!-- AWS One Click-->
  <button class="app-button" formaction="/next/aws-oneclick">
    <img src="https://user-images.githubusercontent.com/1646931/187976316-727d2b75-ff90-43ee-acfb-b63dc4b615f2.png"/>
    <h1>AWS One-Click</h1>
    <p>Install with just one click on AWS</p>
  </button>
  <!-- Azure -->
  <button class="app-button" formaction="/admin/deploy/kubernetes/azure">
    <img src="https://user-images.githubusercontent.com/1646931/187978161-771cfb91-6cb3-4f00-befd-657502b95ed4.png"/>
    <h1>Azure</h1>
    <p>Deploy onto Microsoft Azure</p>
  </button>
  <!-- digital ocean -->
  <button class="app-button" formaction="/admin/deploy/docker-compose/digitalocean">
    <img src="https://res.cloudinary.com/crunchbase-production/image/upload/c_lpad,h_170,w_170,f_auto,b_white,q_auto:eco,dpr_1/v1478792253/gnlwek2zwhq369yryrzv.jpg"/>
    <h1>DigitalOcean</h1>
    <p>Deploy onto Digital Ocean</p>
  </button>
  <!-- Docker Compose -->
  <button class="app-button" formaction="/admin/install/docker-compose">
    <img src="https://user-images.githubusercontent.com/1646931/187978472-1219f3a0-8c89-433c-8a72-223228952814.png"/>
    <h1>Docker Compose</h1>
    <p>Deploy with Docker Compose</p>
  </button>
  <!-- GCP -->
  <button class="app-button" formaction="/admin/deploy/docker-compose/google_cloud">
    <img src="https://user-images.githubusercontent.com/1646931/187977350-3618e506-6fab-47c5-9a7c-286484cbd5a8.png"/>
    <h1>Google Cloud</h1>
    <p>Deploy onto Google Cloud (GCP)</p>
  </button>
  <!-- Kubernetes -->
  <button class="app-button" formaction="/admin/deploy/kubernetes">
    <img src="https://user-images.githubusercontent.com/1646931/187978853-ee9efe0b-a18c-45a1-8375-c6c29647342a.png"/>
    <h1>Kubernetes</h1>
    <p><strong>Enterprise-only</strong>. Helm Chart and Kustomize Overlays supported</p>
  </button>
  <!-- Others -->
  <button class="app-button" formaction="/admin/deploy">
    <img src="https://user-images.githubusercontent.com/1646931/187978634-6c4b2d06-2808-497d-8069-7adbee5bc703.png"/>
    <h1>Private cloud</h1>
    <p>Deploy into a generic cloud environment</p>
  </button>
</form>

## Local Machine

<form class="grid">
  <button class="app-button" formaction="/admin/deploy/docker-single-container">
    <img src="https://user-images.githubusercontent.com/1646931/187978472-1219f3a0-8c89-433c-8a72-223228952814.png"/>
    <h1>Docker Container</h1>
    <p>Spin up a local Sourcegraph instance</p>
  </button>
</form>
