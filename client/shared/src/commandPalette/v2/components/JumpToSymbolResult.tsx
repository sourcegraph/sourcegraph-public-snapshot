import React, { useMemo } from 'react'
import { EMPTY } from 'rxjs'

import { TextDocumentData } from '../../../api/viewerTypes'
import { CommitSymbolsResult, CommitSymbolsVariables } from '../../../graphql-operations'
import { gql } from '../../../graphql/graphql'
import { PlatformContext, PlatformContextProps } from '../../../platform/context'
import { parseRepoURI } from '../../../util/url'
import { useObservable } from '../../../util/useObservable'

import { Message } from './Message'

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

function fetchCommitSymbols(
    { repo, revision, filePath }: { repo: string; revision: string; filePath: string },
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
) {
    const result = platformContext.requestGraphQL<CommitSymbolsResult, CommitSymbolsVariables>({
        request: COMMIT_SYMBOLS_QUERY,
        variables: {
            first: 200,
            repo,
            revision,
            includePatterns: [filePath],
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
    const parsedURI = useMemo(() => (textDocumentData ? parseRepoURI(textDocumentData.uri) : null), [textDocumentData])

    const commitSymbols = useObservable(
        useMemo(
            () =>
                parsedURI
                    ? fetchCommitSymbols(
                          {
                              repo: parsedURI.repoName,
                              revision: parsedURI.revision ?? '',
                              filePath: parsedURI.filePath ?? '',
                          },
                          platformContext
                      )
                    : EMPTY,
            [parsedURI, platformContext]
        )
    )
    console.log({ parsedURI, commitSymbols })
    // Toggle whole repo symbol search? Wouldn't "jump preview" in that case though due to file loading times.
    if (!textDocumentData) {
        return <Message type="muted">Open a text document to jump to symbol</Message>
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
