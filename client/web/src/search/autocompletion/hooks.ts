import { useEffect, useMemo, useState } from 'react'

import { useApolloClient } from '@apollo/client'
import { Fzf } from 'fzf'

import { getDocumentNode } from '@sourcegraph/http-client'
import { ALL_LANGUAGES } from '@sourcegraph/shared/src/search/query/languageFilter'

import { createRepositoryCompletionSource } from './sources'

export function useRepositoryCompletionSource(searchTerm: string): { suggestions: string[]; loading: boolean } {
    const [suggestions, setSuggestions] = useState<string[]>([])
    const [loading, setLoading] = useState(false)

    const client = useApolloClient()

    const repoCache = useMemo(
        () =>
            createRepositoryCompletionSource((query, variables) =>
                client.query({ query: getDocumentNode(query), variables }).then(result => result.data)
            ),
        [client]
    )

    useEffect(() => {
        let isCurrent = true
        setLoading(true)

        if (!searchTerm.trim()) {
            setSuggestions([])
            return
        }

        const result = repoCache.query(searchTerm, suggestions => suggestions.map(suggestion => suggestion.item.name))
        setSuggestions(result.result)

        result
            .next()
            .then(result => {
                setLoading(false)
                if (isCurrent) {
                    setSuggestions(result.result)
                }
            })
            .catch(() => setSuggestions([]))

        return () => {
            isCurrent = false
        }
    }, [searchTerm, repoCache])

    return { suggestions, loading }
}

const languageFzf = new Fzf(ALL_LANGUAGES, {
    fuzzy: false,
})

export function useLanguageCompletionSource(searchTerm: string): { suggestions: string[] } {
    return useMemo(() => {
        if (!searchTerm.trim()) {
            return { suggestions: [] }
        }
        return { suggestions: languageFzf.find(searchTerm).map(match => match.item) }
    }, [searchTerm])
}
