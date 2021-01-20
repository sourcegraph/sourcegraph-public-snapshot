# Tracking existing changesets

Campaigns allow you not only to [publish changesets](publishing_changesets.md) but also to **import and track changesets** that already exist on different code hosts. That allows you to get an overview of the status of multiple changesets, with the ability to filter and drill down into the details of a specific changeset.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/tracking_existing_changesets_overview.png" class="screenshot center">

## Importing changesets into a campaign

To track existing changesets in a campaign you add them to the [campaign spec](../explanations/introduction_to_campaigns.md#campaign-spec) under the `importChangesets` property and apply the campaign spec.

The following example campaign spec tracks multiple existing changesets in different repositories on different code hosts:

```yaml
name: track-important-milestone
description: Track all changesets related to our important milestone

importChangesets:
- repository: github.com/sourcegraph/sourcegraph
  externalIDs: [15397, 15590, 15597, 15583, 15806, 15798]
- repository: github.com/sourcegraph/src-cli
  externalIDs: [378, 373, 374, 369, 368, 361, 380]
- repository: bitbucket.sgdev.org/SOUR/vegeta
  externalIDs: [8]
- repository: gitlab.sgdev.org/sourcegraph/src-cli
  externalIDs: [113, 119]
```

See "[Creating a campaign](creating_a_campaign.md)" on how to create a campaign from the campaign spec.

> NOTE: You can combine the tracking of existing changesets and creating new ones by adding `importChangesets:` to your campaign specs that have `on:`, `steps:` and `changesetTemplate:` properties.

Once you've created the campaign you'll see the existing changeset show up in the list of changesets. The campaign will track the changeset's status and include it in the overall campaign progress (in the same way as if it had been created by the campaign):

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/tracking_existing_changesets_burndown_chart.png" class="screenshot center">
