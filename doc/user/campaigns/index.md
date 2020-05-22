# Campaigns

Campaigns let you make large-scale code changes across many repositories.

> NOTE: Campaigns are in beta.

## About campaigns

A campaign streamlines the creation and tracking of pull requests across many repositories and code hosts. After you create a campaign, you tell it what changes to make (by providing a script that will run in each repository). The campaign lets you create pull requests on all affected repositories, and it tracks their progress until they're all merged. You can preview the changes and update them at any time.

People usually use campaigns to make the following kinds of changes:

- Cleaning up common problems using linters
- Updating uses of deprecated library APIs
- Upgrading dependencies
- Patching critical security issues
- Standardizing build, configuration, and deployment files

For step-by-step instructions to create your first campaign, see [Hello Universe Campaign](TODO) in Sourcegraph Guides.

<!-- TODO(sqs): link to about site for "why use campaigns?"

Why use campaigns?

With campaigns, making large-scale changes becomes:

- Simpler: Just provide a script and select the repositories.
- Easier to follow through on: You can track the progress of all pull requests, including checks and review statuses, to see where to help out and to confirm when everything's merged.
- Less scary: You can preview everything, roll out changes gradually, and update all changes even after creation.
- Collaborative: Other people can see all the changes, including those still in preview, in one place.

-->

<!-- TODO(sqs): Add video here, similar to https://www.youtube.com/aqcCrqRB17w (which will need to be updated for the new campaign flow). -->

## Supported code hosts and changeset types

A single campaign can span many repositories and many code hosts. The generic term **changeset** is used to refer to any of the following:

- GitHub pull requests
- Bitbucket Server pull requests
- Bitbucket Cloud pull requests (not yet supported)
- GitLab merge requests (not yet supported)
- Phabricator diffs (not yet supported)
- Gerrit changes (not yet supported)

## Viewing campaigns

You can view a list of all campaigns by clicking the <img src="campaigns-icon.svg" alt="Campaigns icon" /> campaigns icon in the top navigation bar.

Use the filters to switch between showing all campaigns, open campaigns, or closed campaigns.

If you lack read access to a repository in a campaign, you can only see [limited information about the changes to that repository](managing_access.md#repository-permissions-for-campaigns) (and not the repository name, file paths, or diff).

## Creating a new campaign

> **Creating your first campaign?** See [Hello Universe Campaign](TODO) in Sourcegraph Guides for step-by-step instructions.

1. Click the <img src="campaigns-icon.svg" alt="Campaigns icon" /> campaigns icon in the top navigation bar.
1. Click the **＋ New campaign** button.
1. Type a name for your campaign. The name will be the title of each changeset (e.g., the pull request title).
1. Type an optional description, which will be the description or body of each changeset.
1. Choose a branch name (or use the suggested one). This is the branch on each repository where the campaign's changes will be pushed to.
1. Click the **Create campaign** button.

You've created a new campaign, but it doesn't have any changes yet. Next, you can [generate and upload patches to specify what changes to make](#generating-and-uploading-patches).

If the changesets were already created (outside of campaigns), you can [track existing changesets](#tracking-existing-changesets) in your campaign.

## Generating and uploading patches

After you've [created a campaign](#creating-a-new-campaign), you tell it what changes to make by uploading a list of patches. A patch is a change (in diff format) to a specific repository on a specific branch.

> **Don't worry!** Before any branches are pushed or changesets (e.g., GitHub pull requests) are created, you will see a preview of all changes and can confirm each one before it's published.

1. In your editor, create a [campaign action](actions.md) file, which defines the set of repositories to change and the commands to run in each repository to make the changes.
1. Click the <img src="campaigns-icon.svg" alt="Campaigns icon" /> campaigns icon in the top navigation bar.
1. In the list of campaigns, click the campaign where you'd like to upload patches.
1. In the campaign, click the **Upload patches** button.
1. In your terminal, run the command shown. The command will execute your [campaign action](actions.md) to generate patches and then upload them to the campaign for you to preview and accept.

    > You need [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) for this step. For reference, the command shown by the campaign is the following (with <code><em>CAMPAIGN-ID</em></code> filled in):
   <pre><code>src actions exec -f action.json | src campaign set-patches -preview -id=<em>CAMPAIGN-ID</em></code></pre>
1. Open the preview URL that the command printed out.
1. Examine the preview. Confirm that the patches' diffs are what you intended.
   
    (If not, edit your campaign action and then rerun the command above. Old and unused previews are automatically discarded, so you don't need to manually delete this preview.)
1. Click the **Update campaign** button.

After you've added patches, you can [publish changesets](#publishing-changesets-to-the-code-host) to the code host when you're ready. This will turn the patches into commits, branches, and changesets (such as GitHub pull requests) for others to review and merge.

You can share the link to your campaign with other people if you want their help. Any person on your Sourcegraph instance can [view it in the campaigns list](#viewing-campaigns).

If a person viewing the campaign lacks read access to a repository in the campaign, they can only see [limited information about the changes to that repository](managing_access.md#repository-permissions-for-campaigns) (and not the repository name, file paths, or diff).

You can update a campaign's changes at any time, even after you've published changesets. For more information, see "[Updating a campaign](#updating-a-campaign)".

### Example campaigns

The [example campaigns](examples/index.md) show how to use campaigns to make useful, real-world changes:

* [Using ESLint to automatically migrate to a new TypeScript version](examples/eslint_typescript_version.md)
* [Adding a GitHub action to upload LSIF data to Sourcegraph](examples/lsif_action.md)
* [Refactoring Go code using Comby](examples/refactor_go_comby.md)

## Publishing changesets to the code host

After you've added patches, you can see a preview of the changesets (e.g., GitHub pull requests) that will be created from the patches. Publishing the changesets will, for each repository:

- Create a commit with the changes (from the patches for that repository)
- Push a branch (using the branch name you chose when creating the campaign)
- Create a changeset (e.g., GitHub pull request) on the code host for review and merging

When you're ready, you can publish all of a campaign's changesets, or just an individual changeset.

1. Click the <img src="campaigns-icon.svg" alt="Campaigns icon" /> campaigns icon in the top navigation bar.
1. In the list of campaigns, click the campaign that has changesets you'd like to publish.
1. Click the **Publish** button next to a preview changeset to publish it.
1. To publish all changesets (that have not already been published), click the **Publish all** button. If all changesets are already published, this button is not shown.

You'll see a progress indicator when changesets are being published. Any errors will be shown, and you can retry publishing after you've resolved the problem. You don't need to worry about it creating multiple branches or pull requests when you retry, because it uses the same branch name.

To publish a changeset, you need admin access to the campaign and write access to the changeset's repository (on the code host). For more information, see "[Code host interactions in campaigns](managing_access.md#code-host-interactions-in-campaigns)". [Forking the repository](#known-issues) is not yet supported.

## Tracking campaign progress and changeset statuses

A campaign tracks all of its changesets for updates to:

- Status: open, merged, or closed
- Checks: passed (green), failed (red), or pending (yellow)
- Review status: approved, changes requested, pending, or other statuses (depending on your code host or code review tool)

You can see the overall trend of a campaign in the burndown chart, which shows the proportion of changesets that have been merged over time since the campaign was created.

> TODO(sqs) screenshot

In the list of changesets, you can see the detailed status for each changeset.

> TODO(sqs) screenshot

If you lack read access to a repository, you can only see [limited information about the changes to that repository](managing_access.md#repository-permissions-for-campaigns) (and not the repository name, file paths, or diff).

## Updating a campaign

<!-- TODO(sqs): needs wireframes/mocks -->

You can edit a campaign's name and description, and upload new patches, at any time. If you haven't yet published any changesets, you can also choose a different branch name for the campaign.

To update a campaign, you need [admin access to the campaign](managing_access.md#campaign-access-for-each-permission-level), and [write access to all affected repositories](managing_access.md#repository-permissions-for-campaigns) with published changesets.

1. Click the <img src="campaigns-icon.svg" alt="Campaigns icon" /> campaigns icon in the top navigation bar.
1. In the list of campaigns, click the campaign that you want to edit.
<!-- TODO(sqs): needs to handle edit name case as well, not just upload patches -->
1. In the campaign, click the ***Upload patches** button.
1. In your terminal, run the command shown. The command will execute your [campaign action](actions.md) to generate patches and then upload them to the campaign for you to preview and accept.

    > You need [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) for this step. For reference, the command shown by the campaign is the following (with <code><em>CAMPAIGN-ID</em></code> filled in):
   <pre><code>src actions exec -f action.json | src campaign set-patches -preview -id=<em>CAMPAIGN-ID</em></code></pre>
1. Open the preview URL that the command printed out.
1. Examine the preview. Confirm that the patches' diffs are what you intended.
   
    (If not, edit your campaign action and then rerun the command above. Old and unused previews are automatically discarded, so you don't need to manually delete this preview.)
1. Click the **Update campaign** button.

1. In your editor, create a [campaign action](actions.md) file, which defines the set of repositories to change and the commands to run in each repository to make the changes.


## Tracking existing changesets

<!-- TODO(sqs): needs wireframes/mocks -->

1. Click the <img src="campaigns-icon.svg" alt="Campaigns icon" /> campaigns icon in the top navigation bar.
1. *To use an existing campaign:* In the list of campaigns, click the campaign where you'd like to track existing changesets.

    *To create a new campaign:* Click the **＋ New campaign** button. For more information, see "[Creating a new campaign](#creating-a-new-campaign)".
1. On the right side of the **Changesets** list, use the **＋** menu and select **Existing changeset**.
1. Type in the name of the changeset's repository.

    This is the repository's name on Sourcegraph. If you can visit the repository at `https://sourcegraph.example.com/foo/bar`, the name is `foo/bar`. Depending on the configuration, it may or may not begin with a hostname (such as `github.com/foo/bar`).
1. Type in the changeset number (e.g., the GitHub pull request number).
1. Click **Add**.

You'll see the existing changeset in the list. The campaign will track the changeset's status and include it in the overall campaign progress (in the same way as if it had been created by the campaign). For more information, see "[Tracking campaign progress and changeset statuses](#tracking-campaign-progress-and-changeset-statuses)".

## Closing or deleting a campaign

You can close a campaign when you don't need it anymore, when all changes have been merged, or when you decided not to proceed with making all of the changes. A closed campaign still appears in the [campaigns list](#viewing-campaigns). To completely remove it, you can delete the campaign.

Any person with [admin access to the campaign](managing_access.md#permission-levels-for-campaigns) can close or delete it.

1. Click the <img src="campaigns-icon.svg" alt="Campaigns icon" /> campaigns icon in the top navigation bar.
1. In the list of campaigns, click the campaign that you'd like to close or delete.
1. In the top right, click the **Close** or **Delete** button.
1. Select whether you want to close all of the campaign's changesets (e.g., closing all associated GitHub pull requests on the code host).
1. Click **TODO(sqs)** <!-- decide/confirm button label -->.

## [Managing access to campaigns](managing_access.md)

See "[Managing access to campaigns](managing_access.md)".

## Code host and repository permissions in campaigns

All actions on the code host (such as pushing a branch or opening a changeset) are performed by your individual user account, not by a bot user. For more information, see "[Code host interactions in campaigns](managing_access.md#code-host-interactions-in-campaigns)".

[Repository permissions](../../admin/repo/permission.md) are enforced when campaigns display information. For more information, see "[Repository permissions in campaigns](managing_access.md#repository-permissions-for-campaigns)".

## Site admin configuration for campaigns

Using campaigns requires a [code host connection](../../admin/external_service/index.md) to a supported code host (currently GitHub and Bitbucket Server).

Site admins can also:

- [Allow users to authenticate via the code host](../../admin/auth/index.md#github), which makes it easier for users to authorize [code host interactions in campaigns](managing_access.md#code-host-interactions-in-campaigns)
- [Configure repository permissions](../../admin/repo/permission.md), which campaigns will respect
- [Disable campaigns for all users](managing_access.md#disabling-campaigns-for-all-users)

## Concepts

- A **campaign** is group of related changes to code, along with a title and description.
- You supply a set of **patches** to a campaign. Each patch is a unified diff describing changes to a specific commit and branch in a repository. (To produce the patches, you provide a script that runs in the root of each repository and changes files.)
- The campaign converts the patches into **changesets**, which is a generic term for pull requests, merge requests, or any other reviewable chunk of code. (Code hosts use different terms for this, which is why we chose a generic term.)
- Initially a changeset is just a **preview** and is not actually pushed to or created on the code host.
- You **publish** a changeset when you're ready to push the branch and create the changeset on the code host.

## Roadmap

<!-- TODO(sqs): This section is rough/incomplete/outline-only. -->

### Known issues

<!-- TODO(sqs): This section is rough/incomplete/outline-only. -->

- The only supported code hosts are GitHub and Bitbucket Server. Support for [all other code hosts](../../admin/external_service/index.md) is planned.
- It is not yet possible for a campaign to have multiple changesets in a single repository (e.g., to make changes to multiple subtrees in a monorepo).
- Forking a repository and creating a pull request on the fork is not yet supported. Because of this limitation, you need write access to each repository that your campaign will change (in order to push a branch to it).
