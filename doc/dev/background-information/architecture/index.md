# Sourcegraph architecture overview

This document provides a high level overview of Sourcegraph's architecture so you can understand how our systems fit together.

## Code syncing

At its core, Sourcegraph maintains a persistent cache of all the code that is connected to it. It is persistent, because this data is critical for Sourcegraph to function, but it is ultimately a cache because the code host is the source of truth and our cache is eventually consistent.

- [gitserver](../../../../cmd/gitserver/README.md) is the sharded service that stores the code and makes it accessible to other Sourcegraph services.
- [repo-updater](../../../../cmd/repo-updater/README.md) is the singleton service that is responsible for ensuring all the code in gitserver is as up-to-date as possible while respecting code host rate limits. It is also responsible for syncing code repository metadata from the code host that is stored in the `repo` table of our Postgres database.

If you want to learn more about how code is synchronized, read [Life of a repository](life-of-a-repository.md).

## Search

Devs can search across all the code that is connected to their Sourcegraph instance.

By default, Sourcegraph uses [zoekt](https://github.com/sourcegraph/zoekt) to create a trigram index of the default branch of every repository so that searches are fast. This trigram index is the reason why Sourcegraph search is more powerful and faster than what is usually provided by code hosts. 

- [zoekt-indexserver](https://sourcegraph.com/github.com/sourcegraph/zoekt/-/tree/cmd/zoekt-sourcegraph-indexserver) 
- [zoekt-webserver](https://sourcegraph.com/github.com/sourcegraph/zoekt/-/tree/cmd/zoekt-webserver)

Sourcegraph also has a fast search path for code that isn't indexed yet, or for code that will never be indexed (for example: code that is not on a default branch). Indexing every branch of every repository isn't a pragmatic use of resources for most customers, so this decision balances optimizing the common case (searching all default branches) with space savings (not indexing everything).

- [searcher](../../../../cmd/searcher/README.md) implements the non-indexed search.

Syntax highlighting for any code view, including search results, is provided by [Syntect server](https://sourcegraph.com/github.com/sourcegraph/syntect_server).

If you want to learn more about search:

- [Code search product documentation](../../../code_search/index.md)
- [Life of a search query](life-of-a-search-query.md)
- [Search pagination](search-pagination.md)

## Code intelligence

Code intelligence surfaces data (for example: doc comments for a symbol) and actions (for example: go to definition, find references) based on our semantic understanding of code (unlike search, which is completely text based).

By default, Sourcegraph provides imprecise [search-based code intelligence](../../../code_intelligence/explanations/search_based_code_intelligence.md). This reuses all the architecture that makes search fast, but it can result in false positives (for example: finding two definitions for a symbol, or references that aren't actually references), or false negatives (for example: not able to find the definition or all references). This is the default because it works with no extra configuration and is pretty good for many use cases and languages. We support a lot of languages this way because it only requires writing a few regular expressions.

With some setup, customer can enable [precise code intelligence](../../../code_intelligence/explanations/precise_code_intelligence.md). Repositories add a step to their build pipeline that computes the index for that revision of code and uploads it to Sourcegraph. We have to write language specific indexers, so adding precise code intel support for new languages is a non-trivial task. 

If you want to learn more about code intelligence:

- [Code intelligence product documentation](../../../code_intelligence/index.md)
- [Code intelligence developer documentation](../codeintel/index.md)
- [Available indexers](../../../code_intelligence/references/indexers.md)

## Campaigns

Campaigns create and manage large scale code changes across projects, repositories, and code hosts.

To create a campaign, users write a [campaign spec](../../../campaigns/references/campaign_spec_yaml_reference.md), which is a YAML file that specifies the changes that should be performed, and the repositories that they should be performed upon — either through a Sourcegraph search, or by declaring them directly. This spec is then executed by [src-cli](#src-cli) on the user's machine (or in CI, or some other environment controlled by the user), which results in [changeset specs](../../../campaigns/explanations/introduction_to_campaigns.md#changeset-spec) that are sent to Sourcegraph. These changeset specs are then applied by Sourcegraph to create one or more changesets per repository. (A changeset is a pull request or merge request, depending on the code host.)

Once created, changesets are monitored by Sourcegraph, and their current review and CI status can be viewed on the campaign's page, providing a single pane of glass view of all the changesets created as part of the campaign. The campaign can be updated at any time by re-applying the original campaign spec: this will transparently add or remove changesets in repositories that now match or don't match the original search as needed.

If you want to learn more about campaigns:

- [Campaign product documentation](../../../campaigns/index.md)
- [Campaign design principles](../../../campaigns/explanations/campaigns_design.md)
- [Campaign developer documentation](../campaigns/index.md)
- [How `src` executes a campaign spec](../../../campaigns/explanations/how_src_executes_a_campaign_spec.md)

## Code insights

TODO

## Code monitoring

Code monitoring allows users to get notified of changes to their codebase.

Users can view, edit and create code monitors through the code monitoring UI (`/code-monitoring`). A code monitor comprises a **trigger**, and one or more **actions**.

The **trigger** watches for new data and if there is new data we call this an event. For now, the only supported trigger is a search query of `type:diff` or `type:commit`, run every five minutes by the Go backend with an automatically added `after:` parameter narrowing down the diffs/commits that should be searched. The monitor's configured actions are run when this query returns a non-zero number of results.

The **actions** are run in response to a trigger event. For now, the only supported action is an email notification to the primary email address of the code monitor's owner. In order for this to work, [`email.address` and `email.smtp` must be configured in site configuration](https://docs.sourcegraph.com/admin/observability/alerting#email). Code monitoring actions will be extended in the future to support webhooks.

If you want to learn more about code monitoring:
- [Code monitoring documentation](https://docs.sourcegraph.com/code_monitoring)

## Browser extensions

TODO

## Sourcegraph extension API

TODO

If you want to learn more about our extension API:

- [Sourcegraph extension architecture](sourcegraph-extensions.md)

## src-cli

src-cli, or `src`, is a command line tool that users can run locally to interact with Sourcegraph.

src-cli is written in Go, and distributed as a standalone binary for Windows, macOS, and Linux. Its features include [running searches](../../../cli/references/search.md), managing Sourcegraph, and [executing campaigns](../../../campaigns/quickstart.md#create-the-campaign). src-cli is an integral part of the [campaigns product](#campaigns).

Note that src-cli is not contained within the Sourcegraph monorepo, and has its own release cadence.

If you want to learn more about src-cli:

- [src-cli repository](https://github.com/sourcegraph/src-cli)
- [src-cli documentation](../../../cli/index.md)

## Editor extensions

TODO

## Deployment

Sourcegraph is deployable via three supported methods:

- [pure-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker) for small environments on a single server. Easiest and quickest to setup with a series of bash scripts. Little infrastructure knowledge is required. 
- [docker-compose](https://github.com/sourcegraph/deploy-sourcegraph-docker/docker-compose) is intended to be used for small to medium production deployments, with some customization available. Easy to setup with basic infrastructure and docker knowledge required. An example of a compose deployment.
- [kubernetes](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/master/docker-compose) is intended for all medium to large scale production deployments that require fault tolerance and high availibility. For advanced users only with significant kubernetes experience required.

The [resource estimator](https://docs.sourcegraph.com/admin/install/resource_estimator) can guide you on the requirements for each deployment type. 

If you want to learn more about deploying Sourcegraph:
- [docker-compose](https://docs.sourcegraph.com/admin/install/docker-compose)
- [Kubernetes](https://docs.sourcegraph.com/admin/install/kubernetes)

### Internal examples

The following are working internal deployments that can be used as a reference for various deployment types:

#### Docker-compose

 - [demo.sourcegraph.com](https://demo.sourcegraph.com)
 - [devmanaged.sourcegraph.com](https://devmanaged.sourcegraph.com)

 Both of these are [managed instances](https://about.sourcegraph.com/handbook/engineering/distribution/managed) and are the responsibility of the Distribution team. 

 #### Kubernetes

 - [sourcegraph.com](https://sourcegraph.com) Continuously updated from [this repo](https://github.com/sourcegraph/deploy-sourcegraph-dot-com)
 - [k8s.sgdev.org](https://k8s.sgdev.org) Continously updated from [this repo](https://github.com/sourcegraph/deploy-sourcegraph-dogfood-k8s-2)

 These are two examples of internal Kubernetes deployments.

 If you want to learn more about Internal Deployments:
 - [Internal deployments of Sourcegraph](https://about.sourcegraph.com/handbook/engineering/deployments)




## Observability

Observability encapsulates the monitoring and debugging of Sourcegraph deployments.
Sourcegraph is designed, and ships with, a number of observability tools and capabilities out-of-the box to enable visibility into the health and state of a Sourcegraph deployment.

Monitoring includes [metrics and dashboards](../../../admin/observability/metrics.md), [alerting](../../../admin/observability/alerting.md), and [health checking](../../../admin/observability/health_checks.md) capabilities.
Learn more about monitoring in the [monitoring architecture overview](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture).

- [grafana](../observability/grafana.md) is the frontend for service metrics, and ships with customized dashboards for Sourcegraph services.
- [prometheus](../observability/prometheus.md) handles scraping of service metrics, and ships with recording rules, alert rules, and alerting capabilities.
- [Health checks are provided by each Sourcegraph service](../../../admin/observability/health_checks.md).

Debugging includes [tracing](../../../admin/observability/tracing.md) and [logging](../../../admin/observability/logs.md).

- [jaeger](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/docker-images/jaeger-all-in-one) is the distributed tracing service used by Sourcegraph.
- [jaeger-agent](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/docker-images/jaeger-agent) is a sidecar used in Kubernetes deployments for collecting traces from services.
- [Logs are provided by each Sourcegraph service](../../../admin/observability/logs.md).

If you want to learn more about observability:

- [Observability for site administrators](../../../admin/observability/index.md)
- [Observability developer documentation](../observability/index.md)
- [Observability at Sourcegraph](https://about.sourcegraph.com/handbook/engineering/observability)

<!-- content below here has not been reorganized or refreshed yet -->

## Diagram

You can click on each component to jump to its respective code repository or subtree.

<object data="/dev/background-information/architecture/architecture.svg" type="image/svg+xml" style="width:100%; height: 100%">
</object>

Note that almost every service has a link back to the frontend, from which is gathers configuration updates.
These edges are omitted for clarity.

## Clients

We maintain multiple Sourcegraph clients:

- [Web application](https://github.com/sourcegraph/sourcegraph/tree/main/client/web)
- [Browser extensions](https://github.com/sourcegraph/sourcegraph/tree/main/client/browser)
- [src-cli](https://github.com/sourcegraph/src-cli)
- [Editor integrations](https://docs.sourcegraph.com/integration/editor)
  - [Visual Studio Code](https://github.com/sourcegraph/sourcegraph-vscode)
  - [Atom](https://github.com/sourcegraph/sourcegraph-atom)
  - [JetBrains IDEs](https://github.com/sourcegraph/sourcegraph-jetbrains)
  - [Sublime Text 3](https://github.com/sourcegraph/sourcegraph-sublime)

These clients generally communicate with a Sourcegraph instance (either https://sourcegraph.com or a private customer instance) through our [GraphQL API](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/cmd/frontend/graphqlbackend/schema.graphql). There are also a small number of REST endpoints for specific use-cases.

## Services

Our backend is composed of multiple services:

- Most are Go services found in the [cmd](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/cmd) and [enterprise/cmd](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/enterprise/cmd) folders.

## Infrastructure

- [sourcegraph/infrastructure](https://github.com/sourcegraph/infrastructure) contains Terraform configurations for Cloudflare DNS and Site 24x7 monitoring, as well as build steps for various Docker images. Only private Docker images should be added here, public ones belong in the main repository.
- [sourcegraph/deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) contains YAML that can be used by customers to deploy Sourcegraph to a Kubernetes cluster.
- [sourcegraph/deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker) contains a pure-Docker cluster deployment reference that some one-off customers use to deploy Sourcegraph to a non-Kubernetes cluster.
  - [sourcegraph/deploy-sourcegraph-dot-com](https://github.com/sourcegraph/deploy-sourcegraph-dot-com) is a fork of the above that is used to deploy to the Kubernetes cluster that serves https://sourcegraph.com.

## References

Here are some references to help you understand how multiple systems fit together:

- [Life of a ping](life-of-a-ping.md)
