# Automation

> Automation is currently available in preview for select enterprise customers.

Sourcegraph automations allow large-scale code changes across many repositories and different code hosts.

## Setup

In order to use the Automation preview, a site-admin of your Sourcegraph instance has to enable it in the settings at `/site-admin/configuration` using

```json
{
  "experimentalFeatures": {
      "automation": "enabled"
  }
}
```

When using automatic campaigns that create changesets on your code host, the external service needs to have a token that permits **write access**. (See [more on external services](/admin/external_service))

## Supported campaign types

### Manual:
A campaign that keeps track of manually added changesets from various codehosts. Use this type if you have existing changesets that you want to keep track of.

### Comby search and replace:
A campaign that runs [comby](https://comby.dev), a powerful search and replace tool, over a set of repositories that match the `scopeQuery`.
> For comby campaigns, we currently support GitHub and Bitbucket Server as Codehosts. Other repositories **won't** be matched by the `scopeQuery` parameter. Also, a maximum of 200 repositories applies at this time.

Parameters:

| Name            | Description                                                                                    |
| --------------- | ---------------------------------------------------------------------------------------------- |
| scopeQuery      | Search query to narrow down repositories included in this campaign.                            |
| matchTemplate   | The template to match against in source files. See https://comby.dev/#match-syntax for syntax. |
| rewriteTemplate | The template to use for the replacements. See https://comby.dev/#match-syntax for syntax.      |

## Creating a new campaign

1. Navigate to `/campaigns` or simply click the "Campaigns" entry in the top navbar. (It will only appear when you did the setup correctly).
2. Click "Create new campaign".
3. On the following screen, enter the title and an optional description for your campaign. Next to select is the type of campaign you wish to run.
  b) Comby search and replace requires additional parameters, enter them into the json editor.
4. For automatically generated campaigns, hit the preview button and wait for all repositories to be processed. Once done, you can preview the total diff that was generated, as well as the changesets that will be opened on the repositories. Adjust parameters as needed.
5. Hit create, if the campaign runs automatic changes, they will be applied asynchronously and you can track the progress on that page. Once fully created, the whole list of changesets will be available and you can track the progress of your newly created campaign.

---

If you are looking to run automation on a larger scale in the local dev environment, follow the [guide on local automation development](/dev/local_automation_development).
