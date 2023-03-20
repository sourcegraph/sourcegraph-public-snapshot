# Package repositories

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This feature is experimental and might change or be removed in the future. We've released it as an experimental feature to provide a preview of functionality we're working on.
</p>
</aside>

Sourcegraph package repos can sync dependency sources from dependency artifact hosts such as Rust crates, JVM libraries, Node.js packages, Ruby gems, and more.

## Enable package repositories

Package repositories can be enabled on a per-ecosystem level in your [site configuration](/admin/config/site_config), for example:

```json
{
  // ...
  "experimentalFeatures": {
    "jvmPackages": "enabled"
  }
  // ...
}
```

## Repository syncing

There are generally two ways of syncing package repositories to the Sourcegraph instance.

1. **Indexing** (recommended): package repositories can be added to the Sourcegraph instance when they are referenced in [code graph data uploads](/code_navigation/explanations/uploads). Code graph indexers can derive package repository references during indexing, which will be used on upload to sync them to your instance.
2. **Code host configuration**: each package repository external service supports manually defining versions of packages to sync. See the page specific to the ecosystem you wish to configure. This method can be useful to verify that the credentials are picked up correctly without having to upload an index.

## Filters

bitch banana
