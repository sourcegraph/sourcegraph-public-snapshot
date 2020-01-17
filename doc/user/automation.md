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

## Creating a new campaign

1. After enabling the feature flag, navigate to `sourcegraph.example.com/campaigns` or simply click the "Campaigns" entry in the navigation bar at the top.
1. Click `Create new campaign`.
1. Enter the name and an optional description for your campaign.
1. Select the type of campaign to run. See the list of supported campaign types and functionality below.
1. Configure the campaign type by editing its parameters.
1. Now generate a preview of the changes that would be made by the campaign by clicking `Preview changes`. Wait for all repositories to be processed. This might take a while when the the campaign runs over hundreds of repositories.
1. Once the preview is fully computed, you'll see a list of patches that would be turned into changesets (i.e. pull requests on GitHub) on the code hosts associated with each repository.
1. Feel free to change the arguments and create a new preview.
1. When you're happy with the proposed changes click `Create`. **This will asynchronously create the changesets (i.e. pull requests) on the code hosts**. Once created, changeset progress (e.g., `open`, `merged`) can be tracked in the campaign view.

Automation requires that your [external service](../admin/external_service.md) is using a token with **write access** in order to create changesets on your code host.

## Current limitations

Automation is still in beta and currently has some limitations to keep in mind:

- We currently only support GitHub and Bitbucket Server code hosts in Automation.
- The maximum number of repositories which an Automation campaign can process is 200.

If you have a specific automation workflow in mind that is not covered by our current feature set or exceeds our repository limit, let us know at <support@sourcegraph.com>.

## Supported campaign types and functionality

Our focus is to deliver _general_ functionality (e.g., centralized monitoring of a large set of pull requests on different code hosts) as well as _tailored_ solutions for large-scale code changes and workflows (e.g., detect leaked NPM credentials).

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

### Finding leaked credentials

This campaign type finds possibly leaked credentials across your codebase and creates changesets that remove or replace the credentials.

Parameters:

| Name            | Description                                                                                    |
| --------------- | ---------------------------------------------------------------------------------------------- |
| scopeQuery      | Search query to narrow down repositories to be included in this campaign.                      |
| matchers        | A list of credential "matchers" that are run.                                                  |

Properties of a `matcher`:

| Name            | Description                                                                                    |
| --------------- | ---------------------------------------------------------------------------------------------- |
| type            | The type of the matcher. Currently supported: `"npm"`                   .                      |
| replaceWith     | An optional string that the credentials are replaced with.                                     |


The currently available matcher types:

1. `npm`: finds (and optionally replaces) NPM registry tokens and passwords in `.npmrc` files.

### Manual changeset monitoring

Manual campaigns keep track of existing changesets from various code hosts. You can manually add each changeset you would like to track (such as a GitHub pull request), and can track them to completion.

---

If you are looking to run automation on a larger scale in the local dev environment, follow the [guide on automation development](../dev/automation_development.md).
