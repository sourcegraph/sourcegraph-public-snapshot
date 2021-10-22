// NOTE (@fkling): The use of 'zustand' in this codebase should be considered as
// experimental until we had more time to evaluate this library. General
// application of this library is not recommended at this point.
// It is used here because it solves a very real performance issue
// (see https://github.com/sourcegraph/sourcegraph/issues/21200).
import create from 'zustand'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { appendFilter, updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'

import { QueryState, SubmitSearchParameters, submitSearch, toggleSubquery } from './helpers'

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
          type: 'toggleSubquery'
          value: string
      }

function updateQuery(query: string, updates: QueryUpdate[]): string {
    return updates.reduce((query, update) => {
        switch (update.type) {
            case 'appendFilter':
                if (!update.unique || !filterExists(query, update.field)) {
                    return appendFilter(query, update.field, update.value)
                }
                break
            case 'updateOrAppendFilter':
                return updateFilter(query, update.field, update.value)
            case 'toggleSubquery':
                return toggleSubquery(query, update.value)
        }
        return query
    }, query)
}

export interface NavbarQueryState {
    /**
     * The current search query (usually visible in the main search input).
     */
    queryState: QueryState
    setQueryState: (queryState: QueryStateUpdate) => void
    /**
     * submitSearch makes it possible to submit a new search query by updating
     * the current query via update directives. It won't submit the query if it
     * is empty.
     */
    submitSearch: (parameters: Omit<SubmitSearchParameters, 'query'>, updates?: QueryUpdate[]) => void
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
    submitSearch: (parameters, updates = []) => {
        const query = updateQuery(get().queryState.query, updates)
        if (query !== '') {
            submitSearch({ ...parameters, query })
        }
    },
}))
