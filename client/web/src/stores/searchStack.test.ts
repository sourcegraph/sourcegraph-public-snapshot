import { renderHook, act } from '@testing-library/react-hooks'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { setAct } from '../__mocks__/zustand'

import { useExperimentalFeatures } from './experimentalFeatures'
import {
    removeAllSearchStackEntries,
    removeSearchStackEntry,
    restorePreviousSession,
    SearchStackEntry,
    useSearchStack,
    useSearchStackState,
} from './searchStack'

describe('search stack store', () => {
    beforeAll(() => {
        setAct(act)
    })

    const exampleEntry: SearchStackEntry = {
        id: 0,
        type: 'search',
        query: 'test',
        patternType: SearchPatternType.literal,
        caseSensitive: false,
    }

    describe('adding entries (via useSearchStack)', () => {
        beforeEach(() => {
            useExperimentalFeatures.setState({ enableSearchStack: true })
        })

        it('adds a new entry', () => {
            renderHook(() => useSearchStack(exampleEntry))
            expect(useSearchStackState.getState().entries).toEqual([exampleEntry])
        })

        it('updates an existing query entry', () => {
            const { rerender } = renderHook(({ entry }: { entry: SearchStackEntry }) => useSearchStack(entry), {
                initialProps: { entry: exampleEntry },
            })
            rerender({ entry: { ...exampleEntry, caseSensitive: true } })

            const { entries } = useSearchStackState.getState()
            expect(entries).toHaveLength(1)
            expect(entries[0]).toHaveProperty('caseSensitive', true)
        })

        it('updates an existing file entry', () => {
            const entry: SearchStackEntry = {
                id: 0,
                type: 'file',
                path: 'path/to/file',
                repo: 'test',
                revision: 'master',
                lineRange: null,
            }
            const { rerender } = renderHook(({ entry }: { entry: SearchStackEntry }) => useSearchStack(entry), {
                initialProps: { entry },
            })
            rerender({ entry: { ...entry, lineRange: { startLine: 10, endLine: 11 } } })

            expect(useSearchStackState.getState().entries).toHaveLength(1)
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
