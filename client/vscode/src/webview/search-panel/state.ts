import create, { UseStore } from 'zustand'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql/schema'
import { QueryState } from '@sourcegraph/shared/src/search/helpers'

import { SearchResult } from '../../graphql-operations'
import { VsCodeApi } from '../vsCodeApi'

export const DEFAULT_SEARCH_CONTEXT_SPEC = 'global'

const initialSearchState: SearchState = {
    executed: false,
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    queryState: {
        query: '',
    },
    queryToRun: {
        query: '',
    },
    searchResults: null,
    selectedSearchContextSpec: DEFAULT_SEARCH_CONTEXT_SPEC,
    selectedResultPositionSuffix: undefined,
}

interface SearchState {
    executed: boolean | undefined
    caseSensitive: boolean
    patternType: SearchPatternType
    /** QueryState used for the input. Updated on input. */
    queryState: QueryState
    /** QueryState updated on submission. Run search on change. */
    queryToRun: QueryState
    searchResults: SearchResult | null
    selectedSearchContextSpec: string | undefined
    selectedResultPositionSuffix: string | undefined
}

export interface State {
    state: SearchState
    actions: {
        setQuery: (queryState: QueryState) => void
        submitQuery: (queryState?: QueryState) => void
        updateResults: (searchResults: SearchResult) => void
        setCaseSensitivity: (caseSensitive: boolean) => void
        setPatternType: (patternType: SearchPatternType) => void
        setSelectedSearchContextSpec: (selectedSearchContextSpec: string | undefined) => void
        setSelectedResultPositionSuffix: (selectedResultPositionSuffix: string | undefined) => void
    }
}

export function createUseQueryState(vsCodeApi: VsCodeApi<State['state']>): UseStore<State> {
    const useQueryState = create<State>(set => ({
        state: initialSearchState,
        actions: {
            setQuery: queryState => {
                set(({ state }) => ({ state: { ...state, queryState } }))
            },
            submitQuery: queryState => {
                if (queryState) {
                    set(({ state }) => ({ state: { ...state, queryState, queryToRun: queryState, executed: true } }))
                } else {
                    // Sync queryToRun with current queryState.
                    set(({ state }) => ({ state: { ...state, queryToRun: state.queryState, executed: true } }))
                }
            },
            updateResults: searchResults => {
                set(({ state }) => ({ state: { ...state, searchResults } }))
            },
            setCaseSensitivity: caseSensitive => {
                set(({ state }) => ({ state: { ...state, caseSensitive } }))
            },
            setPatternType: patternType => {
                set(({ state }) => ({ state: { ...state, patternType } }))
            },
            setSelectedSearchContextSpec: selectedSearchContextSpec => {
                set(({ state }) => ({ state: { ...state, selectedSearchContextSpec } }))
            },
            setSelectedResultPositionSuffix: selectedResultPositionSuffix => {
                set(({ state }) => ({ state: { ...state, selectedResultPositionSuffix } }))
            },
        },
    }))

    // Persist latest state.
    useQueryState.subscribe(({ state }) => {
        vsCodeApi.setState(state)
    })

    return useQueryState
}
