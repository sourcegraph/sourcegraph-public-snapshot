# Choosing the right deployment model for Sourcegraph

You don't have to be a Sys Admin or DevOps Engineer to run Sourcegraph. You just need access to a machine that has Docker installed in order to get started.

It's free to deploy for up to 10 monthly users (great for trialing with your team) and you unlock access to all enterprise features by [getting a trial enterprise license key](https://about.sourcegraph.com/contact/request-demo/?form_submission_source=guides).

There are three ways of deploying Sourcegraph:

- **[Single Docker container](../admin/install/docker.md)**<br/>
Single container on a single host. Great for trying Sourcegraph locally or deploying for your team or company if only indexing a few hundred repositories.

- **[Docker Compose](../admin/install/docker-compose.md)**<br />
Individual Sourcegraph services on a single host. Best fit for companies with many hundreds or a thousand plus repositories and dozens of monthly users. Enables per service resource allocation and service replicas to tune and meet load requirements. Perfect for demanding workloads with a relatively simple scaling model.

- **[Kubernetes cluster](../admin/install/cluster.md)**<br/>
Handles the most demanding of deployments from thousands to tens of thousands of repositories. In most instances, it's preferable to start with Docker Compose on a single powerful instance, then deploy to a Kubernetes cluster as part of an official enterprise trial.

> NOTE: If you're still unsure how to best deploy Sourcegraph, [request a demo](https://about.sourcegraph.com/contact/request-demo/?form_submission_source=guides&utm_source=guides) to discuss your requirements with our deployment engineering team.

## Using the Sourcegraph resource estimator

The [Sourcegraph resource estimator](..//admin/install/resource_estimator.md) is a configurable form that enables you to dial-in the specifics of your environment in order get prcise deployment recommendations.

---

[**» Next: Installing Sourcegraph**](installing_sourcegraph.md)
