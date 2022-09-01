---
title: Deploy Sourcegraph
---

<style>
a.app-button-wrapper {
	text-decoration: none;
	color: inherit;
}

.app-button {
	height: 15em;
	width: 15em;
	border-radius: 1em;
	background-color: white;
	border: 1px solid lightgray;
	margin: 0 1em 1em 0;
	padding: 1.5em;
}

.app-button:hover {
	border: 1px solid orange;
	box-shadow: 0 0 10px purple;
}

.app-button.app-button-extended {
	width: 20em;
}

.app-button .body img {
	height: 4em;
}

.app-button .header {
}

.app-button .header h1 {
	font-size: 1.5em;
	font-weight: 100;
	margin: 0 0 0.25em 0;
}
</style>

# Deploy Sourcegraph

## Cloud

<a href="" class="app-button-wrapper">
	<div class="app-button app-button-extended">
		<div class="body">
			<img src="https://handbook.sourcegraph.com/departments/engineering/design/brand_guidelines/logo/versions/Sourcegraph_Logomark_Color.svg"/>
		</div>
		<div class="header">
			<h1>Sourcegraph Cloud</h1>
		</div>
		<div class="detail">
			Create a single-tenant instance managed by Sourcegraph<br/>
			<strong>(Recommended)</strong>
		</div>
	</div>
</a>

## Self-hosted

<div style="display: flex; flex-direction: row; flex-wrap: wrap;">

<a href="/next/aws-oneclick" class="app-button-wrapper">
<div class="app-button">
	<div class="body">
		<img src="https://user-images.githubusercontent.com/1646931/187976316-727d2b75-ff90-43ee-acfb-b63dc4b615f2.png"/>
	</div>
	<div class="header">
		<h1>AWS one-click</h1>
	</div>
	<div class="detail">
		One-click AWS instance
	</div>
</div>
</a>

<a href="/admin/deploy/docker-compose/aws" class="app-button-wrapper">
<div class="app-button">
	<div class="body">
		<img src="https://user-images.githubusercontent.com/1646931/187976316-727d2b75-ff90-43ee-acfb-b63dc4b615f2.png"/>
	</div>
	<div class="header">
		<h1>AWS manual install</h1>
	</div>
	<div class="detail">
		Deploy manually onto AWS EC2
	</div>
</div>
</a>

<a href="/admin/deploy/docker-compose/google_cloud" class="app-button-wrapper">
<div class="app-button">
	<div class="body">
		<img src="https://user-images.githubusercontent.com/1646931/187977350-3618e506-6fab-47c5-9a7c-286484cbd5a8.png"/>
	</div>
	<div class="header">
		<h1>Google Cloud</h1>
	</div>
	<div class="detail">
		Deploy onto GCP
	</div>
</div>
</a>

<a href="/admin/deploy/docker-compose/digitalocean" class="app-button-wrapper">
<div class="app-button">
	<div class="body">
		<img src="https://res.cloudinary.com/crunchbase-production/image/upload/c_lpad,h_170,w_170,f_auto,b_white,q_auto:eco,dpr_1/v1478792253/gnlwek2zwhq369yryrzv.jpg"/>
	</div>
	<div class="header">
		<h1>DigitalOcean</h1>
	</div>
	<div class="detail">
		Deploy onto DigitalOcean
	</div>
</div>
</a>

<a href="/admin/deploy/docker-compose" class="app-button-wrapper">
<div class="app-button">
	<div class="body">
		<img src="https://user-images.githubusercontent.com/1646931/187978161-771cfb91-6cb3-4f00-befd-657502b95ed4.png"/>
	</div>
	<div class="header">
		<h1>Azure</h1>
	</div>
	<div class="detail">
		Deploy onto Microsoft Azure
	</div>
</div>
</a>

<a href="/admin/install/docker-compose" class="app-button-wrapper">
<div class="app-button">
	<div class="body">
		<img src="https://user-images.githubusercontent.com/1646931/187978472-1219f3a0-8c89-433c-8a72-223228952814.png"/>
	</div>
	<div class="header">
		<h1>Docker Compose</h1>
	</div>
	<div class="detail">
		Deploy with Docker Compose
	</div>
</div>
</a>

<a href="/admin/install/docker-compose" class="app-button-wrapper">
<div class="app-button">
	<div class="body">
		<img src="https://user-images.githubusercontent.com/1646931/187978634-6c4b2d06-2808-497d-8069-7adbee5bc703.png"/>
	</div>
	<div class="header">
		<h1>Private cloud</h1>
	</div>
	<div class="detail">
		Deploy into a generic cloud environment
	</div>
</div>
</a>

<a href="/admin/deploy/kubernetes" class="app-button-wrapper">
<div class="app-button">
	<div class="body">
		<img src="https://user-images.githubusercontent.com/1646931/187978853-ee9efe0b-a18c-45a1-8375-c6c29647342a.png"/>
	</div>
	<div class="header">
		<h1>Kubernetes</h1>
	</div>
	<div class="detail">
		Deploy with Kubernetes<br/>
		<strong>(Enterprise-only)</strong>
	</div>
</div>
</a>

</div>

## Local machine

<a href="/admin/deploy/docker-single-container" class="app-button-wrapper">
<div class="app-button">
	<div class="body">
		<img src="https://user-images.githubusercontent.com/1646931/187978472-1219f3a0-8c89-433c-8a72-223228952814.png"/>
	</div>
	<div class="header">
		<h1>Docker container</h1>
	</div>
	<div class="detail">
		Spin up Sourcegraph on your local machine.
	</div>
</div>
</a>
