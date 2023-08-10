import { query, gql, type NodeFromResult } from '$lib/graphql'
import type { BlobResult, BlobVariables, HighlightResult, HighlightVariables, Scalars } from '$lib/graphql-operations'

interface FetchBlobOptions {
    repoID: Scalars['ID']
    commitID: string
    filePath: string
    disableTimeout?: boolean
}

/**
 * Makes sure that default values are applied consistently for the cache key and the `fetchBlob` function.
 */
function applyDefaultValuesToFetchBlobOptions({
    disableTimeout = false,
    ...options
}: FetchBlobOptions): Required<FetchBlobOptions> {
    return {
        ...options,
        disableTimeout,
    }
}

type Highlight = NonNullable<
    NonNullable<NodeFromResult<HighlightResult['node'], 'Repository'>['commit']>['blob']
>['highlight']

export async function fetchHighlight(options: FetchBlobOptions): Promise<Highlight | null> {
    const { repoID, commitID, filePath, disableTimeout } = applyDefaultValuesToFetchBlobOptions(options)

    const data = await query<HighlightResult, HighlightVariables>(
        gql`
            query Highlight($repoID: ID!, $commitID: String!, $filePath: String!, $disableTimeout: Boolean!) {
                node(id: $repoID) {
                    __typename
                    id
                    ... on Repository {
                        commit(rev: $commitID) {
                            id
                            blob(path: $filePath) {
                                canonicalURL
                                highlight(disableTimeout: $disableTimeout, format: JSON_SCIP) {
                                    aborted
                                    lsif
                                }
                            }
                        }
                    }
                }
            }
        `,
        { repoID, commitID, filePath, disableTimeout }
    )

    if (data.node?.__typename !== 'Repository' || !data.node?.commit) {
        throw new Error('Commit not found')
    }

    return data.node.commit.blob?.highlight ?? null
}

export type BlobFileFields = NonNullable<
    NonNullable<NodeFromResult<BlobResult['node'], 'Repository'>['commit']>['blob']
>

export async function fetchBlobPlaintext(options: FetchBlobOptions): Promise<BlobFileFields | null> {
    const { repoID, commitID, filePath } = applyDefaultValuesToFetchBlobOptions(options)

    const data = await query<BlobResult, BlobVariables>(
        gql`
            query Blob($repoID: ID!, $commitID: String!, $filePath: String!) {
                node(id: $repoID) {
                    __typename
                    id
                    ... on Repository {
                        commit(rev: $commitID) {
                            id
                            blob(path: $filePath) {
                                canonicalURL
                                content
                                richHTML
                                lfs {
                                    byteSize
                                }
                                externalURLs {
                                    url
                                    serviceKind
                                }
                            }
                        }
                    }
                }
            }
        `,
        { repoID, commitID, filePath }
    )

    if (data.node?.__typename !== 'Repository' || !data.node?.commit?.blob) {
        throw new Error('Commit or file not found')
    }

    return data.node.commit.blob
}
