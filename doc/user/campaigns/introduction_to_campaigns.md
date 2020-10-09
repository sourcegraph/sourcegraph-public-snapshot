# Introduction to campaigns

<style>

img.screenshot {
    display: block;
    margin: 1em auto;
    max-width: 600px;
    margin-bottom: 0.5em;
    border: 1px solid lightgrey;
    border-radius: 10px;
}

</style>

## Overview

Campaigns let you make large-scale code changes across many repositories and code hosts. The campaign lets you create pull requests on all affected repositories, and it tracks their progress until they're all merged. You can preview the changes and update them at any time.

People usually use campaigns to make the following kinds of changes:

- Cleaning up common problems using linters.
- Updating uses of deprecated library APIs.
- Upgrading dependencies.
- Patching critical security issues.
- Standardizing build, configuration, and deployment files.

A campaign tracks all of its changesets for updates to:

- Status: open, merged, or closed
- Checks: passed (green), failed (red), or pending (yellow)
- Review status: approved, changes requested, pending, or other statuses (depending on your code host or code review tool)

You can see the overall trend of a campaign in the burndown chart, which shows the proportion of changesets that have been merged over time since the campaign was created.

<img src="campaign_tracking_sourcegraph_prs_burndown.png" class="screenshot">

In the list of changesets, you can see the detailed status for each changeset.

<img src="campaign_tracking_sourcegraph_prs.png" class="screenshot">

## Concepts

- A **campaign** is group of related changes to code, along with a title and description.
- {#campaign-spec} A **campaign spec** is a YAML file describing the campaign: repositories to change, commands to run, and a template for changesets and commits. You describe your high-level intent in the campaign spec, such as "lint files in all repositories with a `package.json` file".
- The campaign has associated **changesets**, which is a generic term for pull requests, merge requests, or any other reviewable chunk of code. (Code hosts use different terms for this, which is why we chose a generic term.)
- A **published changeset** means the commit, branch, and changeset have been created on the code host. An **unpublished changeset** is just a preview that you can view in the campaign but does not exist on the code host yet.
- A **spec** (campaign spec or changeset spec) is a "record of intent". When you provide a spec for a thing, the system will continuously try to reconcile the actual thing with your desired intent (as described by the spec). This involves creating, updating, and deleting things as needed.
- A campaign has many **changeset specs**, which are produced by executing the campaign spec (i.e., running the commands on each selected repository) and then using its changeset template to produce a list of changesets, including the diffs, commit messages, changeset title, and changeset body. You don't need to view or edit the raw changeset specs; you will edit the campaign spec and view the changesets in the UI.
- The **campaign controller** reconciles the actual state of the campaign's changesets on the code host so that they match your desired intent (as described in the changeset specs).

To learn about the internals of campaigns, see [Campaigns](../../dev/background-information/campaigns/index.md) in the developer documentation.
