# Gitolite

Site admins can link and sync Git repositories on [Gitolite](https://gitolite.com) with Sourcegraph so that users can search and navigate the repositories.

To set this up, add Gitolite as an external service to Sourcegraph:

1. Go to **User menu > Site admin**.
1. Open the **External services** page.
1. Press **+ Add external service**.
1. Enter a **Display name** (using "Gitolite" is OK if you only have one Gitolite instance).
1. In the **Kind** menu, select **Gitolite**.
1. Configure the connection to Gitolite in the JSON editor. Use Cmd/Ctrl+Space for completion, and [see configuration documentation below](#configuration).
1. Press **Add external service**.

## Configuration

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/gitolite.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/gitolite) to see rendered content.</div>
