# How to address common monorepo performance problems

This document is intended as an explanation of common issues faced by Sourcegraph instances syncing monorepos. Sourcegraph search relies on the retrieval and indexing of git repos. For very large repos, more computational resources are required to perform the necessary operations. As such tuning Sourcegraph's services is often the resolution to bugs and performance issues covered in this how-to.

_This document is targeted at docker-compose and kubernetes deployments, where services can be isolated for individual tuning_

The following bullets provide a general guidline to which service may require more resources:

* `sourcegraph-frontend` CPU/memory resource allocations
* `searcher` CPU/memory resource allocations (allocate enough memory to hold all non-binary files in your repositories)
* `indexedSearch` CPU/memory resource allocations (for the `zoekt-indexserver` pod, allocate enough memory to hold all non-binary files in your largest repository; for the `zoekt-webserver` pod, allocate enough memory to hold ~2.7x the size of all non-binary files in your repositories)
* `symbols` CPU/memory resource allocations
* `gitserver` CPU/memory resource allocations (allocate enough memory to hold your Git packed bare repositories)

## Symbols sidebar - Processing symbols

![Screen Shot 2021-11-15 at 12 35 07 AM](https://user-images.githubusercontent.com/13024338/141749036-95759cbe-abd5-4d78-91eb-618423d2f66c.png)

If you are regularly seeing the `Processing symbols is taking longer than expected. Try again in awhile` warning in your sidebar, its likely that your symbols and/or gitserver services are underprovisioned need more CPU/mem resources.

The [symbols sidebar](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/web/src/repo/RepoRevisionSidebarSymbols.tsx?L42) is dependent on the symbols and gitserver services. When Sourcegraph displays a page associated with a repo, a query is made to the graphQL API to retrieve the symbols associated with the current git revision context, [if an index](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:%5Einternal/search/symbol/symbol%5C.go+if+branch+:%3D+indexedSymbolsBranch%28&patternType=literal) has been created by zoekt (our background indexing service) the query should be resolved quickly. However this query evaluates the currency of the index relative to recent commits, because monorepos are often a main workspace and recieve frequent commits their main branch's zoekt index is often considered out of date. In these cases the symbols query will be resolved by the Symbols service. 

Symbols uses [ctags](https://github.com/universal-ctags/ctags#readme) to process a repo and generate the list of symbols displayed in the symbols sidebar. Ctags processing is lazy, and occurs only when the symbols service is first queried. When symbols recieves a query it [fetchs](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:%5Ecmd/symbols/internal/symbols/fetch%5C.go+fetchRepositoryArchive%28&patternType=literal) a [repository archive](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:%5Einternal/gitserver/client%5C.go+func+%28c+*Client%29+Archive%28&patternType=literal) from gitserver to parse, once the archive is parsed to produce an index, the index is [cached](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24%406f4d327+file:%5Ecmd/symbols/internal/symbols/search%5C.go+s.writeAllSymbolsToNewDB%28&patternType=literal) on-disk in an SQLite database for use in subsequent queries. For large monorepos this indexing can take a few minutes resulting in timeouts.

To address this concern allocate more resources to the symbols service (to provide more processing power for indexing operations) and allocate more resources to the gitserver (to provide for the extra load associated with responding to fetch requests from symbols, and speed up sending the large repo)

Below is an example of a diff to improve symbols performance in a k8s deployment.
```diff
          name: debug
        resources:
          limits:
-           cpu: "3"
-           memory: 6G
          requests:
-           cpu: 500m
-           memory: 5G

          name: debug
        resources:
          limits:
+           cpu: "4"
+           memory: 16G
          requests:
+           cpu: "1"
+           memory: 8G
```
_Learn more about managing resources in [docker-compose](https://docs.sourcegraph.com/admin/install/docker-compose/operations) and [kubernetes](https://docs.sourcegraph.com/admin/install/kubernetes/operations)_

## Slow hover tooltip results
Hovering over a symbol results in queries for the definition, and references of the symbol selected. If the symbol is defined in a repo that has been indexed via lsif these results should be uneffected by the monorepo. However if no lsif index exists on the repo and the repo is unindexed, hover results rely on the symbols service and face the same challenges as described above in [symbols sidebar](#symbols-sidebar---processing-symbols). 

If you are viewing a repository with an lsif index, only symbols defined in that repository will return lsif results, when hovering over a symbol whose declaration exists in another repo that has no lsif index, results will first be derived from a zoekt index. If no such index exists, or the index is out of date, hover defaults to the symbols service for an "on the fly" index.

## Slow history tab and git blame results
![Screen Shot 2021-11-15 at 1 10 16 AM](https://user-images.githubusercontent.com/13024338/141754063-2080c7c6-b5be-43c1-b9db-386e916d2968.png)

Selecting the Show History button while viewing a file initiates a request to [fetch commits](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:%5Eclient/web/src/repo/RepoRevisionSidebarCommits%5C.tsx+function+fetchCommits%28&patternType=literal) for the file. This request is ultimately [resolved by gitserver]() using functionality similar to git log. To improve performance allocate gitserver more CPU.
