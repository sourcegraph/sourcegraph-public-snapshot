import { useCallback, useContext, useEffect, useRef, useState } from 'react'

import { useDebounce } from '@sourcegraph/wildcard/src'

import { InsightsApiContext } from '../../../../core/backend/api-provider'

interface RepositorySuggestion {
    name: string
}

interface UseRepoSuggestionsProps {
    search: string | null
    disable?: boolean
}

interface UseRepoSuggestionsResult {
    searchValue: string | null
    suggestions: RepositorySuggestion[] | Error | undefined
}

/**
 * Returns fetch method for repository suggestions with local cache
 */
function useFetchSuggestions(): (search: string) => Promise<RepositorySuggestion[]> {
    /** Local suggestions cache */
    const cache = useRef<Record<string, RepositorySuggestion[]>>({})
    const { getRepositorySuggestions } = useContext(InsightsApiContext)

    return useCallback(
        search => {
            if (cache?.current?.[search]) {
                return Promise.resolve(cache.current?.[search])
            }

            return getRepositorySuggestions(search).then(suggestions => {
                cache.current[search] = suggestions

                return suggestions
            })
        },
        [getRepositorySuggestions]
    )
}

/**
 * Provides list of repository suggestions.
 */
export function useRepoSuggestions(props: UseRepoSuggestionsProps): UseRepoSuggestionsResult {
    const { search, disable = false } = props

    const [suggestions, setSuggestions] = useState<RepositorySuggestion[] | Error | undefined>([])
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
