import { fetchBlobPlaintext } from '$lib/repo/api/blob'
import { findReadme } from '$lib/repo/tree'

import type { PageLoad } from './$types'

export const load: PageLoad = async ({ parent }) => {
    const { resolvedRevision, deferred } = await parent()

    return {
        deferred: {
            ...deferred,
            readme: deferred.fileTree.then(result => {
                const readme = findReadme(result.root.entries)
                if (!readme) {
                    return null
                }
                return fetchBlobPlaintext({
                    repoID: resolvedRevision.repo.id,
                    commitID: resolvedRevision.commitID,
                    filePath: readme.path,
                }).then(result =>
                    result
                        ? {
                              name: readme.name,
                              content: result.content,
                              richHTML: result.richHTML,
                          }
                        : null
                )
            }),
        },
    }
}
