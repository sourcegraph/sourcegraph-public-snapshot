import React, { useMemo } from 'react'
import { useHistory } from 'react-router'
import { EMPTY, Subject } from 'rxjs'
import { mapTo, tap, throttleTime } from 'rxjs/operators'

import { Range } from '@sourcegraph/extension-api-types'
import { useDebounce } from '@sourcegraph/wildcard'

import { TextDocumentData } from '../../../api/viewerTypes'
import { CommitSymbolsResult, CommitSymbolsVariables, Maybe, SymbolKind } from '../../../graphql-operations'
import { gql } from '../../../graphql/graphql'
import { PlatformContext, PlatformContextProps } from '../../../platform/context'
import { SymbolIcon } from '../../../symbols/SymbolIcon'
import { addLineRangeQueryParameter, parseRepoURI, toPositionOrRangeQueryParameter } from '../../../util/url'
import { useObservable } from '../../../util/useObservable'

import { Message } from './Message'
import { NavigableList } from './NavigableList'
import listStyles from './NavigableList.module.scss'

interface JumpToSymbolResultProps extends PlatformContextProps<'requestGraphQL'> {
    value: string
    onClick: () => void
    textDocumentData: TextDocumentData | null | undefined
}

const COMMIT_SYMBOLS_QUERY = gql`
    query CommitSymbols(
        $repo: String!
        $revision: String!
        $first: Int!
        $query: String!
        $includePatterns: [String!]
    ) {
        repository(name: $repo) {
            commit(rev: $revision) {
                symbols(first: $first, query: $query, includePatterns: $includePatterns) {
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
    { repo, revision, filePath, query }: { repo: string; revision: string; filePath: string; query: string },
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
) {
    const result = platformContext.requestGraphQL<CommitSymbolsResult, CommitSymbolsVariables>({
        request: COMMIT_SYMBOLS_QUERY,
        variables: {
            first: 25, // TODO infinite scroll. For now, narrowing results with search query should be good enough
            repo,
            revision,
            query,
            includePatterns: [filePath],
        },
        mightContainPrivateInfo: true,
    })

    return result
}

interface SymbolNode {
    name: string
    containerName: Maybe<string>
    kind: SymbolKind
    language: string
    location: {
        resource: { path: string }
        range: Maybe<{
            start: { line: number; character: number }
            end: { line: number; character: number }
        }>
    }
}

export const JumpToSymbolResult: React.FC<JumpToSymbolResultProps> = ({
    value,
    onClick,
    textDocumentData,
    platformContext,
}) => {
    const history = useHistory()

    const parsedURI = useMemo(() => (textDocumentData ? parseRepoURI(textDocumentData.uri) : null), [textDocumentData])

    const debouncedValue = useDebounce(value, 300)

    // TODO: cache, shared with sidebar?
    const commitSymbolsResult = useObservable(
        useMemo(
            () =>
                parsedURI
                    ? fetchCommitSymbols(
                          {
                              repo: parsedURI.repoName,
                              revision: parsedURI.revision ?? '',
                              filePath: '^' + (parsedURI.filePath ?? '') + '$', // Only include symbols from this file, not similar files.
                              // TODO: use value for search query
                              query: debouncedValue,
                          },
                          platformContext
                      )
                    : EMPTY,
            [debouncedValue, parsedURI, platformContext]
        )
    )

    const rangeUpdates = useMemo(() => new Subject<Range>(), [])
    useObservable(
        useMemo(
            () =>
                rangeUpdates.pipe(
                    throttleTime(150, undefined, { leading: true, trailing: true }),
                    tap(range => {
                        // TODO: abstract for bext. Could just jump to line but with symbol hints.

                        const searchParameters = addLineRangeQueryParameter(
                            new URLSearchParams(history.location.search),
                            toPositionOrRangeQueryParameter({
                                range,
                            })
                        )
                        console.log('focused symbol', { searchParameters })
                        history.replace({
                            ...history.location,
                            search: searchParameters.toString(),
                        })
                    }),
                    mapTo(undefined)
                ),
            [rangeUpdates, history]
        )
    )

    if (!textDocumentData) {
        return <Message>Navigate to a text document to jump to symbol</Message>
    }

    const onSymbolFocus = (symbol: SymbolNode): void => {
        const range = symbol.location.range
        if (range) {
            // Convert to 1-indexed for code view
            const adjustedRange: Range = {
                start: {
                    line: range.start.line + 1,
                    character: range.start.character + 1,
                },
                end: {
                    line: range.end.line + 1,
                    character: range.end.character + 1,
                },
            }
            rangeUpdates.next(adjustedRange)
        }
    }

    const onSymbolClick = (symbol: SymbolNode): void => {
        onSymbolFocus(symbol)
        onClick()
    }

    // TODO: directory page symbol search (in directory)?
    // TODO: error handling

    if (commitSymbolsResult === undefined) {
        return <Message>Downloading....</Message>
    }

    const symbols = commitSymbolsResult?.data?.repository?.commit?.symbols.nodes ?? []

    return (
        <div>
            <NavigableList items={[null, ...symbols]}>
                {/* First no-op item so that location isn't changed without user input */}
                {(symbol, { active }) => {
                    if (symbol === null) {
                        return (
                            <NavigableList.Item active={active}>
                                <span className={listStyles.itemContainer}>
                                    <SymbolIcon kind={SymbolKind.FILE} className="icon-inline mr-2" />
                                    {parsedURI?.filePath ? `Symbols in ${parsedURI.filePath}` : 'Symbol results'}
                                </span>
                            </NavigableList.Item>
                        )
                    }

                    return (
                        <NavigableList.Item
                            onFocus={() => onSymbolFocus(symbol)}
                            onClick={() => onSymbolClick(symbol)}
                            active={active}
                        >
                            <span className={listStyles.itemContainer}>
                                <SymbolIcon kind={symbol.kind} className="icon-inline mr-2" />
                                {symbol.name}
                            </span>
                        </NavigableList.Item>
                    )
                }}
            </NavigableList>
        </div>
    )
}
