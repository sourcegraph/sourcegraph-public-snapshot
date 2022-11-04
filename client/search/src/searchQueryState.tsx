import React, { createContext, useContext } from 'react'

import { StoreApi, UseBoundStore } from 'zustand'

import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { appendFilter, updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'

import { QueryState, SubmitSearchParameters, toggleSubquery } from '.'

export type SearchQueryStateStore<T extends SearchQueryState = SearchQueryState> = UseBoundStore<T, StoreApi<T>>

export const SearchQueryStateStoreContext = createContext<SearchQueryStateStore | null>(null)

/**
 * React context provider for SearchQueryStateStore.
 * Clients that render `search` package components that depend on a SearchQueryStateStore
 * need to be wrapped with this.
 *
 * Example: Both the VS Code extension and the web app render `<SearchSidebar>`, so it needs to
 * reference the appropriate zustand store through context (provided here).
 */
export const SearchQueryStateStoreProvider: React.FunctionComponent<
    React.PropsWithChildren<{
        useSearchQueryState: SearchQueryStateStore
    }>
> = ({ children, useSearchQueryState }) => (
    <SearchQueryStateStoreContext.Provider value={useSearchQueryState}>
        {children}
    </SearchQueryStateStoreContext.Provider>
)

export const useSearchQueryStateStoreContext = (): SearchQueryStateStore => {
    const context = useContext(SearchQueryStateStoreContext)
    if (context === null) {
        throw new Error('useSearchQueryStateStoreContext must be used within a SearchQueryStateStoreProvider')
    }
    return context
}

/**
 * Describes where settings have been loaded from when the app loads. Higher
 * values have higher precedence, i.e. if settings have been loaded from the
 * URL, user settings should not overwrite them.
 */
export enum InitialParametersSource {
    DEFAULT,
    USER_SETTINGS,
    URL,
}

// Implemented in /web as navbar query state, /vscode as webview query state.
export interface SearchQueryState {
    /**
     * This is used to determine whether a source is allowed to overwrite
     * parameters that have been loaded from another source. For example, user
     * settings (e.g. default pattern type) are not allowed to overwrite the
     * parameters if they have been loaded from a URL (because parameters loaded
     * from a URL are more specific).
     */
    parametersSource: InitialParametersSource

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
    | {
          type: 'replaceQuery'
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
            case 'replaceQuery':
                return update.value
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
