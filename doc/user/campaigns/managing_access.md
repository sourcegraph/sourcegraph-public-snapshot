# Managing access to campaigns

You can customize access to a campaign and propose changes to repositories with varying permission levels. Other people see the campaign's proposed changes to a repository if they can view that repository; otherwise, they can see only limited, non-identifying information about the change.

## Permission levels for campaigns

Any person with a user account can create a campaign.

The permission levels for a campaign are:

- **Read:** For people who need to view the campaign. 
- **Admin:** For people who need full access to the campaign, including editing, closing, and deleting it.

To see the campaign's proposed changes on a repository, a person *also* needs read access to that specific repository. Read or admin access on the campaign does not (by itself) entitle a person to viewing all of the campaign's changes. For more information, see "[Repository permissions for campaigns](#repository-permissions-for-campaigns)".

Site admins have admin permissions on all campaigns.

### Campaign access for each permission level

Campaign action | Read | Admin
--------------- | :--: | :----:
View campaign name and description<br/><small class="text-muted">Also applies to viewing the input branch name, created/updated date, and campaign status</small> | ⬤ | ⬤
View burndown chart (aggregate changeset statuses over time) | ⬤ | ⬤
View list of patches and changesets | ⬤ | ⬤
View diffstat (aggregate count of added/changed/deleted lines) | ⬤ | ⬤
View error messages (related to creating or syncing changesets) |  | ⬤
Edit campaign name, description, and branch name | ⬤ | ⬤
Update campaign patches (and changesets on code hosts) |  | ⬤
Publish changesets to code host |  | ⬤
Add/remove existing changesets to/from campaign |  | ⬤
Refresh changeset statuses |  | ⬤
Close campaign |  | ⬤
Delete campaign |  | ⬤

Authorization for all actions is also subject to [repository permissions](#repository-permissions-for-campaigns).

## Setting and viewing permissions for a campaign

When you create a campaign, you are given admin permissions on the campaign.

All users are automatically given read permissions to a campaign. Granular permissions are not yet supported. Assigning admin permissions on a campaign to another person or to all organization members is not yet supported. Transferring ownership of a campaign is not yet supported.

## Code host interactions in campaigns

All interactions with the code host are performed by your individual user account on the code host, not by a bot user or Sourcegraph machine account. These operations include:

- Pushing a branch with the changes (the Git author and committer will be you, and the Git push will be authenticated with your credentials)
- Creating a changeset (e.g., on GitHub, the pull request author will be you)
- Updating a changeset
- Closing a changeset

If you attempt to perform these operations and haven't yet linked your code host account to your Sourcegraph account, you'll be prompted to do so. You can either:

- Manually enter a personal access token in your user profile
- Sign in via your code host (currently only supported for GitHub; requires a site admin to [configure GitHub user authentication](../../admin/auth/index.md#github))

## Repository permissions for campaigns

Your [repository permissions](../../admin/repo/permission.md) determine what information in a campaign you can view. You can only see a campaign's proposed changes to a repository if you have read access to that repository. Read or admin permissions on a campaign does not (by itself) permit you to view all of the campaign's changes.

When you view a campaign, you can see a list of patches and changesets. For each patch and changeset:

- **If you have read access** to the repository for a patch or changeset, you can see the diff, changeset title, changeset link, detailed status, and other information.
- **If you do not have read access** to the repository for a patch or changeset, you can only see the status, last-updated date, and whether an error occurred (but not the error message). You can't see the diff, changeset title, changeset link, repository name, or any other information.

When you perform any campaign operation that involves repositories or code host interaction, your current repository permissions are taken into account.

- Creating, updating, or publishing a campaign; or publishing a single changeset: You must have access to push a branch to the repositories and create the changesets on your code host (e.g., push branches to and open pull requests on the GitHub repositories).
- Adding existing changesets to a campaign: You must have read access to the existing changesets' repository.
- Closing or deleting a campaign: If you choose to also close associated changesets on the code host, you must have access to do so on the code host. If you do not have access to close a changeset on the code host, the changeset will remain in its current state. A person with repository permissions for the remaining changesets can view them and manually close them.

Your repository permissions can change at any time. If you've already published a changeset to a repository that you no longer have access to, then you won't be able to view its details or update it in your campaign. The changeset on the code host will remain in its current state. A person with permissions to the changeset on the code host will need to manually manage or close it. When possible, you'll be informed of this when updating a campaign that contains changesets you've lost access to.

If you are not permitted to view a repository on Sourcegraph, then you won't be able to perform any operations on it, even if you are authorized on the code host.

## Disabling campaigns for all users

A site admin can completely disable campaigns on an instance by setting the [site configuration](../../admin/config/site_config.md) property `campaigns.disable` to `true`. This setting applies to everyone, including site admins. It prevents all viewing, creation, editing, syncing, and other actions on campaigns. Existing campaign changesets on code hosts are left in their current state and are not affected by this setting.
