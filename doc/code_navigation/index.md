<style>

.markdown-body h2 {
  margin-top: 2em;
}

.markdown-body ul {
  padding-left: 1em;
}

.markdown-body ul li {
  margin: 0.5em 0;
}

.markdown-body .lead-screenshot {
    float: right;
    display: block;
    margin: 1em auto;
    max-width: 500px;
    margin-left: 0.5em;
    border: 1px solid lightgrey;
    border-radius: 10px;
}

</style>

# Code navigation

<p class="subtitle">Navigate your code with tooling that understands it</p>

<div>
<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/code-intel-overview.png" class="lead-screenshot">

<p class="lead">
Code navigation enables
developers to explore code by
</p>


<ul class="lead">
<li>jumping to definitions</li>
<li>finding references</li>
<li>listing implementations</li>
<li>browsing symbols defined in current document or folder</li>
<li>navigate dependencies</li>
<li>documentation in hover tooltips</li>
</ul>
</div>

<div style="display: block; float: clear;"> </div>

<div class="cta-group">
<a class="btn btn-primary" href="https://sourcegraph.com/github.com/dgrijalva/jwt-go/-/blob/token.go?L37:6#tab=references">★ Try it on public code!</a>
<a class="btn" href="#code-intelligence-for-your-code">Enable it for your code</a>
<a class="btn" href="references/precise_examples">Examples</a>
<a class="btn" href="references/indexers">🗂 Indexers</a>
</div>

Code navigation is made up of multiple features that build on top of each other:

- [Search-based code navigation](explanations/search_based_code_navigation.md) works out of the box with all of the most popular programming languages, powered by Sourcegraph's code search and [extensions](https://sourcegraph.com/extensions?query=category%3A%22Programming+languages%22).
- [Precise code navigation](explanations/precise_code_navigation.md) uses code graph data to provide correct code navigation features and accurate cross-repository navigation.
- [Auto-indexing](explanations/auto_indexing.md) uses [Sourcegraph executors](../admin/executors.md) to create indexes for the code in your Sourcegraph instance, giving you up-to-date cross-repository code navigation.
- <span class="badge badge-beta">Beta</span> [Dependency navigation](explanations/features.md#dependency-navigation) allows you to navigate and search through the dependencies of your code, by leveraging precise code navigation and auto-indexing.

## Code navigation for your code

Here's how you go from search-based code navigation to **automatically-updating, precise code navigation across multiple repositories and dependencies**:

1. Navigate code with [search-based code navigation](explanations/search_based_code_navigation.md) and [Sourcegraph extensions](../../../extensions/index.md).

    Included in a standard Sourcegraph installation and works out of the box on the instances connected to the Internet.
    To see how to enable code navigation on the air-gapped instances please check [this guide](how-to/enable_code_intel_on_air_gapped_instances.md).
1. Start using [precise code navigation](explanations/precise_code_navigation.md) by creating an index of a repository and uploading it to your Sourcegraph instance:

    - [Index a Go repository](how-to/index_a_go_repository.md#manual-indexing)
    - [Index a TypeScript or JavaScript repository](how-to/index_a_typescript_and_javascript_repository.md#manual-indexing)
    - [Index a C++ repository](how-to/index_a_cpp_repository.md)
    - [Index a Java, Scala & Kotlin repository](https://sourcegraph.github.io/scip-java/docs/getting-started.html)
    - [Index a Python repository](https://github.com/sourcegraph/scip-python)

1. _Optional_: automate the uploading of indexes by [adding indexing and uploading to your CI setup](how-to/adding_lsif_to_workflows.md).
1. Get **automatic precise code navigation** by [enabling auto-indexing](how-to/enable_auto_indexing.md) which makes Sourcegraph automatically index the your repositories and enable precise code navigation for them.
1. Setup **auto-dependency indexing** to navigate and search through the dependencies used by your code:
    - **Go**: Enable [auto-indexing](explanations/auto_indexing.md) and Sourcegraph will start indexing your dependencies.
    - **Java, Scala, Kotlin**: Enable [auto-indexing](explanations/auto_indexing.md) and setup a [JVM dependencies code host](../../integration/jvm.md).
    - **JavaScript, TypeScript**: Enable [auto-indexing](explanations/auto_indexing.md) and setup a [npm dependencies code host](../../integration/npm.md).

Once setup, code navigation is available for use across popular development tools:

- In the Sourcegraph web UI
- When browsing code on your code host, via [integrations](../../../integration/index.md)
- While looking at diffs in your code review tool, via [integrations](../../../integration/index.md)
- In the [Sourcegraph API](https://docs.sourcegraph.com/api/graphql)

## [Explanations](explanations/index.md)

- [Introduction to code navigation](explanations/introduction_to_code_navigation.md)
- [Precise code navigation](explanations/precise_code_navigation.md)
  - [Code graph data uploads](explanations/uploads.md)
- [Search-based code navigation](explanations/search_based_code_navigation.md)
- [Code navigation features](explanations/features.md)
- <span class="badge badge-beta">Beta</span> [Rockskip: faster search-based code navigation](explanations/rockskip.md)
- [Writing an indexer](explanations/writing_an_indexer.md)
- <span class="badge badge-beta">Beta</span> [Auto-indexing](explanations/auto_indexing.md)
- <span class="badge badge-beta">Beta</span> [Auto-indexing inference](explanations/auto_indexing_inference.md)


## [How-tos](how-to/index.md)

- General
  - [Configure data retention policies](how-to/configure_data_retention.md)
  - [Enable code navigation on the air-gapped instances](how-to/enable_code_intel_on_air_gapped_instances.md)

- Language-specific guides
  - [Index a Go repository](how-to/index_a_go_repository.md)
  - [Index a TypeScript or JavaScript repository](how-to/index_a_typescript_and_javascript_repository.md)
  - [Index a C++ repository](how-to/index_a_cpp_repository.md)
  - [Index a Java, Scala & Kotlin repository](https://sourcegraph.github.io/scip-java/docs/getting-started.html)
- Automate uploading LSIF data
  - [Add LSIF to many repositories](how-to/adding_lsif_to_many_repos.md)
  - [Adding LSIF to CI workflows](how-to/adding_lsif_to_workflows.md)
  - <span class="badge badge-beta">Beta</span> [Enable auto-indexing](how-to/enable_auto_indexing.md)
  - <span class="badge badge-beta">Beta</span> [Configure auto-indexing](how-to/configure_auto_indexing.md)

## [References](references/index.md)

- [Requirements](references/requirements.md)
- [Troubleshooting](references/troubleshooting.md)
- [FAQ](references/faq.md)
- [Sourcegraph recommended indexers](references/indexers.md)
- [Environment variables](references/envvars.md)
- <span class="badge badge-beta">Beta</span> [Auto-indexing configuration](references/auto_indexing_configuration.md)


