# Search configuration

## Indexing

Sourcegraph can index the code in your repositories to speed up search; however, it requires additional resource usage so it is only enabled by default not running as a single docker image. You can enable indexed search when running as a single docker image by configuring [search.index](site_config/all.md#search-index-string-enum). If you have lots of code to index please ensure Sourcegraph is deployed to a well provisioned node.
