import { fetchBlobPlaintext } from '$lib/loader/blob'
import { findReadme } from '$lib/repo/tree'

import type { PageLoad } from './$types'

export const load: PageLoad = async ({ parent }) => {
    const { repoName, revision, deferred } = await parent()

    return {
        deferred: {
            ...deferred,
            readme: deferred.fileTree.then(result => {
                const readme = findReadme(result.root.entries)
                if (!readme) {
                    return null
                }
                return fetchBlobPlaintext({ repoName, revision: revision ?? '', filePath: readme.path })
                    .toPromise()
                    .then(result =>
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
