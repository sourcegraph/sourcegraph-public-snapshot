# (advanced) Loading configuration via the file system or Kubernetes ConfigMap

In teams where Sourcegraph is a critical piece of infrastructure, it can often be desirable to check the Sourcegraph configuration into version control.

As of Sourcegraph v3.4+, this is possible for [site configuration](site_config.md), [critical configuration](critical_config.md), and [external services configuration]()

## Benefits

1. Configuration can be checked into version control (e.g. Git).
2. Edits through the web UI cannot be saved (good for enforcing configuration organization-wide).

## Drawbacks

Loading configuration in this manner has two important drawbacks:

1. You will no longer be able to save configuration edits through the web UI (you can use the web UI as scratch space, though).
2. Sourcegraph sometimes performs automatic migrations of configuration when upgrading versions. This process will now be more manual for you (see below).

## Getting started

Simply add the relevant environment variable below to all `frontend` containers (cluster deployment) or to the `server` container (single-container Docker deployment):

```sh
CRITICAL_CONFIG_FILE=critical.json
SITE_CONFIG_FILE=site.json
EXTSVC_CONFIG_FILE=extsvc.json
```

You should also add to the `management-console` container (cluster deployment) or to the `server` container (single-container Docker deployment) the following:

```sh
DISABLE_CONFIG_UPDATES=true
```

- `critical.json` is literally the [management console](../management_console.md) configuration.
- `site.json` is literally the [site configuration](site_config.md)
- `extsvc.json` is _all_ of your external services in a single JSONC file like so:

```jsonc

{
  "GITHUB": [
    {
      // First GitHub external service configuration
      "authorization": {},
      "url": "https://github.com",
      "token": "...",
      "repositoryQuery": ["affiliated"]
    },
    {
      // Second GitHub external service configuration
      "authorization": {},
      "url": "https://github.com",
      "token": "...",
      "repositoryQuery": ["affiliated"]
    },
  ],
  "PHABRICATOR": [
    {
      // First Phabricator external service configuration
    },
  ]
}
```

You can find a full list of [valid top-level keys here](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@b7ebb9024e3a95109fdedfb8057795b9a7c638bc/-/blob/cmd/frontend/graphqlbackend/schema.graphql#L1104-1110).

## Upgrades & Migrations

As mentioned earlier, when configuration is loaded via this manner Sourcegraph can no longer persist the automatic migrations to configuration it sometimes performs on upgrades.

It will still perform such migrations on the configuration loaded from file, it just cannot persist such migrations _back to file_ and we only guarantee such migrations stick around for two minor versions.

When you upgrade Sourcegraph versions, to ensure your configurations do not become invalid, you should:

1. Upgrade Sourcegraph to the new version
2. Visit each configuration page in the web UI (management console, site configuration, each external service)
3. Copy the (now migrated) configuration from those pages into your JSON files.

We're planning to improve this by having Sourcegraph notify you as a site admin when you should do the above, since today it is not actually required in most upgrades. See https://github.com/sourcegraph/sourcegraph/issues/4650 for details.

## Kubernetes ConfigMap

Currently, site admins are responsible for creating the ConfigMap resource that maps the above environment variables to refer to the ConfigMap'd files on disk.

(If you need assistance with this, please contact us.)
