import { act } from '@testing-library/react'
import { beforeAll, describe, expect, it } from 'vitest'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { setAct } from '../__mocks__/zustand'

import {
    addNotepadEntry,
    removeAllNotepadEntries,
    removeFromNotepad,
    restorePreviousSession,
    type NotepadEntry,
    type NotepadEntryInput,
    useNotepadState,
} from './notepad'

describe('notepad store', () => {
    beforeAll(() => {
        setAct(act)
    })

    const exampleEntryInput: NotepadEntryInput = {
        type: 'search',
        query: 'test',
        patternType: SearchPatternType.standard,
        caseSensitive: false,
    }

    const exampleEntry: NotepadEntry = {
        ...exampleEntryInput,
        id: 0,
    }

    describe('adding entries', () => {
        it('adds a new entry', () => {
            addNotepadEntry(exampleEntry)
            expect(useNotepadState.getState().entries).toEqual([exampleEntry])
        })

        it('adds a new entry as file', () => {
            addNotepadEntry(
                { type: 'file', path: 'path/', lineRange: { startLine: 0, endLine: 1 }, repo: 'repo', revision: 'rev' },
                'file'
            )
            expect(useNotepadState.getState().entries[0]).toHaveProperty('lineRange', null)
        })

        it('adds a new entry as line range', () => {
            addNotepadEntry(
                { type: 'file', path: 'path/', lineRange: { startLine: 0, endLine: 1 }, repo: 'repo', revision: 'rev' },
                'range'
            )
            expect(useNotepadState.getState().entries[0]).toHaveProperty('lineRange', { startLine: 0, endLine: 1 })
        })
    })

    it('restores previous session entries', () => {
        useNotepadState.setState({ entries: [], previousEntries: [exampleEntry], canRestoreSession: true })
        restorePreviousSession()

        const state = useNotepadState.getState()
        expect(state.entries).toEqual([exampleEntry])
        expect(state.canRestoreSession).toBe(false)
    })

    it('removes individual entries', () => {
        useNotepadState.setState({ entries: [exampleEntry, { ...exampleEntry, id: 1 }] })
        removeFromNotepad(exampleEntry.id)

        const state = useNotepadState.getState()
        expect(state.entries).toHaveLength(1)
    })

    it('removes all entries', () => {
        useNotepadState.setState({ entries: [exampleEntry, { ...exampleEntry, id: 1 }] })
        removeAllNotepadEntries()

        const state = useNotepadState.getState()
        expect(state.entries).toHaveLength(0)
    })
})
