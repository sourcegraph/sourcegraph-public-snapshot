# Pagerank generator for lsif indexes

This directory contains a quick-and-dirty tool for using an lsif index to run the Google PageRank algorithm on the symbol graph.

## Motivation

The intuition is that by treating symbols as "virtual web pages", and references between symbols as virtual links, then running pagerank on the resulting reference graph will bubble the most important symbols to the top.

The symbol-pagerank ranking could potentially be used as an additional signal for scoring file results in any context; e.g., in regular zoekt queries.

## Approach

Pagerank is computed from a graph of directed, unweighted edges. Our edges are from the symbol-reference graph, which only contains as much information as is recorded by the indexers.

The lsif graph model is not a symbol database, and we do not record very much information about symbol scopes. Hence for now, we consider all symbols to be file-scoped. The lsif-pagerank tool walks the references in an input lsif file, computes the edges between file vertices for the referenced symbol and its definition location(s), and feeds those edges into the PageRank calculator.

The result is a sorted list of all the file paths from the index, ranked by their computed PageRank.

## Limitations

- PageRank itself is a fairly simplistic model for estimating importance, and it's probably only useful as a tie-breaking signal during regular ranking.
- Running the algorithm on code graphs is currently non-deterministic: the rankings change each time you run it, though the results always look pretty reasonable. It tends to surface core library routines over application/tool code, which is the desired behavior.
- PageRank should probably be computed at upload time; this manual tool-based design is a POC.
- This tool assumes the whole lsif file is in memory--the computation should eventually be sharded.
- We only have file-level ranking, which is less accurate/fine-grained than symbol-level ranking.
- PageRank scores should be cached in the database or blob store, and added to the codeintel APIs.
- There should be tests, although perhaps not until it is actually integrated properly.

## Future directions

The current design and implementation of Symbol PageRank are naive and could be improved in several ways, including improving the reference-graph modeling and fidelity, adding cross-repo pagerank calculation, and generally studying how the algorithm might be tweaked and tuned for best results on code graphs.

### Integration with Zoekt

This PageRank library emits scores between 0.0 and 1.0, with real-world scores ranging from perhaps 0.1-0.2 max down to 10e-4 at the low end.

Zoekt scores are integral and range from tens to perhaps hundreds of thousands. PageRank might be integrated in as an extra file scoring factor of, say, 1+PR(E). Normally a search result gets a 10.0 score for its file weight. If we factor in the PR for this page, and it is, say, 0.05 (a rather high score), then the file match score factor for that file would be 10.0 * (1 + 0.05) = 10.5.

## PageRank 3P Library

The PageRank implementation used by the lsif-pagerank tool is from [dcadenas on GitHub](https://github.com/dcadenas/pagerank), MIT licensed, and has not been tested or vetted for correctness or performance.
