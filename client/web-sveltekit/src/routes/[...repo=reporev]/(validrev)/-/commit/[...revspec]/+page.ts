import { fetchRepoCommit, queryRepositoryComparisonFileDiffs } from '$lib/repo/api/commits'

import type { PageLoad } from './$types'

export const load: PageLoad = async ({ parent, params }) => {
    const { resolvedRevision } = await parent()
    const commit = fetchRepoCommit(resolvedRevision.repo.id, params.revspec).then(data => {
        if (data?.node?.__typename === 'Repository') {
            return { commit: data.node.commit, repo: resolvedRevision.repo }
        }
        return { commit: null, repo: resolvedRevision.repo }
    })

    return {
        deferred: {
            commit: commit.then(result => result?.commit ?? null),
            diff: commit.then(result => {
                if (!result.commit?.oid || !result.commit.parents[0]?.oid) {
                    return null
                }
                return queryRepositoryComparisonFileDiffs({
                    repo: result.repo.id,
                    base: result.commit?.parents[0].oid,
                    head: result.commit?.oid,
                    paths: [],
                    first: null,
                    after: null,
                })
            }),
        },
    }
}
