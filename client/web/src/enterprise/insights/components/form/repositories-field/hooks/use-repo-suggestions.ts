import { useContext, useEffect, useMemo, useState } from 'react'

import { useDebounce } from '@sourcegraph/wildcard'

import { CodeInsightsBackendContext } from '../../../../core/backend/code-insights-backend-context'
import { RepositorySuggestionData } from '../../../../core/backend/code-insights-backend-types'
import { memoizeAsync } from '../utils/memoize-async'

interface UseRepoSuggestionsProps {
    search: string | null
    disable?: boolean
}

interface UseRepoSuggestionsResult {
    searchValue: string | null
    suggestions: RepositorySuggestionData[] | Error | undefined
}

/**
 * Returns fetch method for repository suggestions with local cache
 */
function useFetchSuggestions(): (search: string) => Promise<RepositorySuggestionData[]> {
    const { getRepositorySuggestions } = useContext(CodeInsightsBackendContext)

    return useMemo(
        // memoizeAsync adds local result cache
        () => memoizeAsync<string, RepositorySuggestionData[]>(getRepositorySuggestions, query => query),
        [getRepositorySuggestions]
    )
}

/**
 * Provides list of repository suggestions.
 */
export function useRepoSuggestions(props: UseRepoSuggestionsProps): UseRepoSuggestionsResult {
    const { search, disable = false } = props

    const [suggestions, setSuggestions] = useState<RepositorySuggestionData[] | Error | undefined>([])
    const debouncedSearchTerm = useDebounce(search, 1000)
    const fetchSuggestions = useFetchSuggestions()

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
                    setSuggestions(suggestions)
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
    }, [fetchSuggestions, disable, debouncedSearchTerm])

    return { searchValue: debouncedSearchTerm, suggestions }
}
