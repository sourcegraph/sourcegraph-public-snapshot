import { map } from 'rxjs/operators'

import { fetchHighlight, fetchBlobPlaintext } from '$lib/loader/blob'
import { fetchRepoCommits, queryRepositoryComparisonFileDiffs } from '$lib/repo/api/commits'

import type { PageLoad } from './$types'

export const load: PageLoad = async ({ parent, params, url }) => {
    const revisionToCompare = url.searchParams.get('rev')
    const { resolvedRevision } = await parent()

    return {
        deferred: {
            blob: fetchBlobPlaintext({
                filePath: params.path,
                repoName: resolvedRevision.repo.name,
                revision: resolvedRevision.commitID,
            }).toPromise(),
            highlights: fetchHighlight({
                filePath: params.path,
                repoName: resolvedRevision.repo.name,
                revision: resolvedRevision.commitID,
            })
                .pipe(map(highlight => highlight?.lsif))
                .toPromise(),
            compare: revisionToCompare
                ? {
                      revisionToCompare,
                      diff: fetchRepoCommits({
                          repoID: resolvedRevision.repo.id,
                          revision: revisionToCompare,
                          filePath: params.path,
                          first: 1,
                          pageInfo: { hasNextPage: true, endCursor: '1' },
                      })
                          .then(history =>
                              queryRepositoryComparisonFileDiffs({
                                  repo: resolvedRevision.repo.id,
                                  base: history.nodes[0]?.oid ?? null,
                                  head: revisionToCompare,
                                  paths: [params.path],
                                  first: null,
                                  after: null,
                              }).toPromise()
                          )
                          .then(result => result?.nodes[0]),
                  }
                : null,
        },
    }
}
