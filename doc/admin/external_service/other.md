# Other Git repository hosts

Site admins can sync Git repositories on any Git repository host (by Git clone URL) with Sourcegraph so that users can search and navigate the repositories. Use this method only when your repository host is not named as a supported [external service](index.md).

To add Git repositories from any Git repository host:

1. Go to **User menu > Site admin**.
1. Open the **External services** page.
1. Press **+ Add external service**.
1. Enter a **Display name** (such as the human-readable name of the repository host).
1. In the **Kind** menu, select **Other**.
1. Configure the repositories to clone in the JSON editor. Use Cmd/Ctrl+Space for completion, and [see configuration documentation below](#configuration).
1. Press **Add external service**.

## Configuration

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/other_external_service.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/other) to see rendered content.</div>

