# Rockskip: fast symbol sidebar and search-based code intelligence on monorepos

Rockskip is an alternative symbol indexing and query engine for the symbol service intended to improve performance of the symbol sidebar and search-based code intelligence on big monorepos. It was added in Sourcegraph 3.38.

## When should I use Rockskip?

![still processing symbols error](https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/symbol-sidebar-timeout.png)

![hover popover spinner](https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/hover-popover-spinner.png)

If you regularly see the above error or slow hover popovers (when not using LSIF), it probably means that the default implementation (which copies SQLite DBs for each commit) is not fast enough and that Rockskip might help.

A very rough way to gauge if Rockskip will help is if your repository has a 2GB+ `.git` directory, 50K+ commits, or 50K+ files in the working tree.

You can always try Rockskip for a while and if it doesn't help then you can disable it.

## How do I enable Rockskip?

**Step 1:** Give your `codeintel-db` has a few extra GB of RAM and set environment variables on the `symbols` container:

For Kubernetes:

```yaml
# base/symbols/symbols.Deployment.yaml
spec:
  template:
    spec:
      containers:
      - name: symbols
        env:
        # 👇 Enables Rockskip
        - name: USE_ROCKSKIP
          value: "true"
        # 👇 Uses Rockskip for the repositories in the comma separated list
        - name: ROCKSKIP_REPOS
          value: "github.com/torvalds/linux,github.com/pallets/flask"
```

```yaml
# base/codeintel-db/codeintel-db.Deployment.yaml
spec:
  template:
    spec:
      containers:
      - name: pgsql
        resources:
          limits:
            memory: 8Gi # 👈 Increase RAM from 4g to 8g
          requests:
            memory: 8Gi # 👈 Increase RAM from 4g to 8g
```

For Docker Compose:

```yaml
services:

  symbols-0:
    environment:
      # 👇 Enables Rockskip
      - USE_ROCKSKIP=true
      # 👇 Uses Rockskip for the repositories in the comma separated list
      - ROCKSKIP_REPOS=github.com/torvalds/linux,github.com/pallets/flask

  codeintel-db:
    mem_limit: '8g' # 👈 Increase RAM from 2g to 8g
```

For other deployments, make sure that:

- The `symbols` service has access to the codeintel DB
- The `symbols` service has the environment variables set
- The `codeintel-db` has a few extra GB of RAM

**Step 2:** Kick off indexing

1. Visit your repository in the Sourcegraph UI
1. Click on the branch selector, click **Commits**, and select the second most recent commit (this avoids routing the request to Zoekt)
1. Open the symbols sidebar to kick off indexing (it's ok to see a loading spinner, that probably means indexing is in progress)

**Step 3:** Check the indexing status by following the [instructions below](#how-do-i-check-the-indexing-status).

**Step 4:** Open the symbols sidebar again and the symbols should appear quickly. Hover popovers and jump-to-definition via search-based code intelligence should also respond quickly.

That's it! New commits will be indexed automatically when users visit them.

## How long does indexing take?

The initial indexing takes roughly 4 hours per GB of the `.git` directory (you can check the size with `du -sch .git`). Once the full repository has been indexed, indexing new commits takes less than 1 second most of the time.

## What resources does Rockskip use?

Rockskip stores all data in Postgres, and the tables it creates use roughly 3x as much space as your `.git` directory, so make sure your Postgres instance has enough free disk. Rockskip indexes every symbol in the entire history of your repository and makes heavy use of Postgres indexes to make all kinds of queries fast, including: path prefix queries, regex queries with trigram index optimization, file extension queries, and the internal commit visibility queries.

Rockskip is completely single-threaded when indexing a repository, but multiple repositories can be indexed at a time. The concurrency is limited by `MAX_CONCURRENTLY_INDEXING`, which defaults to 4.

Rockskip heavily relies on gitserver for data. Rockskip issues very long-running `git log` commands, as well as many `git archive` calls.

## How do I check the indexing status?

The symbols container responds to GET requests on the `localhost:3184/status` endpoint with the following info:

- Repository count
- Size of the symbols table in Postgres
- Most recently searched repositories
- List of in-flight indexing and search requests

For Kubernetes, find the symbols pod and `exec` a `curl` command in it:

```
$ kubectl get pods | grep symbols
symbols-5ff7c67b57-mr5h4

$ kubectl exec -ti symbols-5ff7c67b57-mr5h4 -- curl localhost:3184/status
This is the symbols service status page.

Number of repositories: 1
Size of symbols table: 3253 MB

Most recently searched repositories (at most 5 shown)
  2022-03-11 05:48:58.890765 +0000 UTC github.com/sgtest/megarepo

Here are all in-flight requests:

indexing github.com/sgtest/megarepo@14a3d9849ba587d667efbc542cf0c7cd106c3e72
    progress 9.53% (indexed 49151 of 515574 commits), 36h55m18.227079912s remaining
    Tasks (14006.77s total, current AppendHop+): AppendHop+ 44.76% 49152x, InsertSymbol 18.67% 1997101x, AppendHop- 12.94% 49151x, UpdateSymbolHops 7.78% 825380x, parse 4.01% 369401x, GetCommitByHash 2.73% 515574x, get hops 2.39% 49152x, ArchiveEach 2.26% 98302x, GetSymbol 1.83% 325351x, CommitTx 1.26% 49151x, DeleteRedundant 0.79% 49151x, InsertCommit 0.30% 49152x, Log 0.28% 1x, RevList 0.00% 1x, iLock 0.00% 1x, idle 0.00% 1x,
    holding iLock
```

In this example you can see there's 1 repository and the symbols service has indexed 9% of all commits with an ETA of 36H from now. There's also a breakdown of tasks that are part of Rockskip's internal workings mostly for Sourcegraph engineers, so you can ignore that.

## How does it work?

For a deeper dive into the index and query structures, check out the [explanatory RFC](https://docs.google.com/document/d/1sDDpZaWdGtIaiNLNB8QsLwHTvH10fhEKpEa4qcog5vg/edit?usp=sharing).
