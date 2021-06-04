import { useCallback, useEffect, useRef, useState } from 'react'

import { useDebounce } from '@sourcegraph/wildcard/src'

import { fetchRepositorySuggestions } from '../../../../../../core/backend/requests/fetch-repository-suggestions'

interface RepositorySuggestion {
    name: string
}

interface UseRepoSuggestionsProps {
    search: string | null
    disable?: boolean
}

interface UseRepoSuggestionsResult {
    suggestions: RepositorySuggestion[] | Error | undefined
}

function useFetchSuggestions(): (search: string) => Promise<RepositorySuggestion[]> {
    /** Local suggestions cache */
    const cache = useRef<Record<string, RepositorySuggestion[]>>({})

    return useCallback(search => {
        if (cache?.current?.[search]) {
            return Promise.resolve(cache.current?.[search])
        }

        return fetchRepositorySuggestions(search)
            .toPromise()
            .then(suggestions => {
                cache.current[search] = suggestions

                return suggestions
            })
    }, [])
}

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

    return { suggestions }
}
