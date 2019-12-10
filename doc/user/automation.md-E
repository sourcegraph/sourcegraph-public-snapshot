# Automation

> Automation is currently available in private beta for select enterprise customers.

[Sourcegraph automation](https://about.sourcegraph.com/product/automation) allows large-scale code changes across many repositories and different code hosts.

## Configuration

In order to use the Automation preview, a site-admin of your Sourcegraph instance must enable it in the site configuration settings e.g. `sourcegraph.example.com/site-admin/configuration`

```json
{
  "experimentalFeatures": {
      "automation": "enabled"
  }
}
```

Automation requires that your [external service](../admin/external_service.md) is using a token with **write access** in order to create changesets on your code host.

## Supported campaign types

### Manual

Manual campaigns keep track of existing changesets from various code hosts. You will manually add each changeset you would like to track (such as a GitHub pull request), and can track them to completion.

### Comby search and replace

Comby search and replace campaigns run [Comby](https://comby.dev), a powerful search and replace tool, over a set of repositories that match a specified query scope.

> Currently GitHub and Bitbucket Server are supported Codehosts for this campaign type. Other repositories **won't** be matched by the `scopeQuery` parameter. Also, a maximum of 200 repositories applies at this time.

Parameters:

| Name            | Description                                                                                    |
| --------------- | ---------------------------------------------------------------------------------------------- |
| scopeQuery      | Search query to narrow down repositories to be included in this campaign.                            |
| matchTemplate   | The template to match against in source files. See the [Comby documentation](https://comby.dev/#match-syntax) for syntax. |
| rewriteTemplate | The template to use for the replacements. See the [Comby documentation](https://comby.dev/#match-syntax) for syntax.      |

## Creating a new campaign

1. Navigate to `sourcegraph.example.com/campaigns` or simply click the "Campaigns" entry in the top navbar. (It will only appear when correctly configured).
1. Click "Create new campaign".
1. Enter the title and an optional description for your campaign.
1. Select the type of campaign you wish to run.
   1. Comby search and replace requires additional parameters, enter them into the json editor.
1. For automatically generated campaigns, preview the changes by selecting the 'preview' button and wait for all repositories to be processed. After the preview has finished loading, you can preview the complete set of diffs that were generated, as well as the changesets that will be opened as a result.
1. Adjust parameters as needed.
1. Select 'create'. If the campaign runs automatic changes, they will be applied asynchronously and you can track the progress on that page. Once fully created, the whole list of changesets will be available and you can track the progress of your newly created campaign.

---

If you are looking to run automation on a larger scale in the local dev environment, follow the [guide on automation development](../dev/automation_development.md).
