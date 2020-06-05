# Installing Sourcegraph on a cluster

| Deployment Type                                       | Suggested for                                       | Setup time | Multi-machine? | Auto healing? | Monitoring? |
|-------------------------------------------------------|-----------------------------------------------------|------------|----------------|---------------|-------------|
| [Single-container server](../install/docker/index.md) | Local testing                                       | 60 seconds | Impossible     | No            | No          |
| [Docker Compose](../install/docker-compose/index.md)  | Small & medium production deployments               | 5 minutes  | Possible       | No            | Yes         |
| [Kubernetes](../install/kubernetes/index.md)          | Medium & large highly-available cluster deployments | 30 minutes | Easily         | Yes           | Yes         |

For cluster deployments, we recommend installing Sourcegraph on Kubernetes. See the [Kubernetes installation guide](../install/kubernetes/index.md) for more information.

If you cannot use Kubernetes or prefer using your own container infrastructure, check out our [pure-Docker deployment reference](https://github.com/sourcegraph/deploy-sourcegraph-docker).

## Resource estimator

Use the [resource estimator](resource_estimator.md) to find a good starting point for your deployment.
