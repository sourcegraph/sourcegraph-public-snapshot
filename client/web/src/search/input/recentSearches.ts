import { useCallback, useEffect, useMemo } from 'react'

import { Completion, insertCompletionText } from '@codemirror/autocomplete'
import { EditorView } from '@codemirror/view'
import { from, Observable, of } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { logger } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { StandardSuggestionSource } from '@sourcegraph/search-ui'
import { stringHuman } from '@sourcegraph/shared/src/search/query/printer'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { useObservable } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../backend/graphql'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import { SearchHistoryEventLogsQueryResult, SearchHistoryEventLogsQueryVariables } from '../../graphql-operations'

const MAX_RECENT_SEARCHES = 20

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
} {
    const [showSearchHistory] = useFeatureFlag('search-input-show-history')
    const [recentSearches, setRecentSearches] = useTemporarySetting('search.input.recentSearches', [])

    // If recentSearches from temporary settings is empty, fetch recent searches from the event log
    // and populate temporary settings with that instead.
    const recentSearchesFromEventLog = useObservable(
        useMemo(
            () =>
                showSearchHistory && recentSearches && recentSearches.length === 0
                    ? getRecentSearchesFromEventLog()
                    : of(null),
            [recentSearches, showSearchHistory]
        )
    )

    useEffect(() => {
        if (recentSearchesFromEventLog && recentSearches && recentSearches.length === 0) {
            setRecentSearches(recentSearchesFromEventLog)
        }
    }, [recentSearches, recentSearchesFromEventLog, setRecentSearches])

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

    return { recentSearches: showSearchHistory ? recentSearches : [], addRecentSearch }
}

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

function getRecentSearchesFromEventLog(): Observable<RecentSearch[] | null> {
    return from(
        requestGraphQL<SearchHistoryEventLogsQueryResult, SearchHistoryEventLogsQueryVariables>(
            SEARCH_HISTORY_EVENT_LOGS_QUERY,
            {
                first: MAX_RECENT_SEARCHES,
            }
        )
    ).pipe(
        map(dataOrThrowErrors),
        map(({ currentUser }) => {
            if (currentUser?.__typename !== 'User') {
                return []
            }
            const searches = currentUser.recentSearchLogs.nodes
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
        }),
        catchError(error => {
            logger.error(error)
            return []
        })
    )
}
