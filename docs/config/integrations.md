+++
title = "Integrations"
+++

## JIRA

Sourcegraph includes an integration with JIRA that allows you to link to JIRA issues by mentioning them in the commit messages or description of a changeset. To enable this feature, set the CLI flag `--jira.url` to your JIRA instance's domain (e.g. "jira.mycompany.com"). If your JIRA instance uses TLS, set the CLI flag `--jira.tls` to `true`.

The integration can also automatically create links to Sourcegraph inside JIRA issues that have been mentioned in a changeset. To enable this, set the CLI flag `--jira.credentials` to the basic auth information of a JIRA user on your instance in the format `username:password`.

To mention a JIRA issue within a commit message or changeset description, create a line that begins with `JIRA-Issues: ` followed by any number of JIRA issue IDs e.g. `MYPROJECT-1`.

## GitHub

Sourcegraph can import private GitHub repositories, enabling a limited set of Sourcegraph features on the externally hosted repository. To set this up, navigate to `http://src.mycompany.com/~USERNAME/.settings/integrations` and follow the instructions to create and add a GitHub personal access token to your Sourcegraph. This will fetch the list of private repositories. Enable a repository to mirror the upstream GitHub repository on Sourcegraph.

To mirror public GitHub repositories, run this command:

`src repo create -m --clone-url https://github.com/mycompany/project project`

### GitHub Enterprise

To import repositories from a GitHub Enterprise instance, set the commandline flag `--github.host` to your GitHub instance's domain name, eg `--github.host=ghe.mycompany.org`. For Ubuntu Linux or cloud installations, add the following lines to the config file `/etc/sourcegraph/config.ini`:

```
[serve.GitHub]
GitHubHost = ghe.mycompany.org
```

Restart the server with `sudo restart src`. Now you can follow the steps from the previous section to import repositories from the GitHub Enterprise instance.

Your GitHub Enterprise instance must have TLS enabled. You can't simultaneously import repos from github.com and GitHub Enterprise.