+++
title = "Integrations"
+++

## JIRA

Sourcegraph includes an integration with JIRA that allows you to link to JIRA issues by mentioning them in the commit messages or description of a changeset. To enable this feature, set the CLI flag `--jira.url` to your JIRA instance's domain (e.g. "jira.mycompany.com"). If your JIRA instance uses TLS, set the CLI flag `--jira.tls` to `true`.

The integration can also automatically create links to Sourcegraph inside JIRA issues that have been mentioned in a changeset. To enable this, set the CLI flag `--jira.credentials` to the basic auth information of a JIRA user on your instance in the format `username:password`.

To mention a JIRA issue within a commit message or changeset description, create a line that begins with `JIRA-Issues: ` followed by any number of JIRA issue IDs e.g. `MYPROJECT-1`.

## GitHub

Sourcegraph can import private GitHub repositories, enabling a limited set of Sourcegraph features on the externally hosted repository. To set this up, navigate to `APP_URL/~USERNAME/.settings/integrations`, and follow the instructions to create and add a GitHub personal access token to your Sourcegraph. This will fetch the list of private repositories. Enable a repository to mirror the upstream GitHub repository on Sourcegraph.

### GitHub Enterprise

To import repositories from a GitHub Enterprise instance, set the commandline flag `--github.host` to your GitHub instance's domain name, eg `--github.host=ghe.mycompany.org`. Now you can follow the steps above to import private repositories from the GitHub Enterprise instance. To import both public and private repositories from the GitHub Enterprise instance, set the command line flag `--github.import-public`.

Note that your GitHub Enterprise instance must have TLS enabled.