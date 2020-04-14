# Updating a campaign

You can also apply a new patch set to an existing campaign. Following the creation of the patch set with the `src campaign patchset create-from-patches` command, a URL will be output that will guide you to the web UI to allow you to change an existing campaign's patch set.

On this page, click "Preview" for the campaign that will be updated. From there, the delta of existing and new changesets will be displayed. Click "Update" to finalize the proposed changes.

Edits to the name and description of a campaign can also be made in the web UI with the changes reflected in each changeset. The branch name of a draft campaign with a patch set can also be edited, but only if the campaign doesn't contain any published changesets.
