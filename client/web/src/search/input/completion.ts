import { from } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { StandardSuggestionSource } from '@sourcegraph/search-ui'
import { stringHuman } from '@sourcegraph/shared/src/search/query/printer'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'

import { requestGraphQL } from '../../backend/graphql'
import { SearchHistoryQueryResult, SearchHistoryQueryVariables } from '../../graphql-operations'
import { EventLogResult } from '../backend'

interface RecentSearch {
    count: number
    searchText: string
    timestamp: string
}

const SEARCH_HISTORY_QUERY = gql`
    query SearchHistoryQuery($userId: ID!) {
        node(id: $userId) {
            __typename
            ... on User {
                recentSearchLogs: eventLogs(first: 20, eventName: "SearchResultsQueried") {
                    nodes {
                        argument
                        timestamp
                        url
                    }
                    totalCount
                    pageInfo {
                        hasNextPage
                    }
                }
            }
        }
    }
`

export function searchQueryHistorySource({
    userId,
    selectedSearchContext,
}: {
    userId: string
    selectedSearchContext?: string
}): StandardSuggestionSource {
    return async (_context, tokens) => {
        if (tokens.length > 0) {
            return null
        }
        // If there are no tokens we must be at position 0

        try {
            const searches = processRecentSearches(
                (await from(
                    requestGraphQL<SearchHistoryQueryResult, SearchHistoryQueryVariables>(SEARCH_HISTORY_QUERY, {
                        userId,
                    })
                )
                    .pipe(
                        map(dataOrThrowErrors),
                        map(({ node }) => {
                            if (node?.__typename !== 'User') {
                                return null
                            }
                            return node.recentSearchLogs
                        })
                    )
                    .toPromise()) ?? undefined
            )

            if (!searches) {
                return null
            }

            return {
                from: 0,
                filter: false,
                options: searches
                    .map(search => {
                        let query = search.searchText

                        {
                            const result = scanSearchQuery(search.searchText)
                            if (result.type === 'success') {
                                query = stringHuman(
                                    result.term.filter(term => {
                                        switch (term.type) {
                                            case 'filter':
                                                if (
                                                    term.field.value === 'context' &&
                                                    term.value?.value === selectedSearchContext
                                                ) {
                                                    return false
                                                }
                                                return true
                                            default:
                                                return true
                                        }
                                    })
                                )
                            }
                            // TODO: filter out invalid searches
                        }

                        return {
                            label: query,
                            type: 'searchhistory',
                        }
                    })
                    .filter(item => item.label.trim() !== ''),
            }
        } catch {
            return null
        }
    }
}

function processRecentSearches(eventLogResult?: EventLogResult): RecentSearch[] | null {
    if (!eventLogResult) {
        return null
    }

    const recentSearches: Map<string, RecentSearch> = new Map()

    for (const node of eventLogResult.nodes) {
        if (node.argument) {
            const parsedArguments = JSON.parse(node.argument)
            const searchText: string | undefined = parsedArguments?.code_search?.query_data?.combined

            if (searchText) {
                if (recentSearches.has(searchText)) {
                    recentSearches.get(searchText)!.count += 1
                } else {
                    recentSearches.set(searchText, {
                        count: 1,
                        searchText,
                        timestamp: node.timestamp,
                    })
                }
            }
        }
    }

    return Array.from(recentSearches.values()).sort((a, b) => b.timestamp.localeCompare(a.timestamp))
}
