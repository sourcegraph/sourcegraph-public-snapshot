+++
title = "GitHub Enterprise"
+++

To mirror repositories from a GitHub Enterprise instance on Sourcegraph, set the commandline flag `--github.host` to your
GitHub Enterprise instance's domain name, eg `--github.host=ghe.mycompany.org`. Or, for Ubuntu Linux or cloud installations,
add the following lines to the config file `/etc/sourcegraph/config.ini` and restart your Sourcegraph server with
`sudo restart src`:

```
[serve.GitHub]
GitHubHost = ghe.mycompany.org
```

Now you can follow steps for [importing private GitHub repositories]({{< relref "integrations/GitHub.md" >}}).

**Note:** Your GitHub Enterprise instance must have TLS enabled. You cannot simultaneously import repos from
GitHub.com and GitHub Enterprise.
