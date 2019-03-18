# Sourcegraph documentation

Sourcegraph is a code search and browsing tool with code intelligence that helps developers write and review code. Learn more about Sourcegraph at [about.sourcegraph.com](https://about.sourcegraph.com).

Sourcegraph development is open source at [github.com/sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph). If these docs don't solve your problem, check the [issue tracker](https://github.com/sourcegraph/sourcegraph/issues).

## Quickstart

Set up a private Sourcegraph instance for your team in seconds with the command below. Visit the [site administrator documentation](admin/index.md) to learn more about installation options, or start [adding repositories](admin/repo/add.md) and searching!

**Prerequisites:** [Docker](https://docs.docker.com/engine/installation/) is required.

```
docker run \
  --publish 7080:7080 --rm \
  --volume ~/.sourcegraph/config:/etc/sourcegraph \
  --volume ~/.sourcegraph/data:/var/opt/sourcegraph \
  --volume /var/run/docker.sock:/var/run/docker.sock \
  sourcegraph/server:2.12.2
```

When Sourcegraph is ready, continue at http://localhost:7080.

### Next steps

- [Add repositories](admin/repo/add.md)
- [Configure your Sourcegraph instance](admin/site_config/index.md)
- [Configure code intelligence](extensions/language_servers/index.md)
- [Deploy Sourcegraph on AWS](admin/install/docker/aws.md)
- [Deploy Sourcegraph on Google Cloud Platform](admin/install/docker/google_cloud.md)
- [Deploy Sourcegraph on Digital Ocean](admin/install/docker/digitalocean.md)

## For users

The [user documentation](user/index.md) is about how to use Sourcegraph. The most read docs are:

- [Overview](user/index.md): What is Sourcegraph?
- [Tour](user/tour.md): A walkthrough of Sourcegraph's features, with real-world example use cases.
- [Code search](user/search/index.md)
- [Integrations](integration/index.md)
  - [Browser extension](integration/browser_extension.md)
- [GraphQL API](api/graphql.md)

## For site administrators

The [site administration documentation](admin/index.md) is about deploying and managing a Sourcegraph self-hosted instance.

## For contributors to Sourcegraph

See the [contributor documentation](dev/index.md) and the [main Sourcegraph repository](https://github.com/sourcegraph/sourcegraph) (open-source).

## Sourcegraph roadmap

The [Sourcegraph roadmap](dev/roadmap.md) describes what's coming next.

<!-- TODO(sqs): Add link to ./graphbook when it has more content. -->

## Sourcegraph subscriptions

You can use Sourcegraph in 2 ways:

- Self-hosted Sourcegraph: Deploy and manage your own Sourcegraph instance.
- [Sourcegraph.com](https://sourcegraph.com): For public code only. No signup or installation required.

For self-hosted Sourcegraph instances, you run a Docker image or Kubernetes cluster on-premises or on your preferred cloud provider. There are [3 tiers](https://about.sourcegraph.com/pricing): Core (free), Enterprise Starter, and Enterprise. Enterprise features require a [Sourcegraph subscription](https://sourcegraph.com/user/subscriptions).

## Other links

- [Sourcegraph open-source repository](https://github.com/sourcegraph/sourcegraph)
- [Issue tracker](https://github.com/sourcegraph/sourcegraph/issues)
- [about.sourcegraph.com](https://about.sourcegraph.com) (general information about Sourcegraph)
- [@srcgraph on Twitter](https://twitter.com/srcgraph)
