// NOTE (@fkling): The use of 'zustand' in this codebase should be considered as
// experimental until we had more time to evaluate this library. General
// application of this library is not recommended at this point.
// It is used here because it solves a very real performance issue
// (see https://github.com/sourcegraph/sourcegraph/issues/21200).
import { StateCreator } from 'zustand'

import { SearchQueryState, updateQuery } from '@sourcegraph/shared/src/search/searchQueryState'

import { submitSearch, canSubmitSearch } from '../search/helpers'

export interface NavbarQueryState extends SearchQueryState {}

export const createNavbarQueryStateStore: StateCreator<NavbarQueryState> = (set, get) => ({
    queryState: { query: '' },
    setQueryState: queryStateUpdate => {
        if (typeof queryStateUpdate === 'function') {
            set({ queryState: queryStateUpdate(get().queryState) })
        } else {
            set({ queryState: queryStateUpdate })
        }
    },
    submitSearch: (parameters, updates = []) => {
        const query = updateQuery(get().queryState.query, updates)
        if (canSubmitSearch(query, parameters.selectedSearchContextSpec)) {
            submitSearch({ ...parameters, query })
        }
    },
})
