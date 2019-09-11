# Cross-repo data model

This document outlines the data model used to correlate multiple LSIF dumps. The definition of the cross-repo database tables can be found in `models.xrepo.ts`.

## Database values

**`packages` table**

This table links a package manager-specific identifier and version to the repository and commit that _provides_ the package. The scheme, name, and version values are correlated with a moniker and its package information from an LSIF dump.

| id  | scheme | name   | version | repository                    | commit                                   |
| --- | ------ | ------ | ------- | ----------------------------- | ---------------------------------------- |
| 1   | npm    | sample | 0.1.0   | github.com/sourcegraph/sample | e58d28c98a43f97112299ad6e590e5846b241763 |

This table enables cross-repository jump-to-definition. When a range has no definition result but does have an _import_ moniker, the scheme, name, and version of the moniker can be queried in this table to get the repository and commit of the package that should contain that moniker's definition.

**`references` table**

This table links a repository and commit to the set of packages on which it depends. This table shares common columns with the `packages` table, which are documented above. In addition, this table also has a `filter` column, which encodes a [bloom filter](https://en.wikipedia.org/wiki/Bloom_filter) populated with the set of identifiers that the commit imports from the dependent package.

| id  | scheme | name      | version | repository                    | commit                                   | filter                       |
| --- | ------ | --------- | ------- | ----------------------------- | ---------------------------------------- | ---------------------------- |
| 1   | npm    | left-pad  | 0.1.0   | github.com/sourcegraph/sample | e58d28c98a43f97112299ad6e590e5846b241763 | _gzipped_ and _json-encoded_ |
| 2   | npm    | right-pad | 1.2.3   | github.com/sourcegraph/sample | e58d28c98a43f97112299ad6e590e5846b241763 | _gzipped_ and _json-encoded_ |
| 2   | npm    | left-pad  | 0.1.0   | github.com/sourcegraph/sample | 9f6e6ec73509159714606ec77e1c55be75235346 | _gzipped_ and _json-encoded_ |
| 2   | npm    | right-pad | 1.2.4   | github.com/sourcegraph/sample | 9f6e6ec73509159714606ec77e1c55be75235346 | _gzipped_ and _json-encoded_ |

This table enables global find-references. When finding all references o fa definition that has an _export_ moniker, the set of repositories and commits that depend on the package of that moniker are queried. We want to open only the databases that import this particular symbol (not all projects depending on this package import the identifier under query). To do this, the bloom filter is deserialized and queried for the identifier under query. A positive response from a bloom filter indicates that the identifier may be present in the set; a negative response from the bloom filter indicates that the identifier is _definitely_ not in the set. We only open the set of databases for which the bloom filter query responds positively.
