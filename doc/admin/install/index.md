# Installing Sourcegraph

You can opt to use Sourcegraph as a [self-hosted](#self-hosted) deployment or [managed instance](#managed-instances).
## Managed instances

The Sourcegraph team can host and manage a Sourcegraph instance for you. This makes them a simple choice for customers that want to utilize Sourcegraph but do not wish to manage its deployment and maintenance. You can find more details in its [installation page](managed.md).

## Self-hosted

| Deployment Type                                       | Suggested for                                       | Setup time | Multi-machine? | Auto healing? | Monitoring? |
|-------------------------------------------------------|-----------------------------------------------------|------------|----------------|---------------|-------------|
| [Single-container server](../install/docker/index.md) | Local testing                                       | 60 seconds | No             | No            | No          |
| [Docker Compose](../install/docker-compose/index.md)  | Small & medium production deployments               | 5 minutes  | Not Supported  | No            | Yes         |
| [Kubernetes](../install/kubernetes/index.md)          | Medium & large highly-available cluster deployments | 30 minutes | Yes            | Yes           | Yes         |


* If you're just starting out, we recommend [running Sourcegraph locally](docker/index.md). It takes only a few minutes and lets you try out all of the features.
* If you need scalability and high-availability beyond what a single-node [Docker Compose](https://docs.docker.com/compose/) can offer, use the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph), instead.

### Resource estimator

Use the [resource estimator](resource_estimator.md) to find a good starting point for your deployment.
