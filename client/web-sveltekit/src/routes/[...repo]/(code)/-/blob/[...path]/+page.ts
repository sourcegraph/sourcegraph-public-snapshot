import { map } from 'rxjs/operators'

import { isErrorLike } from '$lib/common'
import { fetchHighlight, fetchBlobPlaintext } from '$lib/loader/blob'
import { fetchRepoCommits, queryRepositoryComparisonFileDiffs } from '$lib/repo/api/commits'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'

export const load: PageLoad = ({ parent, params, url }) => {
    const { repoName, revision } = parseRepoRevision(params.repo)
    const revisionToCompare = url.searchParams.get('rev')

    return {
        deferred: {
            blob: fetchBlobPlaintext({
                filePath: params.path,
                repoName,
                revision: revision ?? '',
            }).toPromise(),
            highlights: fetchHighlight({ filePath: params.path, repoName, revision: revision ?? '' })
                .pipe(map(highlight => highlight?.lsif))
                .toPromise(),
            compare: revisionToCompare
                ? {
                      revisionToCompare,
                      diff: parent()
                          .then(({ resolvedRevision }) =>
                              !isErrorLike(resolvedRevision)
                                  ? fetchRepoCommits({
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
                                    )
                                  : null
                          )
                          .then(result => result?.nodes[0]),
                  }
                : null,
        },
    }
}
