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

interface UseRepoSearchResult {
    searchText: string
    setSearchText: (text: string) => void
    clearSearchText: () => void
    loading: boolean
    results: ContextSelectorRepoFields[]
}

export const useRepoSearch = (
    transcriptHistory: TranscriptJSON[],
    authenticatedUser: AuthenticatedUser | null = null
): UseRepoSearchResult => {
    const [searchText, _setSearchText] = useState('')
    const [debouncedSearchText, setDebouncedSearchText] = useState('')

    const [search, { data, loading: queryLoading, stopPolling }] = useLazyQuery<
        ReposSelectorSearchResult,
        ReposSelectorSearchVariables
    >(ReposSelectorSearchQuery, {})

    const debouncedSearch = useMemo(
        () =>
            debounce((text: string) => {
                setDebouncedSearchText(text)
                search({
                    variables: { query: text, includeJobs: !!authenticatedUser?.siteAdmin },
                    pollInterval: 5000,
                }).catch(noop)
            }, 300),
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

    const searchResults = useMemo(() => data?.repositories.nodes || [], [data])

    return {
        searchText,
        setSearchText,
        clearSearchText,
        loading,
        results: searchResults,
    }
}
