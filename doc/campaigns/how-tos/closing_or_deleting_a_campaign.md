# Closing or deleting a campaign

You can close a campaign when you don't need it anymore, when all changes have been merged, or when you decide not to proceed with making changes. A closed campaign still appears in the [campaigns list](viewing_campaigns.md). To completely remove it, you can delete the campaign.

Any person with [admin access to the campaign](../explanations/permissions_in_campaigns.md#permission-levels-for-campaigns) can close or delete it.

## Closing a campaign

1. Click the <img src="../campaigns-icon.svg" alt="Campaigns icon" /> campaigns icon in the top navigation bar.

    <img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/campaigns_icon_in_menu.png" class="screenshot">
1. In the list of campaigns, click the campaign that you'd like to close or delete.
1. In the top right, click the **Close** button.

    <img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/closing_campaigns_close_icon.png" class="screenshot">
1. Select whether you want to close all of the campaign's open changesets (e.g., closing all associated GitHub pull requests on the code host).

    <img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/closing_campaigns_close_changesets.png" class="screenshot">
1. Click **Close campaign**.

Once a campaign is closed it can't be updated or reopened anymore.

## Delete a campaign

1. First, close the campaign.
1. Instead of a "Close campaign" button you'll now see a **Delete** button.

    <img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/closing_campaigns_deleting_campaign.png" class="screenshot">
1. Click **Delete**.

The campaign is now deleted from the Sourcegraph instance. The changesets it created (and possibly closed) will still exist on the code hosts, since most code hosts don't support deleting changesets.
