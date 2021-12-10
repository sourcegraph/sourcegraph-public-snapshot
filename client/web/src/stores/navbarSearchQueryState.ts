// NOTE (@fkling): The use of 'zustand' in this codebase should be considered as
// experimental until we had more time to evaluate this library. General
// application of this library is not recommended at this point.
// It is used here because it solves a very real performance issue
// (see https://github.com/sourcegraph/sourcegraph/issues/21200).
import create from 'zustand'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql/schema'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import {
    BuildSearchQueryURLParameters,
    SearchQueryState,
    updateQuery,
} from '@sourcegraph/shared/src/search/searchQueryState'
import { Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { parseSearchURL } from '../search'
import { submitSearch, canSubmitSearch } from '../search/helpers'
import { defaultCaseSensitiveFromSettings, defaultPatternTypeFromSettings } from '../util/settings'

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

export interface NavbarQueryState extends SearchQueryState {}

export const useNavbarQueryState = create<NavbarQueryState>((set, get) => ({
    queryState: { query: '' },
    searchCaseSensitivity: false,
    searchPatternType: SearchPatternType.literal,

    setQueryState: queryStateUpdate => {
        if (typeof queryStateUpdate === 'function') {
            set({ queryState: queryStateUpdate(get().queryState) })
        } else {
            set({ queryState: queryStateUpdate })
        }
    },

    submitSearch: (parameters, updates = []) => {
        const {
            queryState: { query },
            searchCaseSensitivity: caseSensitive,
            searchPatternType: patternType,
        } = get()
        const updatedQuery = updateQuery(query, updates)
        if (canSubmitSearch(query, parameters.selectedSearchContextSpec)) {
            submitSearch({ ...parameters, query: updatedQuery, caseSensitive, patternType })
        }
    },
}))

export function setSearchPatternType(searchPatternType: SearchPatternType): void {
    useNavbarQueryState.setState({ searchPatternType })
}

export function setSearchCaseSensitivity(searchCaseSensitivity: boolean): void {
    useNavbarQueryState.setState({ searchCaseSensitivity })
}

/**
 * Update or initialize query state related data from URL search parameters
 */
export function setQueryStateFromURL(urlParameters: string): void {
    // This will be updated with the default in settings when the web app mounts.
    const parsedSearchURL = parseSearchURL(urlParameters)
    if (parsedSearchURL.query) {
        // Only update flags if the URL contains a search query.
        useNavbarQueryState.setState(state => ({
            searchCaseSensitivity: parsedSearchURL.caseSensitive,
            searchPatternType: parsedSearchURL.patternType ?? state.searchPatternType,
        }))
    }
}

/**
 * Update or initialize query state related data from settings
 */
export function setQueryStateFromSettings(settings: SettingsCascadeOrError<Settings>): void {
    const newState: Partial<Pick<NavbarQueryState, 'searchPatternType' | 'searchCaseSensitivity'>> = {}

    const caseSensitive = defaultCaseSensitiveFromSettings(settings)
    if (caseSensitive) {
        newState.searchCaseSensitivity = caseSensitive
    }

    const searchPatternType = defaultPatternTypeFromSettings(settings)
    if (caseSensitive) {
        newState.searchPatternType = searchPatternType
    }

    // The way Zustand is designed makes it difficult to build up a partial new
    // state object, hence the cast to any here.
    useNavbarQueryState.setState(newState as any)
}

/**
 * This function wraps 'buildSearchURLQuery' and uses values from the global
 * query state for parameters that have not been supplied. This provides a
 * simple way for components to compose new queries with global query state
 * without needing access to that state themselves.
 */
export function buildSearchURLQueryFromQueryState(parameters: BuildSearchQueryURLParameters): string {
    const currentState = useNavbarQueryState.getState()

    return buildSearchURLQuery(
        parameters.query,
        parameters.patternType ?? currentState.searchPatternType,
        parameters.caseSensitive ?? currentState.searchCaseSensitivity,
        parameters.searchContextSpec,
        parameters.searchParametersList
    )
}
