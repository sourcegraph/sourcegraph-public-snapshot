import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { appendFilter, updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'

import { QueryState, SubmitSearchParameters, toggleSubquery } from '.'

// Implemented in /web as navbar query state, /vscode as webview query state.
export interface SearchQueryState {
    // DATA
    /**
     * The current seach query and auxiliary information needed by the
     * MonacoQueryInput component. You most likely don't have to read this value
     * directly.
     * See {@link QueryState} for more information.
     */
    queryState: QueryState
    searchCaseSensitivity: boolean
    searchPatternType: SearchPatternType
    searchQueryFromURL: string

    // ACTIONS
    /**
     * setQueryState updates `queryState`
     */
    setQueryState: (queryState: QueryStateUpdate) => void

    /**
     * submitSearch makes it possible to submit a new search query by updating
     * the current query via update directives. It won't submit the query if it
     * is empty.
     * Note that this won't update `queryState` directly.
     */
    submitSearch: (
        parameters: Omit<SubmitSearchParameters, 'query' | 'caseSensitive' | 'patternType'>,
        updates?: QueryUpdate[]
    ) => void
}

export type QueryStateUpdate = QueryState | ((queryState: QueryState) => QueryState)

export type QueryUpdate =
    | /**
     * Appends a filter to the current search query. If the filter is unique and
     * already exists in the query, the update is ignored.
     */
    {
          type: 'appendFilter'
          field: FilterType
          value: string
          /**
           * If true, the filter will only be appended a filter with the same name
           * doesn't already exist in the query.
           */
          unique?: true
      }
    /**
     * Appends or updates a filter to/in the query.
     */
    | {
          type: 'updateOrAppendFilter'
          field: FilterType
          value: string
      }
    // Only exists for the filters from the search sidebar since they come in
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

export interface BuildSearchQueryURLParameters {
    query: string
    patternType?: SearchPatternType
    caseSensitive?: boolean
    searchContextSpec?: string
    searchParametersList?: { key: string; value: string }[]
}
