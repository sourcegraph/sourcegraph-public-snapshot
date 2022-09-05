import { useEffect } from 'react'

import create from 'zustand'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { IHighlightLineRange } from '@sourcegraph/shared/src/schema'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'

export type NotepadEntryID = number
export interface SearchEntry {
    type: 'search'
    /**
     * The ID is primarily used to let the UI uniquily identifiy each entry.
     */
    id: NotepadEntryID
    query: string
    caseSensitive: boolean
    searchContext?: string
    patternType: SearchPatternType
    annotation?: string
}

export interface FileEntry {
    type: 'file'
    /**
     * The ID is primarily used to let the UI uniquily identifiy each entry.
     */
    id: NotepadEntryID
    path: string
    repo: string
    revision: string
    lineRange: IHighlightLineRange | null
    annotation?: string
}

export type NotepadEntry = SearchEntry | FileEntry
export type NotepadEntryInput = Omit<SearchEntry, 'id'> | Omit<FileEntry, 'id'>

export interface NotepadStore {
    /**
     * If a page/component has information that can be added to the notepad, it
     * should set this value.
     */
    addableEntry: NotepadEntryInput | null
    entries: NotepadEntry[]
    previousEntries: NotepadEntry[]
    canRestoreSession: boolean
}

const NOTEPAD_SESSION_KEY = 'search:notepad:session'
// TODO (@fkling): Remove fallback to old name after ~2 releases.
const SEARCH_STACK_SESSION_KEY = 'search:search-stack:session'
/**
 * Uniquly identifies each entry.
 */
let nextEntryID = 0

/**
 * Hook to get the notepad's current state. Used by the Notepad
 * component itself and by internal functions to add a new entry to the notepad.
 * The current entries persist in local and session storage. Currently this
 * doesn't work well with multiple tabs.
 */
export const useNotepadState = create<NotepadStore>(() => {
    // We have to get data for the current and previous session here (and retain
    // them) because those entries might get overwritten immediately if a page
    // is loaded that calls addNotepadEntry
    const entriesFromSession = restoreSession(sessionStorage)
    const entriesFromPreviousSession = restoreSession(localStorage)

    return {
        addableEntry: null,
        entries: entriesFromSession,
        previousEntries: entriesFromPreviousSession,
        canRestoreSession: entriesFromSession.length === 0 && entriesFromPreviousSession.length > 0,
    }
})

/**
 * Hook to make a new entry available for adding to the notepad. Use
 * `useMemo` to avoid unnecessary triggers and to properly remove the entry when
 * the component gets unmounted.
 */
export function useNotepad(newEntry: NotepadEntryInput | null): void {
    const [enableNotepad] = useTemporarySetting('search.notepad.enabled')
    useEffect(() => {
        if (!enableNotepad || !newEntry) {
            return
        }

        let entry: NotepadEntryInput = newEntry

        switch (entry.type) {
            case 'search': {
                // `query` most likely contains a 'context' filter that we don't
                // want to show (this information is kept separately in
                // `searchContext`).
                let processedQuery = entry.query
                const contextFilter = findFilter(entry.query, FilterType.context, FilterKind.Global)
                if (contextFilter) {
                    processedQuery = omitFilter(entry.query, contextFilter)
                }
                entry = { ...entry, query: processedQuery }
                break
            }
        }
        useNotepadState.setState({ addableEntry: entry })

        // We have to "remove" the entry if the component unmounts.
        return () => {
            const currentState = useNotepadState.getState()
            if (currentState.addableEntry === entry) {
                useNotepadState.setState({ addableEntry: null })
            }
        }
    }, [newEntry, enableNotepad])
}

/**
 * Adds the current value of addableEntry to the list of items.
 * If that value is a file entry, then a hint can be provided to control whether
 * the whole file or the line range should be added.
 */
export function addNotepadEntry(newEntry: NotepadEntryInput, hint?: 'file' | 'range'): void {
    const { entries } = useNotepadState.getState()

    let entry = newEntry
    if (entry.type === 'file' && entry.lineRange && hint === 'file') {
        entry = { ...entry, lineRange: null }
    }

    const newState = {
        entries: [...entries, { ...entry, id: nextEntryID++ }],
        canRestoreSession: entries.length === 0,
    }

    persistSession(newState.entries)
    useNotepadState.setState(newState)
}

export function restorePreviousSession(): void {
    if (useNotepadState.getState().canRestoreSession) {
        useNotepadState.setState(state =>
            // TODO (@fkling): Merge current and previous session?
            ({ entries: state.previousEntries, canRestoreSession: false })
        )
    }
}

export function removeFromNotepad(idsToDelete: NotepadEntryID | NotepadEntryID[]): void {
    useNotepadState.setState(currentState => {
        if (!Array.isArray(idsToDelete)) {
            idsToDelete = [idsToDelete]
        }
        const entries = [...currentState.entries]
        for (const id of idsToDelete) {
            entries.splice(
                entries.findIndex(entry => entry.id === id),
                1
            )
        }
        persistSession(entries)
        return { entries }
    })
}

export function removeAllNotepadEntries(): void {
    persistSession([])
    useNotepadState.setState({ entries: [] })
}

export function setEntryAnnotation(entry: NotepadEntry, annotation: string): void {
    useNotepadState.setState(state => {
        const index = state.entries.indexOf(entry)
        if (index > -1) {
            const entriesCopy = state.entries.slice()
            entriesCopy.splice(index, 1, { ...state.entries[index], annotation })
            return { entries: entriesCopy }
        }
        return state
    })
}

function restoreSession(storage: Storage): NotepadEntry[] {
    return (
        JSON.parse(storage.getItem(NOTEPAD_SESSION_KEY) ?? storage.getItem(SEARCH_STACK_SESSION_KEY) ?? '[]')
            // We always "re-id" restored entries. This makes things easier (no need
            // to track which IDs have already been used)
            .map((entry: NotepadEntry) => ({ ...entry, id: nextEntryID++ }))
    )
}

function persistSession(entries: NotepadEntry[]): void {
    // We store notepad data in both local and session storage: This feature
    // should really be considered to be session related but at the same time we
    // want to make it possible to restore information from the previous session
    // (e.g. in case the page was accidentally closed).
    // Storing the entries in local storage allows us to do that (see
    // useNotepadState above).
    const serializedEntries = JSON.stringify(entries)
    localStorage.setItem(NOTEPAD_SESSION_KEY, serializedEntries)
    sessionStorage.setItem(NOTEPAD_SESSION_KEY, serializedEntries)
}
