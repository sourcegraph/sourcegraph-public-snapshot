# Creating multiple changesets in large repositories

<style>
.markdown-body h2 { margin-top: 50px; }
.markdown-body pre.chroma { font-size: 0.75em; }
</style>

<aside class="experimental">
<span class="badge badge-experimental">Experimental</span> This feature is experimental and might change or be removed in the future. We've released it as an experimental feature to provide a preview of functionality we're working on. We're very much looking for feedback.
<br />
It's available in Sourcegraph 3.23 with <a href="https://github.com/sourcegraph/src-cli">Sourcegraph CLI</a> 3.23.0 and later.
</aside>

Campaigns can produce a lot of changes in a single repository and it might make sense to split up the changes into multiple changesets, so that each changeset can be reviewed and merged independently.

By using [`transformChanges`](../references/campaign_spec_yaml_reference.md#transformchanges) in your campaign spec you can group the changes produced in one repository by directory and create a changeset for each group:

The following campaign spec uses the `transformChanges` property to create multiple changesets in a single repository by grouping the changes made in different directories:

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

# Transform the changes produced in each repository.
transformChanges:
  # Group the file diffs by directory and produce one additional changeset per group.
  # Changes that haven't been grouped will be be in the standard changeset.
  group:
    - directory: client
      branch: hello-world-client # will replace the `branch` in the `changesetTemplate`
    - directory: docker-images
      repository: github.com/sourcegraph/sourcegraph # Optional: only apply the rule in this repository
      branch: hello-world-infra
    - directory: monitoring
      repository: github.com/sourcegraph/sourcegraph
      branch: hello-world-monitoring

# Describe the changeset (e.g., GitHub pull request) you want for each repository.
changesetTemplate:
  title: Hello World
  body: My first campaign!
  branch: hello-world # This branch is the default branch and will be
                      # overwritten for each additional changeset.
  commit:
    message: Append Hello World to all README.md files
  published: false # Do not publish any changes to the code hosts yet
```

This campaign spec will produce up to 4 changesets in the `github.com/sourcegraph/sourcegraph` repository: one changeset with all the changes in the `client` directory, one changeset for the changes in `docker-images`, one changeset for changes in `monitoring`, and one changeset for the changes in the other directories.

Since code hosts and git don't allow creating multiple, _different_ changesets on the same branch, we need to specify the `branch` that will be used for the additional changesets. That `branch` will overwrite the default branch specified in `changesetTemplate`.

In case no changes have been made in a `directory` specified in a `group`, no additional changeset will be produced.

If the optional `repository` property is specified only the changes in that repository will be grouped.
