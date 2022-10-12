import { useCallback, useEffect, useState } from 'react'

import { logger } from '@sourcegraph/common'
import { gql, useLazyQuery } from '@sourcegraph/http-client'
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

// Returns all recent searches from temporary settings and a function to add a new search to the list.
// If no recent searches exist, the temporary settings is initialized with
// the user's recent searches from the event log.
export function useRecentSearches(): {
    recentSearches: RecentSearch[] | undefined
    addRecentSearch: (query: string) => void
    state: 'loading' | 'success'
} {
    const [recentSearches, setRecentSearches] = useTemporarySetting('search.input.recentSearches', [])
    const [state, setState] = useState<'loading' | 'success'>('loading')

    // If recentSearches from temporary settings is empty, fetch recent searches from the event log
    // and populate temporary settings with that instead.
    // This is a temporary solution to get users some starting data for this feature.
    // Once this feature is enabled for a while we can remove this.
    const [loadFromEventLog] = useLazyQuery<SearchHistoryEventLogsQueryResult, SearchHistoryEventLogsQueryVariables>(
        SEARCH_HISTORY_EVENT_LOGS_QUERY,
        {
            // Note: It's possible that we end up with less than MAX_RECENT_SEARCHES after removing duplicates.
            // This should be fine, since this is meant to be a starting point
            // for when the user first gets this feature.
            variables: { first: MAX_RECENT_SEARCHES },
        }
    )

    useEffect(() => {
        if (recentSearches) {
            if (recentSearches && recentSearches.length > 0) {
                setState('success')
            } else {
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
                        setState('success') // Ignore the error and use the empty list from temporary settings
                    })
            }
        }
    }, [recentSearches, loadFromEventLog, setRecentSearches])

    // Adds a new search to the top of the recent searches list.
    // If the search is already in the recent searches list, it moves it to the top.
    // If the list is full, the oldest search is removed.
    const addOrMoveRecentSearchToTop = useCallback(
        (recentSearch: RecentSearch) => {
            setRecentSearches(recentSearches => {
                const newRecentSearches = recentSearches?.filter(search => search.query !== recentSearch.query) || []
                newRecentSearches.unshift(recentSearch)
                // Truncate array if it's too long
                if (newRecentSearches.length > MAX_RECENT_SEARCHES) {
                    newRecentSearches.splice(MAX_RECENT_SEARCHES)
                }
                return newRecentSearches
            })
        },
        [setRecentSearches]
    )

    const [pendingAdditions, setPendingAdditions] = useState<RecentSearch[]>([])

    // If the search is being added after the list is finished loading,
    // add it immediately.
    // If the search is being added before the list is finished loading,
    // queue it to be added after loading is complete.
    const addRecentSearch = useCallback(
        (query: string) => {
            const recentSearch = { query, timestamp: new Date().toISOString() }

            if (state === 'success') {
                addOrMoveRecentSearchToTop(recentSearch)
            } else {
                setPendingAdditions(pendingAdditions => pendingAdditions.concat(recentSearch))
            }
        },
        [addOrMoveRecentSearchToTop, state]
    )

    // Process the queue of pending additions after the list is finished loading.
    useEffect(() => {
        if (state === 'success' && pendingAdditions.length > 0) {
            for (const pendingAddition of pendingAdditions) {
                addOrMoveRecentSearchToTop(pendingAddition)
            }
            setPendingAdditions([])
        }
    }, [addOrMoveRecentSearchToTop, pendingAdditions, state])

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
