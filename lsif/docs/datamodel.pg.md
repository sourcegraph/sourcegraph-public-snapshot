# Cross-repo data model

This document outlines the data model used to correlate multiple LSIF dumps. The definition of the cross-repo database tables can be found in [pg.ts](../src/shared/models/pg.ts).

In the following document, commits have been abbreviated to 7 characters for readability.

## Database tables

**`lsif_commits` table**

This table contains all commits known for a repository for which LSIF data has been uploaded. Each commit consists of one or more rows indicating their parent. If a commit has no parent, then the parentCommit field is an empty string.

| id  | repository_id | commit    | parent_commit |
| --- | ------------- | --------- | ------------- |
| 1   | 6             | `a360643` | `f4fb066`     |
| 2   | 6             | `f4fb066` | `4c8d9dc`     |
| 3   | 6             | `313082b` | `4c8d9dc`     |
| 4   | 6             | `4c8d9dc` | `d67b8de`     |
| 5   | 6             | `d67b8de` | `323e23f`     |
| 6   | 6             | `323e23f` |               |

This table allows us to ues recursive CTEs to find ancestor and descendant commits with a particular property (as indicated by the existence of an entry in the `lsif_dumps` table) and enables closest commit functionality.

**`lsif_uploads` table**

This table contains an entry for each LSIF upload. An upload is inserted with the state `queued` and is processed asynchronously by a dump processor. The `root` field indicates the directory for which this upload provides code intelligence. The `indexer` field indicates the tool that generated the input. The `visible_at_tip` field indicates whether this a (completed) upload that is closest to the tip of the default branch.

| id  | repository_id | commit    | root | indexer | state     | visible_at_tip |
| --- | ------------- | --------- | ---- | ------- | --------- | -------------- |
| 1   | 6             | `a360643` |      | lsif-go | completed | true           |
| 2   | 6             | `f4fb066` |      | lsif-go | completed | false          |
| 3   | 6             | `4c8d9dc` | cmd  | lsif-go | completed | true           |
| 4   | 6             | `323e23f` |      | lsif-go | completed | false          |

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

This table links a package manager-specific identifier and version to the LSIF upload _provides_ the package. The scheme, name, and version values are correlated with a moniker and its package information from an LSIF dump.

| id  | scheme | name   | version | dump_id |
| --- | ------ | ------ | ------- | ------- |
| 1   | npm    | sample | 0.1.0   | 6       |

This table enables cross-repository jump-to-definition. When a range has no definition result but does have an _import_ moniker, the scheme, name, and version of the moniker can be queried in this table to get the repository and commit of the package that should contain that moniker's definition.

**`lsif_references` table**

This table links an LSIF upload to the set of packages on which it depends. This table shares common columns with the `lsif_packages` table, which are documented above. In addition, this table also has a `filter` column, which encodes a [bloom filter](https://en.wikipedia.org/wiki/Bloom_filter) populated with the set of identifiers that the commit imports from the dependent package.

| id  | scheme | name      | version | filter                       | dump_id |
| --- | ------ | --------- | ------- | ---------------------------- | ------- |
| 1   | npm    | left-pad  | 0.1.0   | _gzipped_ and _json-encoded_ | 6       |
| 2   | npm    | right-pad | 1.2.3   | _gzipped_ and _json-encoded_ | 6       |
| 2   | npm    | left-pad  | 0.1.0   | _gzipped_ and _json-encoded_ | 7       |
| 2   | npm    | right-pad | 1.2.4   | _gzipped_ and _json-encoded_ | 7       |

This table enables global find-references. When finding all references of a definition that has an _export_ moniker, the set of repositories and commits that depend on the package of that moniker are queried. We want to open only the databases that import this particular symbol (not all projects depending on this package import the identifier under query). To do this, the bloom filter is deserialized and queried for the identifier under query. A positive response from a bloom filter indicates that the identifier may be present in the set; a negative response from the bloom filter indicates that the identifier is _definitely_ not in the set. We only open the set of databases for which the bloom filter query responds positively.
