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
2. **Code host configuration**: each package repository external service supports manually defining versions of packages to sync. See the page specific to the ecosystem you wish to configure. This method can be useful to verify that the credentials are picked up correctly without having to upload an index, as we'll as to poke holes in the filters (detailed below) if necessary.

## Filters

Package repository filters allow you to limit the package repositories and versions that are synced to the Sourcegraph instance. Using glob patterns, you can block or allow certain package repositories or versions on a per-ecosystem level.

There are two kinds of filters:

1. **Package name filters**: these match a glob pattern to only the package repository name, and will apply to all versions of any matched names.
2. **Package version filters**: these match a glob pattern to only the versions for a specific package repository. It cannot apply to multiple package repositories.

Filters can also have one of two behaviours:

1. **Block filters**: package repositories or versions matching these filters won't be synced.
2. **Allow filters**: only package repositories or versions matching these filters (that don't match a block filter) will be synced.

In other words, if there is a non-empty blocklist and an empty allowlist, anything not matching the blocklist will be synced. If there is a non-empty or empty blocklist and a non-empty allowlist, only package repositories matching the allowlist will be synced, if not matching the blocklist.
