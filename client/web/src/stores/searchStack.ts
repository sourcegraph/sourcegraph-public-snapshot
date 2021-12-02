import { useEffect } from 'react'
import create from 'zustand'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/validate'

import { useExperimentalFeatures } from './experimentalFeatures'

export interface SearchEntry {
    type: 'search'
    query: string
    caseSensitive: boolean
    searchContext?: string
    patternType: SearchPatternType
}

export interface FileEntry {
    type: 'file'
    path: string
    repo: string
    revision: string
}

type Entry = SearchEntry | FileEntry

const SEARCH_STACK_SESSION_KEY = 'search:search-stack:session'

export interface SearchStackStore {
    entries: Entry[]
    previousEntries: Entry[]
    canRestoreSession: boolean
}

/**
 * Hook to get the search stack's current state. Used by the SearchStack
 * component itself and by internal functions to add a new entry to the stack.
 */
export const useSearchStackState = create<SearchStackStore>(() => {
    // We have to get data for the current and previous session here (and retain
    // them) because those entries might get overwritten immediately if a page
    // is loaded that calls addSearchStackEntry
    const entriesFromSession = restoreSession(sessionStorage)
    const entriesFromPreviousSession = restoreSession(localStorage)

    return {
        entries: entriesFromSession,
        previousEntries: entriesFromPreviousSession,
        canRestoreSession: entriesFromSession.length === 0 && entriesFromPreviousSession.length > 0,
    }
})

/**
 * Hook to add a new entry to the search stack. Use `useMemo` to avoid
 * unnecessary triggers. This hook will *update* an existing entry if
 * necessary.
 */
export function useSearchStack(newEntry: Entry | null): void {
    const enableSearchStack = useExperimentalFeatures(features => features.enableSearchStack)
    useEffect(() => {
        if (enableSearchStack && newEntry) {
            switch (newEntry.type) {
                case 'file':
                    addSearchStackEntry(
                        newEntry,
                        existingEntry =>
                            existingEntry.type === 'file' &&
                            existingEntry.repo === newEntry.repo &&
                            existingEntry.path === newEntry.path
                    )
                    break
                case 'search': {
                    // `query` most likely contains a 'context' filter that we don't
                    // want to show (this information is kept separately in
                    // `searchContext`).
                    let processedQuery = newEntry.query
                    const contextFilter = findFilter(newEntry.query, FilterType.context, FilterKind.Global)
                    if (contextFilter) {
                        processedQuery = omitFilter(newEntry.query, contextFilter)
                    }
                    addSearchStackEntry(
                        { ...newEntry, query: processedQuery },
                        existingEntry => existingEntry.type === 'search' && existingEntry.query === processedQuery
                    )
                    break
                }
            }
        }
    }, [newEntry, enableSearchStack])
}

function addSearchStackEntry(entry: Entry, update?: (entry: Entry) => boolean): void {
    useSearchStackState.setState(state => {
        // If the list contains more than one entry we disable restoring from
        // the previous session
        const canRestoreSession = state.entries.length === 0

        if (update) {
            const existingEntry = state.entries.find(update)
            if (existingEntry) {
                const entriesCopy = [...state.entries]
                const index = entriesCopy.indexOf(existingEntry)
                entriesCopy[index] = { ...existingEntry, ...entry }
                return { entries: entriesCopy, canRestoreSession }
            }
        }
        const newState = {
            entries: [...state.entries, entry],
            canRestoreSession,
        }
        // We store search stack data in both local and session storage: This
        // feature should really be considered to be session related but at the
        // same time we want to make it possible to restore information from the
        // previous session (e.g. in case the page was accidentally closed).
        const serializedEntries = JSON.stringify(newState.entries)
        localStorage.setItem(SEARCH_STACK_SESSION_KEY, serializedEntries)
        sessionStorage.setItem(SEARCH_STACK_SESSION_KEY, serializedEntries)

        return newState
    })
}

export function restorePreviousSession(): void {
    if (useSearchStackState.getState().canRestoreSession) {
        useSearchStackState.setState(state =>
            // TODO (@fkling): Merge current and previous session?
            ({ entries: state.previousEntries, canRestoreSession: false })
        )
    }
}

function restoreSession(storage: Storage): Entry[] {
    return JSON.parse(storage.getItem(SEARCH_STACK_SESSION_KEY) ?? '[]')
}
