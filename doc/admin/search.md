# Search configuration

## Indexing

Indexed search allows Sourcegraph to quickly text search all of your repositories. However, it requires additional resource usage so is only enabled by default when deploying via Kubernetes. For the single docker image it requires enabling via the configuration [search.index](site_config/all.md#search-index-string-enum). If you have lots of code to index please ensure Sourcegraph is deployed to a well provisioned node.
