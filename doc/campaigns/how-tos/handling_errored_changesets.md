# Handling errored changesets

Publishing a changeset can result in an error for different reasons.

Sometimes the problem can be fixed by automatically retrying to publish the changeset, but other errors require the user to take some action.

Errored changesets that are marked as **Retrying** are being automatically retried:

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/retrying_changeset.png" class="screenshot">

Changesets that are marked as **Failed** can be [retried manually](#manual-retrying-of-errored-changesets):

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/campaigns/failed_changeset_retry.png" class="screenshot">

## Types of errors

Examples of errors that can be fixed by [automatically retrying](#automatic-retrying-of-errored-changesets):

- Connecting to the code host failed
- Code host responds with an error when trying to open a pull request
- Internal network errors
- ...

Examples of errors that requires [manual retrying](#manual-retrying-by-re-applying-the-campaign-spec):

- No [campaigns credentials](configuring_user_credentials.md) have been setup for the affected code host
- The configured code host connection needs a different type of credentials (e.g. SSH keys, which are currently not supported)
- A pull request for the specified branch already exists in another campaign
- ...

## Automatic retrying of errored changesets

When Sourcegraph campaigns marks a changeset as **Retrying** it's automatically going to retry publishing it for up to 60 times.

No user action is needed.

## Manual retrying of errored changesets

Changesets that are marked as **Failed** won't be retried automatically. That's either because the number of automatic retries has been exhausted, or because retrying won't fix the error without user intervention.

When a changeset failed publishing, the user can click _Retry_ on the error message. No re-applying needed.

Additionally, in order to retry all **Failed** (or even **Retrying**) changesets manually, you can re-apply the campaign spec.

**Option 1:** Preview and re-apply the campaign spec in the UI by running

```bash
src campaign preview -f YOUR_CAMPAIGN_SPEC.campaign.yaml
```

and clicking on the printed URL to apply the uploaded campaign spec.

**Option 2:** Re-apply directly by running the following:

```bash
src campaign apply -f YOUR_CAMPAIGN_SPEC.campaign.yaml
```

See "[Creating a campaign](creating_a_campaign.md)" for more information on these commands.
