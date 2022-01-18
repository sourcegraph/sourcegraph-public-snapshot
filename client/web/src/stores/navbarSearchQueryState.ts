// NOTE (@fkling): The use of 'zustand' in this codebase should be considered as
// experimental until we had more time to evaluate this library. General
// application of this library is not recommended at this point.
// It is used here because it solves a very real performance issue
// (see https://github.com/sourcegraph/sourcegraph/issues/21200).
import { Subscription } from 'rxjs'
import create, { GetState, SetState } from 'zustand'
import { StoreApiWithSubscribeWithSelector, subscribeWithSelector } from 'zustand/middleware'

import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { getGlobalSearchContextFilter } from '@sourcegraph/shared/src/search/query/query'
import { appendFilter, omitFilter, updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'
import { Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { SettingsExperimentalFeatures } from '../schema/settings.schema'
import { getAvailableSearchContextSpecOrDefault, ParsedSearchURL, parseSearchURL } from '../search'
import { QueryState, SubmitSearchParameters, submitSearch, toggleSubquery, canSubmitSearch } from '../search/helpers'
import { defaultCaseSensitiveFromSettings, defaultPatternTypeFromSettings } from '../util/settings'

import { useExperimentalFeatures } from './experimentalFeatures'

type QueryStateUpdate = QueryState | ((queryState: QueryState) => QueryState)

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

const LAST_SEARCH_CONTEXT_KEY = 'sg-last-search-context'

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

let subscription: Subscription | null = null
/**
 * Helper function to validate a given search context and keeps track of
 * previous requests so that they can be cancelled.
 */
function getAvailableSearchContext(
    spec: string,
    defaultSpec: string,
    callback: (availableSearchContextSpecOrDefault: string) => void
): void {
    subscription?.unsubscribe()
    subscription = getAvailableSearchContextSpecOrDefault({ spec, defaultSpec }).subscribe(callback)
}

export interface NavbarQueryState {
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
    /**
     * The currently active/submitted query.
     */
    searchQueryFromURL: string
    /**
     * The currently selected search context. Don't read this value directly
     * from the store, use the `useSelectedSearchContext` hook instead. This
     * hook also takes into account whether search contexts are enabled or not.
     */
    selectedSearchContext: string
    defaultSearchContext: string
    /**
     * Used to determine whether or not search contexts are available at all.
     * Should be set at app initialization time.
     */
    searchContextsEnabled: boolean

    /**
     * Used for determining whether or not show the search context CTA. Set in
     * the application root.
     */
    hasUserAddedRepositories: boolean

    /**
     * Used for determining whether or not show the search context CTA. Set in
     * the application root.
     */
    hasUserSyncedPublicRepositories: boolean

    /**
     * Used in the search context CTA. Set in the application root.
     */
    hasUserAddedExternalServices: boolean

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
        parameters: Omit<
            SubmitSearchParameters,
            'query' | 'caseSensitive' | 'patternType' | 'selectedSearchContextSpec'
        >,
        updates?: QueryUpdate[]
    ) => void
}

export const useNavbarQueryState = create<
    NavbarQueryState,
    SetState<NavbarQueryState>,
    GetState<NavbarQueryState>,
    StoreApiWithSubscribeWithSelector<NavbarQueryState>
>(
    subscribeWithSelector(
        (set, get): NavbarQueryState => ({
            queryState: { query: '' },
            searchCaseSensitivity: false,
            searchPatternType: SearchPatternType.literal,
            searchQueryFromURL: '',
            searchContextsEnabled: false,
            defaultSearchContext: 'global', // users will be able to set the default in the future
            selectedSearchContext: '',

            hasUserAddedRepositories: false,
            hasUserSyncedPublicRepositories: false,
            hasUserAddedExternalServices: false,

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
                    selectedSearchContext,
                } = get()
                const updatedQuery = updateQuery(query, updates)
                if (canSubmitSearch(query, selectedSearchContext)) {
                    submitSearch({
                        ...parameters,
                        query: updatedQuery,
                        caseSensitive,
                        patternType,
                        selectedSearchContextSpec: selectedSearchContext,
                    })
                }
            },
        })
    )
)

const featureSelector = (features: SettingsExperimentalFeatures): boolean => features.showSearchContext ?? false
const contextSelector = (state: NavbarQueryState): NavbarQueryState['selectedSearchContext'] =>
    state.selectedSearchContext
const contextEnabledSelector = (state: NavbarQueryState): NavbarQueryState['searchContextsEnabled'] =>
    state.searchContextsEnabled

export const useSelectedSearchContext = (): string | undefined => {
    const showSearchContext = useExperimentalFeatures(featureSelector)
    const selectedSearchContext = useNavbarQueryState(contextSelector)
    const searchContextEnabled = useNavbarQueryState(contextEnabledSelector)

    return searchContextEnabled && showSearchContext ? selectedSearchContext : undefined
}

export function setQuery(query: string): void {
    useNavbarQueryState.setState({ queryState: { query } })
}

export function setSearchPatternType(searchPatternType: SearchPatternType): void {
    useNavbarQueryState.setState({ searchPatternType })
}

export function setSearchCaseSensitivity(searchCaseSensitivity: boolean): void {
    useNavbarQueryState.setState({ searchCaseSensitivity })
}

export function setSelectedSearchContext(spec: string): void {
    const { searchContextsEnabled, defaultSearchContext } = useNavbarQueryState.getState()

    if (!searchContextsEnabled) {
        return
    }

    getAvailableSearchContext(spec, defaultSearchContext, availableSearchContextSpecOrDefault => {
        useNavbarQueryState.setState({ selectedSearchContext: availableSearchContextSpecOrDefault })
        localStorage.setItem(LAST_SEARCH_CONTEXT_KEY, availableSearchContextSpecOrDefault)
    })
}

/**
 * initQueryState restores the previously selected search context
 * from local storage and initializes the query state from the passed query
 * string.
 * It should be called when the application is initialized.
 */
export function initQueryState(urlParameters: string, searchContextsEnabled: boolean): void {
    useNavbarQueryState.setState({ searchContextsEnabled })

    const parsedSearchURL = parseSearchURL(urlParameters)

    if (parsedSearchURL.query) {
        setQueryStateFromURL(parsedSearchURL, searchContextsEnabled)
    } else {
        // If no query is present (e.g. search page, settings page), select the last saved
        // search context from localStorage as currently selected search context.
        useNavbarQueryState.setState({
            selectedSearchContext: localStorage.getItem(LAST_SEARCH_CONTEXT_KEY) || 'global',
        })
    }
}

/**
 * Update or initialize query state related data from URL search parameters
 */
export function setQueryStateFromURL(parsedSearchURL: ParsedSearchURL, showSearchContext: boolean): void {
    // This will be updated with the default in settings when the web app mounts.
    const newState: Partial<
        Pick<
            NavbarQueryState,
            | 'queryState'
            | 'searchPatternType'
            | 'searchCaseSensitivity'
            | 'searchQueryFromURL'
            | 'selectedSearchContext'
        >
    > = {}
    const { searchContextsEnabled, selectedSearchContext, defaultSearchContext } = useNavbarQueryState.getState()

    const query = parsedSearchURL.query ?? ''
    newState.searchQueryFromURL = query

    if (query) {
        newState.queryState = { query }

        // Only update flags if the URL contains a search query.
        newState.searchCaseSensitivity = parsedSearchURL.caseSensitive
        if (parsedSearchURL.patternType !== undefined) {
            newState.searchPatternType = parsedSearchURL.patternType
        }

        if (showSearchContext && searchContextsEnabled) {
            const newSearchContext = getGlobalSearchContextFilter(query)

            if (newSearchContext) {
                const queryWithoutContext = omitFilter(query, newSearchContext.filter)

                // While we wait for the result of the
                // `getAvailableSearchContext` call, we assume the context
                // is available to prevent flashing and moving content in the query bar.
                // This optimizes for the most common use case where user selects a
                // search context from the dropdown. See
                // https://github.com/sourcegraph/sourcegraph/issues/19918 for more
                // info.
                newState.queryState.query = queryWithoutContext

                if (newSearchContext.spec !== selectedSearchContext) {
                    getAvailableSearchContext(
                        newSearchContext.spec,
                        defaultSearchContext,
                        availableSearchContextSpecOrDefault => {
                            newState.queryState = {
                                // If a global search context spec is available to the user, we omit it from the
                                // query and move it to the search contexts dropdown.
                                query:
                                    availableSearchContextSpecOrDefault === newSearchContext.spec
                                        ? queryWithoutContext
                                        : query,
                            }

                            newState.selectedSearchContext = availableSearchContextSpecOrDefault

                            useNavbarQueryState.setState(newState as any)
                            localStorage.setItem(LAST_SEARCH_CONTEXT_KEY, availableSearchContextSpecOrDefault)
                        }
                    )
                }
            } else {
                // If a context filter does not exist in the query, we have to switch the selected context
                // to global to match the UI with the backend semantics (if no context is specified in the query,
                // the query is run in global context).
                newState.selectedSearchContext = 'global'
            }
        }
    }

    // The way Zustand is designed makes it difficult to build up a partial new
    // state object, hence the cast to any here.
    useNavbarQueryState.setState(newState as any)
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

interface BuildSearchQueryURLParameters {
    query: string
    patternType?: SearchPatternType
    caseSensitive?: boolean
    searchContextSpec?: string
    searchParametersList?: { key: string; value: string }[]
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
        parameters.searchContextSpec ?? currentState.selectedSearchContext,
        parameters.searchParametersList
    )
}

type SubmitSearchParametersFromState = 'query' | 'caseSensitive' | 'patternType' | 'selectedSearchContextSpec'

export function submitSearchWithGlobalQueryState(
    parameters: Partial<Pick<SubmitSearchParameters, SubmitSearchParametersFromState>> &
        Omit<SubmitSearchParameters, SubmitSearchParametersFromState>
): void {
    const currentState = useNavbarQueryState.getState()

    const completeParameters: SubmitSearchParameters = {
        query: currentState.queryState.query,
        caseSensitive: currentState.searchCaseSensitivity,
        patternType: currentState.searchPatternType,
        selectedSearchContextSpec: currentState.selectedSearchContext,
        ...parameters,
    }

    if (canSubmitSearch(completeParameters.query, completeParameters.selectedSearchContextSpec)) {
        submitSearch(completeParameters)
    }
}
