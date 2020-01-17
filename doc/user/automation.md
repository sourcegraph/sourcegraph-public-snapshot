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

Our focus is to deliver _general_ functionality (e.g., centralized monitoring of a large set of pull requests on different code hosts) as well as _tailored_ solutions for large-scale code changes and workflows (e.g., detect leaked NPM credentials). If you have a specific automation workflow in mind that is not covered by our current feature set, please reach out to us at <support@sourcegraph.com>.

We currently support GitHub and Bitbucket Server code hosts. A maximum of 200 repositories can be processed at a time (please reach out to us if you have a use case that exceeds this number).

### Regex search and replace

We support regular expression search and replace using [RE2 syntax](https://github.com/google/re2/wiki/Syntax). A scope query allows to narrow the set of repositories to process. Specific files can also be filtered by changing the `file:` parameter in the scope query.

| Name            | Description                                                                                    |
| --------------- | ---------------------------------------------------------------------------------------------- |
| scopeQuery      | Search query to narrow down repositories to be included in this campaign.                      |
| regexMatch      | A regular expression. We support [RE2 syntax](https://github.com/google/re2/wiki/Syntax)
| textReplace     | Replacement text for `regexMatch`. You may refer to match groups using `$1` or `${1}` syntax.  |

### Comby search and replace

We support search and replace functionality using [Comby](https://comby.dev), which is a tailored solution for syntactic, lint-like code changes.

Parameters:

| Name            | Description                                                                                    |
| --------------- | ---------------------------------------------------------------------------------------------- |
| scopeQuery      | Search query to narrow down repositories to be included in this campaign.                            |
| matchTemplate   | The template to match against in source files. See the [Comby documentation](https://comby.dev/#match-syntax) for syntax. |
| rewriteTemplate | The template to use for the replacements. See the [Comby documentation](https://comby.dev/#match-syntax) for syntax.      |

Note: the `scopeQuery` filter for `comby` narrows the set of repositories and will currently run on _all_ files in the repository. Future improvements will allow to restrict changes using the `file:` filter.

### Manual changeset monitoring

Manual campaigns keep track of existing changesets from various code hosts. You can manually add each changeset you would like to track (such as a GitHub pull request), and can track them to completion.

## Creating a new campaign

1. Navigate to `sourcegraph.example.com/campaigns` or simply click the "Campaigns" entry in the top navbar. (It will only appear when correctly configured).
1. Click `Create new campaign`.
1. Enter the title and an optional description for your campaign.
1. Select the type of campaign to run. For example, enter regex patterns to match and replace text using the JSON editor. Note: special characters like `\` and `"` need to be escaped inside the string.
1. To generate a preview campaign, click `Preview changes` and wait for all repositories to be processed. When the preview is ready, you'll see a list of diffs.
1. Feel free to change the search and replace patterns to create a new preview.
1. Once the preview looks good, click `create`. **This will create a pull request of the changset on your code host**. Once created, changeset progress (e.g., `open`, `merged`) can be tracked in the campaign view.

---

If you are looking to run automation on a larger scale in the local dev environment, follow the [guide on automation development](../dev/automation_development.md).
