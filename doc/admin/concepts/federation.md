# Federation

Federation refers to using multiple Sourcegraph servers together, each of which is responsible for a subset of repositories.

Currently the only supported federation use case is [redirecting your Sourcegraph instance's users to Sourcegraph.com for _public_ repositories](public_repositories.md) (instead of mirroring, analyzing, and indexing public repositories on your Sourcegraph instance).

## Future plans

We plan to enhance federation in the future to support merging data (such as cross-references) from multiple Sourcegraph instances, sharing user accounts, etc. [Post an issue](https://github.com/sourcegraph/sourcegraph/issues) if you have a specific feature request for federation.
