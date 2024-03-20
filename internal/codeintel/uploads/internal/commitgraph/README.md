# commitgraph

Problem: Given a repository id and a commit hash, find me the set of index uploads most relevant to that commit.

There are various reasons for why a commit might not have associated indexes uploaded
- It is a monorepo moving at high commit velocity relative to indexing speed
- It's not covered by any indexing policy. For example, some customers only index commits on the `main` branch, or on release tags.
- The upload is in a queued, processing, or errored/failed state.
- The index was deleted, by a retention policy or an explicit user action.

To solve this problem we build an [annotated copy of the git commit graph](#what-is-an-annotated-commit-graph) in Postgres, so that we can return so called "visibleUploads" quickly.

## High-level architecture

- repository listing (table `lsif_dirty_repositories`), isDirty flag is set whenever an upload is completed
- Worker process updates commit graph for dirty repositories and clears flag
- TODO: When information is requested for an unknown commit in a repository we ask git server for a fragment of the commit graph that includes the commit and "graft" it onto our existing graph

## Updating the commit graph
1. Load data
2. Compute new commit graph
3. Diff old vs new and do incremental updates

### Loading data

For the marked as dirty repository:
- Get full commit graph (non-annotated) from git server. All commits and their relationships
- Get metadata for all uploaded indices (indexer name, commit, root directory)

### What is an annotated commit graph

A fully annotated commit graph contains a set of visible uploads for every commit. We can never see multiple visible uploads from the same indexer _and_ root at the same commit.

Conceptually:

`commitgraph : Map<Commit, Map<(Indexer, Root), (UploadId, Distance)>>`

Here's what an example entry for a commit `deadbeef` might look like for our conceptual model:

```go
commitgraph = {
  "deadbeef" => {
    ("scip-go", "backend/") => (UploadId(23), 0),
    ("scip-go", "worker/") => (UploadId(24), 1),
    ("scip-typescript", "frontend/") => (UploadId(33), 1),
  }
}
```

We'll look at a couple of examples with increasing complexity to understand what constitutes visibility.

#### Visibility rules

Our examples will be drawings of commit graphs.

We will use `I1`..`In` to denote indexer+root combinations or "index keys". For example `I1` could correspond to indices created by `"scip-go"` in the `"backend/"` root directory.

We will use single uppercase letters in boxes with rounded corners to denote commits.
Colored boxes mark commits with uploads, and black boxes those with no uploads.
A commit with multiple uploads will have additional colored outlines.

Boxes with sharp corners record the set of visible uploads for a commit. An entry like `I1 : A(3)` means this commit sees an upload for `I1` made at commit `A` at a depth of 3.

![commitgraph_0.png](https://storage.googleapis.com/sourcegraph-assets/dev-docs/commitgraph/commitgraph_0.png)

**Rule 1: If an index key is uploaded at a particular commit, that commit will consider it visible at depth 0**

For our first example we'll start with a graph of a single commit B, and indices for both `I1` and `I2` have been uploaded for `B`.
According to our first rule both `I1` and `I2` are visible to B at depth 0.

**Rule 2: An upload is visible to a commit at depth N + 1 if it's visible to its parent at depth N**

![commitgraph_1.png](https://storage.googleapis.com/sourcegraph-assets/dev-docs/commitgraph/commitgraph_1.png)

Because there are no uploads at `A` and no uploads in `A`'s ancestors it does not have any visible uploads.
There are no uploads for `C`, but because there are uploads at `B` and `C` is its child, it has visibility onto both uploads at `B` at depth 1.

**Rule 3: Uploads for the same index key shadow uploads at greater depth**

![commitgraph_2.png](https://storage.googleapis.com/sourcegraph-assets/dev-docs/commitgraph/commitgraph_2.png)

In this example `C` now has an upload for `I1`, which means it sees the upload at depth 0, rather than the one from `B` at depth 1.

`E` sees the upload for `I2` at `D` at a distance of 1, rather than `B` because the upload at `B` has a distance of 2.
The notion of shadowing can be a bit unintuitive here, as `D` does *not* need to be on the path from `B` to `E` in order to shadow `B`.

_Implementation note: If there are two uploads at the same depth we pick the smaller upload_id. This choice is arbitrary but deterministic._

### Algorithm for computing the visibility graph

TODO: Input data
  - Topo-sorted list of commits and their relationships
  - Full list of uploaded indices

TODO: "Relevant" commits
  - Commits with an upload
  - "Merge" commits
  - All parents of "merge" commits

TODO: Traversal

### Updating database tables

TODO: Streaming data out of the computed annotated commit graph

TODO: Temporary tables and diffing
