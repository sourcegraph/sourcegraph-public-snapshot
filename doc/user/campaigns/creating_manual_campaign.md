# Creating a manual campaign

Manual campaigns provide the ability to manage and monitor changesets (pull requests) that already exist on code hosts.

<div style="max-width: 450px;" class="mx-auto">
  <figure class="figure">
    <div class="figure-img">
      <img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/manual_campaign.png" />
    </div>
    <figcaption class="figure-caption text-right">A campaign tracking multiple changesets in different repositories.</figcaption>
  </figure>
</div>

In order to create a manual campaign, follow these steps:

1. Go to `/campaigns` on your Sourcegraph instance and click on the **New campaign** button.
1. Click on **Track existing changesets**.
1. Fill in a title for the campaign and a description.
1. Click **Create**.
1. Add changesets by specifying the name of the repository they belong to and their external ID (e.g. the number of a pull request on GitHub) in the **Add changeset** form.
