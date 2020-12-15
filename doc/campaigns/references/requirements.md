# Requirements

Campaigns has requirements for the Sourcegraph server version, its connected code hosts and developer environments. 


## Code hosts

Campaigns is compatible with the following code hosts and versions:


- Github.com
- Github Enterprise 2.20 and later
- GitLab 12.7 and later (burndown charts are only supported with 13.2 and later)
- Bitbucket server 5.7 and later

We **highly recommend** enabling webhooks to increase performance with large campaigns:

- [GitHub](../../admin/external_service/github.md#webhooks)
- [Bitbucket Server](../../admin/external_service/bitbucket_server.md#webhooks)
- [GitLab](../../admin/external_service/gitlab.md#webhooks)


## Sourcegraph server

While the latest version of Sourcegraph server is always recommended, version 3.22 or greater is the minimum version required to run campaigns. 

## Requirements for developers creating and running campaigns

- Latest version of the [Sourcegraph CLI `src`](../../cli/index.md)
  - `src` is supported on Linux or macOS , Windows support is experimental
- Docker
- Git
