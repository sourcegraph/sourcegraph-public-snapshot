# Updating base images in Dockerfiles

In this tutorial we'll create a campaign that allows us to update base in
Dockerfiles across all of our repositories.

## Campaign spec

```yaml
name: update-golang-base-images
description: This campaign updates the golang builder images in Dockerfiles to Go 1.15.

on:
  # Find all repositories that contain Dockerfiles with `FROM golang:1.MINOR-alpine [...]` in it,
  # where the MINOR version can be 10 to 14.
  - repositoriesMatchingQuery: FROM golang:1.1:[minor~[0-4]]-alpine file:Dockerfile patternType:structural
  # and optionally specify a sha256 hash
  - repositoriesMatchingQuery: FROM golang:1.1:[minor~[0-4]]-alpine@sha256::[hash~[a-f0-9]+] file:Dockerfile patternType:structural

# In each repository
steps:
  # we use comby to update the base images with the sha256 suffix
  - run: |
      comby \
        -in-place \
        'FROM golang::[version]-alpine@sha256::[hash~[a-f0-9]+]' \
        'FROM golang:1.15-alpine@sha256:df0119b970c8e5e9f0f5c40f6b55edddf616bab2b911927ebc3b361c469ea29c' \
        Dockerfile
    container: comby/comby
  # and use comby to replace the ones without it:
  - run: |
      comby \
        -in-place \
        'FROM golang::[version]-alpine' \
        'FROM golang:1.15-alpine' \
        Dockerfile
    container: comby/comby

# Describe the changeset (e.g., GitHub pull request) you want for each repository.
changesetTemplate:
  title: Update golang base images in Dockerfiles to 1.15
  body: This updates golang base images used in Dockerfiles from golang:1.10 to 1.14 to use golang:1.15.
  branch: campaigns/golang-15-base-images # Push the commit to this branch.
  commit:
    message: Update golang base images in Dockerfiles to 1.15
  published:
    - github.com*: draft
```

## Instructions

1. Ensure that [you have write permissions for the repositories in which you want to change code](../quickstart.md#configure-code-host-connections).
1. [Install the Sourcegraph CLI](../quickstart.md#install-the-sourcegraph-cli) and use `src login https://YOUR-SOURCEGRAPH-INSTANCE` to configure it.
1. Save the campaign spec above as `YOUR_CAMPAIGN_SPEC.campaign.yaml`.
1. Create a campaign from the campaign spec by running the following [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) command:

    <pre><code>src campaign preview -f <em>YOUR_CAMPAIGN_SPEC.campaign.yaml</em> -namespace USERNAME_OR_ORG</code></pre>

1. Open the preview URL that the command printed out.
1. Examine the preview. Confirm that the changesets are the ones you intended to track. If not, edit the campaign spec and then rerun the command above.
1. Click the **Create campaign** button.
