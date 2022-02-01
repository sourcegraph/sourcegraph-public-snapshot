# Site Administration Quickstart 
Administrating and managing a Sourcegraph instance is handled by Site Admins. These admins are typically responsible for deploying, managing, and configuring Sourcegraph for users on their instance. Site Admins have [elevated permissions](https://docs.sourcegraph.com/admin/privileges) within their Sourcegraph instance. 

This guide will walk you through the features and functionalities available to you as a Site Admin and show how you can get started managing and maintaining your Sourcegraph instance. For detailed, in-depth information, please reference the [administration guides and docs](https://docs.sourcegraph.com/admin "Administration guides and documentation").

## Getting Started:

### What is the best deployment option for me?
We recommend Docker Compose for most initial production deployments. You can [migrate to a different deployment method](https://docs.sourcegraph.com/admin/updates#migrating-to-a-new-deployment-type) later on if needed.

If you need a deployment option that offers a higher level of scalability and availability, the [Kubernetes deployment](https://docs.sourcegraph.com/admin/install/kubernetes) is recommended. 

To help give you a starting point on choosing a deployment option and allocating resources to it, check out our [resource estimator](https://docs.sourcegraph.com/admin/install/resource_estimator).

For a comphrensive deployment guide for each option, check out our in-depth documentation for both [Docker Compose](https://docs.sourcegraph.com/admin/install/docker-compose) and [Kubernetes](https://docs.sourcegraph.com/admin/install/kubernetes).

### Deployment options 
| Deployment Type                                             | Suggested for                                           | Setup time        | Resource isolation | Auto-healing | Multi-machine |
| ----------------------------------------------------------- | ------------------------------------------------------- | ----------------- | :----------------: | :----------: | :-----------: |
| [**★ Docker Compose**](https://docs.sourcegraph.com/admin/install/docker-compose) | **Small & medium** production deployments               | 🟢 5 minutes     |         ✅         |      ✅      |      ❌       |
| [**★ Kubernetes**](https://docs.sourcegraph.com/admin/install/kubernetes)         | **Medium & large** highly-available cluster deployments | 🟠 30-90 minutes |         ✅         |      ✅      |      ✅       |
| [Single-container](https://docs.sourcegraph.com/admin/install/docker)              | Local testing                                           | 🟢 1 minute      |         ❌         |      ❌      |      ❌       |

> **NOTE: Some features for production deployments [require a Sourcegraph license](https://about.sourcegraph.com/pricing/)**.

### Self-hosted vs. Managed instances
Regardless of the deployment option you choose, Sourcegraph can be self-hosted locally or with the cloud provider of your choice. We also offer [managed instances](https://docs.sourcegraph.com/admin/install/managed) (we handle deployment, updates, and management of the instance for you). Please [contact us](https://about.sourcegraph.com/contact/sales) if you are interested in learning more about managed instances. 


## Updating your instance 
New versions of Sourcegraph are released monthly (with patches released in between, as needed). New updates are announced in the [Sourcegraph blog](https://about.sourcegraph.com/blog), and comprehensive update notes are available in the [changelog](https://docs.sourcegraph.com/CHANGELOG). 

Regardless of the deployment type you choose, the following update rules apply: 
- **Update one minor version at a time**, e.g., v3.26 –> v3.27 –> v3.28.
    - Patches (e.g., vX.X.4 vs. vX.X.5) do not have to be adopted when moving between vX.X versions.
- **Check the [update notes](https://docs.sourcegraph.com/admin/updates#update-notes) for your deployment type for any required manual actions** before updating.
- Check [out of band migration status](https://docs.sourcegraph.com/admin/migration) before updating to avoid a necessary rollback while the migration finishes.

To check the current version of your instance, go to **User menu > Site admin > Updates**.

For more details on updating your instance, please refer to the [update docs](https://docs.sourcegraph.com/admin/updates).

## Configuration
As a Site Admin, you have the ability to control and configure the various aspects of your instance. Including the code host connection(s), SSO, repository indexing, and the functionality of the instance itself. 

### Site configuration 
At the heart of managing your Sourcegraph instance is Site configuration. Site config is a JSON file that defines how the various features and functionality within Sourcegraph are set up and configured. 

To access site config, go to **User menu > Site admin > Site configuration**.

### Connecting to code hosts
Sourcegraph supports connections to and repository syncing from any Git code host. Once connected, Sourcegraph will clone and index your repos so that users can search and navigate across them. To get started, go to **User menu > Site Admin > Manage code hosts > Add code host**.

Please reference the code host [documentation](https://docs.sourcegraph.com/admin/external_service "documentation") for detailed instructions on connecting to your code host.

### Setting up user authentication and repository permissions
In addition to connecting to your repositories, you can also configure Sourcegraph to use your preferred SSO or sign-in method and inherit and enforce repository permissions for users.

#### Auth & SSO:
Sourcegraph supports several different authentication methods: OAuth (for GitHub or GitLab), OpenID Connect (Google Workspace), and SAML; and also provides a built-in authentication method via email, if needed. 

To get started setting up user authentication and SSO, please reference our [auth documentation](https://docs.sourcegraph.com/admin/auth "auth documentation").

#### Repository Permissions:
In addition to configuring user authentication to Sourcegraph, you may also want to ensure that users can only view repositories that they would have access to on your code host. Sourcegraph supports the ability to inherit and enforce these repository permissions on a per-user basis and can be configured for connections to GitHub, GitLab, and Bitbucket Server.

For more info, check out our complete [repository permission documentation.](https://docs.sourcegraph.com/admin/repo/permissions#repository-permissions "repository permission documentation.")


### External services 
By default, Sourcegraph bundles the services it needs to operate into installations. These services include PostgreSQL, Redis, and MinIO. 

Your Sourcegraph instance can also be configured to use existing external services if you wish. For more information on configuring Sourcegraph to use your external services, please reference this [documentation.](https://docs.sourcegraph.com/admin/external_services)

## Observability 
One key component to managing a Sourcegraph instance is having the ability to observe, monitor, and analyze the health of your instance. Sourcegraph ships with [Grafana](https://grafana.com/) for dashboards; [Prometheus](https://prometheus.io/) for metrics and alerting; as well as a [built-in alerting system.](https://docs.sourcegraph.com/admin/observability/alerting)

### Viewing instance health and metrics
Alerts and metrics can be viewed and monitored in Grafana. To access the Grafana dashboard bundled with your Sourcegraph instance, go to **User menu > Site admin > Monitoring**.

We also have an exhaustive [reference guide](https://docs.sourcegraph.com/admin/observability/dashboards) for understanding the available dashboards, and an [alert solutions guide](https://docs.sourcegraph.com/admin/observability/alert_solutions) with descriptions and possible solutions for each alert that fires in Grafana. 

