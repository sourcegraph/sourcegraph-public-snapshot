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

Without any further configuration, campaigns are **only accessible to site-admins.** If you want to grant read-only access to non-site-admins, use the following site configuration setting:

```json
{
  "campaigns.readAccess.enabled": true
}
```
