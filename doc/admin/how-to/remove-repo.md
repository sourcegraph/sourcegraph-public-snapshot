# How to remove a repository from Sourcegraph

This document will walk you through the steps of removing a repository from Sourcegraph. 

## Prerequisites

This document assumes that you have:

* site-admin level permissions on your Sourcegraph instance
* access to your Sourcegraph deployment

## Steps to remove a repository from Sourcegraph

1. Add the repository name to the [exclude list](https://docs.sourcegraph.com/admin/external_service/github#exclude) in your [code host configuration](https://docs.sourcegraph.com/admin/external_service).
1. Wait for the repository to disappear from the Repository Status Page located in your Site Admin panel.

## Remove corrupted repository data from Sourcegraph

1. Add the repository name to the [exclude list](https://docs.sourcegraph.com/admin/external_service/github#exclude) in your [code host configuration](https://docs.sourcegraph.com/admin/external_service).
1. Wait for the repository to disappear from the Repository Status Page located in your Site Admin panel.
1. Once you have confirmed the previous step has been completed, you will then exec into Gitserver (for docker-compose and kubernetes deployments) to locate the files that are associated with the repository.
1. Look for a directory with the name of the repository in the Gitserver. It should be located in the following file path: `data/repos/{name-of-code-host}/{name-of-repo}`
1. Delete the directory for that repo from the previous step.

## To reclone a removed repository

1. Remove the repository from the [exclude list](https://docs.sourcegraph.com/admin/external_service/github#exclude)
2. The reclone process should start in the next syncing cycle

## Manually purge deleted repository data from disk

After a repository is deleted from Sourcegraph in the database, its data still remains on disk on gitserver so that in the event the repository is added again it doesn't need to be recloned. These repos are automatically removed when disk space is low on gitserver. However, it is possible to manually trigger removal of deleted repos in the following way:

**NOTE:** This is not available on Docker Compose deployments.

1. Browse to Site Admin -> Instrumentation -> Repo Updater -> Manual Repo Purge
2. You'll be at a url similar to `https://sourcegraph-instance/-/debug/proxies/repo-updater/manual-purge`
3. You need to specify a limit parameter which specifies the upper limit to the number of repos that will be removed, for example: `https://sourcegraph-instance/-/debug/proxies/repo-updater/manual-purge?limit=1000`
4. This will trigger a background process to delete up to `limit` repos, rate limited at 1 per second. 

It's possible to see the number of repos that can be cleaned up on disk in Grafana using this query:

```
max(src_repoupdater_purgeable_repos)
```
