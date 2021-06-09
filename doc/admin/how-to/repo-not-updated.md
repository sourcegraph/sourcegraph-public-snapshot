# How to troubleshoot a repo that is not being updated on Sourcegraph

A repository on Sourcegraph that fails to be updated could have different reasons. For example, the repository is a bad repository, the number of repositories causes significant load on the code host, a large repository, or code host permission issues etc. This guide will walk you through the steps you can take to investigate further when encountering this issue. 

## Prerequisites

This document assumes that you are a [site administrator](https://docs.sourcegraph.com/admin) and **do not** have `disableAutoGitUpdates` set to `true` in your site configuration.

## Steps to investigate

1. First of all, you should identify how long has it been since the repo was last updated on Sourcegraph by going to the repository page > Settings > Mirroring
1. From there you can check:
    1. Last refreshed: Time when the repo was last synced
    1. Next scheduled update: Estimated time of when the repo will be updated next (this could change as it is determined by a [smart heuristic](https://docs.sourcegraph.com/admin/repo/update_frequency#repository-update-frequency))
    1. Queued for update: Its position in queue to be updated next
    1. Connection: Connection status to the repository
 1. If clicking on the `Refresh Now` button has triggered the repository to be updated instantly for you then congratulations! You can now move on from this troubleshooting guide!
 1. If clicking on the `Refresh Now` button does not work for you, try using webhooks following the instructions detailed in our [Repository Webhooks Docs](https://docs.sourcegraph.com/admin/repo/webhooks#webhook-for-manually-telling-sourcegraph-to-update-a-repository)
 1. Look for errors related to this repository in gitserver logs, which should help you to determine the next best course of action.
 
## FAQs:

### What is a bad repository?

A bad repository is simiply a repository that Sourcegraph cannot handle. For example:
1. A large repository
    1. This could result in gitserver running out of memory.
1. An empty/locked/disabled repository
    1. For example, if you've created a repo repository GitHub and did not click the "initialize the repository for me" button, the repository would then become an empty repository as it has no commits at all.
1. A corrupted repository
    1. The repository could be corrupted on the code host, or corrupted on disk.

### Looks like Sourcegraph is trying to clone the same bad repository over and over again, what should I do?

Please upgrade your instance to [3.27.0](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/CHANGELOG.md#3-27-0) where the bug has been addressed.
