import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'

import { requestGraphQL } from '../../backend/graphql'
import { BlobFileFields, BlobResult, BlobVariables, HighlightResponseFormat } from '../../graphql-operations'

/**
 * Makes sure that default values are applied consistently for the cache key and the `fetchBlob` function.
 */
const applyDefaultValuesToFetchBlobOptions = ({
    disableTimeout = false,
    format = HighlightResponseFormat.HTML_HIGHLIGHT,
    startLine = null,
    endLine = null,
    visibleIndexID = null,
    scipSnapshot = false,
    ...options
}: FetchBlobOptions): Required<FetchBlobOptions> => ({
    ...options,
    disableTimeout,
    format,
    startLine,
    endLine,
    visibleIndexID,
    scipSnapshot,
})

function fetchBlobCacheKey(options: FetchBlobOptions): string {
    const { disableTimeout, format, scipSnapshot, visibleIndexID } = applyDefaultValuesToFetchBlobOptions(options)

    return `${makeRepoURI(
        options
    )}?disableTimeout=${disableTimeout}&=${format}&snap=${scipSnapshot}&visible=${visibleIndexID}`
}

interface FetchBlobOptions {
    repoName: string
    revision: string
    filePath: string
    disableTimeout?: boolean
    format?: HighlightResponseFormat
    startLine?: number | null
    endLine?: number | null
    scipSnapshot?: boolean
    visibleIndexID?: string | null
}

export const fetchBlob = memoizeObservable(
    (
        options: FetchBlobOptions
    ): Observable<(BlobFileFields & { snapshot?: { offset: number; data: string }[] | null }) | null> => {
        const {
            repoName,
            revision,
            filePath,
            disableTimeout,
            format,
            startLine,
            endLine,
            scipSnapshot,
            visibleIndexID,
        } = applyDefaultValuesToFetchBlobOptions(options)

        return requestGraphQL<BlobResult, BlobVariables>(
            gql`
                query Blob(
                    $repoName: String!
                    $revision: String!
                    $filePath: String!
                    $disableTimeout: Boolean!
                    $format: HighlightResponseFormat!
                    $startLine: Int
                    $endLine: Int
                    $snapshot: Boolean!
                    $visibleIndexID: ID!
                ) {
                    repository(name: $repoName) {
                        commit(rev: $revision) {
                            file(path: $filePath) {
                                ...BlobFileFields
                            }
                            blob(path: $filePath) @include(if: $snapshot) {
                                lsif {
                                    snapshot(indexID: $visibleIndexID) {
                                        offset
                                        data
                                    }
                                }
                            }
                        }
                    }
                }

                fragment BlobFileFields on File2 {
                    __typename
                    content(startLine: $startLine, endLine: $endLine)
                    richHTML(startLine: $startLine, endLine: $endLine)
                    highlight(
                        disableTimeout: $disableTimeout
                        format: $format
                        startLine: $startLine
                        endLine: $endLine
                    ) {
                        aborted
                        lsif
                    }
                    totalLines
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
            {
                repoName,
                revision,
                filePath,
                disableTimeout,
                format,
                startLine,
                endLine,
                snapshot: scipSnapshot,
                visibleIndexID: visibleIndexID ?? '',
            }
        ).pipe(
            map(dataOrThrowErrors),
            map(data => {
                if (!data.repository?.commit) {
                    throw new Error('Commit not found')
                }

                if (!data.repository.commit.file) {
                    throw new Error('File not found')
                }

                return {
                    ...data.repository.commit.file,
                    snapshot: data.repository.commit.blob?.lsif?.snapshot,
                }
            })
        )
    },
    fetchBlobCacheKey
)
