import { useCallback, useMemo, useState } from 'react'

import { debounce, noop } from 'lodash'

import type { TranscriptJSON } from '@sourcegraph/cody-shared/dist/chat/transcript'
import { useLazyQuery } from '@sourcegraph/http-client'
import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'

import { useUserHistory } from '../../../components/useUserHistory'
import type {
    ContextSelectorRepoFields,
    ReposSelectorSearchResult,
    ReposSelectorSearchVariables,
} from '../../../graphql-operations'

import { ReposSelectorSearchQuery } from './backend'
import { extractAndOrderRepoNames } from './useRepoSuggestions'

interface UseRepoSearchResult {
    searchText: string
    setSearchText: (text: string) => void
    clearSearchText: () => void
    loading: boolean
    error: Error | undefined
    results: ContextSelectorRepoFields[]
}

/**
 * useRepoSearch is a custom hook that manages the input search text state and generates
 * repository search results for the context scope selector.
 *
 * Repositories that the user has recently interacted with will be automatically ranked
 * higher on the list of results, based on most recent interaction.
 *
 * @param transcriptHistory the current user's chat transcript history from the store
 * @param authenticatedUser the current authenticated user
 * @returns a `UseRepoSearchResult` object
 */
export const useRepoSearch = (
    transcriptHistory: TranscriptJSON[],
    authenticatedUser: AuthenticatedUser | null = null
): UseRepoSearchResult => {
    const [searchText, _setSearchText] = useState('')
    const [debouncedSearchText, setDebouncedSearchText] = useState('')

    const [search, { data, loading: queryLoading, error, stopPolling }] = useLazyQuery<
        ReposSelectorSearchResult,
        ReposSelectorSearchVariables
    >(ReposSelectorSearchQuery, {})

    // Debounce the search query to avoid making too many requests.
    const debouncedSearch = useMemo(
        () =>
            debounce((text: string) => {
                setDebouncedSearchText(text)
                search({
                    variables: { query: text, includeJobs: !!authenticatedUser?.siteAdmin },
                    pollInterval: 5000,
                }).catch(noop)
            }, 500),
        [search, authenticatedUser?.siteAdmin]
    )

    const setSearchText = useCallback(
        (text: string) => {
            _setSearchText(text)
            debouncedSearch(text)
        },
        [debouncedSearch]
    )

    const clearSearchText = useCallback(() => {
        setSearchText('')
        stopPolling()
    }, [setSearchText, stopPolling])

    const loading = queryLoading || debouncedSearchText !== searchText

    const userHistory = useUserHistory(authenticatedUser?.id, false)
    const userHistoryEntries = useMemo(() => userHistory.loadEntries(), [userHistory])
    const recentRepos = useMemo(
        () => extractAndOrderRepoNames(transcriptHistory, userHistoryEntries),
        [userHistoryEntries, transcriptHistory]
    )

    const searchResults = useMemo(() => {
        if (!data?.repositories) {
            return []
        }

        const nodes = [...data.repositories.nodes]
        nodes.sort((a, b) => {
            const aIsRecent = recentRepos.includes(a.name)
            const bIsRecent = recentRepos.includes(b.name)
            // If a was recently accessed and b was not, a should be ranked higher than b.
            if (aIsRecent && !bIsRecent) {
                return -1
            }
            // If b was recently accessed and a was not, b should be ranked higher than a.
            if (!aIsRecent && bIsRecent) {
                return 1
            }
            // If both a and b were recently accessed, they should be ranked according to
            // their order in the recentRepos array. Alternatively, if neither a nor b were
            // recently accessed, they should be left in the order returned by the search.
            // recentRepos.indexOf({a,b}.name) will return -1 if the repository is not in
            // the recentRepos array, which will result in -1 - (-1) = 0 (no change).
            return recentRepos.indexOf(a.name) - recentRepos.indexOf(b.name)
        })
        return nodes
    }, [data, recentRepos])

    return {
        searchText,
        setSearchText,
        clearSearchText,
        loading,
        error,
        results: searchResults,
    }
}
