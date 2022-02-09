import { renderHook, act } from '@testing-library/react-hooks'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { setAct } from '../__mocks__/zustand'

import { useExperimentalFeatures } from './experimentalFeatures'
import {
    addSearchStackEntry,
    removeAllSearchStackEntries,
    removeSearchStackEntry,
    restorePreviousSession,
    SearchStackEntry,
    SearchStackEntryInput,
    useSearchStack,
    useSearchStackState,
} from './searchStack'

describe('search stack store', () => {
    beforeAll(() => {
        setAct(act)
    })

    const exampleEntryInput: SearchStackEntryInput = {
        type: 'search',
        query: 'test',
        patternType: SearchPatternType.literal,
        caseSensitive: false,
    }

    const exampleEntry: SearchStackEntry = {
        ...exampleEntryInput,
        id: 0,
    }

    describe('adding entries (via useSearchStack)', () => {
        beforeEach(() => {
            useExperimentalFeatures.setState({ enableSearchStack: true })
        })

        it('adds a new entry', () => {
            renderHook(() => useSearchStack(exampleEntry))
            addSearchStackEntry()
            expect(useSearchStackState.getState().entries).toEqual([exampleEntry])
        })
    })

    it('restores previous session entries', () => {
        useSearchStackState.setState({ entries: [], previousEntries: [exampleEntry], canRestoreSession: true })
        restorePreviousSession()

        const state = useSearchStackState.getState()
        expect(state.entries).toEqual([exampleEntry])
        expect(state.canRestoreSession).toBe(false)
    })

    it('removes individual entries', () => {
        useSearchStackState.setState({ entries: [exampleEntry, { ...exampleEntry }] })
        removeSearchStackEntry(exampleEntry)

        const state = useSearchStackState.getState()
        expect(state.entries).toHaveLength(1)
    })

    it('removes all entries', () => {
        useSearchStackState.setState({ entries: [exampleEntry, { ...exampleEntry }] })
        removeAllSearchStackEntries()

        const state = useSearchStackState.getState()
        expect(state.entries).toHaveLength(0)
    })
})
