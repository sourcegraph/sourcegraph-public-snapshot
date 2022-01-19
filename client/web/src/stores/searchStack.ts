import { useEffect } from 'react'
import create from 'zustand'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { IHighlightLineRange } from '@sourcegraph/shared/src/schema'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'

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
    lineRange: IHighlightLineRange | null
}

export type SearchStackEntry = SearchEntry | FileEntry

const SEARCH_STACK_SESSION_KEY = 'search:search-stack:session'

export interface SearchStackStore {
    entries: SearchStackEntry[]
    previousEntries: SearchStackEntry[]
    canRestoreSession: boolean
}

/**
 * Hook to get the search stack's current state. Used by the SearchStack
 * component itself and by internal functions to add a new entry to the stack.
 * The current entries persist in local and session storage. Currently this
 * doesn't work well with multiple tabs.
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
 * necessary:
 * - A search entry is considered the same if the query is the same (search
 * type, case and context are updated)
 * - A file entry is considered the same if the repo and the path are the same
 * (revison and line range are updated)
 */
export function useSearchStack(newEntry: SearchStackEntry | null): void {
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

function addSearchStackEntry(entry: SearchStackEntry, update?: (entry: SearchStackEntry) => boolean): void {
    useSearchStackState.setState(state => {
        if (update) {
            const existingEntry = state.entries.find(update)
            if (existingEntry) {
                const entriesCopy = [...state.entries]
                const index = entriesCopy.indexOf(existingEntry)
                entriesCopy[index] = { ...existingEntry, ...entry }
                // If the list contains more than one entry we disable restoring from
                // the previous session
                return { entries: entriesCopy, canRestoreSession: entriesCopy.length <= 1 }
            }
        }
        const newState = {
            entries: [...state.entries, entry],
            canRestoreSession: state.entries.length === 0,
        }
        // We store search stack data in both local and session storage: This
        // feature should really be considered to be session related but at the
        // same time we want to make it possible to restore information from the
        // previous session (e.g. in case the page was accidentally closed).
        // Storing the entries in local storage allows us to do that (see
        // useSearchStackState above).
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

function restoreSession(storage: Storage): SearchStackEntry[] {
    return JSON.parse(storage.getItem(SEARCH_STACK_SESSION_KEY) ?? '[]')
}
