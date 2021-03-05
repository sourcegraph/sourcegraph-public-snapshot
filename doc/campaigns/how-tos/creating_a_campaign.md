# Creating a campaign

Campaigns are created by writing a [campaign spec](../references/campaign_spec_yaml_reference.md) and executing that campaign spec with the [Sourcegraph CLI](https://github.com/sourcegraph/src-cli) `src`.

## Requirements

- Sourcegraph instance with repositories in it. See the "[Quickstart](../../index.md#quickstart)" guide on how to setup a Sourcegraph instance.
- Installed and configured [Sourcegraph CLI](https://github.com/sourcegraph/src-cli) (see "[Install the Sourcegraph CLI](../quickstart.md#install-the-sourcegraph-cli)" in the campaigns quickstart for detailed instructions).
- Configured user credentials for the code host(s) that you'll be creating changesets on. See "[Configuring user credentials](configuring_user_credentials.md)" for a guide on how to add and manage your user credentials.

## Writing a campaign spec

In order to create a campaign, you need a campaign spec that describes the campaign. Here is an example campaign spec that describes a campaign to add "Hello World" to all `README.md` files:

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
  published: false # Do not publish any changes to the code hosts yet
```

See the ["Campaign spec YAML reference"](../references/campaign_spec_yaml_reference.md) and the [tutorials](../tutorials/index.md) for more details on how to write campaign specs.

## Creating a campaign after previewing

After writing a campaign spec you use the [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) to execute the campaign spec and upload it to Sourcegraph, where you can preview the changes and apply the campaign spec to create a campaign:

1. Run the following command in your terminal:

    <pre><code>src campaign preview -f <em>YOUR_CAMPAIGN_SPEC.campaign.yaml</em></code></pre>

    > **Don't worry!** Before any branches are pushed or changesets (e.g., GitHub pull requests) are created, you will see a preview of all changes and can confirm each one before proceeding.
    > NOTE: Campaigns's default behavior is to stop if computing changes in a repository errors. You can choose to ignore errors instead by adding the [`skip-errors`](../../cli/references/campaigns/preview.md) flag : `src campaign preview -f spec.campaign.yml -skip-errors`

1. Wait for it to run and compute the changes for each repository (using the repositories and commands in the campaign spec).
    <img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/src_campaign_preview_waiting.png" class="screenshot">
1. Open the preview URL that the command printed out.
    <img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/src_campaign_preview_link.png" class="screenshot">
1. Examine the preview. This is the result of executing the campaign spec. Confirm that the changes are what you intended. If not, edit the campaign spec and then rerun the command above.
    <img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/browser_campaign_preview.png" class="screenshot">
1. Click the **Apply spec** button to create the campaign.

After you've applied a campaign spec, you can [publish changesets](publishing_changesets.md) to the code host when you're ready. This will turn the patches into commits, branches, and changesets (such as GitHub pull requests) for others to review and merge.

You can share the link to your campaign with other people if you want their help. Any person on your Sourcegraph instance can [view it in the campaigns list](viewing_campaigns.md).

If a person viewing the campaign lacks read access to a repository in the campaign, they can only see [limited information about the changes to that repository](../explanations/permissions_in_campaigns.md#repository-permissions-for-campaigns) (and not the repository name, file paths, or diff).

You can update a campaign's changes at any time, even after you've published changesets. For more information, see [Updating a campaign](updating_a_campaign.md).

## Applying a campaign spec without preview

You can use [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) to directly apply a campaign spec to create or [update](updating_a_campaign.md) a campaign without having to use the UI.

Instead of running `src campaign preview` you run the following:

```bash
src campaign apply -f YOUR_CAMPAIGN_SPEC.campaign.yaml
```

This command won't print a link to a preview. It will create or update the campaign it describes directly.

That can be useful if you just want to update a single field in the campaign spec, i.e. the `description` or the `changesetTemplate.body`, or if you want to continously update a campaign by running `src` in a CI workflow.

## Creating a campaign in a different namespace

Campaigns are uniquely identified by their name and namespace. The namespace can be any Sourcegraph username or the name of a Sourcegraph organization.

By default, campaigns will use your username on Sourcegraph as your namespace. To create campaigns in a different namespace use the `-namespace` flag when previewing or applying a campaign spec:

```
src campaign preview -f your_campaign_spec.campaign.yaml -namespace USERNAME_OR_ORG
```
