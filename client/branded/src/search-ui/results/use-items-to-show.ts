import { useCallback, useEffect, useState } from 'react'

export const DEFAULT_INITIAL_ITEMS_TO_SHOW = 15
export const INCREMENTAL_ITEMS_TO_SHOW = 10

/**
 * Used to save number of search items user saw previously
 * to restore scroll position on page-refresh or navigating back to search results.
 */
enum SessionStorageKeys {
    itemsToShow = 'search-items-to-show',
    query = 'search-items-to-show-for-query',
}

/**
 * If a user browsed search results for the same query and scrolled through them ->
 * return the number of items he saw to restore scroll position without multiple UI jumps.
 *
 * Otherwise, return the initial default value.
 */
function getInitialItemsToShowValue(query: string): number {
    const previousQuery = sessionStorage.getItem(SessionStorageKeys.query)
    const itemsToShow = sessionStorage.getItem(SessionStorageKeys.itemsToShow)

    // In case of receiving new query — reset the state saved in sessionStorage.
    if (previousQuery !== query) {
        sessionStorage.setItem(SessionStorageKeys.query, query)
        sessionStorage.setItem(SessionStorageKeys.itemsToShow, String(DEFAULT_INITIAL_ITEMS_TO_SHOW))

        return DEFAULT_INITIAL_ITEMS_TO_SHOW
    }

    return Number(itemsToShow) || DEFAULT_INITIAL_ITEMS_TO_SHOW
}

interface UseItemsToShowResult {
    itemsToShow: number
    handleBottomHit: () => void
}

/**
 * Preserves the number of items to show for the query by saving it in the session storage.
 */
export function useItemsToShow(query: string, resultsNumber: number): UseItemsToShowResult {
    const [itemsToShow, setItemsToShow] = useState(() => getInitialItemsToShowValue(query))

    const handleBottomHit = useCallback(
        () =>
            setItemsToShow(items => {
                const itemsToShow = Math.min(resultsNumber, items + INCREMENTAL_ITEMS_TO_SHOW)
                sessionStorage.setItem(SessionStorageKeys.itemsToShow, String(itemsToShow))

                return itemsToShow
            }),
        [resultsNumber]
    )

    // Reset scroll visibility state when new search is started
    useEffect(() => {
        setItemsToShow(getInitialItemsToShowValue(query))
    }, [query])

    return {
        itemsToShow,
        handleBottomHit,
    }
}
