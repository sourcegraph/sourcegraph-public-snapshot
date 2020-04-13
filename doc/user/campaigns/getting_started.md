## Getting Started

1. Make sure that the Campaigns feature flag is enabled: [Configuration](./configuration.md)
1. Optional, but highly recommended for optimal syncing performance between your code host and Sourcegraph, setup the webhook integration:
  * GitHub: [Configuring GitHub webhooks](https://docs.sourcegraph.com/admin/external_service/github#webhooks).
  * Bitbucket Server: [Setup the `bitbucket-server-plugin`](https://github.com/sourcegraph/bitbucket-server-plugin), [create a webhook](https://github.com/sourcegraph/bitbucket-server-plugin/blob/master/src/main/java/com/sourcegraph/webhook/README.md#create) and configure the `"plugin"` settings for your [Bitbucket Server code host connection](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#configuration).
1. Setup the `src` CLI on your machine: [Installation and setup instructions](https://github.com/sourcegraph/src-cli/#installation)
1. Create your first campaign from a set of patches: [Creating a Campaign from Patches](./creating_campaign_from_patches.md)
1. Take a look at example campaigns: [Example campaigns](./examples/index.md)
