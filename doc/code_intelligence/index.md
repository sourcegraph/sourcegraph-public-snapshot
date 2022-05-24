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

# Code intelligence

<p class="subtitle">Navigate your code with tooling that understands it</p>

<div>
<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/code-intel-overview.png" class="lead-screenshot">

<p class="lead">
Code Intelligence adds advanced code navigation to Sourcegraph, enabling
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
</div>

Code Intelligence is made up of multiple features that build on top of each other:

- [Search-based code intelligence](explanations/search_based_code_intelligence.md) works out of the box with all of the most popular programming languages, powered by Sourcegraph's code search and [extensions](https://sourcegraph.com/extensions?query=category%3A%22Programming+languages%22).
- [Precise code intelligence](explanations/precise_code_intelligence.md) uses [LSIF indexes](https://lsif.dev/) to provide correct code intelligence features and accurate cross-repository navigation.
- [Auto-indexing](explanations/auto_indexing.md) uses [Sourcegraph executors](../admin/executors.md) to create LSIF indexes for the code in your Sourcegraph instance, giving you up-to-date cross-repository code intelligence.
- [Dependency navigation](explanations/features.md#dependency-navigation) allows you to navigate and search through the dependencies of your code, by leveraging precise code intelligence and auto-indexing.

## Code Intelligence for your code

Here's how you go from search-based code intelligence to **automatically-updating, precise code intelligence across multiple repositories and dependencies**:

1. Navigate code with [search-based code intelligence](explanations/search_based_code_intelligence.md) and [Sourcegraph extensions](../../../extensions/index.md):

    Included in a standard Sourcegraph installation and works out of the box!
1. Start using [precise code intelligence](explanations/precise_code_intelligence.md) by creating an LSIF index of a repository and uploading it to your Sourcegraph instance:

    - [Index a Go repository](how-to/index_a_go_repository.md#manual-indexing)
    - [Index a TypeScript or JavaScript repository](how-to/index_a_typescript_and_javascript_repository.md#manual-indexing)
    - [Index a C++ repository](how-to/index_a_cpp_repository.md)
    - [Index a Java, Scala & Kotlin repository](https://sourcegraph.github.io/scip-java/docs/getting-started.html)

    See the [tutorials](#tutorials) for more step-by-step instructions.
1. _Optional_: automate the uploading of LSIF indexes by [adding LSIF indexing and uploading to your CI setup](how-to/adding_lsif_to_workflows.md).
1. Get **automatic precise code intelligence** by [enabling auto-indexing](how-to/enable_auto_indexing.md) which makes Sourcegraph automatically index the your repositories and enable precise code intelligence for them.
1. Setup **auto-dependency indexing** to navigate and search through the dependencies used by your code:
    - **Go**: Enable [auto-indexing](explanations/auto_indexing.md) and Sourcegraph will start indexing your dependencies.
    - **Java, Scala, Kotlin**: Enable [auto-indexing](explanations/auto_indexing.md) and setup a [JVM dependencies code host](../../integration/jvm.md).
    - **JavaScript, TypeScript**: Enable [auto-indexing](explanations/auto_indexing.md) and setup a [npm dependencies code host](../../integration/npm.md).

Once setup, code intelligence is available for use across popular development tools:

- In the Sourcegraph web UI
- When browsing code on your code host, via [integrations](../../../integration/index.md)
- While looking at diffs in your code review tool, via [integrations](../../../integration/index.md)
- In the [Sourcegraph API](https://docs.sourcegraph.com/api/graphql)

## [Explanations](explanations/index.md)

- [Introduction to code intelligence](explanations/introduction_to_code_intelligence.md)
- [Precise code intelligence](explanations/precise_code_intelligence.md)
  - [Precise code intelligence uploads](explanations/uploads.md)
- [Search-based code intelligence](explanations/search_based_code_intelligence.md)
- [Code navigation features](explanations/features.md)
- <span class="badge badge-experimental">Experimental</span> [Rockskip: faster search-based code intelligence](explanations/rockskip.md)
- [Writing an indexer](explanations/writing_an_indexer.md)
- <span class="badge badge-experimental">Experimental</span> [Auto-indexing](explanations/auto_indexing.md)
- <span class="badge badge-experimental">Experimental</span> [Auto-indexing inference](explanations/auto_indexing_inference.md)


## [How-tos](how-to/index.md)

- General
  - [Configure data retention policies](how-to/configure_data_retention.md)
- Language-specific guides
  - [Index a Go repository](how-to/index_a_go_repository.md)
  - [Index a TypeScript or JavaScript repository](how-to/index_a_typescript_and_javascript_repository.md)
  - [Index a C++ repository](how-to/index_a_cpp_repository.md)
  - [Index a Java, Scala & Kotlin repository](https://sourcegraph.github.io/scip-java/docs/getting-started.html)
- Automate uploading LSIF data
  - [Add LSIF to many repositories](how-to/adding_lsif_to_many_repos.md)
  - [Adding LSIF to CI workflows](how-to/adding_lsif_to_workflows.md)
  - <span class="badge badge-experimental">Experimental</span> [Enable auto-indexing](how-to/enable_auto_indexing.md)
  - <span class="badge badge-experimental">Experimental</span> [Configure auto-indexing](how-to/configure_auto_indexing.md)

## [Tutorials](tutorials/index.md)

- [Manually index a popular Go repository](tutorials/indexing_go_repo.md)
- [Manually index a popular TypeScript repository](tutorials/indexing_ts_repo.md)


## [References](references/index.md)

- [Requirements](references/requirements.md)
- [Troubleshooting](references/troubleshooting.md)
- [FAQ](references/faq.md)
- [Sourcegraph recommended indexers](references/indexers.md)
- [Environment variables](references/envvars.md)
- <span class="badge badge-experimental">Experimental</span> [Auto-indexing configuration](references/auto_indexing_configuration.md)


