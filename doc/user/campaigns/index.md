# Campaigns

>NOTE: **Campaigns are currently in beta.** We're actively building out the feature set and improving the user experience with every update. Let us know what you think! [File an issue](https://github.com/sourcegraph/sourcegraph) with feedback/problems/questions, or [contact us directly](https://about.sourcegraph.com/contact).

## What are campaigns?

Campaigns are a part of [Sourcegraph code change management](https://about.sourcegraph.com/product/code-change-management) that allow you to make large-scale code changes across many repositories and different code hosts, and monitor their progress.

<div class="text-center">
  <iframe
      width="560"
      height="315"
      src="https://www.youtube.com/embed/aqcCrqRB17w"
      frameborder="0"
      allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture"
      allowfullscreen="true"
  ></iframe>
</div>

## Are you a first time user of campaigns?

If you are a first-time user of campaigns, we recommend that you read through the following sections of the documentation:

1. Read through the **[How it works](#how-it-works)** section below and **watch the video** to get an understanding of how campaigns work.
1. Go through the "[Getting started](./getting_started.md)" instructions to setup your Sourcegraph instance for campaigns.
1. Create your first campaign from a set of patches by reading "[Creating a campaign from patches](./creating_campaign_from_patches.md)".
1. Create a manual campaign to track the progress of existing pull requests on your code host by reading "[Creating a manual campaign](./creating_manual_campaign.md)".

At this point you're ready to explore the [**example campaigns**](./examples/index.md) and [create your own action definitions](./actions.md) and campaigns.

## When should I use campaigns?

You should use campaigns if you want to:

* run code to make changes across a large number of repositories.
* keep track of a large number of pull requests and their status on GitHub or Bitbucket Server instances.
* execute commands to upgrade dependencies in multiple repositories.
* use Sourcegraph's search and replace matches by running code in the matched repositories.

Campaigns allow you to **leverage Sourcegraph's search powers** and **execute code and Docker containers in all the repositories** yielded by a single Sourcegraph search query.

The created set of patches can then be turned into multiple **changesets** (a generic name for what some code hosts call _pull requests_ and others _merge requests_) on different code hosts by creating a **campaign**.

<div style="max-width: 300px;" class="float-none float-xl-right ml-xl-3 mx-auto">
  <figure class="figure">
    <div class="figure-img">
      <a href="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/manual_campaign.png">
        <img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/manual_campaign.png" />
      </a>
    </div>
    <figcaption class="figure-caption text-right">A campaign tracking multiple changesets in different repositories.</figcaption>
  </figure>
</div>

Once the campaign is created, you can track the **review state, CI status,** and **open/closed/merged lifecycle** of each changeset in the Sourcegraph UI.

<div class="clearfix"></div>

## How it works

See this video for a demonstration of lifecycle of a campaign:

<div style="max-width: 450px;" class="mx-auto">
  <figure class="figure">
    <div class="figure-img">
      <iframe src="https://player.vimeo.com/video/398878670?color=0CB6F4&title=0&byline=0&portrait=0" style="max-height: 250px; width:100%;height:100%;" frameborder="0" webkitallowfullscreen mozallowfullscreen allowfullscreen></iframe>
    </div>
    <figcaption class="figure-caption text-right">Campaign: Running <code>gofmt</code> in each repository containing a <code>go.mod</code> file.</figcaption>
  </figure>
</div>

1. The `src` CLI **generates a set of patches** by running `gofmt` over every repository that has a `go.mod` file, leveraging Sourcegraph's search capabilities.

    This is called **executing an _action_** (an _action_ is a series of commands and Docker containers to run in each repository). It yields a **set of patches**, one for each repository, which you can inspect either in the CLI or in the Sourcegraph UI.
1. The patches are then used to **create a draft campaign**.
1. At this point, since it's a draft camapaign, no changesets have been created on the code host.
1. The user then selectively **creates GitHub pull requests** by publishing single patches.

<div class="clearfix"></div>

## Requirements

* Sourcegraph instance [configured for campaigns](./configuration.md).
* `src` CLI: [Installation and setup instructions](https://github.com/sourcegraph/src-cli/#installation)

## Limitations

Campaigns currently support **GitHub**, **GitLab** and **Bitbucket Server** repositories. If you're interested in using campaigns on other code hosts, [let us know](https://about.sourcegraph.com/contact).
