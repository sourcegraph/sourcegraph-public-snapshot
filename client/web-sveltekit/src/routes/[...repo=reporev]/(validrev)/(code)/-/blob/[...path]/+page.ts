import { fetchHighlight, fetchBlobPlaintext } from '$lib/repo/api/blob'
import { fetchDiff } from '$lib/repo/api/commits'

import type { PageLoad } from './$types'

export const load: PageLoad = async ({ parent, params, url }) => {
    const revisionToCompare = url.searchParams.get('rev')
    const { resolvedRevision } = await parent()

    return {
        deferred: {
            blob: fetchBlobPlaintext({
                filePath: params.path,
                repoID: resolvedRevision.repo.id,
                commitID: resolvedRevision.commitID,
            }),
            highlights: fetchHighlight({
                filePath: params.path,
                repoID: resolvedRevision.repo.id,
                commitID: resolvedRevision.commitID,
            }).then(highlight => highlight?.lsif),
            compare: revisionToCompare
                ? {
                      revisionToCompare,
                      diff: fetchDiff(resolvedRevision.repo.id, revisionToCompare, [params.path]).then(
                          nodes => nodes[0]
                      ),
                  }
                : null,
        },
    }
}
