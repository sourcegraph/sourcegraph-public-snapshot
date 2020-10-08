# Site configuration reference

All site configuration options and their default values are shown below.

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/reference/site.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/config/site_config) to see rendered content.</div>

## Known bugs

The following site configuration options require the server to be restarted for the changes to take effect:

```
auth.accessTokens
auth.sessionExpiry
git.cloneURLToRepositoryName
searchScopes
extensions
disablePublicRepoRedirects
```
