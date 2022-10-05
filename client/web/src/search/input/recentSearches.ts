import { useCallback, useEffect, useState } from 'react'

import { Completion, insertCompletionText } from '@codemirror/autocomplete'
import { EditorView } from '@codemirror/view'

import { logger } from '@sourcegraph/common'
import { gql, useLazyQuery } from '@sourcegraph/http-client'
import { StandardSuggestionSource } from '@sourcegraph/search-ui'
import { stringHuman } from '@sourcegraph/shared/src/search/query/printer'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'

import { SearchHistoryEventLogsQueryResult, SearchHistoryEventLogsQueryVariables } from '../../graphql-operations'

const MAX_RECENT_SEARCHES = 20

export const SEARCH_HISTORY_EVENT_LOGS_QUERY = gql`
    query SearchHistoryEventLogsQuery($first: Int!) {
        currentUser {
            __typename
            recentSearchLogs: eventLogs(first: $first, eventName: "SearchResultsQueried") {
                nodes {
                    argument
                    timestamp
                }
            }
        }
    }
`

export function searchHistorySource({
    searches,
    selectedSearchContext,
    onSelection,
}: {
    searches: RecentSearch[] | undefined
    selectedSearchContext?: string
    onSelection: (index: number) => void
}): StandardSuggestionSource {
    return (_context, tokens) => {
        if (tokens.length > 0) {
            return null
        }
        // If there are no tokens we must be at position 0

        try {
            if (!searches) {
                return null
            }

            const createApplyCompletion = (index: number) => (
                view: EditorView,
                completion: Completion,
                from: number,
                to: number
            ) => {
                onSelection(index)
                view.dispatch(insertCompletionText(view.state, completion.label, from, to))
            }

            return {
                from: 0,
                filter: false,
                options: searches
                    .map(
                        (search): Completion => {
                            let query = search.query

                            {
                                const result = scanSearchQuery(search.query)
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
                        }
                    )
                    .filter(completion => completion.label.trim() !== '')
                    .map((completion, index) => {
                        // This is here not in the .map call above so we can use
                        // the correct index after filtering out empty entries
                        completion.apply = createApplyCompletion(index)
                        return completion
                    }),
            }
        } catch {
            return null
        }
    }
}

// Returns all recent searches from temporary settings.
// If no recent searches exist, the temporary settings is initialized with
// the user's recent searches from the event log.
export function useRecentSearches(): {
    recentSearches: RecentSearch[] | undefined
    addRecentSearch: (query: string) => void
    state: 'loading' | 'error' | 'success'
} {
    const [recentSearches, setRecentSearches] = useTemporarySetting('search.input.recentSearches', [])
    const [state, setState] = useState<'loading' | 'error' | 'success'>('loading')

    // If recentSearches from temporary settings is empty, fetch recent searches from the event log
    // and populate temporary settings with that instead.
    const [loadFromEventLog] = useLazyQuery<SearchHistoryEventLogsQueryResult, SearchHistoryEventLogsQueryVariables>(
        SEARCH_HISTORY_EVENT_LOGS_QUERY,
        {
            variables: { first: MAX_RECENT_SEARCHES },
        }
    )

    useEffect(() => {
        if (recentSearches && recentSearches.length === 0) {
            loadFromEventLog()
                .then(result => {
                    if (result.data) {
                        const processedLogs = processEventLogs(result.data)
                        setRecentSearches(processedLogs)
                    }
                    setState('success')
                })
                .catch(() => {
                    logger.error('Error fetching recent searches from event log')
                    setState('error')
                })
        }
    }, [recentSearches, loadFromEventLog, setRecentSearches])

    useEffect(() => {
        if (recentSearches && recentSearches.length > 0) {
            setState('success')
        }
    }, [recentSearches])

    // Adds a new search to the top of the recent searches list.
    // If the search is already in the recent searches list, it moves it to the top.
    const addRecentSearch = useCallback(
        (query: string) => {
            setRecentSearches(recentSearches => {
                const newRecentSearches = recentSearches?.filter(search => search.query !== query) || []
                newRecentSearches.unshift({ query, timestamp: new Date().toISOString() })
                return newRecentSearches
            })
        },
        [setRecentSearches]
    )

    return { recentSearches, addRecentSearch, state }
}

function processEventLogs(data: SearchHistoryEventLogsQueryResult): RecentSearch[] {
    if (data.currentUser?.__typename !== 'User') {
        return []
    }
    const searches = data.currentUser.recentSearchLogs.nodes
        .filter(node => node.argument && node.timestamp)
        .map(node => ({
            // This JSON.parse is safe, silence any TS linting warnings.
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-non-null-assertion
            query: JSON.parse(node.argument!)?.code_search?.query_data?.combined,
            timestamp: node.timestamp,
        }))
        .filter(search => search.query)
        .filter(
            // Remove duplicates
            // Items are sorted by timestamp, so the first item is the most recent.
            // If a search appears earlier in the list, it is a duplicate.
            (search, index, self) => index === self.findIndex(item => item.query === search.query)
        )

    return searches
}
