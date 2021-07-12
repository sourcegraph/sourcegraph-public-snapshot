import { useContext, useEffect, useMemo, useState } from 'react'

import { useDebounce } from '@sourcegraph/wildcard/src'

import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { RepositorySuggestion } from '../../../../core/backend/requests/fetch-repository-suggestions'
import { useDistinctValue } from '../../../../hooks/use-distinct-value'
import { memoizeAsync } from '../utils/memoize-async'

interface UseRepoSuggestionsProps {
    search: string | null
    disable?: boolean
    excludedItems?: string[]
}

interface UseRepoSuggestionsResult {
    searchValue: string | null
    suggestions: RepositorySuggestion[] | Error | undefined
}

/**
 * Returns fetch method for repository suggestions with local cache
 */
function useFetchSuggestions(): (search: string) => Promise<RepositorySuggestion[]> {
    const { getRepositorySuggestions } = useContext(InsightsApiContext)

    return useMemo(
        // memoizeAsync adds local result cache
        () => memoizeAsync<string, RepositorySuggestion[]>(getRepositorySuggestions, query => query),
        [getRepositorySuggestions]
    )
}

/**
 * Provides list of repository suggestions.
 */
export function useRepoSuggestions(props: UseRepoSuggestionsProps): UseRepoSuggestionsResult {
    const { search, disable = false, excludedItems = [] } = props

    const [suggestions, setSuggestions] = useState<RepositorySuggestion[] | Error | undefined>([])
    const debouncedSearchTerm = useDebounce(search, 1000)
    const fetchSuggestions = useFetchSuggestions()

    // To not trigger use effect with fetching each render call
    // we compare prev and next value for excludedItems and return
    // prev value if value wasn't changed
    const distinctExcludedItems = useDistinctValue(excludedItems)

    useEffect(() => {
        if (disable || !debouncedSearchTerm) {
            setSuggestions([])
            return
        }

        let wasCanceled = false

        // Start fetching repository suggestions
        setSuggestions(undefined)

        fetchSuggestions(debouncedSearchTerm)
            .then(suggestions => {
                if (!wasCanceled) {
                    setSuggestions(suggestions.filter(suggestion => !distinctExcludedItems.includes(suggestion.name)))
                }
            })
            .catch(error => {
                if (!wasCanceled) {
                    setSuggestions(error)
                }
            })

        return () => {
            wasCanceled = true
        }
    }, [distinctExcludedItems, fetchSuggestions, disable, debouncedSearchTerm])

    return { searchValue: debouncedSearchTerm, suggestions }
}
