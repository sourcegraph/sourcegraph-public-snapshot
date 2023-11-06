import { type FC, useCallback, useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import { useLocation, useNavigate } from 'react-router-dom'
import type { Observable } from 'rxjs'

import { limitHit, StreamingProgress, StreamingSearchResultsList } from '@sourcegraph/branded'
import { asError } from '@sourcegraph/common'
import type { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import type { FilePrefetcher } from '@sourcegraph/shared/src/components/PrefetchableFile'
import { HighlightResponseFormat, SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { QueryUpdate, SearchContextProps } from '@sourcegraph/shared/src/search'
import { collectMetrics } from '@sourcegraph/shared/src/search/query/metrics'
import { sanitizeQueryForTelemetry, updateFilters } from '@sourcegraph/shared/src/search/query/transformer'
import {
    type AlertKind,
    LATEST_VERSION,
    type SmartSearchAlertKind,
    type StreamSearchOptions,
} from '@sourcegraph/shared/src/search/stream'
import { type SettingsCascadeProps, useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import type { SearchAggregationProps, SearchStreamingProps } from '..'
import type { AuthenticatedUser } from '../../auth'
import type { CodeMonitoringProps } from '../../codeMonitoring'
import { PageTitle } from '../../components/PageTitle'
import { formatUrlOverrideFeatureFlags } from '../../featureFlags/lib/parseUrlOverrideFeatureFlags'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import { useFeatureFlagOverrides } from '../../featureFlags/useFeatureFlagOverrides'
import type { CodeInsightsProps } from '../../insights/types'
import type { OwnConfigProps } from '../../own/OwnConfigProps'
import { fetchBlob } from '../../repo/blob/backend'
import { SavedSearchModal } from '../../savedSearches/SavedSearchModal'
import { isSearchJobsEnabled } from '../../search-jobs/utility'
import {
    buildSearchURLQueryFromQueryState,
    setSearchMode,
    useDeveloperSettings,
    useNavbarQueryState,
    useNotepad,
} from '../../stores'
import { GettingStartedTour } from '../../tour/GettingStartedTour'
import { useShowOnboardingTour } from '../../tour/hooks'
import { submitSearch } from '../helpers'
import { useRecentSearches } from '../input/useRecentSearches'
import { DidYouMean } from '../suggestion/DidYouMean'
import { SmartSearch, smartSearchEvent } from '../suggestion/SmartSearch'

import { AggregationUIMode, SearchAggregationResult, useAggregationUIMode } from './components/aggregation'
import { SearchResultsCsvExportModal } from './export/SearchResultsCsvExportModal'
import { SearchAlert } from './SearchAlert'
import { useCachedSearchResults } from './SearchResultsCacheProvider'
import { SearchResultsInfoBar } from './SearchResultsInfoBar'
import { SearchFiltersSidebar } from './sidebar/SearchFiltersSidebar'
import { UnownedResultsAlert } from './UnownedResultsAlert'

import styles from './StreamingSearchResults.module.scss'

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

    // Feature flags
    const prefetchFileEnabled = useExperimentalFeatures(features => features.enableSearchFilePrefetch ?? false)
    const [enableSearchResultsKeyboardNavigation] = useFeatureFlag('search-results-keyboard-navigation', true)
    const [enableRepositoryMetadata] = useFeatureFlag('repository-metadata', true)
    const [sidebarCollapsed, setSidebarCollapsed] = useTemporarySetting('search.sidebar.collapsed', false)

    const showOnboardingTour = useShowOnboardingTour({ authenticatedUser, isSourcegraphDotCom })

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

    // Local state
    const [showMobileSidebar, setShowMobileSidebar] = useState(false)

    // Derived state
    const trace = useMemo(() => new URLSearchParams(location.search).get('trace') ?? undefined, [location.search])
    const { searchOptions } = useDeveloperSettings(settings => settings.zoekt)
    const featureOverrides = useFeatureFlagOverrides()
    const { addRecentSearch } = useRecentSearches()

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

    const resultsLength = results?.results.length || 0
    const logSearchResultClicked = useCallback(
        (index: number, type: string) => {
            telemetryService.log('SearchResultClicked')
            // This data ends up in Prometheus and is not part of the ping payload.
            telemetryService.log('search.ranking.result-clicked', {
                index,
                type,
                resultsLength,
            })
        },
        [telemetryService, resultsLength]
    )

    // Log view event on first load
    useEffect(
        () => {
            telemetryService.logViewEvent('SearchResults')
        },
        // Only log view on initial load
        // eslint-disable-next-line react-hooks/exhaustive-deps
        []
    )

    // Log search query event when URL changes
    useEffect(() => {
        const metrics = submittedURLQuery ? collectMetrics(submittedURLQuery) : undefined

        telemetryService.log(
            'SearchResultsQueried',
            {
                code_search: {
                    query_data: {
                        query: metrics,
                        combined: submittedURLQuery,
                        empty: !submittedURLQuery,
                    },
                },
            },
            {
                code_search: {
                    query_data: {
                        // 🚨 PRIVACY: never provide any private query data in the
                        // { code_search: query_data: query } property,
                        // which is also potentially exported in pings data.
                        query: metrics,

                        // 🚨 PRIVACY: Only collect the full query string for unauthenticated users
                        // on Sourcegraph.com, and only after sanitizing to remove certain filters.
                        combined:
                            !authenticatedUser && isSourcegraphDotCom
                                ? sanitizeQueryForTelemetry(submittedURLQuery)
                                : undefined,
                        empty: !submittedURLQuery,
                    },
                },
            }
        )
        // Only log when the query changes
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [submittedURLQuery])

    // Log events when search completes or fails
    useEffect(() => {
        if (results?.state === 'complete') {
            telemetryService.log('SearchResultsFetched', {
                code_search: {
                    // 🚨 PRIVACY: never provide any private data in { code_search: { results } }.
                    query_data: {
                        combined: submittedURLQuery,
                    },
                    results: {
                        results_count: results.progress.matchCount,
                        limit_hit: limitHit(results.progress),
                        any_cloning: results.progress.skipped.some(skipped => skipped.reason === 'repository-cloning'),
                        alert: results.alert ? results.alert.title : null,
                    },
                },
            })
            if (results.results.length > 0) {
                telemetryService.log('SearchResultsNonEmpty')
            }
        } else if (results?.state === 'error') {
            telemetryService.log('SearchResultsFetchFailed', {
                code_search: { error_message: asError(results.error).message },
            })
        }
    }, [results, submittedURLQuery, telemetryService])

    useEffect(() => {
        if (results?.state === 'complete') {
            // Add the recent search in the next queue execution to avoid updating a React component while rendering another component.
            setTimeout(
                () => addRecentSearch(submittedURLQuery, results.progress.matchCount, limitHit(results.progress)),
                0
            )
        }
    }, [addRecentSearch, results, submittedURLQuery])

    useEffect(() => {
        if (
            (results?.alert?.kind === 'smart-search-additional-results' ||
                results?.alert?.kind === 'smart-search-pure-results') &&
            results?.alert?.title &&
            results.alert.proposedQueries
        ) {
            const events = smartSearchEvent(
                results.alert.kind,
                results.alert.title,
                results.alert.proposedQueries.map(entry => entry.description || '')
            )
            for (const event of events) {
                telemetryService.log(event)
            }
        }
    }, [results, telemetryService])

    // Reset expanded state when new search is started
    useEffect(() => {
        setAllExpanded(false)
    }, [location.search])

    useNotepad(
        useMemo(
            () =>
                results?.state === 'complete'
                    ? {
                          type: 'search',
                          query: submittedURLQuery,
                          caseSensitive,
                          patternType,
                          searchContext: props.selectedSearchContextSpec,
                      }
                    : null,
            [results, submittedURLQuery, patternType, caseSensitive, props.selectedSearchContextSpec]
        )
    )

    // Expand/contract all results
    const [allExpanded, setAllExpanded] = useState(false)
    const onExpandAllResultsToggle = useCallback(() => {
        setAllExpanded(oldValue => !oldValue)
        telemetryService.log(allExpanded ? 'allResultsExpanded' : 'allResultsCollapsed')
    }, [allExpanded, telemetryService])
    useEffect(() => {
        setAllExpanded(false) // Reset expanded state when new search is started
    }, [location.search])

    // Save search
    const [showSavedSearchModal, setShowSavedSearchModal] = useState(false)
    const onSaveQueryClick = useCallback(() => setShowSavedSearchModal(true), [])
    const onSaveQueryModalClose = useCallback(() => {
        setShowSavedSearchModal(false)
        telemetryService.log('SavedQueriesToggleCreating', { queries: { creating: false } })
    }, [telemetryService])

    // Export results to CSV
    const [showCsvExportModal, setShowCsvExportModal] = useState(false)
    const onExportCsvClick = useCallback(() => setShowCsvExportModal(true), [])

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

    const hasResultsToAggregate = results?.state === 'complete' ? (results?.results.length ?? 0) > 0 : true

    // Show aggregation panel only if we're in Enterprise versions and hide it in OSS and
    // when search doesn't have any matches
    const showAggregationPanel = searchAggregationEnabled && hasResultsToAggregate

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

    const prefetchFile: FilePrefetcher = useCallback(
        params =>
            fetchBlob({
                ...params,
                format: HighlightResponseFormat.JSON_SCIP,
            }),
        []
    )

    return (
        <div className={classNames(styles.container, sidebarCollapsed && styles.containerWithSidebarHidden)}>
            <PageTitle key="page-title" title={submittedURLQuery} />

            <SearchFiltersSidebar
                liveQuery={liveQuery}
                submittedURLQuery={submittedURLQuery}
                patternType={patternType}
                filters={results?.filters}
                showAggregationPanel={showAggregationPanel}
                selectedSearchContextSpec={props.selectedSearchContextSpec}
                aggregationUIMode={aggregationUIMode}
                settingsCascade={props.settingsCascade}
                telemetryService={props.telemetryService}
                caseSensitive={caseSensitive}
                className={classNames(styles.sidebar, showMobileSidebar && styles.sidebarShowMobile)}
                onNavbarQueryChange={setQueryState}
                onSearchSubmit={handleSidebarSearchSubmit}
                setSidebarCollapsed={setSidebarCollapsed}
            >
                {showOnboardingTour && (
                    <GettingStartedTour
                        className="mb-1"
                        telemetryService={props.telemetryService}
                        authenticatedUser={authenticatedUser}
                    />
                )}
            </SearchFiltersSidebar>

            {aggregationUIMode === AggregationUIMode.SearchPage && (
                <SearchAggregationResult
                    query={submittedURLQuery}
                    patternType={patternType}
                    caseSensitive={caseSensitive}
                    aria-label="Aggregation results panel"
                    className={styles.contents}
                    onQuerySubmit={handleSearchAggregationBarClick}
                    telemetryService={props.telemetryService}
                />
            )}

            {aggregationUIMode !== AggregationUIMode.SearchPage && (
                <>
                    <SearchResultsInfoBar
                        {...props}
                        patternType={patternType}
                        caseSensitive={caseSensitive}
                        query={submittedURLQuery}
                        results={results}
                        options={options}
                        enableCodeMonitoring={codeMonitoringEnabled}
                        className={styles.infobar}
                        allExpanded={allExpanded}
                        onExpandAllResultsToggle={onExpandAllResultsToggle}
                        onSaveQueryClick={onSaveQueryClick}
                        onExportCsvClick={onExportCsvClick}
                        onShowMobileFiltersChanged={show => setShowMobileSidebar(show)}
                        sidebarCollapsed={!!sidebarCollapsed}
                        setSidebarCollapsed={setSidebarCollapsed}
                        stats={
                            <StreamingProgress
                                query={`${submittedURLQuery} patterntype:${patternType}`}
                                progress={results?.progress || { durationMs: 0, matchCount: 0, skipped: [] }}
                                state={results?.state || 'loading'}
                                onSearchAgain={onSearchAgain}
                                showTrace={!!trace}
                                isSearchJobsEnabled={isSearchJobsEnabled()}
                                telemetryService={props.telemetryService}
                            />
                        }
                    />

                    <div className={styles.contents}>
                        <DidYouMean
                            telemetryService={props.telemetryService}
                            query={submittedURLQuery}
                            patternType={patternType}
                            caseSensitive={caseSensitive}
                            selectedSearchContextSpec={props.selectedSearchContextSpec}
                        />

                        {results?.alert?.kind && isSmartSearchAlert(results.alert.kind) && (
                            <SmartSearch alert={results?.alert} onDisableSmartSearch={onDisableSmartSearch} />
                        )}

                        <GettingStartedTour.Info
                            className="mt-2 mb-3"
                            isSourcegraphDotCom={props.isSourcegraphDotCom}
                        />

                        {showSavedSearchModal && (
                            <SavedSearchModal
                                {...props}
                                navigate={navigate}
                                patternType={patternType}
                                query={submittedURLQuery}
                                authenticatedUser={authenticatedUser}
                                onDidCancel={onSaveQueryModalClose}
                            />
                        )}
                        {showCsvExportModal && (
                            <SearchResultsCsvExportModal
                                query={submittedURLQuery}
                                options={options}
                                results={results}
                                sourcegraphURL={platformContext.sourcegraphURL}
                                telemetryService={telemetryService}
                                onClose={() => setShowCsvExportModal(false)}
                            />
                        )}
                        {results?.alert && (!results?.alert.kind || !isSmartSearchAlert(results.alert.kind)) && (
                            <div className={classNames(styles.alertArea, 'mt-4')}>
                                {results?.alert?.kind === 'unowned-results' ? (
                                    <UnownedResultsAlert
                                        alertTitle={results.alert.title}
                                        alertDescription={results.alert.description}
                                        queryState={queryState}
                                        patternType={patternType}
                                        caseSensitive={caseSensitive}
                                        selectedSearchContextSpec={props.selectedSearchContextSpec}
                                    />
                                ) : (
                                    <SearchAlert
                                        alert={results.alert}
                                        caseSensitive={caseSensitive}
                                        patternType={patternType}
                                    />
                                )}
                            </div>
                        )}

                        <StreamingSearchResultsList
                            {...props}
                            enableRepositoryMetadata={enableRepositoryMetadata}
                            results={results}
                            allExpanded={allExpanded}
                            executedQuery={location.search}
                            prefetchFileEnabled={prefetchFileEnabled}
                            prefetchFile={prefetchFile}
                            enableKeyboardNavigation={enableSearchResultsKeyboardNavigation}
                            showQueryExamplesOnNoResultsPage={true}
                            queryState={queryState}
                            setQueryState={setQueryState}
                            buildSearchURLQueryFromQueryState={buildSearchURLQueryFromQueryState}
                            searchMode={searchMode}
                            setSearchMode={setSearchMode}
                            submitSearch={submitSearch}
                            caseSensitive={caseSensitive}
                            searchQueryFromURL={submittedURLQuery}
                            selectedSearchContextSpec={props.selectedSearchContextSpec}
                            logSearchResultClicked={logSearchResultClicked}
                        />
                    </div>
                </>
            )}
        </div>
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

function isSmartSearchAlert(kind: AlertKind): kind is SmartSearchAlertKind {
    switch (kind) {
        case 'smart-search-additional-results':
        case 'smart-search-pure-results': {
            return true
        }
    }
    return false
}
