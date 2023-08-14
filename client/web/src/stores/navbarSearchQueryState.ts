// NOTE (@fkling): The use of 'zustand' in this codebase should be considered as
// experimental until we had more time to evaluate this library. General
// application of this library is not recommended at this point.
// It is used here because it solves a very real performance issue
// (see https://github.com/sourcegraph/sourcegraph/issues/21200).
import create from 'zustand'

import {
    type BuildSearchQueryURLParameters,
    canSubmitSearch,
    type SearchQueryState,
    updateQuery,
    InitialParametersSource,
    SearchMode,
} from '@sourcegraph/shared/src/search'
import type { Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { SearchPatternType } from '../graphql-operations'
import type { ParsedSearchURL } from '../search'
import { submitSearch } from '../search/helpers'
import {
    defaultCaseSensitiveFromSettings,
    defaultPatternTypeFromSettings,
    defaultSearchModeFromSettings,
} from '../util/settings'

export interface NavbarQueryState extends SearchQueryState {}

export const useNavbarQueryState = create<NavbarQueryState>((set, get) => ({
    parametersSource: InitialParametersSource.DEFAULT,
    queryState: { query: '' },
    searchCaseSensitivity: false,
    searchPatternType: SearchPatternType.standard,
    searchMode: SearchMode.SmartSearch,
    searchQueryFromURL: '',

    setQueryState: queryStateUpdate => {
        if (typeof queryStateUpdate === 'function') {
            set({ queryState: queryStateUpdate(get().queryState) })
        } else {
            set({ queryState: queryStateUpdate })
        }
    },

    submitSearch: (parameters, updates = []) => {
        const {
            queryState,
            searchCaseSensitivity: caseSensitive,
            searchPatternType: patternType,
            searchMode: searchMode,
        } = get()
        const query = parameters.query ?? queryState.query
        const updatedQuery = updateQuery(query, updates)
        if (canSubmitSearch(query, parameters.selectedSearchContextSpec)) {
            submitSearch({ ...parameters, query: updatedQuery, caseSensitive, patternType, searchMode })
        }
    },
}))

export function setSearchPatternType(searchPatternType: SearchPatternType): void {
    useNavbarQueryState.setState({ searchPatternType })
}

export function setSearchCaseSensitivity(searchCaseSensitivity: boolean): void {
    useNavbarQueryState.setState({ searchCaseSensitivity })
}

export function setSearchMode(searchMode: SearchMode): void {
    useNavbarQueryState.setState({ searchMode })
}

/**
 * Update or initialize query state related data from URL search parameters.
 *
 * @param parsedSearchURL contains the information extracted from a URL
 * @param query can be used to specify the query to use when it differs from
 * the one contained in the URL (e.g. when the context:... filter got removed)
 */
export function setQueryStateFromURL(parsedSearchURL: ParsedSearchURL, query = parsedSearchURL.query ?? ''): void {
    if (useNavbarQueryState.getState().parametersSource > InitialParametersSource.URL) {
        return
    }

    // This will be updated with the default in settings when the web app mounts.
    const newState: Partial<
        Pick<
            NavbarQueryState,
            | 'queryState'
            | 'searchPatternType'
            | 'searchCaseSensitivity'
            | 'searchQueryFromURL'
            | 'parametersSource'
            | 'searchMode'
        >
    > = {}

    if (parsedSearchURL.query) {
        // Only update flags if the URL contains a search query.
        newState.parametersSource = InitialParametersSource.URL
        newState.searchCaseSensitivity = parsedSearchURL.caseSensitive
        if (parsedSearchURL.patternType !== undefined) {
            newState.searchPatternType = parsedSearchURL.patternType
        }
        newState.queryState = { query }
        newState.searchQueryFromURL = parsedSearchURL.query
        newState.searchMode = parsedSearchURL.searchMode
    }

    // The way Zustand is designed makes it difficult to build up a partial new
    // state object, hence the cast to any here.
    useNavbarQueryState.setState(newState as any)
}

/**
 * Update or initialize query state related data from settings
 */
export function setQueryStateFromSettings(settings: SettingsCascadeOrError<Settings>): void {
    if (useNavbarQueryState.getState().parametersSource > InitialParametersSource.USER_SETTINGS) {
        return
    }

    const newState: Partial<
        Pick<NavbarQueryState, 'searchPatternType' | 'searchCaseSensitivity' | 'parametersSource' | 'searchMode'>
    > = {
        parametersSource: InitialParametersSource.USER_SETTINGS,
    }

    const caseSensitive = defaultCaseSensitiveFromSettings(settings)
    if (caseSensitive !== undefined) {
        newState.searchCaseSensitivity = caseSensitive
    }

    const searchMode = defaultSearchModeFromSettings(settings)
    if (searchMode !== undefined) {
        newState.searchMode = searchMode
    }

    const searchPatternType = defaultPatternTypeFromSettings(settings)
    if (searchPatternType) {
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
        parameters.searchMode ?? currentState.searchMode
    )
}
