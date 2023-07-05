import { writable } from 'svelte/store'

import { isErrorLike } from '$lib/common'
import type { BlobFileFields } from '$lib/graphql-operations'
import { fetchHighlight, fetchBlobPlaintext } from '$lib/loader/blob'

interface BlobStoreParams {
    filePath: string
    repoName: string
    revision: string
}

/**
 * Helper store for coordinating loading blob and highlighting data.
 */
export function createBlobStore() {
    let currentParams: BlobStoreParams | null = null
    let { subscribe, set, update } = writable<
        { blob?: BlobFileFields | null; highlights?: string; loading: false } | { loading: true; highlights?: string }
    >({ loading: false })

    return {
        subscribe,
        fetch: (params: BlobStoreParams | null) => {
            if (params === null) {
                set({ loading: false })
            } else if (
                currentParams?.filePath !== params.filePath ||
                currentParams.repoName !== params.repoName ||
                currentParams.revision !== params.revision
            ) {
                currentParams = params
                set({ loading: true })
                fetchBlobPlaintext(params)
                    .toPromise()
                    .then(result => {
                        if (params === currentParams && result && !isErrorLike(result)) {
                            update(blobInfo => ({ ...blobInfo, loading: false, blob: result }))
                        }
                    })
                fetchHighlight(params)
                    .toPromise()
                    .then(result => {
                        if (params === currentParams && result && !isErrorLike(result)) {
                            update(blobInfo => ({ ...blobInfo, highlights: result.lsif }))
                        }
                    })
            }
        },
    }
}
