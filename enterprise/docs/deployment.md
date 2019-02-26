# Deployment

This document describes how Sourcegraph is deployed and released.

## Deployment types

There are two ways customers deploy Sourcegraph:

- **Single Docker container (from the `sourcegraph/server` Docker image)** ([code](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/cmd/server) | [docs](https://docs.sourcegraph.com/#quickstart)): A Docker image that can be
  run on a single node with a simple command that is documented on our home page. It is the free and
  easy way for a customer to start using Sourcegraph.

- **Kubernetes cluster** ([code](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph) |
  [docs](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph/-/blob/README.md)):
  This deployment is a paid upgrade and allows customers to deploy Sourcegraph onto a
  Kubernetes cluster.

## Deployment locations

Sourcegraph is deployed to multiple locations:

1.  **Customers**: When customers want Sourcegraph to work on their private code, they deploy it to their own infrastructure using our [public documentation](https://docs.sourcegraph.com/#quickstart).
1.  **Dogfood** ([sourcegraph.sgdev.org](https://sourcegraph.sgdev.org)): A Kubernetes cluster that
    runs an instance of Sourcegraph for all of our private code. It is identical to what
    customers run, except that it includes a few extra services, like a Phabricator instance for
    testing purposes.
1.  **Production** ([sourcegraph.com](https://sourcegraph.com)): Production is a public demonstration
    of Sourcegraph for all public code on GitHub. We take shortcuts to make it work at that scale
    (tens of millions of repos). Our primary focus is making Sourcegraph work at customer scale (tens
    of thousands of repos). Production uses Kubernetes (but is on an old deployment scheme because tech debt).

## Releases

### Dogfood and production

Commits to the master branch of github.com/sourcegraph/sourcegraph are
continuously deployed to all of our services in production and in dogfood.

If you need to update more than just the Docker images (i.e., if you need to update the
Kubernetes configuration), refer to the appropriate link below:

- [Production](https://github.com/sourcegraph/infrastructure/blob/master/kubernetes/README.prod.md)
- [Dogfood](https://github.com/sourcegraph/infrastructure/blob/master/datacenter/README.md#updating-a-live-cluster-including-dogfood)

### To our customers

We ship to our customers minor feature releases monthly (e.g. 2.7, 2.8, 2.9), and patch releases on an as-needed basis (e.g. 2.8.1).

- [Docker image release process](https://sourcegraph.sgdev.org/github.com/sourcegraph/sourcegraph/-/blob/cmd/server/README.md)
  - Before a Server release is published ot customers, it should first be tested via `docker run ...` on your machine.
- [Kubernetes cluster release process](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph/-/blob/README.dev.md)
  - Before a Kubernetes release is published to customers, it should first be deployed to and tested on dogfood.

Important notes:

- The Docker image and Kubernetes cluster release processes are completely independent. Either one can be done
  without the other.
- Versioning: the major and minor version of the Docker image and Kubernetes cluster are updated on the same monthly
  release schedule, **but the patch version of one has no relation to that of the other.** That is
  to say, the patch versions are completely independent.
