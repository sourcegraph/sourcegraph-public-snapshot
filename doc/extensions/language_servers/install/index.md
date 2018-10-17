# Installing code intelligence

These instructions walk you through setting up [code intelligence](../index.md) for single-node `sourcegraph/server` deployments of Sourcegraph. For Sourcegraph cluster deployments on Kubernetes, see [cluster code intelligence instructions](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/install.md#add-language-servers-for-code-intelligence).)

## Automatic installation (default)

As of Sourcegraph 2.7, language servers are automatically started for supported languages that we detect in repositories that you have added.

If you do not wish to use code intelligence, you can disable language servers in the **Code intelligence** section of the site admin area.

## Manual Installation

If you do not want to give Sourcegraph access to the Docker socket, then you may [install code intelligence manually](index.md).

---

## Next steps

To get code intelligence on your code host and/or code review tool, see the [browser extension documentation](../../../integration/browser_extension.md).
