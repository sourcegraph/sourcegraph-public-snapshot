import { catchError } from 'rxjs/operators'

import { asError, isErrorLike, type ErrorLike } from '$lib/common'
import { fetchBlobPlaintext } from '$lib/loader/blob'
import { fetchRepoCommits, queryRepositoryComparisonFileDiffs } from '$lib/repo/api/commits'
import { fetchTreeEntries } from '$lib/repo/api/tree'
import { findReadme } from '$lib/repo/tree'

import type { PageLoad } from './$types'

export const load: PageLoad = async ({ params, parent, url }) => {
    const revisionToCompare = url.searchParams.get('rev')
    const { resolvedRevision, revision, repoName } = await parent()

    const treeEntries = fetchTreeEntries({
        repoName,
        commitID: resolvedRevision.commitID,
        revision: revision ?? '',
        filePath: params.path,
        first: 2500,
    })
        .pipe(catchError((error): [ErrorLike] => [asError(error)]))
        .toPromise()
        .then(commit => (isErrorLike(commit) ? null : commit.tree))

    return {
        deferred: {
            treeEntries,
            readme: treeEntries.then(result => {
                if (!result) {
                    return null
                }
                const readme = findReadme(result.entries)
                if (!readme) {
                    return null
                }
                return fetchBlobPlaintext({ repoName, revision: revision ?? '', filePath: readme.path })
                    .toPromise()
                    .then(result => ({
                        name: readme.name,
                        ...result,
                    }))
            }),
            compare: revisionToCompare
                ? {
                      revisionToCompare,
                      diff: fetchRepoCommits({
                          repoID: resolvedRevision.repo.id,
                          revision: revisionToCompare,
                          filePath: params.path,
                          first: 1,
                          pageInfo: { hasNextPage: true, endCursor: '1' },
                      }).then(history =>
                          queryRepositoryComparisonFileDiffs({
                              repo: resolvedRevision.repo.id,
                              base: history.nodes[0]?.oid ?? null,
                              head: revisionToCompare,
                              paths: [params.path],
                              first: null,
                              after: null,
                          }).toPromise()
                      ),
                  }
                : null,
        },
    }
}
