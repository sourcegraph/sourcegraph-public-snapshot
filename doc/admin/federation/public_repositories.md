# Federation: redirecting to Sourcegraph.com for public repositories

Sourcegraph instances can be configured to redirect users to [Sourcegraph.com](https://sourcegraph.com) for public repositories. This offloads the work of mirroring, analyzing, and indexing public repositories from your Sourcegraph instance.

Example: If federation is enabled, a user on your Sourcegraph instance who tries to access a public repository at `https://sourcegraph.example.com/github.com/my/publicrepo` will be redirected to the same repository on Sourcegraph.com `https://sourcegraph.com/github.com/my/publicrepo`.

Regardless of whether federation is enabled, private repositories are always handled entirely by your own Sourcegraph instance.

### Benefits

Enabling federation has the following benefits:

- It offloads the work of mirroring, analyzing, and indexing public code to Sourcegraph.com, so your own instance's performance and resource consumption are unaffected.
- Users get full code navigation for all supported languages on Sourcegraph.com, even if your instance only has code navigation enabled for a subset of languages.
- Sourcegraph.com will show users more cross-repository references to code (via the "Find references" feature) than your instance because Sourcegraph.com's index is already very large. Building a comparable index of public code on your instance would require a lot of time and resources.
- It eliminates the risk of cloning and building public, untrusted code on your own instance (which may be running inside your private network). Sourcegraph.com applies strict isolation and resource quotas to mitigate this risk on our own infrastructure.

### Configuration

The `disablePublicRepoRedirects` [site configuration](../config/site_config.md) option disables redirection.
