# Updating base images in Dockerfiles

<style>

.markdown-body pre.chroma {
  font-size: 0.9em;
}

</style>

<p class="lead">
Create a campaign to update Dockerfiles in every one of your repositories.
</p>

### Introduction

Campaigns, combined with [comby](https://comby.dev) and [structural search](../../code_search/reference/structural.md), allow us to update the base images used in our Dockerfiles in just a few commands.

This tutorial shows you how to create [a campaign spec](../explanations/introduction_to_campaigns#campaign-spec) that

1. finds `Dockerfile`s that make use of `golang:1.x` base images and 
2. changes those `Dockerfiles` to use `golang:1.15`

The campaign spec and instructions here can easily be adapted to update other
base images.

### Prerequisites

We recommend that use the latest version of Sourcegraph when working with campaigns and that you have a basic understanding of how to create campaign specs and run them. See the following documents for more information:

1. ["Quickstart"](../quickstart.md)
1. ["Introduction to campaigns"](../explanations/introduction_to_campaigns.md)

### Create the campaign spec

Save the following campaign spec YAML as `YOUR_CAMPAIGN_SPEC.campaign.yaml`:

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
  - run: comby -in-place 'FROM golang::[version]-alpine' 'FROM golang:1.15-alpine' Dockerfile
    container: comby/comby

# Describe the changeset (e.g., GitHub pull request) you want for each repository.
changesetTemplate:
  title: Update golang base images in Dockerfiles to 1.15
  body: This updates golang base images used in Dockerfiles from golang:1.10 to 1.14 to use golang:1.15.
  branch: campaigns/golang-15-base-images # Push the commit to this branch.
  commit:
    message: Update golang base images in Dockerfiles to 1.15
  published: false
```

### Create the campaign

1. In your terminal, run this command:

    <pre>src campaign preview -f hello-world.campaign.yaml -namespace <em>USERNAME_OR_ORG</em></pre>

    > The `namespace` is either your Sourcegraph username or the name of a Sourcegraph organisation under which you want to create the campaign. If you're not sure what to choose, use your username.
1. Open the preview URL that the command printed out.
1. Examine the preview. Confirm that the changesets are the ones you intended to track. If not, edit the campaign spec and then rerun the command above.
1. Click the **Create campaign** button.

### Updating other base images

In order to update other base images in `Dockerfile`s you can either extend the campaign spec or create another campaign spec in which you change the `repositoriesMatchingQuery` and the `steps.run` properties.

Here's how you would update `alpine` base images from `alpine:3.9`, `alpine:3.10`, `alpine:3.11` to `alpine:3.12`:

```yaml
# [...]

on:
  - repositoriesMatchingQuery: FROM alpine:3.:[minor~[91]+[0-1]?] file:Dockerfile patternType:structural
  - repositoriesMatchingQuery: FROM alpine:3.:[minor~[91]+[0-1]?]@sha256::[hash~[a-f0-9]+] file:Dockerfile patternType:structural

steps:
  - run: |
      comby \
        -in-place \
        'FROM alpine:3.:[minor~[91]+[0-1]?]@sha256::[hash~[a-f0-9]+]' \
        'FROM alpine:3.12@ sha256:c0e9560cda118f9ec63ddefb4a173a2b2a0347082d7dff7dc14272e7841a5b5a' \
        Dockerfile
    container: comby/comby
  - run: comby -in-place 'alpine:3.:[minor~[91]+[0-1]?]' 'FROM alpine:3.12' Dockerfile
    container: comby/comby

# [...]
```
