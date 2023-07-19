import { isErrorLike } from '$lib/common'
import { fetchRepoCommit, queryRepositoryComparisonFileDiffs } from '$lib/repo/api/commits'

import type { PageLoad } from './$types'

export const load: PageLoad = ({ parent, params }) => {
    const commit = parent().then(({ resolvedRevision }) => {
        if (!isErrorLike(resolvedRevision)) {
            return fetchRepoCommit(resolvedRevision.repo.id, params.revspec)
                .toPromise()
                .then(result => {
                    if (result.data?.node?.__typename === 'Repository') {
                        return { commit: result.data.node.commit, repo: resolvedRevision.repo }
                    }
                    return { commit: null, repo: resolvedRevision.repo }
                })
        }
        return null
    })

    return {
        deferred: {
            commit: commit.then(result => result?.commit ?? null),
            diff: commit.then(result => {
                if (!result?.commit?.oid || !result?.commit.parents[0]?.oid) {
                    return null
                }
                return queryRepositoryComparisonFileDiffs({
                    repo: result.repo.id,
                    base: result.commit?.parents[0].oid,
                    head: result.commit?.oid,
                    paths: [],
                    first: null,
                    after: null,
                }).toPromise()
            }),
        },
    }
}
