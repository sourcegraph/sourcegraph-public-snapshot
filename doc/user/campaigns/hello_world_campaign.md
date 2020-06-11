# Hello World Campaign

> TODO(sqs): This will eventually go in a new "Sourcegraph Guides" docs section, but it lives here for now.

Have you ever needed to make the same kind of change to many repositories at once? Campaigns make this much easier. To get you started, let's run a very simple campaign: adding the line `Hello World` to all of your repositories' `README.md` files. After completing this exercise, you'll be able to create your own campaigns to make useful changes, fixes, refactors, and more.

You'll learn how to:

- Create a new campaign
- Select the repositories to work with
- Write the script to make changes to each repository
- Publish and monitor the status of your proposed changes

For more detailed information, see "[Campaigns](index.md)" in Sourcegraph documentation.

## What is a campaign?

A campaign lets you make many related code changes, creating many branches and changesets (such as GitHub pull requests) across many repositories. You can track the progress as they are reviewed and merged. See "[About campaigns](index.md#about-campaigns)" for more information.

To use campaigns, you need to:

- [Set up a Sourcegraph instance](../../index.md#quickstart) and add some repositories to it.
- [Install Sourcegraph CLI](https://github.com/sourcegraph/src-cli) (`src`).

## Step 1. Create a new campaign

<!-- TODO(sqs): keep these steps in sync with index.md#creating-a-new-campaign -->

1. Click the <img src="campaigns-icon.svg" alt="Campaigns icon" /> campaigns icon in the top navigation bar on Sourcegraph.
1. Click the **＋ New campaign** button.
1. Name your campaign `Hello World`.
1. In the description, write `My first campaign!`.
1. Click the **Create campaign** button.

Now you have a campaign, but it's empty. We'll create a campaign template to say what changes to make (and to which repositories).

## Step 2. Make a campaign template

A **campaign template** defines:

- The set of repositories to change
- Commands to run in each repository to make the changes

Save the following campaign template as `hello-world.campaign.yaml`:

``` yaml
# Find all repositories that contain a README.md file.
roots:
  - query: file:README.md

# In each repository, run this command. Each repository's resulting diff is captured.
steps:
  - name: Append "Hello World" to all README.md files
    run: echo Hello World | tee -a $(find -name README.md)
    container: alpine:3

# The name of the branch where the commit will be pushed.
branch: hello-world
```

## Step 3. Generate and apply the changes

Let's see the changes that will be made. Don't worry---no commits, branches, or changesets will be published yet (the repositories on your code host will be untouched).

1. In your terminal, run this command (replacing <code><em>CAMPAIGN-ID</em></code> with the ID of the campaign you created):

    <pre><code>src campaign apply -template=hello-world.campaign.yml -preview -campaign=<em>CAMPAIGN-ID</em></code></pre>
1. Wait for it to run and capture the changes in each repository.
1. When it's done, click the displayed link to see all of the changes that will be made.
1. Make sure the changes look right.

    > If you want to run the campaign on fewer repositories, change the roots query in `hello-world.campaign.yml` to something like `file:README.md repo:myproject` (to only match repositories whose name contains `myproject`).
1. Click the **Apply** button.

Your campaign now has changesets! They are still unpublished, which means they exist only on Sourcegraph and haven't been pushed to your code host yet.

## Step 4. Publish the changes (optional)

In a real campaign, the next step is publishing, which causes commits, branches, and changesets to be created on your code host. You probably don't want to publish these toy "Hello World" changesets to actively developed repositories, because that might confuse people ("Why did you add this line to our READMEs?").

## Congratulations!

You've created your first campaign! 🎉🎉

Now, you can customize your campaign template and experiment with making other types of changes. Here are some [example campaigns](examples/index.md) to help:

- [Using ESLint to automatically migrate to a new TypeScript version](examples/eslint_typescript_version.md)
- [Adding a GitHub action to upload LSIF data to Sourcegraph](examples/lsif_action.md)
- [Refactoring Go code using Comby](examples/refactor_go_comby.md)

To learn what else you can do with campaigns, see "[Campaigns](index.md)" in Sourcegraph documentation.
