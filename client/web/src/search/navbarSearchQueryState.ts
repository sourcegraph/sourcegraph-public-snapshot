// NOTE (@fkling): The use of 'zustand' in this codebase should be considered as
// experimental until we had more time to evaluate this library. General
// application of this library is not recommended at this point.
// It is used here because it solves a very real performance issue
// (see https://github.com/sourcegraph/sourcegraph/issues/21200).
import create from 'zustand'

import { QueryState, SubmitSearchParameters, submitSearch } from './helpers'

type QueryStateUpdate = QueryState | ((queryState: QueryState) => QueryState)

export interface NavbarQueryState {
    queryState: QueryState
    setQueryState: (queryState: QueryStateUpdate) => void
    /**
     * submitSearch makes it possible to submit a new search query by updating
     * the current query via the callback.
     */
    submitSearch: (
        updateQuery: (currentQuery: string) => string,
        parameters: Omit<SubmitSearchParameters, 'query'>
    ) => void
}
export const useNavbarQueryState = create<NavbarQueryState>((set, get) => ({
    queryState: { query: '' },
    setQueryState: queryStateUpdate => {
        if (typeof queryStateUpdate === 'function') {
            set({ queryState: queryStateUpdate(get().queryState) })
        } else {
            set({ queryState: queryStateUpdate })
        }
    },
    submitSearch: (updateQuery, parameters) => {
        submitSearch({ ...parameters, query: updateQuery(get().queryState.query) })
    },
}))
