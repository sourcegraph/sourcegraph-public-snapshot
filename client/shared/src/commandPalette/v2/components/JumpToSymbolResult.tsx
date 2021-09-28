import React, { useMemo } from 'react'

import { TextDocumentData } from '../../../api/viewerTypes'
import { CommitSymbolsResult, CommitSymbolsVariables } from '../../../graphql-operations'
import { gql } from '../../../graphql/graphql'
import { PlatformContext, PlatformContextProps } from '../../../platform/context'
import { useObservable } from '../../../util/useObservable'

interface JumpToSymbolResultProps extends PlatformContextProps<'requestGraphQL'> {
    value: string
    onClick: () => void
    textDocumentData: TextDocumentData | null | undefined
}

const COMMIT_SYMBOLS_QUERY = gql`
    query CommitSymbols($repo: String!, $revision: String!, $first: Int!, $includePatterns: [String!]) {
        repository(name: $repo) {
            commit(rev: $revision) {
                symbols(first: $first, query: "", includePatterns: $includePatterns) {
                    __typename
                    pageInfo {
                        hasNextPage
                    }
                    nodes {
                        name
                        containerName
                        kind
                        language
                        location {
                            resource {
                                path
                            }
                            range {
                                start {
                                    line
                                    character
                                }
                                end {
                                    line
                                    character
                                }
                            }
                        }
                    }
                }
            }
        }
    }
`

function fetchCommitSymbols(platformContext: Pick<PlatformContext, 'requestGraphQL'>) {
    const result = platformContext.requestGraphQL<CommitSymbolsResult, CommitSymbolsVariables>({
        request: COMMIT_SYMBOLS_QUERY,
        variables: {
            first: 10,
            repo: 'github.com/sourcegraph/sourcegraph',
            includePatterns: ['client/extension-api/src/sourcegraph.d.ts'],
            revision: 'main',
        },
        mightContainPrivateInfo: true,
    })

    return result
}

export const JumpToSymbolResult: React.FC<JumpToSymbolResultProps> = ({
    value,
    onClick,
    textDocumentData,
    platformContext,
}) => {
    console.log('TODO')

    const commitSymbols = useObservable(useMemo(() => fetchCommitSymbols(platformContext), [platformContext]))
    console.log({ commitSymbols })
    // Toggle whole repo symbol search? Wouldn't "jump preview" in that case though due to file loading times.
    if (!textDocumentData) {
        return (
            <div>
                <h3>Open a text document to jump to symbol</h3>
            </div>
        )
    }
    // Can apollo cache this/share request w/ symbols sidebar?

    // Fetch a large number of symbols and do client side fuzzy search

    // If there's no next page when query is empty, don't issue a new request, just fuzzy search over existing results.

    return (
        <div>
            <h1>{value}</h1>
        </div>
    )
}
