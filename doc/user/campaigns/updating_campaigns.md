# Updating a campaign

## Updating campaign attributes

In order to update the title and the description of a campaign, simply click the **Edit** button on the top right when viewing a campaign. You can then update both attributes.

When you click on **Save** and the campaign will be updated.

If the campaign was created from a patch set and includes changesets that have already been created on the code host, the title and description of those changesets will be updated on the code host, too.

The branch name of a campaign can also be edited, but only if the campaign doesn't contain any published changesets.

## Updating the patch set of a campaign

You can also apply a new patch set to an existing campaign and update its patches and, if already created, the diff of the changesets on the code hosts.

To do that, you need to use the [`src` CLI](https://github.com/sourcegraph/src-cli) to create a new patch set that reflects the desired state of all patches/changesets in the campaign:

```
$ src action exec -f new-action-definition.json -create-patchset

# Or:

$ src action exec -f new-action-definition.json | src campaign patchset create-from-patches
```

For example, the `new-action-definition.json` could have a `"scopeQuery"` that yields _more_ repositories and thus produces _more_ patches.

Following the creation of the patch set with one of the two commands above, a URL will be printed that will guide you to the Sourcegraph web UI.

In the UI you can then select which campaign should be updated to use the new patch set:

<div style="max-width: 500px;" class="mx-auto">
  <figure class="figure">
    <div class="figure-img">
    <img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/update_select_campaign.png" width="500px"/>
    </div>
    <figcaption class="figure-caption text-center">Select which campaign should be updated to use the new patch set.</figcaption>
  </figure>
</div>

On this page, click **Preview** to select the campaign that should be updated.

You'll then see which changesets will be created, updated (on the code host), deleted and removed from the campaign or left untouched:

<div style="max-width: 500px;" class="mx-auto">
  <figure class="figure">
    <div class="figure-img">
    <img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/update_preview_changes.png" width="500px"/>
    </div>
    <figcaption class="figure-caption text-center">Preview of updating a campaign's patch set.</figcaption>
  </figure>
</div>

When you click on **Update** the patches and changesets in a campaign will be updated:

* Patches (unpublished changesets) will be updated if their diff has changed.
* Published changesets will be updated on the code host if their diff or the campaign's title or description has changed.
* Published changesets will be closed on the code host and detached from the campaign if the new patch set doesn't contain a patch for their repositories.
* Published changesets will be left untouched if the new patch set contain the exact same patch for their repositories and the campaigns title and description have not been changed.
* Published changesets that are already merged or closed will not be updated and kept attached to the campaign. If a the patch set contains an new patch for a repository for which the campaign already has a merged changeset, a new changeset will be created.
