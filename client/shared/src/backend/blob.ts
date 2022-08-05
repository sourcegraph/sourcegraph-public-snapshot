import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { ParsedRepoURI, makeRepoURI } from '@sourcegraph/shared/src/util/url'

import { BlobFileFields, BlobResult, BlobVariables } from '../graphql-operations'
import { PlatformContext } from '../platform/context'

function fetchBlobCacheKey(parsed: ParsedRepoURI & { disableTimeout?: boolean; formatOnly?: boolean }): string {
    return `${makeRepoURI(parsed)}?disableTimeout=${parsed.disableTimeout}&formatOnly=${parsed.formatOnly}`
}

interface FetchBlobArguments {
    repoName: string
    commitID: string
    filePath: string
    disableTimeout?: boolean
    formatOnly?: boolean
}

export const fetchBlob = memoizeObservable(
    ({
        requestGraphQL,
        repoName,
        commitID,
        filePath,
        disableTimeout = false,
        formatOnly = false,
    }: FetchBlobArguments & Pick<PlatformContext, 'requestGraphQL'>): Observable<BlobFileFields | null> =>
        requestGraphQL<BlobResult, BlobVariables>({
            request: gql`
                query Blob(
                    $repoName: String!
                    $commitID: String!
                    $filePath: String!
                    $disableTimeout: Boolean!
                    $formatOnly: Boolean!
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
                    highlight(disableTimeout: $disableTimeout, formatOnly: $formatOnly) {
                        aborted
                        html
                        lsif
                    }
                }
            `,
            variables: { repoName, commitID, filePath, disableTimeout, formatOnly },
            mightContainPrivateInfo: true,
        }).pipe(
            map(dataOrThrowErrors),
            map(data => {
                if (!data.repository?.commit) {
                    throw new Error('Commit not found')
                }
                return data.repository.commit.file
            })
        ),
    fetchBlobCacheKey
)
