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

## Supported campaign types and functionality

Our focus is to deliver _general_ functionality (e.g., centralized monitoring of a large set of pull requests on different codehosts) as well as _tailored_ solutions for large-scale code changes and workflows (e.g., detect leaked NPM credentials). If you have a specific automation workflow in mind that is not covered by our current feature set, please reach out to us at <support@sourcegraph.com>.

### Manual changeset monitoring

Manual campaigns keep track of existing changesets from various code hosts. You will manually add each changeset you would like to track (such as a GitHub pull request), and can track them to completion.

### Comby search and replace

We currently offer search and replace functionality using [Comby](https://comby.dev), which is a tailored solution for syntactic, lint-like code changes (if you're interested in general text or regex search and replace then stay tuned, it's in the works!)

> Run Comby across all of your repositories that match a specified scope query. Currently GitHub and Bitbucket Server are supported Codehosts for this campaign type. Other repositories **won't** be matched by the `scopeQuery` parameter. Also, a maximum of 200 repositories applies at this time.

Parameters:

| Name            | Description                                                                                    |
| --------------- | ---------------------------------------------------------------------------------------------- |
| scopeQuery      | Search query to narrow down repositories to be included in this campaign.                            |
| matchTemplate   | The template to match against in source files. See the [Comby documentation](https://comby.dev/#match-syntax) for syntax. |
| rewriteTemplate | The template to use for the replacements. See the [Comby documentation](https://comby.dev/#match-syntax) for syntax.      |

Note: the `scopeQuery` filter narrows the set of repositories to run on, and will then run across all files in the repository. Future improvements will allow further scoping by file name and file contents.

### Upcoming: Regex search and replace

We are actively working on campaign support for regex search and replace. Let us know if this is a feature you are especially excited for!

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
