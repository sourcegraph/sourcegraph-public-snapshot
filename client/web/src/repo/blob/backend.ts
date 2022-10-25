import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'

import { requestGraphQL } from '../../backend/graphql'
import {
    BlobFileFields,
    BlobResult,
    BlobStencilFields,
    BlobVariables,
    HighlightResponseFormat,
} from '../../graphql-operations'
import { useExperimentalFeatures } from '../../stores'

/**
 * Makes sure that default values are applied consistently for the cache key and the `fetchBlob` function.
 */
const applyDefaultValuesToFetchBlobOptions = ({
    disableTimeout = false,
    stencil = false,
    format = HighlightResponseFormat.HTML_HIGHLIGHT,
    ...options
}: FetchBlobOptions): Required<FetchBlobOptions> => ({
    ...options,
    disableTimeout,
    format,
    stencil,
})

function fetchBlobCacheKey(options: FetchBlobOptions): string {
    const { disableTimeout, format, stencil } = applyDefaultValuesToFetchBlobOptions(options)

    return `${makeRepoURI(options)}?disableTimeout=${disableTimeout}&=${format}&stencil=${stencil}`
}

interface FetchBlobOptions {
    repoName: string
    revision: string
    filePath: string
    disableTimeout?: boolean
    format?: HighlightResponseFormat
    stencil?: boolean
}

interface FetchBlobResponse {
    blob: BlobFileFields | null
    stencil?: BlobStencilFields[]
}

export const fetchBlob = memoizeObservable((options: FetchBlobOptions): Observable<FetchBlobResponse> => {
    const { repoName, revision, filePath, disableTimeout, format, stencil } = applyDefaultValuesToFetchBlobOptions(
        options
    )

    // We only want to include HTML data if explicitly requested. We always
    // include LSIF because this is used for languages that are configured
    // to be processed with tree sitter (and is used when explicitly
    // requested via JSON_SCIP).
    const html = [HighlightResponseFormat.HTML_PLAINTEXT, HighlightResponseFormat.HTML_HIGHLIGHT].includes(format)

    return requestGraphQL<BlobResult, BlobVariables>(
        gql`
            query Blob(
                $repoName: String!
                $revision: String!
                $filePath: String!
                $disableTimeout: Boolean!
                $format: HighlightResponseFormat!
                $html: Boolean!
                $stencil: Boolean!
            ) {
                repository(name: $repoName) {
                    commit(rev: $revision) {
                        blob(path: $filePath) @include(if: $stencil) {
                            lsif {
                                stencil {
                                    ...BlobStencilFields
                                }
                            }
                        }
                        file(path: $filePath) {
                            ...BlobFileFields
                        }
                    }
                }
            }

            fragment BlobFileFields on File2 {
                content
                richHTML
                highlight(disableTimeout: $disableTimeout, format: $format) {
                    aborted
                    html @include(if: $html)
                    lsif
                }
            }

            fragment BlobStencilFields on Range {
                start {
                    line
                    character
                }
                end {
                    line
                    character
                }
            }
        `,
        { repoName, revision, filePath, disableTimeout, format, html, stencil }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.repository?.commit) {
                throw new Error('Commit not found')
            }

            return {
                blob: data.repository.commit.file,
                stencil: data.repository.commit.blob?.lsif?.stencil,
            }
        })
    )
}, fetchBlobCacheKey)

/**
 * Returns the preferred blob prefetch format.
 *
 * Note: This format should match the format used when the blob is 'normally' fetched. E.g. in `BlobPage.tsx`.
 */
export const usePrefetchBlobFormat = (): HighlightResponseFormat => {
    const enableCodeMirror = useExperimentalFeatures(features => features.enableCodeMirrorFileView ?? false)
    const enableLazyHighlighting = useExperimentalFeatures(
        features => features.enableLazyBlobSyntaxHighlighting ?? false
    )

    /**
     * Highlighted blobs (Fast)
     *
     * TODO: For large files, `PLAINTEXT` can still be faster, this is another potential UX improvement.
     * Outstanding issue before this can be enabled: https://github.com/sourcegraph/sourcegraph/issues/41413
     */
    if (enableCodeMirror) {
        return HighlightResponseFormat.JSON_SCIP
    }

    /**
     * Plaintext blobs (Fast)
     */
    if (enableLazyHighlighting) {
        return HighlightResponseFormat.HTML_PLAINTEXT
    }

    /**
     * Highlighted blobs (Slow)
     */
    return HighlightResponseFormat.HTML_HIGHLIGHT
}
