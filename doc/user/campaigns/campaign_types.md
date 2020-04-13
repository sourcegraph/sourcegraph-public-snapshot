# Campaign Types

There are two types of campaigns:

### Campaigns created from a set of patches

When a Campaign is created from a set of patches, one per repository, Sourcegraph will create changesets (pull requests) on the associated code hosts and track their progress in the newly created campaign, where you can manage them.

With the `src` CLI tool, you can not only create the campaign from an existing set of patches, but you can also _generate the patches_ for a number of repositories.

### Manual campaigns

Manual campaigns provide the ability to manage and monitor changesets (pull requests) that already exist on code hosts by manually adding them to a campaign.
