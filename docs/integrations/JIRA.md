+++
title = "JIRA"
+++

Sourcegraph includes an integration with JIRA that allows you to link to JIRA issues by mentioning them in the commit messages or description of a changeset. To enable this feature, set the CLI flag `--jira.url` to your JIRA instance's URL (e.g. "http://jira.mycompany.com").

The integration can also automatically create links to Sourcegraph inside JIRA issues that have been mentioned in a changeset. To enable this, set the CLI flag `--jira.credentials`, or the environment variable `SG_JIRA_CREDENTIALS` to the basic auth information of a JIRA user on your instance in the format `username:password`.

To mention a JIRA issue within a commit message or changeset description, create a line that begins with `JIRA Issues: ` followed by any number of JIRA issue IDs e.g. `MYPROJECT-1`.
