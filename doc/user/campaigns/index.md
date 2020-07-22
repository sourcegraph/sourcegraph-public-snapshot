# Campaigns

> NOTE: This documentation describes the current work-in-progress version of campaigns. [Click here](https://docs.sourcegraph.com/@3.18/user/campaigns) to read the documentation for campaigns in Sourcegraph 3.18.

Campaigns let you make large-scale code changes across many repositories.

## About campaigns

A campaign streamlines the creation and tracking of pull requests across many repositories and code hosts. After you create a campaign, you tell it what changes to make (by providing a list of repositories and a script to run in each). The campaign lets you create pull requests on all affected repositories, and it tracks their progress until they're all merged. You can preview the changes and update them at any time.

People usually use campaigns to make the following kinds of changes:

- Cleaning up common problems using linters
- Updating uses of deprecated library APIs
- Upgrading dependencies
- Patching critical security issues
- Standardizing build, configuration, and deployment files

For step-by-step instructions to create your first campaign, see [Hello World Campaign](hello_world_campaign.md) in Sourcegraph Guides.

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

The generic term **changeset** is used to refer to any of the following:

- GitHub pull requests
- Bitbucket Server pull requests
- Bitbucket Cloud pull requests (not yet supported)
- GitLab merge requests (not yet supported)
- Phabricator diffs (not yet supported)
- Gerrit changes (not yet supported)

A single campaign can span many repositories and many code hosts.

## Viewing campaigns

You can view a list of all campaigns by clicking the <img src="campaigns-icon.svg" alt="Campaigns icon" /> campaigns icon in the top navigation bar.

Use the filters to switch between showing all campaigns, open campaigns, or closed campaigns.

If you lack read access to a repository in a campaign, you can only see [limited information about the changes to that repository](managing_access.md#repository-permissions-for-campaigns) (and not the repository name, file paths, or diff).

### Campaign specs

You can create or update a campaign from a [campaign spec](#campaign-spec), which is a YAML file that defines a campaign.

See the "[Creating a campaign](#creating-a-campaign)" section for an example campaign spec YAML file.

For more information, see:

- [Creating a campaign](#creating-a-campaign) from a campaign spec
- [Updating a campaign](#updating-a-campaign) from a campaign spec
<!-- - TODO(sqs) <u>Campaign spec YAML reference</u> -->
- [Example campaign specs](examples/index.md)

## Creating a campaign

> **Creating your first campaign?** See [Hello World Campaign](hello_world_campaign.md) in Sourcegraph Guides for step-by-step instructions.

You can create a campaign from a [campaign spec](#campaign-spec), which is a YAML file that describes your campaign.

The following example campaign spec adds "Hello World" to all `README.md` files:

```yaml
name: hello-world
description: Add Hello World to READMEs

# Find all repositories that contain a README.md file.
on:
  - repositoriesMatchingQuery: file:README.md

# In each repository, run this command. Each repository's resulting diff is captured.
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    container: alpine:3

# Describe the changeset (e.g., GitHub pull request) you want for each repository.
changesetTemplate:
  title: Hello World
  body: My first campaign!
  branch: hello-world # Push the commit to this branch.
  commit:
    message: Append Hello World to all README.md files
  published: false
```

1. Create a campaign from the campaign spec by running the following [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) command:

    <pre><code>src campaign apply -f <em>YOUR_CAMPAIGN_SPEC.campaign.yaml</em> -preview</code></pre>

    > **Don't worry!** Before any branches are pushed or changesets (e.g., GitHub pull requests) are created, you will see a preview of all changes and can confirm each one before proceeding.
1. Wait for it to run and compute the changes for each repository (using the repositories and commands in the campaign spec).
1. Open the preview URL that the command printed out.
1. Examine the preview. Confirm that the changes are what you intended. (If not, edit the campaign spec and then rerun the command above.)
1. Click the **Create campaign** button.

After you've applied a campaign spec, you can [publish changesets](#publishing-changesets-to-the-code-host) to the code host when you're ready. This will turn the patches into commits, branches, and changesets (such as GitHub pull requests) for others to review and merge.

You can share the link to your campaign with other people if you want their help. Any person on your Sourcegraph instance can [view it in the campaigns list](#viewing-campaigns).

If a person viewing the campaign lacks read access to a repository in the campaign, they can only see [limited information about the changes to that repository](managing_access.md#repository-permissions-for-campaigns) (and not the repository name, file paths, or diff).

You can update a campaign's changes at any time, even after you've published changesets. For more information, see "[Updating a campaign](#updating-a-campaign)".

### Example campaigns

The [example campaigns](examples/index.md) show how to use campaigns to make useful, real-world changes:

- [Using ESLint to automatically migrate to a new TypeScript version](examples/eslint_typescript_version.md)
- [Adding a GitHub action to upload LSIF data to Sourcegraph](examples/lsif_action.md)
- [Refactoring Go code using Comby](examples/refactor_go_comby.md)

## Publishing changesets to the code host

After you've added patches, you can see a preview of the changesets (e.g., GitHub pull requests) that will be created from the patches. Publishing the changesets will, for each repository:

- Create a commit with the changes (from the patches for that repository)
- Push a branch (using the branch name you chose when creating the campaign)
- Create a changeset (e.g., GitHub pull request) on the code host for review and merging

When you're ready, you can publish some or all of a campaign's changesets.

<!-- > TODO(sqs): add steps for updating campaign spec's `changesetTemplate` to publish -->

You'll see a progress indicator when changesets are being published. Any errors will be shown, and you can retry publishing after you've resolved the problem. You don't need to worry about it creating multiple branches or pull requests when you retry, because it uses the same branch name.

To publish a changeset, you need admin access to the campaign and write access to the changeset's repository (on the code host). For more information, see "[Code host interactions in campaigns](managing_access.md#code-host-interactions-in-campaigns)". [Forking the repository](#known-issues) is not yet supported.

## Tracking campaign progress and changeset statuses

A campaign tracks all of its changesets for updates to:

- Status: open, merged, or closed
- Checks: passed (green), failed (red), or pending (yellow)
- Review status: approved, changes requested, pending, or other statuses (depending on your code host or code review tool)

You can see the overall trend of a campaign in the burndown chart, which shows the proportion of changesets that have been merged over time since the campaign was created.

<!-- > TODO(sqs) screenshot -->

In the list of changesets, you can see the detailed status for each changeset.

<!-- > TODO(sqs) screenshot -->

If you lack read access to a repository, you can only see [limited information about the changes to that repository](managing_access.md#repository-permissions-for-campaigns) (and not the repository name, file paths, or diff).

## Updating a campaign

<!-- TODO(sqs): needs wireframes/mocks -->

You can edit a campaign's name, description, and any other part of its campaign spec at any time.

To update a campaign, you need [admin access to the campaign](managing_access.md#campaign-access-for-each-permission-level), and [write access to all affected repositories](managing_access.md#repository-permissions-for-campaigns) with published changesets.

1. In your terminal, run the [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) command shown. The command will execute your campaign spec to generate changes and then upload them to the campaign for you to preview and accept.

    <pre><code>src campaign apply -f <em>YOUR_CAMPAIGN_SPEC.campaign.yaml</em> -preview</code></pre>

    > **Don't worry!** Before any branches or changesets are modified, you will see a preview of all changes and can confirm before proceeding.
1. Open the preview URL that the command printed out.
1. Examine the preview. Confirm that the changes are what you intended. (If not, edit your campaign spec and then rerun the command above.)
1. Click the **Update campaign** button.

All of the changesets on your code host will be updated to the desired state that was shown in the preview.

## Tracking existing changesets

<!-- TODO(sqs): needs wireframes/mocks -->

1. Click the <img src="campaigns-icon.svg" alt="Campaigns icon" /> campaigns icon in the top navigation bar.
1. *To use an existing campaign:* In the list of campaigns, click the campaign where you'd like to track existing changesets.

    *To create a new campaign:* Click the **＋ New campaign** button. For more information, see "[Creating a new campaign](#creating-a-new-campaign)".
1. Click the **Track existing changeset** button in the top right of the **Changesets** list.
1. Type in the name of the changeset's repository.

    This is the repository's name on Sourcegraph. If you can visit the repository at `https://sourcegraph.example.com/foo/bar`, the name is `foo/bar`. Depending on the configuration, it may or may not begin with a hostname (such as `github.com/foo/bar`).
1. Type in the changeset number (e.g., the GitHub pull request number).
1. Click **Add**. <!-- TODO(sqs): button label -->

You'll see the existing changeset in the list. The campaign will track the changeset's status and include it in the overall campaign progress (in the same way as if it had been created by the campaign). For more information, see "[Tracking campaign progress and changeset statuses](#tracking-campaign-progress-and-changeset-statuses)".

## Closing or deleting a campaign

You can close a campaign when you don't need it anymore, when all changes have been merged, or when you decided not to proceed with making all of the changes. A closed campaign still appears in the [campaigns list](#viewing-campaigns). To completely remove it, you can delete the campaign.

Any person with [admin access to the campaign](managing_access.md#permission-levels-for-campaigns) can close or delete it.

1. Click the <img src="campaigns-icon.svg" alt="Campaigns icon" /> campaigns icon in the top navigation bar.
1. In the list of campaigns, click the campaign that you'd like to close or delete.
1. In the top right, click the **Close**.
1. Select whether you want to close all of the campaign's changesets (e.g., closing all associated GitHub pull requests on the code host).
1. Click **TODO(sqs)** <!-- decide/confirm button label -->.

## [Managing access to campaigns](managing_access.md)

See "[Managing access to campaigns](managing_access.md)".

## Code host and repository permissions in campaigns

All actions on the code host (such as pushing a branch or opening a changeset) are performed by your individual user account, not by a bot user. For more information, see "[Code host interactions in campaigns](managing_access.md#code-host-interactions-in-campaigns)".

[Repository permissions](../../admin/repo/permissions.md) are enforced when campaigns display information. For more information, see "[Repository permissions in campaigns](managing_access.md#repository-permissions-for-campaigns)".

## Site admin configuration for campaigns

Using campaigns requires a [code host connection](../../admin/external_service/index.md) to a supported code host (currently GitHub and Bitbucket Server).

Site admins can also:

- [Allow users to authenticate via the code host](../../admin/auth/index.md#github), which makes it easier for users to authorize [code host interactions in campaigns](managing_access.md#code-host-interactions-in-campaigns)
- [Configure repository permissions](../../admin/repo/permissions.md), which campaigns will respect
- [Disable campaigns for all users](managing_access.md#disabling-campaigns-for-all-users)

## Concepts

- A **campaign** is group of related changes to code, along with a title and description.
- The campaign has associated **changesets**, which is a generic term for pull requests, merge requests, or any other reviewable chunk of code. (Code hosts use different terms for this, which is why we chose a generic term.)
- A **published changeset** means the commit, branch, and changeset have been created on the code host. An **unpublished changeset** is just a preview that you can view in the campaign but does not exist on the code host yet.
- A **spec** (campaign spec or changeset spec) is a "record of intent". When you provide a spec for a thing, the system will continuously try to reconcile the actual thing with your desired intent (as described by the spec). This involves creating, updating, and deleting things as needed.
- {#campaign-spec} A **campaign spec** is a YAML file describing the campaign: repositories to change, commands to run, and a template for changesets and commits. You describe your high-level intent in the campaign spec, such as "lint files in all repositories with a `package.json` file".
- A campaign has many **changeset specs**, which are produced by executing the campaign spec (i.e., running the commands on each selected repository) and then using its changeset template to produce a list of changesets, including the diffs, commit messages, changeset title, and changeset body. You don't need to view or edit the raw changeset specs; you will edit the campaign spec and view the changesets in the UI.
- The **campaign controller** reconciles the actual state of the campaign's changesets on the code host so that they match your desired intent (as described in the changeset specs).

To learn about the internals of campaigns, see "[Campaigns](../../dev/campaigns_development.md)" in the developer documentation.

## Roadmap

<!-- TODO(sqs): This section is rough/incomplete/outline-only. -->

### Known issues

<!-- TODO(sqs): This section is rough/incomplete/outline-only. -->

- Campaigns currently support **GitHub**, **GitLab** and **Bitbucket Server** repositories. If you're interested in using campaigns on other code hosts, [let us know](https://about.sourcegraph.com/contact).
- It is not yet possible for a campaign to have multiple changesets in a single repository (e.g., to make changes to multiple subtrees in a monorepo).
- Forking a repository and creating a pull request on the fork is not yet supported. Because of this limitation, you need write access to each repository that your campaign will change (in order to push a branch to it).
- Campaign steps are run locally (in the [Sourcegraph CLI](https://github.com/sourcegraph/src-cli)). Sourcegraph does not yet support executing campaign steps (which can be arbitrary commands) on the server. For this reason, the APIs for creating and updating a campaign require you to upload all of the changeset specs (which are produced by executing the campaign spec locally). {#server-execution}
