# Requirements

Campaigns has requirements for the Sourcegraph server version, its connected code hosts and developer environments.

## Code hosts

Campaigns is compatible with the following code hosts:

* Github.com
* Github Enterprise 2.20 and later
* GitLab 12.7 and later (burndown charts are only supported with 13.2 and later)
* Bitbucket server 5.7 and later

### Campaigns effect on code host rate limits

Each changeset makes multiple API requests to its code host so that its status and other information is current. Sourcegraph intelligently schedules these requests to avoid overwhelming the code host's rate limits. In environments with many open campaigns, this can result in outdated changesets as they await their turn in the update queue.

We **highly recommend** enabling webhooks to increase performance and reduce api requests accociated with large campaigns:

* [GitHub](../../admin/external_service/github.md#webhooks)
* [Bitbucket Server](../../admin/external_service/bitbucket_server.md#webhooks)
* [GitLab](../../admin/external_service/gitlab.md#webhooks)

### A note on campaigns effect on CI systems

Campaigns makes it possible to create changesets in tens or thousands of repositories. These changesets may trigger many checks or continuous integration jobs that may stress the resources allotted to these systems. Campaigns supports partial publishing for changesets to help mitigate these issues. You may also consider publishing your changesets at times of low activity.  

## Sourcegraph server

While the latest version of Sourcegraph server is always recommended, version 3.22 or greater is the minimum version required to run campaigns.

## Requirements for developers creating campaigns

* Latest version of the [Sourcegraph CLI `src`](../../cli/index.md)
  * `src` is supported on Linux or macOS, Windows support is experimental
* Docker
  * MacOS:
    * If using Docker 3.x, ensure your version is at least 3.0.1
    * In 3.x versions, the gRPC setting must be disabled
* Disk space
  * Disk equal to each campaign's largest repository plus any dependencies or requirements required by the run steps, times the number of parallel jobs is required.
    * The default number of parallel jobs is equivalent to the number of cores in the system running src-cli. This setting is configurable.
* Disk space is also required for the generated patches. This requirement *is* cumulative across each repository altered.
* Git
