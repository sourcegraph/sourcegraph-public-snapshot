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

| Deployment Type | Suggested for | Setup time | Resource isolation | Auto-healing | Multi-machine | Complexity |
| ---------------------------------------------------------------------- | ---------------------------------------------------------------- | ----------------- | :----------------: | :----------: | :-----------: | :--------: |
| [**★ Kubernetes with Helm **](../install/kubernetes/helm.md) | Production deployments of any size | 5 - 90 minutes | ✅ | ✅ | ✅ | 🟢 - 🔴 |
| [** Docker Compose**](../install/docker-compose/index.md) | Production deployments where Kubernetes with Helm is not viable | 5 - 30 minutes | ✅ | ✅ | ❌ | 🟢 - 🟠 |
| [** Kubernetes without Helm **](../install/kubernetes/index.md) | Production deployments of any size | 30 - 90 minutes | ✅ | ✅ | ✅ | 🟠 - 🔴 |
| [Single-container](../install/docker/index.md) | Local testing (Not recommended for production) | 1 minute | ❌ | ❌ | ❌ | 🟢 |

<span class="virtual-br"></span>

> NOTE: Setup times vary based on the level and complexity of customizations required.

> WARNING: Some features for self-hosted deployments [require a Sourcegraph license](https://about.sourcegraph.com/pricing/).

### Tips

* **We recommend Kubernetes with Helm for most production deployments**.
   *  Kubernetes provides resource isolation (from other services or applications), automated-healing, and far greater ability to scale.
   *  Helm provides a simple mechanism for deployment customizations, as well as a much simpler upgrade experience.
* If you are unable to use Helm to deploy, but still want to use Kubernetes, see the [Kubernetes guide](kubernetes/index.md). 
* Note that for a Kubernetes deployments, more advanced customizations and use of Kubernetes without Helm both make it more necessary to have existing Kubernetes expertise within your company. If in any doubt about your team's ability to support this, please either stick to use of Helm or speak to your Sourcegraph contact about using Docker Compose instead.

### Resource estimator

Use the [resource estimator](resource_estimator.md) to find a good starting point for your deployment.
