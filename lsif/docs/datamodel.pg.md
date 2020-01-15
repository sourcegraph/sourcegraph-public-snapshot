# Cross-repo data model

This document outlines the data model used to correlate multiple LSIF dumps. The definition of the cross-repo database tables can be found in [pg.ts](../src/shared/models/pg.ts).

In the following document, commits have been abbreviated to 7 characters for readibility.

## Database tables

**`lsif_commits` table**

This table contains all commits known for a repository for which LSIF data has been uploaded. Each commit consists of one or more rows indicating their parent. If a commit has no parent, then the parentCommit field is an empty string.

| id  | repository                    | commit    | parentCommit |
| --- | ----------------------------- | --------- | ------------ |
| 1   | github.com/sourcegraph/sample | `a360643` | `f4fb066`    |
| 2   | github.com/sourcegraph/sample | `f4fb066` | `4c8d9dc`    |
| 3   | github.com/sourcegraph/sample | `313082b` | `4c8d9dc`    |
| 4   | github.com/sourcegraph/sample | `4c8d9dc` | `d67b8de`    |
| 5   | github.com/sourcegraph/sample | `d67b8de` | `323e23f`    |
| 6   | github.com/sourcegraph/sample | `323e23f` |              |

This table allows us to ues recursive CTEs to find ancestor and descendant commits with a particular property (as indicated by the existence of an entry in the `lsif_dumps` table) and enables closest commit functionality.

**`lsif_uploads` table**

This table contains an entry for each LSIF upload. An upload is inserted with the state `queued` and is processed asynchronously by a worker. The `root` indicates the directory for which this upload provides code intelligence. The `visible_at_tip` indicates whether this a (completed) upload that is closest to the tip of the default branch.

| id  | repository                    | commit    | root | state     | visible_at_tip |
| --- | ----------------------------- | --------- | ---- | --------- | -------------- |
| 1   | github.com/sourcegraph/sample | `a360643` |      | completed | true           |
| 2   | github.com/sourcegraph/sample | `f4fb066` |      | completed | false          |
| 3   | github.com/sourcegraph/sample | `4c8d9dc` | cmd  | completed | true           |
| 4   | github.com/sourcegraph/sample | `323e23f` |      | completed | false          |

The view `lsif_dumps` selects all uploads with a state of `completed`.

Additional fields are not shown in the table above which do not affect code intelligence queries in a meaningful way.

- `filename`: The filename of the raw upload.
- `uploaded_at`: The time the record was inserted.
- `started_at`: The time the conversion was started.
- `finished_at`: The time the conversion was finished.
- `failure_summary`: The message of the error that occurred during conversion.
- `failure_stacktrace`: The stacktrace of the error that occurred during conversion.
- `tracing_context`: The tracing context from the `/upload` endpoint. Used to trace the entire span of work from the upload to the end of conversion.

**`lsif_packages` table**

This table links a package manager-specific identifier and version to the repository and commit that _provides_ the package. The scheme, name, and version values are correlated with a moniker and its package information from an LSIF dump.

| id  | scheme | name   | version | repository                    | commit    |
| --- | ------ | ------ | ------- | ----------------------------- | --------- |
| 1   | npm    | sample | 0.1.0   | github.com/sourcegraph/sample | `e58d28c` |

This table enables cross-repository jump-to-definition. When a range has no definition result but does have an _import_ moniker, the scheme, name, and version of the moniker can be queried in this table to get the repository and commit of the package that should contain that moniker's definition.

**`lsif_references` table**

This table links a repository and commit to the set of packages on which it depends. This table shares common columns with the `lsif_packages` table, which are documented above. In addition, this table also has a `filter` column, which encodes a [bloom filter](https://en.wikipedia.org/wiki/Bloom_filter) populated with the set of identifiers that the commit imports from the dependent package.

| id  | scheme | name      | version | repository                    | commit    | filter                       |
| --- | ------ | --------- | ------- | ----------------------------- | --------- | ---------------------------- |
| 1   | npm    | left-pad  | 0.1.0   | github.com/sourcegraph/sample | `e58d28c` | _gzipped_ and _json-encoded_ |
| 2   | npm    | right-pad | 1.2.3   | github.com/sourcegraph/sample | `e58d28c` | _gzipped_ and _json-encoded_ |
| 2   | npm    | left-pad  | 0.1.0   | github.com/sourcegraph/sample | `9f6e6ec` | _gzipped_ and _json-encoded_ |
| 2   | npm    | right-pad | 1.2.4   | github.com/sourcegraph/sample | `9f6e6ec` | _gzipped_ and _json-encoded_ |

This table enables global find-references. When finding all references o fa definition that has an _export_ moniker, the set of repositories and commits that depend on the package of that moniker are queried. We want to open only the databases that import this particular symbol (not all projects depending on this package import the identifier under query). To do this, the bloom filter is deserialized and queried for the identifier under query. A positive response from a bloom filter indicates that the identifier may be present in the set; a negative response from the bloom filter indicates that the identifier is _definitely_ not in the set. We only open the set of databases for which the bloom filter query responds positively.
