import { QueryState, SubmitSearchParameters, toggleSubquery } from './helpers'
import { FilterType } from './query/filters'
import { appendFilter, updateFilter } from './query/transformer'
import { filterExists } from './query/validate'

// Implemented in /web as navbar query state, /vscode as webview query state.
export interface SearchQueryState {
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

export function updateQuery(query: string, updates: QueryUpdate[]): string {
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
