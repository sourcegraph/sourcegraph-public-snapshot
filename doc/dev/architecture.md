# Sourcegraph Architecture Overview

This is a high level overview of our architecture at Sourcegraph so you can understand how our services fit together.

![Sourcegraph architecture](img/architecture.svg)

**Note**: Non-sighted users can view a [text-representation of this diagram](architecture-mermaid.md).

<!--
Updating the architecture image

TODO: Automate this or replace mermaidjs diagrams

TLDR: Get @ryan-blunden to render a new svg after making changes to architecture.mermaid.

After changing architecture.mermaid, render the new diagram at https://mermaidjs.github.io/mermaid-live-editor/, set "theme" to be "neutral" in the config textarea, then download and replace img/architecture.svg. But there's one more step.

if you try rendering the downloaded SVG as is, the text is cut off in most boxes. This is because the  downloaded SVG is missing font styles that were present in the live editor page.

To fix, open the new architecture.svg, then add the following to the first class (`#mermaid-numbers .label`).

  font-size: 14px;
  font-variant: tabular-nums;
  line-height: 1.5;

Save architecture.svg, view architecture.md and the labels should now render correctly.
-->

## Services

Here are the services that compose Sourcegraph.

### frontend ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/frontend))

The frontend serves our [web app](web_app.md) and hosts our [GraphQL API](../api/graphql/index.md).

Application data is stored in our Postgresql database.

Session data is stored in Redis.

#### Scaling

Typically there are multiple replicas running in production to scale with load.

frontend tends to use a large amount of memory. For example our search architecture does a scatter and gather amongst the search backends in the frontend. The gathering of results can result in a lot of memory usage, even though the final result set returned to the user is much smaller. There are a few more examples of these since our frontend has a monolithic architecture. Additionally we haven't optimized for memory usage since it hasn't caused us issues in production since we can just scale it out.

### github-proxy ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/github-proxy))

Proxies all requests to github.com to keep track of rate limits and prevent triggering abuse mechanisms.

There is only one replica running in production. However, we can have multiple replicas to increase our rate limits (rate limit is per IP).

### gitserver ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/gitserver))

Mirrors repositories from their code host. All other Sourcegraph services talk to gitserver when they need data from git. Requests for fetch operations, however, should go through repo-updater.

#### Scaling

gitserver's memory usage consists of short lived git subprocesses.

This is an IO and compute heavy service since most Sourcegraph requests will trigger 1 or more git commands. As such we shard requests for a repo to a specific replica. This allows us to horizontally scale out the service.

The service is stateful (maintaining git clones). However, it only contains data mirrored from upstream code hosts.

### Sourcegraph extensions

[Sourcegraph extensions](../extensions/index.md) add features to Sourcegraph, including language support. Many extensions rely, in turn, on language servers (implementing the [Language Server Protocol](https://microsoft.github.io/language-server-protocol/)) to provide code intelligence (hover tooltips, jump to definition, find references).

### query-runner ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/query-runner))

Periodically runs saved searches and sends notification emails. Only one replica should be running.

### repo-updater ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/repo-updater))

Repo-updater (which may get renamed since it does more than that) tracks the state of repos, and is responsible for automatically scheduling updates ("git fetch" runs) using gitserver. Other apps which desire updates or fetches should be telling repo-updater, rather than using gitserver directly, so repo-updater can take their changes into account. Only one replica should be running.

### searcher ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/searcher))

Provides on-demand search for repositories. It scans through a git archive fetched from gitserver to find results.

This service should be scaled up the more on-demand searches that need to be done at once. For a search the frontend will scatter the search for each repo@commit across the replicas. The frontend will then gather the results. Like gitserver this is an IO and compute bound service. However, its state is a cache which can be lost at anytime.

### indexed-search/zoekt ([code](https://github.com/sourcegraph/zoekt))

Provides search results for repositories that have been indexed.

This service can only have one replica. Typically large customers provision a large node for it since it is memory and CPU heavy. Note: We could shard across multiple replicas to scale out. However, we haven't had a customer were this is necessary yet so haven't written the code for it yet.

We forked [zoekt](https://github.com/google/zoekt) to add some Sourcegraph specific integrations. See our [fork's README](https://github.com/sourcegraph/zoekt/blob/master/README.md) for details.

### symbols ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/symbols))

Indexes symbols in repositories using Ctags. Similar in architecture to searcher, except over ctags output.

### syntect ([code](https://github.com/sourcegraph/syntect_server))

Syntect is a Rust service that is responsible for syntax highlighting.

Horizontally scalable, but typically only one replica is necessary.

### Browser extensions ([code](https://github.com/sourcegraph/sourcegraph/tree/master/browser) | [docs](https://docs.sourcegraph.com/integration/browser_extension))

We publish browser extensions for Chrome, Firefox, and Safari, that provide code intelligence (hover tooltips, jump to definition, find references) when browsing code on code hosts. By default it works for open-source code, but it also works for private code if your company has a Sourcegraph deployment.

It uses GraphQL APIs exposed by the frontend to fetch data.

### Editor extensions ([docs](https://docs.sourcegraph.com/integration/editor))

Our editor extensions provide lightweight hooks into Sourcegraph, currently.

### When to create a new service

Sourcegraph is composed of several smaller services (gitserver, repo-updater, symbols, etc.) and a single monolithic service (the frontend). When thinking of adding a new service, it is important to think through the following questions carefully:

- Does the code belong in an existing service?
    - If yes, it most likely belongs in that service container. For example, don't introduce a seperate container to cleanup gitserver repositories if gitserver itself could reasonably perform that same work.
- Instead of being a seperate service, could it reasonably live inside the frontend as a singleton background worker?
- Does it rely heavily on the APIs that exist in another service?
- If done in an existing container, would it substantially increase the complexity of the task?
    - For example, the service you are writing _must_ be written in language X and it is impossible/very difficult to integrate language X into one of our existing Go services.
- Does it need its own resource constraints and scaling?
   - For example, the service you are creating needs its own CPU / memory resource constraints, or must be able to scale horizontally across machines.

If after asking the above questions to yourself you still believe introducing a new service is the best approach, propose it to the rest of the team by answering the above questions and explaining why you think a seperate service is warranted.

#### Complexity of introducing new services for us and users

Whether implemented in a new service container or not, introducing a new long-running service means we must:

- Create Prometheus metrics and a Grafana dashboard to monitor it.
- Define and set up a clear set of Prometheus metrics to alert on, both for us on Sourcegraph.com and for our customers.

When introducing a new service _container_, additional complexity is involved:

- It needs [its own Kubernetes YAML](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/base).
- It must be [integrated into the single-container `sourcegraph/server` deployment mode](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/server).
- It needs to be documented [how it scales](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/scale.md) alongside other services for cluster deployments.
- Our [architecture diagram](https://github.com/sourcegraph/sourcegraph/blob/master/doc/dev/architecture.md) and documentation must be updated.
- We must update [the docker-compose file we ship to some customers (near future)](https://github.com/sourcegraph/deploy-sourcegraph-docker/pull/38), and advise them they must update it when upgrading.
- We must [update deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker) and test that it works in that deployment mode and denote any container <-> container environment variables that must be set.
    - As well, we must email customers using that deployment mode and advise them exactly which new container has been added with a short blurb about what it does, why we've added it, and how to deploy it (remember, these users are not just running the scripts in our repository -- they are effectively deploying these containers individually and manually).

In general, introducing a new Docker container only makes sense when there is a clear need to do so (such as applying seperate resource constraints on the code, or scaling it horizontally across multiple machines). Do not add new _containers_ just to seperate code and avoid technical debt (instead, find ways to refactor code so that we have appropriate seperate logical units as you desire -- but within an existing appropriate container).
