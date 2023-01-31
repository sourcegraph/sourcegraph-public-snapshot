# Deploying Sourcegraph executors on Kubernetes

<aside class="beta">
<p>
<span class="badge badge-beta">Beta</span> This feature is in beta and might change in the future.
</p>

<p><b>We're very much looking for input and feedback on this feature.</b> You can either <a href="https://about.sourcegraph.com/contact">contact us directly</a>, <a href="https://github.com/sourcegraph/sourcegraph">file an issue</a>, or <a href="https://twitter.com/sourcegraph">tweet at us</a>.</p>
</aside>

[Kubernetes manifests](https://github.com/sourcegraph/deploy-sourcegraph) are provided to deploy Sourcegraph Executors on a running Kubernetes cluster. If you are deploying Sourcegraph with helm, charts are available [here](https://github.com/sourcegraph/deploy-sourcegraph-helm).

## Deployment

Executors on kubernetes machines require privileged access to a container runtime daemon in order to operate correctly. In order to ensure maximum capability across Kubernetes versions and container runtimes, a [Docker in Docker](https://www.docker.com/blog/docker-can-now-run-within-docker/) side car is deployed with each executor pod to avoid accessing the host container runtime directly.

