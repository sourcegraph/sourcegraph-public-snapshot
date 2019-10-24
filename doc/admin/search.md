# Search configuration

See "[Code search overview](../user/search/index.md)" for general information about Sourcegraph's code search.

## Indexed search

Sourcegraph can index the code on the default branch of each repository. This speeds up searches that hit many repositories at once. It also increases the memory and storage requirements for Sourcegraph, so it is disabled by default when running Sourcegraph on a single node.

To enable indexed search when running Sourcegraph on a single node, set the `search.index.enabled` [site configuration](config/site_config.md) property to `true`. Ensure the node is well provisioned. The resource requirements vary considerably based on the text contents of your repositories, but a good estimate is that the node should have enough memory to hold the entire text contents of the default branch of each repository.
