# Using the `customGitFetch` env var

Some organizations maintain their own fork of `git` in order to use custom optimization of the `git fetch` for handling large monorepos.

In such use cases Sourcegraph's `gitserver` can be configured to use a custom command in place of a common git command. This is done by the specification of `CUSTOM_GIT_FETCH_CONF`, an environment variable set in gitserver.

> *NOTE: This command assumes the use of a forked gitserver image containing a custom git binary, maintained by the Sourcegraph instance's org.*
---
## Usage

In `gitserver` the `CUSTOM_GIT_FETCH_CONF` file is set with the path to a json file containing a mapping of git clone URL domain/path (`domainPath`), to custom `git fetch` command (`fetch`).

For example:
```json
[
	{
		"domainPath": "git.latveria.bot/vi/src",
		"fetch": "/doom-git/git.Linux.x86_64/bin/git -c remote.origin.url=https://git.latveria.bot/vi/src -c remote.origin.mirror=true -c remote.origin.fetch='+refs/*:refs/*' journal-fetch origin"
	},
    {
        "domainPath": "github.com/foo/absolute",
        "fetch": "/foo/bar/git fetch things"
    }
]
```
Where the env var is set `CUSTOM_GIT_FETCH_CONF=/path/to/config.json`

Code contained for this option can be found [here](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/blob/cmd/gitserver/server/customfetch.go).