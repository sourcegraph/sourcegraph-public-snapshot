import { type FC, useCallback, useEffect, useMemo, useState } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'
import type { Observable } from 'rxjs'

import { limitHit } from '@sourcegraph/branded'
import type { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { QueryUpdate, SearchContextProps } from '@sourcegraph/shared/src/search'
import { updateFilters } from '@sourcegraph/shared/src/search/query/transformer'
import { LATEST_VERSION, type StreamSearchOptions } from '@sourcegraph/shared/src/search/stream'
import { type SettingsCascadeProps, useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import type { SearchAggregationProps, SearchStreamingProps } from '..'
import type { AuthenticatedUser } from '../../auth'
import type { CodeMonitoringProps } from '../../codeMonitoring'
import { formatUrlOverrideFeatureFlags } from '../../featureFlags/lib/parseUrlOverrideFeatureFlags'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import { useFeatureFlagOverrides } from '../../featureFlags/useFeatureFlagOverrides'
import type { CodeInsightsProps } from '../../insights/types'
import type { OwnConfigProps } from '../../own/OwnConfigProps'
import { useDeveloperSettings, useNavbarQueryState } from '../../stores'
import { submitSearch } from '../helpers'
import { useRecentSearches } from '../input/useRecentSearches'

import { useAggregationUIMode } from './components/aggregation'
import { NewSearchContent } from './components/new-search-content/NewSearchContent'
import { SearchContent } from './components/search-content/SearchContent'
import { useCachedSearchResults } from './SearchResultsCacheProvider'
import { useStreamingSearchPings } from './useStreamingSearchPings'

export interface StreamingSearchResultsProps
    extends SearchStreamingProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec' | 'searchContextsEnabled'>,
        SettingsCascadeProps,
        PlatformContextProps<'settings' | 'requestGraphQL' | 'sourcegraphURL'>,
        TelemetryProps,
        CodeInsightsProps,
        SearchAggregationProps,
        CodeMonitoringProps,
        OwnConfigProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

export const StreamingSearchResults: FC<StreamingSearchResultsProps> = props => {
    const {
        streamSearch,
        authenticatedUser,
        telemetryService,
        isSourcegraphDotCom,
        searchAggregationEnabled,
        codeMonitoringEnabled,
        platformContext,
    } = props

    const location = useLocation()
    const navigate = useNavigate()
    const { addRecentSearch } = useRecentSearches()
    const featureOverrides = useFeatureFlagOverrides()

    // Feature flags
    const [enableRepositoryMetadata] = useFeatureFlag('repository-metadata', true)
    const newSearchNavigation = useExperimentalFeatures(features => features.newSearchNavigationUI ?? false)

    // Global state
    const caseSensitive = useNavbarQueryState(state => state.searchCaseSensitivity)
    const patternType = useNavbarQueryState(state => state.searchPatternType)
    const searchMode = useNavbarQueryState(state => state.searchMode)
    const liveQuery = useNavbarQueryState(state => state.queryState.query)
    const submittedURLQuery = useNavbarQueryState(state => state.searchQueryFromURL)
    const queryState = useNavbarQueryState(state => state.queryState)
    const setQueryState = useNavbarQueryState(state => state.setQueryState)
    const submitQuerySearch = useNavbarQueryState(state => state.submitSearch)
    const [aggregationUIMode] = useAggregationUIMode()

    // Derived state
    const trace = useMemo(() => new URLSearchParams(location.search).get('trace') ?? undefined, [location.search])
    const { searchOptions } = useDeveloperSettings(settings => settings.zoekt)

    const options: StreamSearchOptions = useMemo(
        () => ({
            version: LATEST_VERSION,
            patternType: patternType ?? SearchPatternType.standard,
            caseSensitive,
            trace,
            featureOverrides: formatUrlOverrideFeatureFlags(featureOverrides),
            searchMode,
            chunkMatches: true,
            zoektSearchOptions: searchOptions,
        }),
        [patternType, caseSensitive, trace, featureOverrides, searchMode, searchOptions]
    )
    const results = useCachedSearchResults(streamSearch, submittedURLQuery, options, telemetryService)

    const { logSearchResultClicked } = useStreamingSearchPings({
        telemetryService,
        isSourcegraphDotCom,
        results,
        isAuauthenticated: !!authenticatedUser,
    })

    useEffect(() => {
        if (results?.state === 'complete') {
            // Add the recent search in the next queue execution to avoid updating a React component while rendering another component.
            setTimeout(
                () => addRecentSearch(submittedURLQuery, results.progress.matchCount, limitHit(results.progress)),
                0
            )
        }
    }, [addRecentSearch, results, submittedURLQuery])

    // Expand/contract all results
    const [allExpanded, setAllExpanded] = useState(false)

    const onExpandAllResultsToggle = useCallback(() => {
        setAllExpanded(oldValue => !oldValue)
        telemetryService.log(allExpanded ? 'allResultsExpanded' : 'allResultsCollapsed')
    }, [allExpanded, telemetryService])

    useEffect(() => {
        setAllExpanded(false) // Reset expanded state when new search is started
    }, [location.search])

    const handleSidebarSearchSubmit = useCallback(
        /**
         * The `updatedSearchQuery` is required in case we synchronously update the search
         * query in the event handlers and want to submit a new search. Without this argument,
         * the `handleSidebarSearchSubmit` function uses the outdated location reference
         * because the component was not re-rendered yet.
         *
         * Example use-case: search-aggregation result bar click where we first update the URL
         * by settings the `groupBy` search param to `null` and then synchronously call `submitSearch`.
         */
        (updates: QueryUpdate[], updatedSearchQuery?: string) =>
            submitQuerySearch(
                {
                    selectedSearchContextSpec: props.selectedSearchContextSpec,
                    historyOrNavigate: navigate,
                    location: {
                        ...location,
                        search: updatedSearchQuery || location.search,
                    },
                    source: 'filter',
                },
                updates
            ),
        [submitQuerySearch, props.selectedSearchContextSpec, navigate, location]
    )

    const onSearchAgain = useCallback(
        (additionalFilters: string[]) => {
            telemetryService.log('SearchSkippedResultsAgainClicked')

            const { selectedSearchContextSpec } = props
            submitSearch({
                historyOrNavigate: navigate,
                location,
                selectedSearchContextSpec,
                caseSensitive,
                patternType,
                query: applyAdditionalFilters(submittedURLQuery, additionalFilters),
                source: 'excludedResults',
            })
        },
        [telemetryService, props, navigate, location, caseSensitive, patternType, submittedURLQuery]
    )

    /**
     * The `updatedSearchQuery` is required in case we synchronously update the search
     * query in the event handlers and want to submit a new search. Without this argument,
     * the `handleSidebarSearchSubmit` function uses the outdated location reference
     * because the component was not re-rendered yet.
     *
     * Example use-case: search-aggregation result bar click where we first update the URL
     * by settings the `groupBy` search param to `null` and then synchronously call `submitSearch`.
     */
    const handleSearchAggregationBarClick = (query: string, updatedSearchQuery: string): void => {
        const { selectedSearchContextSpec } = props
        submitSearch({
            historyOrNavigate: navigate,
            location: { ...location, search: updatedSearchQuery },
            selectedSearchContextSpec,
            caseSensitive,
            patternType,
            query,
            source: 'nav',
        })
    }

    const onDisableSmartSearch = useCallback(() => {
        const { selectedSearchContextSpec } = props
        submitSearch({
            historyOrNavigate: navigate,
            location,
            selectedSearchContextSpec,
            caseSensitive,
            patternType: SearchPatternType.standard,
            query: submittedURLQuery,
            source: 'smartSearchDisabled',
        })
    }, [caseSensitive, location, navigate, props, submittedURLQuery])

    const hasResultsToAggregate = results?.state === 'complete' ? (results?.results.length ?? 0) > 0 : true
    const showAggregationPanel = searchAggregationEnabled && hasResultsToAggregate

    return !newSearchNavigation ? (
        <SearchContent
            submittedURLQuery={submittedURLQuery}
            queryState={queryState}
            liveQuery={liveQuery}
            allExpanded={allExpanded}
            searchMode={searchMode}
            trace={!!trace}
            searchContextsEnabled={props.searchContextsEnabled}
            patternType={patternType}
            results={results}
            showAggregationPanel={showAggregationPanel}
            selectedSearchContextSpec={props.selectedSearchContextSpec}
            aggregationUIMode={aggregationUIMode}
            caseSensitive={caseSensitive}
            authenticatedUser={authenticatedUser}
            isSourcegraphDotCom={isSourcegraphDotCom}
            enableRepositoryMetadata={enableRepositoryMetadata}
            options={options}
            codeMonitoringEnabled={codeMonitoringEnabled}
            fetchHighlightedFileLineRanges={props.fetchHighlightedFileLineRanges}
            onNavbarQueryChange={setQueryState}
            onSearchSubmit={handleSidebarSearchSubmit}
            onQuerySubmit={handleSearchAggregationBarClick}
            onExpandAllResultsToggle={onExpandAllResultsToggle}
            onSearchAgain={onSearchAgain}
            onDisableSmartSearch={onDisableSmartSearch}
            onLogSearchResultClick={logSearchResultClicked}
            settingsCascade={props.settingsCascade}
            telemetryService={telemetryService}
            platformContext={platformContext}
        />
    ) : (
        <NewSearchContent
            submittedURLQuery={submittedURLQuery}
            queryState={queryState}
            liveQuery={liveQuery}
            allExpanded={allExpanded}
            searchMode={searchMode}
            trace={!!trace}
            searchContextsEnabled={props.searchContextsEnabled}
            patternType={patternType}
            results={results}
            showAggregationPanel={showAggregationPanel}
            selectedSearchContextSpec={props.selectedSearchContextSpec}
            aggregationUIMode={aggregationUIMode}
            caseSensitive={caseSensitive}
            authenticatedUser={authenticatedUser}
            isSourcegraphDotCom={isSourcegraphDotCom}
            enableRepositoryMetadata={enableRepositoryMetadata}
            options={options}
            codeMonitoringEnabled={codeMonitoringEnabled}
            fetchHighlightedFileLineRanges={props.fetchHighlightedFileLineRanges}
            onNavbarQueryChange={setQueryState}
            onSearchSubmit={handleSidebarSearchSubmit}
            onQuerySubmit={handleSearchAggregationBarClick}
            onExpandAllResultsToggle={onExpandAllResultsToggle}
            onSearchAgain={onSearchAgain}
            onDisableSmartSearch={onDisableSmartSearch}
            onLogSearchResultClick={logSearchResultClicked}
            settingsCascade={props.settingsCascade}
            telemetryService={telemetryService}
            platformContext={platformContext}
        />
    )
}

const applyAdditionalFilters = (query: string, additionalFilters: string[]): string => {
    let newQuery = query
    for (const filter of additionalFilters) {
        const fieldValue = filter.split(':', 2)
        newQuery = updateFilters(newQuery, fieldValue[0], fieldValue[1])
    }
    return newQuery
}
