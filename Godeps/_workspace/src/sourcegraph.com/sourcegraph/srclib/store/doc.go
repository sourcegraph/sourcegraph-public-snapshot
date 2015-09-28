/*

Package store handles srclib data access, importing, querying, and
storage.

It uses the filesystem as its underlying storage. Access to the
filesystem is abstracted through a virtual filesystem (VFS) interface.

This package writes two types of files: index files and data
files. Index files are per-source unit or per-repo and support fast
lookups by certain fields or criteria.

* Indexes are consulted by the indexedXyz stores in indexed.go (which
  implement the same interface as their underlying, non-indexed store
  and delegate to the underlying store when the index does not cover
  the query).

  Indexes are constructed during the index process using a variety of
  methods, depending on the operation they need to support. For
  example, the search query index (which maps prefix queries by def
  names to the defs that match, repo-wide) builds a MAFSA and writes
  it to the def_query index file. Check the *_index.go files for the
  full set of indexes in use.

  See the godoc for Index for more information.

* Data files consist of packed, varint-length-encoded protobufs of a
  certain type (defs, refs, etc.). The indexes typically contain byte
  offsets that refer to positions in the data files.


CONVENTION - ZERO VALUES FOR FIELDS OUTSIDE OF A STORE'S SCOPE

We adopt the convention that all items in a store contain zero
values in fields pertaining to things outside the store's
scope. For example, remember that all items in a TreeStore are, by
definition, at the same commit. So, the TreeStore need not be aware
of commits; commits are out of a TreeStore's scope. Thus all items
in a TreeStore or returned by its methods have their CommitID
fields set to "", but when a higher-level store (MultiRepoStore or
RepoStore) receives data from a TreeStore, it sets those CommitID
fields to the TreeStore's commit ID.

Why? This lets us avoid reduce allocations and assignments and save
storage space.

The following demonstrates why this approach does not lead to
incorrect results from filters. There are 2 ways filters can be
applied. Filters can either:

* Narrow the scope of the search to only certain stores, excluding
  those that we're certain can't contain what we're looking for
  (e.g., when we're looking for a specific repo, only that repo's
  RepoStore needs to be opened). If the filters implement any of
  the ByXyzFilter types, this occurs. OR

* Filter each item individually by having their SelectXyz methods
  called on each item.

If a filter is used to narrow the scope (the first filter
application method above), then when it is called to filter items
in lower-level stores, the data in those lower-level stores will
have the zero value for the field by which we narrowed the
scope. We could set the field value on each item after narrowing
the scope but before calling our filter on each item, but that
would be both time-consuming and unnecessary.

E.g., Suppose we have a ByReposFilter and immediately narrow our
scope to a single RepoStore. All items in a RepoStore by definition
are from the same repo, so it's unnecessary for any of their Repo
fields to be set. Therefore their Repo fields are "", but we know
that they have already satisfied the filter.

Filters should still apply criteria in their SelectXyz funcs even if
they also implement corresponding ByXyzFilter interfaces. This allows
the filters to be combined arbitrarily (even if the AND of their
ByXyzFilters can't be used to narrow the scope) and allows scope
narrowing to use techniques that may yield false positives (e.g.,
bloom filters).


DEBUGGING

The `srclib store` subcommands allow you to perform most store
operations interactively.

Set the env var V=1 while running tests or `srclib store` commands for
more output.

*/
package store

// TODO(sqs): pass a list of ...XyzFilter instead of a single filter,
// so we can apply multiple filters but still be able to type-assert
// them.
