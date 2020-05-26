# Updating a campaign

## Updating campaign attributes

In order to update the title and the description of a campaign, simply click the **Edit** button on the top right when viewing a campaign. You can then update both attributes. The branch name of a campaign can also be edited, but only if the campaign was created from a patch set and doesn't contain any published changesets.

When you click on **Save** and the campaign will be updated.

If the campaign was created from a patch set and includes changesets that have already been created on the code host, the title and description of those changesets will be updated on the code host, too.

## Updating the patch set of a campaign

You can also apply a new patch set to an existing campaign and update its patches and, if already created, the diff of the changesets on the code hosts.

1. Use the [`src` CLI](https://github.com/sourcegraph/src-cli) to create a new patch set that reflects the desired state of all patches/changesets in the campaign:

    ```
    $ src action exec -f new-action-definition.json -create-patchset

    # Or:

    $ src action exec -f new-action-definition.json | src campaign patchset create-from-patches
    ```

    For example, the `new-action-definition.json` could have a `"scopeQuery"` that yields _more_ repositories and thus produces _more_ patches.

2. Following the creation of the patch set with one of the two commands above, a URL will be printed that will guide you to the Sourcegraph web UI:

    <div style="max-width: 500px;" class="mx-auto">
      <figure class="figure">
        <div class="figure-img border">
        <img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/update_patchset.png" width="500px"/>
        </div>
      </figure>
    </div>

3. In the UI you can then select which campaign should be updated to use the new patch set.

    Click **Preview** to select the campaign that should be updated.

    <div style="max-width: 500px;" class="mx-auto">
      <figure class="figure">
        <div class="figure-img border">
        <img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/update_select_campaign.png" width="500px"/>
        </div>
        <figcaption class="figure-caption text-center">Select which campaign should be updated to use the new patch set.</figcaption>
      </figure>
    </div>

4. You'll then see which changesets will be created, updated (on the code host), deleted and removed from the campaign or left untouched:

    <div style="max-width: 500px;" class="mx-auto">
      <figure class="figure">
        <div class="figure-img border">
        <img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/update_preview_changes.png" width="500px"/>
        </div>
        <figcaption class="figure-caption text-center">Preview of updating a campaign's patch set.</figcaption>
      </figure>
    </div>

5. Click on **Update**.

The patches and changesets in a campaign will be like this:

* Patches (unpublished changesets) will be updated if their diff has changed.
* Published changesets will be updated on the code host if their diff or the campaign's title or description has changed.
* Published changesets will be closed on the code host and detached from the campaign if the new patch set doesn't contain a patch for their repositories.
* Published changesets will be left untouched if the new patch set contain the exact same patch for their repositories and the campaigns title and description have not been changed.
* Published changesets that are already merged or closed will not be updated and kept attached to the campaign. If a the patch set contains an new patch for a repository for which the campaign already has a merged changeset, a new changeset will be created.

### Example: Extending the scope of an campaign

A common reason for updating campaigns is to widen or narrow their scope, wanting more or fewer changesets to be created on a code host. In order to do that, one needs to update the patch set of an existing campaign with a patch set that contains the desired amount of patches.

Say you have successfully created a campaign based on the patches yielded by running the following action:

```json
{
  "scopeQuery": "repo:github.com/sourcegraph/sourcegraph$",
  "steps": [
    {
      "type": "docker",
      "image": "golang:1.14-alpine",
      "args": ["sh", "-c", "cd /work && go fmt ./..."]
    }
  ]
}
```

The campaign now has a single changeset, for the `github.com/sourcegraph/sourcegraph` repository, which was the only repository yielded by the `"scopeQuery"` above.

Now you want to run the same action over more repositories and extend the existing campaign by creating additional changesets on the code hosts.

To do that, first change the action definition's `"scopeQuery"` to yield more repositories:

```json
{
  "scopeQuery": "repo:github.com/sourcegraph",
  "steps": [
    {
      "type": "docker",
      "image": "golang:1.14-alpine",
      "args": ["sh", "-c", "cd /work && go fmt ./..."]
    }
  ]
}
```

This will now run the `go fmt` in every repository in the `github.com/sourcegraph` organization.

Execute the action and create a new patch set:

```
$ src action exec -f extended-action.json -create-patchset
```

After the command ran successfully, a URL to continue in the web UI is printed.

Open it to select which campaign you want to update. Select your existing campaign. The preview now shows you the additional changesets that will be created when you update the campaign and, if something changed in that repository, how the changeset that already exists will be updated.

Click **Update** to create the additional changesets.
