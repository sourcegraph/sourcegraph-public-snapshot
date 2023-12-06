import { fetchBlobPlaintext } from '$lib/repo/api/blob'
import { fetchDiff } from '$lib/repo/api/commits'
import { fetchTreeEntries } from '$lib/repo/api/tree'
import { findReadme } from '$lib/repo/tree'

import type { PageLoad } from './$types'

export const load: PageLoad = async ({ params, parent, url }) => {
    const revisionToCompare = url.searchParams.get('rev')
    const { resolvedRevision } = await parent()

    const treeEntries = fetchTreeEntries({
        repoID: resolvedRevision.repo.id,
        commitID: resolvedRevision.commitID,
        filePath: params.path,
        first: null,
    }).then(
        commit => commit.tree,
        () => null
    )

    return {
        filePath: params.path,
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
                return fetchBlobPlaintext({
                    repoID: resolvedRevision.repo.id,
                    commitID: resolvedRevision.commitID,
                    filePath: readme.path,
                }).then(result => ({
                    name: readme.name,
                    ...result,
                }))
            }),
            compare: revisionToCompare
                ? {
                      revisionToCompare,
                      diff: fetchDiff(resolvedRevision.repo.id, revisionToCompare, [params.path]),
                  }
                : null,
        },
    }
}
