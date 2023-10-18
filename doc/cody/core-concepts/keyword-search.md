# Keyword Search

<p class="subtitle">Learn how Cody makes use to Keyword Search to gather context.</p>

Keyword search is the traditional approach to text search. It splits content into terms and builds a mapping from terms to documents. At query time, it extracts terms from the query and uses the mapping to retrieve your documents.

Both Cody chat and completions use Keyword Search. It comes out of the box without any additional setup. While powerful, this method can be used as a fallback solution when a codebase lacks embeddings with enhanced context quality and decreased latency over time.

An example of a Keyword Search would look like this:

![keyword-search-example](https://storage.googleapis.com/sourcegraph-assets/Docs/keyword-search-example.png)
