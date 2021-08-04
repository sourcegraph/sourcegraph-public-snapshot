# Quickstart for Batch Changes

Get started and create your first [batch change](index.md) in 10 minutes or less.

## Introduction

In this guide, you'll create a Sourcegraph batch change that appends text to all `README.md` files in all of your repositories.

For more information about Batch Changes see the ["Batch Changes"](index.md) documentation and watch the [Batch Changes demo video](https://www.youtube.com/watch?v=EfKwKFzOs3E).

## Requirements

- A Sourcegraph instance with some repositories in it. See "[Quick install](../index.md#quick-install)" on how to set up a Sourcegraph instance.
- A local environment matching "[Requirements](./references/requirements.md)" to create batch changes with the Sourcegraph CLI.

## Install the Sourcegraph CLI

In order to create batch changes we need to [install the Sourcegraph CLI](../cli/index.md) (`src`).

1. Install the version of `src` that's compatible with your Sourcegraph instance:

    **macOS**:
    ```
    curl -L https://YOUR-SOURCEGRAPH-INSTANCE/.api/src-cli/src_darwin_amd64 -o /usr/local/bin/src
    chmod +x /usr/local/bin/src
    ```
    **Linux**:
    ```
    curl -L https://YOUR-SOURCEGRAPH-INSTANCE/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
    chmod +x /usr/local/bin/src
    ```
2. Authenticate `src` with your Sourcegraph instance by running **`src login`** and following the instructions:

    ```
    src login https://YOUR-SOURCEGRAPH-INSTANCE
    ```
    <img src="https://sourcegraphstatic.com/docs/images/batch_changes/src_login_success.png" class="screenshot">


Once `src login` reports that you're authenticated, we're ready for the next step.

## Write a batch spec

A **batch spec** is a YAML file that defines a batch change. It specifies which changes should be made in which repositories and how those should be published on the code host.

See the ["batch spec YAML reference"](references/batch_spec_yaml_reference.md) for details.

Save the following batch spec as `hello-world.batch.yaml`:

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
  body: My first batch change!
  branch: hello-world # Push the commit to this branch.
  commit:
    message: Append Hello World to all README.md files
  published: false
```

## Create the batch change

Let's see the changes that will be made. Don't worry---no commits, branches, or changesets will be published yet (the repositories on your code host will be untouched).

1. In your terminal, run this command:

    <pre>src batch preview -f hello-world.batch.yaml</pre>
1. Wait for it to run and compute the changes for each repository.
    <img src="https://sourcegraphstatic.com/docs/images/batch_changes/src_batch_preview_waiting.png" class="screenshot">
1. When it's done, click the displayed link to see all of the changes that will be made.
    <img src="https://sourcegraphstatic.com/docs/images/batch_changes/src_batch_preview_link.png" class="screenshot">
1. Make sure the changes look right.
    <img src="https://sourcegraphstatic.com/docs/images/batch_changes/browser_batch_preview.png" class="screenshot">
1. If you want to modify which changes are made, edit the `hello-world.batch.yaml` file, rerun the `src batch preview` command and open the newly generated preview URL.

    >NOTE: If you want to run the batch change on fewer repositories, change the `repositoriesMatchingQuery` in `hello-world.batch.yaml` to something like `file:README.md repo:myproject` (to only match repositories whose name contains `myproject`).
1. Click the **Apply spec** button to create the batch change. You should see something like this:
    <img src="https://sourcegraphstatic.com/docs/images/batch_changes/browser_batch_created.png" class="screenshot">

You created your first batch change! The batch change's changesets are still unpublished, which means they exist only on Sourcegraph and haven't been pushed to your code host yet.

## Publish the changes

So far, nothing has been created on the code hosts yet. For that to happen, we need to publish the changesets in our batch change.

Publishing causes commits, branches, and pull requests/merge requests to be created on your code host.

_You probably don't want to publish these toy "Hello World" changesets to actively developed repositories, because that might confuse people ("Why did you add this line to our READMEs?")._

### Configure code host credentials

Batch Changes needs permission to open changesets on your behalf. To grant permission, you will need to [add a personal access token](how-tos/configuring_credentials.md#adding-a-token) for each code host you'll be publishing changesets on.

This is a one-time operation that you don't need to do for each batch change. You can also ask the administrators of your Sourcegraph instance to [configure global credentials](how-tos/configuring_credentials.md#global-service-account-tokens) instead.

Once you have successfully added a token, Sourcegraph will have everything it needs to publish changesets to that code host!

### Publishing changesets

Now that you have credentials set up, you can publish the changesets in the batch change. On a real batch change, you would do the following:

1. Change the `published: false` in `hello-world.batch.yaml` to `published: true`.
    <img src="https://sourcegraphstatic.com/docs/images/batch_changes/batch_publish_true.png" class="screenshot">

    > NOTE: Change [`published` to an array](references/batch_spec_yaml_reference.md#publishing-only-specific-changesets) to publish only some of the changesets, or set [`'draft'` to create changesets as drafts on code hosts that support drafts](references/batch_spec_yaml_reference.md#changesettemplate-published).
1. Run the `src batch preview` command again and open the URL.
    <img src="https://sourcegraphstatic.com/docs/images/batch_changes/src_rerun_preview.png" class="screenshot">
1. On the preview page you can confirm that changesets will be published when the spec is applied.
    <img src="https://sourcegraphstatic.com/docs/images/batch_changes/browser_batch_preview_publish.png" class="screenshot">
1. Click the **Apply spec** button and those changesets will be published on the code host.
    <img src="https://sourcegraphstatic.com/docs/images/batch_changes/browser_batch_async.png" class="screenshot">

    > NOTE: You can also create or update a batch change by running `src batch apply`. This skips the preview stage, and is especially useful when updating an existing batch change.

> NOTE: You can also publish directly from Sourcegraph by omitting the `published` field from your batch spec. This is described in more detail in "[Publishing changesets to the code host](how-tos/publishing_changesets.md#publishing-changesets)".

## Congratulations!

You've created your first batch change! 🎉🎉

You can customize your batch spec and experiment with making other types of changes.

To update your batch change, edit `hello-world.batch.yaml` and run `src batch preview` again. (As before, you'll see a preview before any changes are applied.)

To learn what else you can do with Batch Changes, see "[Batch Changes](index.md)" in Sourcegraph documentation.
