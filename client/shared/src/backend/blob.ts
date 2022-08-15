import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { ParsedRepoURI, makeRepoURI } from '@sourcegraph/shared/src/util/url'

import { BlobFileFields, BlobResult, BlobVariables, HighlightResponseFormat } from '../graphql-operations'
import { PlatformContext } from '../platform/context'

function fetchBlobCacheKey(parsed: ParsedRepoURI & { disableTimeout?: boolean; format?: string }): string {
    return `${makeRepoURI(parsed)}?disableTimeout=${parsed.disableTimeout}&=${parsed.format}`
}
interface FetchBlobArguments {
    repoName: string
    commitID?: string
    filePath: string
    disableTimeout?: boolean
    format?: HighlightResponseFormat
}

export const fetchBlob = memoizeObservable(
    ({
        requestGraphQL,
        repoName,
        commitID = '',
        filePath,
        disableTimeout = false,
        format = HighlightResponseFormat.HTML_HIGHLIGHT,
    }: FetchBlobArguments & Pick<PlatformContext, 'requestGraphQL'>): Observable<BlobFileFields | null> => {
        // We only want to include HTML data if explicitly requested. We always
        // include LSIF because this is used for languages that are configured
        // to be processed with tree sitter (and is used when explicitly
        // requested via JSON_SCIP).
        const html =
            format === HighlightResponseFormat.HTML_PLAINTEXT || format === HighlightResponseFormat.HTML_HIGHLIGHT

        return requestGraphQL<BlobResult, BlobVariables>({
            request: gql`
                query Blob(
                    $repoName: String!
                    $commitID: String!
                    $filePath: String!
                    $disableTimeout: Boolean!
                    $format: HighlightResponseFormat!
                    $html: Boolean!
                ) {
                    repository(name: $repoName) {
                        commit(rev: $commitID) {
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
            `,
            variables: { repoName, commitID, filePath, disableTimeout, format, html },
            mightContainPrivateInfo: true,
        }).pipe(
            map(dataOrThrowErrors),
            map(data => {
                if (!data.repository?.commit) {
                    throw new Error('Commit not found')
                }
                return data.repository.commit.file
            })
        )
    },
    fetchBlobCacheKey
)
