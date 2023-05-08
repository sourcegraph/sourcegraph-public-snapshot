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
<a class="btn" href="#code-navigation-for-your-code">Enable it for your code</a>
<a class="btn" href="references/precise_examples">Examples</a>
<a class="btn" href="references/indexers">🗂 Indexers</a>
</div>

Code navigation is made up of multiple features that build on top of each other:

- [Search-based code navigation](explanations/search_based_code_navigation.md) works out of the box with all of the most popular programming languages, powered by Sourcegraph's code search.
- [Precise code navigation](explanations/precise_code_navigation.md) uses code graph data to provide precise code navigation features and accurate cross-repository navigation. We currently have precise code navigation support for the languages below. See the [indexers page](references/indexers.md) for a detailed breakdown of each indexer's status.
<table>
   <thead>
      <tr>
        <th>Language</th>
        <th>Indexer</th>
        <th>Status</th>
      </tr>
   </thead>
   <tbody>
      <tr>
        <td>Go</td>
        <td><a href="https://sourcegraph.com/github.com/sourcegraph/lsif-go">lsif-go</a></td>
        <td>🟢 GA</td>
      </tr>
      <tr>
        <td>TypeScript/JavaScript</td>
        <td><a href="https://sourcegraph.com/github.com/sourcegraph/scip-typescript">scip-typescript</a></td>
        <td>🟢 GA</td>
      </tr>
      <tr>
        <td>C/C++</td>
        <td><a href="https://sourcegraph.com/github.com/sourcegraph/lsif-clang">lsif-clang</a></td>
        <td>🟡 Partially available</td>
      </tr>
      <tr>
         <td>Java</td>
        <td><a href="https://sourcegraph.com/github.com/sourcegraph/scip-java">scip-java</a></td>
        <td>🟢 GA</td>
      </tr>
      <tr>
        <td>Scala</td>
        <td><a href="https://sourcegraph.com/github.com/sourcegraph/scip-java">scip-java</a></td>
        <td>🟢 GA</td>
      </tr>
      <tr>
        <td>Kotlin</td>
        <td><a href="https://sourcegraph.com/github.com/sourcegraph/scip-java">scip-java</a></td>
        <td>🟢 GA</td>
      </tr>
      <tr>
        <td>Rust</td>
        <td><a href="https://sourcegraph.com/github.com/rust-lang/rust-analyzer">rust-analyzer</a></td>
        <td>🟢 GA</td>
      </tr>
     <tr>
        <td>Python</td>
        <td><a href="https://sourcegraph.com/github.com/sourcegraph/scip-python">scip-python</a></td>
        <td>🟢 GA</td>
      </tr>
     <tr>
        <td>Ruby</td>
        <td><a href="https://sourcegraph.com/github.com/sourcegraph/scip-ruby">scip-ruby</a></td>
        <td>🟢 GA</td>
      </tr>
      <tr>
        <td>C#</td>
        <td><a href="https://github.com/sourcegraph/scip-dotnet">scip-dotnet</a></td>
        <td>🟡 Partially available</td>
      </tr>
   </tbody>
</table>

- [Auto-indexing](explanations/auto_indexing.md) uses [Sourcegraph executors](../admin/executors/index.md) to create indexes for the code in your Sourcegraph instance, giving you up-to-date, cross-repository code navigation.
- <span class="badge badge-beta">Beta</span> [Dependency navigation](explanations/features.md#dependency-navigation) allows you to navigate and search through the dependencies of your code, leveraging precise code navigation and auto-indexing.

## Code navigation for your code

Here's how you go from search-based code navigation to **automatically updating, precise code navigation across multiple repositories and dependencies**:

1. Navigate code with [search-based code navigation](explanations/search_based_code_navigation.md).
1. Start using [precise code navigation](explanations/precise_code_navigation.md) by creating an index of a repository and uploading it to your Sourcegraph instance:

    - [Index a Go repository](how-to/index_a_go_repository.md#manual-indexing)
    - [Index a TypeScript or JavaScript repository](how-to/index_a_typescript_and_javascript_repository.md#manual-indexing)
    - [Index a Java, Scala, or Kotlin repository](https://sourcegraph.github.io/scip-java/docs/getting-started.html)
    - [Index a Python repository](https://sourcegraph.com/github.com/sourcegraph/scip-python)
    - [Index a Ruby repository](https://sourcegraph.com/github.com/sourcegraph/scip-ruby)
    - [Index a C++ repository](how-to/index_a_cpp_repository.md)


1. Optionally automate index uploading by [adding indexing and uploading to your CI setup](how-to/adding_lsif_to_workflows.md).
1. [Enable auto-indexing](how-to/enable_auto_indexing.md) on your Sourcegraph instance to get **automatic precise code navigation**. Sourcegraph will automatically index your repositories and enable precise code navigation for them.
1. Set up **auto-dependency indexing** to navigate and search through the dependencies your code uses:
    - **Go**: Enable [auto-indexing](explanations/auto_indexing.md) and Sourcegraph will start indexing your dependencies.
    - **JavaScript, TypeScript**: Enable [auto-indexing](explanations/auto_indexing.md) and set up an [npm dependencies code host](../../integration/npm.md).
    - **Java, Scala, Kotlin**: Enable [auto-indexing](explanations/auto_indexing.md) and set up a [JVM dependencies code host](../../integration/jvm.md).
    

Once set up, you can use code navigation with popular development tools:

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
- <span class="badge badge-beta">Beta</span> [Rockskip: Faster search-based code navigation](explanations/rockskip.md)
- [Writing an indexer](explanations/writing_an_indexer.md)
- [Auto-indexing](explanations/auto_indexing.md)
- [Auto-indexing inference](explanations/auto_indexing_inference.md)


## [How-to guides](how-to/index.md)

- General
  - [Configure data retention policies](how-to/configure_data_retention.md)

- Language-specific guides
  - [Index a Go repository](how-to/index_a_go_repository.md)
  - [Index a TypeScript or JavaScript repository](how-to/index_a_typescript_and_javascript_repository.md)
  - [Index a C++ repository](how-to/index_a_cpp_repository.md)
  - [Index a Java, Scala, or Kotlin repository](https://sourcegraph.github.io/scip-java/docs/getting-started.html)
- Automate uploading code graph data
  - [Add code graph data to many repositories](how-to/adding_lsif_to_many_repos.md)
  - [Adding code graph data to CI workflows](how-to/adding_lsif_to_workflows.md)
  - [Enable auto-indexing](how-to/enable_auto_indexing.md)
  - [Configure auto-indexing](how-to/configure_auto_indexing.md)

## [Reference](references/index.md)

- [Requirements](references/requirements.md)
- [Troubleshooting](references/troubleshooting.md)
- [FAQ](references/faq.md)
- [Sourcegraph-recommended indexers](references/indexers.md)
- [Environment variables](references/envvars.md)
- [Auto-indexing configuration](references/auto_indexing_configuration.md)


