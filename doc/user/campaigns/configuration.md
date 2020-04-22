# Configuration

## Enabling Campaigns

In order to use Campaigns, a site-admin of your Sourcegraph instance must enable it in the site configuration settings e.g. `sourcegraph.example.com/site-admin/configuration`

```json
{
  "experimentalFeatures": {
      "automation": "enabled"
  }
}
```

## Read-access for non-site-admins

Without any further configuration, campaigns are **only accessible to site admins.** If you want to grant read-only access to non-site-admins, use the following site configuration setting:

```json
{
  "campaigns.readAccess.enabled": true
}
```

## Code host configuration

When using campaigns with repositories hosted on GitHub, make sure that the GitHub connection configured in Sourcegraph uses a token with the [required token scopes](../../admin/external_service/github.md#github-api-token-and-access). Otherwise campaigns won't be able to create changesets (pull requests) on the configured GitHub instance and sync them back to Sourcegraph.

The user associated with the token also needs to have write-access to the repository in order to create changesets when creating campaigns.
