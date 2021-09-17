// NOTE (@fkling): The use of 'zustand' in this codebase should be considered as
// experimental until we had more time to evaluate this library. General
// application of this library is not recommended at this point.
// It is used here because it solves a very real performance issue
// (see https://github.com/sourcegraph/sourcegraph/issues/21200).
import create from 'zustand'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { appendFilter, updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'

import { QueryState, SubmitSearchParameters, submitSearch, toggleSearchFilter } from './helpers'

type QueryStateUpdate = QueryState | ((queryState: QueryState) => QueryState)

export type QueryUpdate =
    | {
          type: 'appendFilter'
          field: FilterType
          value: string
          /**
           * If true, the filter will only be appended a filter with the same name
           * doesn't already exist in the query.
           */
          unique?: true
      }
    | {
          type: 'updateOrAppendFilter'
          field: FilterType
          value: string
      }
    // Only exists for the filters from the serach sidebar since they come in
    // filter:value form. Should not be used elsewhere.
    | {
          type: 'toggleSubstring'
          value: string
      }

function updateQuery(query: string, updates: QueryUpdate[]): string {
    for (const update of updates) {
        switch (update.type) {
            case 'appendFilter':
                if (!update.unique || !filterExists(query, update.field)) {
                    query = appendFilter(query, update.field, update.value)
                }
                break
            case 'updateOrAppendFilter':
                query = updateFilter(query, update.field, update.value)
                break
            case 'toggleSubstring':
                query = toggleSearchFilter(query, update.value)
                break
        }
    }
    return query
}

export interface NavbarQueryState {
    queryState: QueryState
    setQueryState: (queryState: QueryStateUpdate) => void
    /**
     * submitSearch makes it possible to submit a new search query by updating
     * the current query via the callback.
     */
    submitSearch: (updates: QueryUpdate[], parameters: Omit<SubmitSearchParameters, 'query'>) => void
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
    submitSearch: (updates, parameters) => {
        submitSearch({ ...parameters, query: updateQuery(get().queryState.query, updates) })
    },
}))
