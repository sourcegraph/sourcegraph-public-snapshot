# Install Sourcegraph

<p class="lead">
Sourcegraph can be installed in a variety of methods to set up a deployment for your private code.
</p>

If you're just starting out, you can [**try Sourcegraph Cloud**](https://sourcegraph.com) or [run Sourcegraph locally](docker/index.md).

<div class="cta-group">
<a class="btn btn-primary" href="#self-hosted">★ Self-hosted instance</a>
<a class="btn" href="managed">Managed instance</a>
<a class="btn" href="../../#get-help">Get help</a>
</div>

## Self-hosted

| Deployment Type                                       | Suggested for                                       | Setup time    | Multi-machine | Auto healing | Monitoring |
| ----------------------------------------------------- | --------------------------------------------------- | ------------- | -------------- | ------------- | ----------- |
| [**★ Docker Compose**](../install/docker-compose/index.md)  | **Small & medium** production deployments               | 🟢 5 minutes     | ❌             | ❌            | ✅         |
| [**★ Kubernetes**](../install/kubernetes/index.md)          | **Medium & large** highly-available cluster deployments | 🟠 30-90 minutes | ✅            | ✅           | ✅         |
| [Single-container server](../install/docker/index.md) | Local testing                                       | 🟢 1 minute    | ❌             | ❌            | ❌          |

<span class="virtual-br"></span>

> NOTE: Some features for self-hosted deployments [require a Sourcegraph license](https://about.sourcegraph.com/pricing/).

### Tips

* **We recommend Docker Compose for most initial production deployments**. You can [migrate to a different deployment method](../updates/index.md#migrating-to-a-new-deployment-type) later on if needed.
* Note that **for a Kubernetes deployment, you are expected to have a team that is familiar with operating Kubernetes clusters**, including but not limited to the use of persistent storage. If in any doubt about your team's ability to support this, please speak to your Sourcegraph contact about using Docker Compose instead.
* Don't want to worry about managing a Sourcegraph deployment? Consider a [managed instance](./managed.md).

### Resource estimator

Use the [resource estimator](resource_estimator.md) to find a good starting point for your deployment.
