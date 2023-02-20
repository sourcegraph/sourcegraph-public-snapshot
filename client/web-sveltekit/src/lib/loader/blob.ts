import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

export { fetchHighlightedFileLineRanges } from '@sourcegraph/shared/src/backend/file'

import { memoizeObservable } from '$lib/common'
import type {
    BlobFileFields,
    HighlightingFields,
    BlobResult,
    BlobVariables,
    HighlightResult,
    HighlightVariables,
} from '$lib/graphql-operations'
import { dataOrThrowErrors, gql } from '$lib/http-client'
import { makeRepoURI } from '$lib/shared'
import { requestGraphQL } from '$lib/web'

interface FetchBlobOptions {
    repoName: string
    revision: string
    filePath: string
    disableTimeout?: boolean
}

/**
 * Makes sure that default values are applied consistently for the cache key and the `fetchBlob` function.
 */
const applyDefaultValuesToFetchBlobOptions = ({
    disableTimeout = false,
    ...options
}: FetchBlobOptions): Required<FetchBlobOptions> => ({
    ...options,
    disableTimeout,
})

function fetchBlobCacheKey(options: FetchBlobOptions): string {
    const { disableTimeout } = applyDefaultValuesToFetchBlobOptions(options)

    return `${makeRepoURI(options)}?disableTimeout=${disableTimeout}`
}

export const fetchHighlight = memoizeObservable(
    (options: FetchBlobOptions): Observable<HighlightingFields['highlight'] | null> => {
        const { repoName, revision, filePath, disableTimeout } = applyDefaultValuesToFetchBlobOptions(options)

        return requestGraphQL<HighlightResult, HighlightVariables>(
            gql`
                query Highlight($repoName: String!, $revision: String!, $filePath: String!, $disableTimeout: Boolean!) {
                    repository(name: $repoName) {
                        commit(rev: $revision) {
                            file(path: $filePath) {
                                ...HighlightingFields
                            }
                        }
                    }
                }

                fragment HighlightingFields on File2 {
                    __typename
                    highlight(disableTimeout: $disableTimeout, format: JSON_SCIP) {
                        aborted
                        lsif
                    }
                }
            `,
            { repoName, revision, filePath, disableTimeout }
        ).pipe(
            map(dataOrThrowErrors),
            map(data => {
                if (!data.repository?.commit) {
                    throw new Error('Commit not found')
                }

                return data.repository.commit.file?.highlight ?? null
            })
        )
    },
    fetchBlobCacheKey
)

export const fetchBlobPlaintext = memoizeObservable((options: FetchBlobOptions): Observable<BlobFileFields | null> => {
    const { repoName, revision, filePath } = applyDefaultValuesToFetchBlobOptions(options)

    return requestGraphQL<BlobResult, BlobVariables>(
        gql`
            query Blob($repoName: String!, $revision: String!, $filePath: String!) {
                repository(name: $repoName) {
                    commit(rev: $revision) {
                        file(path: $filePath) {
                            ...BlobFileFields
                        }
                    }
                }
            }

            fragment BlobFileFields on File2 {
                __typename
                content
                richHTML
                ... on GitBlob {
                    lfs {
                        byteSize
                    }
                    externalURLs {
                        url
                        serviceKind
                    }
                }
            }
        `,
        { repoName, revision, filePath }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.repository?.commit) {
                throw new Error('Commit not found')
            }

            return data.repository.commit.file
        })
    )
}, fetchBlobCacheKey)
