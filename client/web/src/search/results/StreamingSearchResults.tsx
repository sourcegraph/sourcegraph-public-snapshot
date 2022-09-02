import React, { useCallback, useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { useHistory } from 'react-router'
import { Observable } from 'rxjs'

import { asError } from '@sourcegraph/common'
import { QueryUpdate, SearchContextProps } from '@sourcegraph/search'
import {
    FetchFileParameters,
    SidebarButtonStrip,
    StreamingProgress,
    StreamingSearchResultsList,
} from '@sourcegraph/search-ui'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { collectMetrics } from '@sourcegraph/shared/src/search/query/metrics'
import { sanitizeQueryForTelemetry, updateFilters } from '@sourcegraph/shared/src/search/query/transformer'
import { LATEST_VERSION, StreamSearchOptions } from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { SearchStreamingProps } from '..'
import { AuthenticatedUser } from '../../auth'
import { PageTitle } from '../../components/PageTitle'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import { CodeInsightsProps } from '../../insights/types'
import { isCodeInsightsEnabled } from '../../insights/utils/is-code-insights-enabled'
import { SavedSearchModal } from '../../savedSearches/SavedSearchModal'
import { useExperimentalFeatures, useNavbarQueryState, useNotepad } from '../../stores'
import { GettingStartedTour } from '../../tour/GettingStartedTour'
import { submitSearch } from '../helpers'
import { DidYouMean } from '../suggestion/DidYouMean'
import { LuckySearch, luckySearchEvent } from '../suggestion/LuckySearch'

import { AggregationUIMode, SearchAggregationResult, useAggregationUIMode } from './components/aggregation'
import { SearchAlert } from './SearchAlert'
import { useCachedSearchResults } from './SearchResultsCacheProvider'
import { SearchResultsInfoBar } from './SearchResultsInfoBar'
import { SearchFiltersSidebar } from './sidebar/SearchFiltersSidebar'

import styles from './StreamingSearchResults.module.scss'

export interface StreamingSearchResultsProps
    extends SearchStreamingProps,
        Pick<ActivationProps, 'activation'>,
        Pick<SearchContextProps, 'selectedSearchContextSpec' | 'searchContextsEnabled'>,
        SettingsCascadeProps,
        ExtensionsControllerProps<'executeCommand' | 'extHostAPI'>,
        PlatformContextProps<'settings' | 'requestGraphQL' | 'sourcegraphURL'>,
        TelemetryProps,
        ThemeProps,
        CodeInsightsProps {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

export const StreamingSearchResults: React.FunctionComponent<
    React.PropsWithChildren<StreamingSearchResultsProps>
> = props => {
    const {
        streamSearch,
        location,
        authenticatedUser,
        telemetryService,
        codeInsightsEnabled,
        isSourcegraphDotCom,
        extensionsController,
    } = props

    const history = useHistory()
    // Feature flags
    // Log lucky search events. To be removed at latest by 12/2022.
    const [luckySearchEnabled] = useFeatureFlag('ab-lucky-search')
    const enableCodeMonitoring = useExperimentalFeatures(features => features.codeMonitoring ?? false)
    const showSearchContext = useExperimentalFeatures(features => features.showSearchContext ?? false)
    const [selectedTab] = useTemporarySetting('search.sidebar.selectedTab', 'filters')

    // Global state
    const caseSensitive = useNavbarQueryState(state => state.searchCaseSensitivity)
    const patternType = useNavbarQueryState(state => state.searchPatternType)
    const liveQuery = useNavbarQueryState(state => state.queryState.query)
    const submittedURLQuery = useNavbarQueryState(state => state.searchQueryFromURL)
    const setQueryState = useNavbarQueryState(state => state.setQueryState)
    const submitQuerySearch = useNavbarQueryState(state => state.submitSearch)
    const [aggregationUIMode] = useAggregationUIMode()

    // Local state
    const [allExpanded, setAllExpanded] = useState(false)
    const [showSavedSearchModal, setShowSavedSearchModal] = useState(false)
    const [showMobileSidebar, setShowMobileSidebar] = useState(false)

    // Derived state
    const extensionHostAPI =
        extensionsController !== null && window.context.enableLegacyExtensions ? extensionsController.extHostAPI : null
    const trace = useMemo(() => new URLSearchParams(location.search).get('trace') ?? undefined, [location.search])

    const options: StreamSearchOptions = useMemo(
        () => ({
            version: LATEST_VERSION,
            patternType: patternType ?? SearchPatternType.standard,
            caseSensitive,
            trace,
        }),
        [caseSensitive, patternType, trace]
    )

    const results = useCachedSearchResults(streamSearch, submittedURLQuery, options, extensionHostAPI, telemetryService)
    const resultsFound = useMemo<boolean>(() => (results ? results.results.length > 0 : false), [results])

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
                    results: {
                        results_count: results.results.length,
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
    }, [results, telemetryService])

    useEffect(() => {
        if (luckySearchEnabled && results?.state === 'complete') {
            telemetryService.log('SearchResultsFetchedAuto')
            if (results.results.length > 0) {
                telemetryService.log('SearchResultsNonEmptyAuto')
            }
        }
        if (
            luckySearchEnabled &&
            results?.alert?.kind === 'lucky-search-queries' &&
            results?.alert?.title &&
            results.alert.proposedQueries
        ) {
            const events = luckySearchEvent(
                results.alert.title,
                results.alert.proposedQueries.map(entry => entry.description || '')
            )
            for (const event of events) {
                telemetryService.log(event)
            }
        }
    }, [results, luckySearchEnabled, telemetryService])

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

    const onExpandAllResultsToggle = useCallback(() => {
        setAllExpanded(oldValue => !oldValue)
        telemetryService.log(allExpanded ? 'allResultsExpanded' : 'allResultsCollapsed')
    }, [allExpanded, telemetryService])

    const onSaveQueryClick = useCallback(() => setShowSavedSearchModal(true), [])

    const onSaveQueryModalClose = useCallback(() => {
        setShowSavedSearchModal(false)
        telemetryService.log('SavedQueriesToggleCreating', { queries: { creating: false } })
    }, [telemetryService])

    // Reset expanded state when new search is started
    useEffect(() => {
        setAllExpanded(false)
    }, [location.search])

    const handleSidebarSearchSubmit = useCallback(
        (updates: QueryUpdate[]) =>
            submitQuerySearch(
                {
                    activation: props.activation,
                    selectedSearchContextSpec: props.selectedSearchContextSpec,
                    history,
                    source: 'filter',
                },
                updates
            ),
        [submitQuerySearch, props.activation, props.selectedSearchContextSpec, history]
    )

    const onSearchAgain = useCallback(
        (additionalFilters: string[]) => {
            telemetryService.log('SearchSkippedResultsAgainClicked')
            submitSearch({
                ...props,
                caseSensitive,
                patternType,
                query: applyAdditionalFilters(submittedURLQuery, additionalFilters),
                source: 'excludedResults',
            })
        },
        [submittedURLQuery, telemetryService, patternType, caseSensitive, props]
    )

    const handleSearchAggregationBarClick = (query: string): void => {
        submitSearch({
            ...props,
            caseSensitive,
            patternType,
            query,
            source: 'nav',
        })
    }

    return (
        <div className={classNames(styles.container, selectedTab !== 'filters' && styles.containerWithSidebarHidden)}>
            <PageTitle key="page-title" title={submittedURLQuery} />

            <SidebarButtonStrip className={styles.sidebarButtonStrip} />

            <SearchFiltersSidebar
                liveQuery={liveQuery}
                submittedURLQuery={submittedURLQuery}
                patternType={patternType}
                filters={results?.filters}
                selectedSearchContextSpec={props.selectedSearchContextSpec}
                aggregationUIMode={aggregationUIMode}
                settingsCascade={props.settingsCascade}
                telemetryService={props.telemetryService}
                className={classNames(styles.sidebar, showMobileSidebar && styles.sidebarShowMobile)}
                onNavbarQueryChange={setQueryState}
                onSearchSubmit={handleSidebarSearchSubmit}
            >
                <GettingStartedTour
                    className="mb-1"
                    isSourcegraphDotCom={props.isSourcegraphDotCom}
                    telemetryService={props.telemetryService}
                    isAuthenticated={!!props.authenticatedUser}
                />
            </SearchFiltersSidebar>

            {aggregationUIMode === AggregationUIMode.SearchPage && (
                <SearchAggregationResult
                    query={submittedURLQuery}
                    patternType={patternType}
                    aria-label="Aggregation results panel"
                    className={styles.contents}
                    onQuerySubmit={handleSearchAggregationBarClick}
                />
            )}

            {aggregationUIMode !== AggregationUIMode.SearchPage && (
                <>
                    <SearchResultsInfoBar
                        {...props}
                        patternType={patternType}
                        caseSensitive={caseSensitive}
                        query={submittedURLQuery}
                        enableCodeInsights={codeInsightsEnabled && isCodeInsightsEnabled(props.settingsCascade)}
                        enableCodeMonitoring={enableCodeMonitoring}
                        resultsFound={resultsFound}
                        className={classNames('flex-grow-1', styles.infobar)}
                        allExpanded={allExpanded}
                        onExpandAllResultsToggle={onExpandAllResultsToggle}
                        onSaveQueryClick={onSaveQueryClick}
                        onShowFiltersChanged={show => setShowMobileSidebar(show)}
                        stats={
                            <StreamingProgress
                                progress={results?.progress || { durationMs: 0, matchCount: 0, skipped: [] }}
                                state={results?.state || 'loading'}
                                onSearchAgain={onSearchAgain}
                                showTrace={!!trace}
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

                        {results?.alert?.kind && <LuckySearch alert={results?.alert} />}

                        <GettingStartedTour.Info
                            className="mt-2 mb-3"
                            isSourcegraphDotCom={props.isSourcegraphDotCom}
                        />

                        {showSavedSearchModal && (
                            <SavedSearchModal
                                {...props}
                                patternType={patternType}
                                query={submittedURLQuery}
                                authenticatedUser={authenticatedUser}
                                onDidCancel={onSaveQueryModalClose}
                            />
                        )}
                        {results?.alert && !results?.alert.kind && (
                            <div className={classNames(styles.alertArea, 'mt-4')}>
                                <SearchAlert
                                    alert={results.alert}
                                    caseSensitive={caseSensitive}
                                    patternType={patternType}
                                />
                            </div>
                        )}

                        <StreamingSearchResultsList
                            {...props}
                            results={results}
                            allExpanded={allExpanded}
                            showSearchContext={showSearchContext}
                            assetsRoot={window.context?.assetsRoot || ''}
                            executedQuery={location.search}
                            luckySearchEnabled={luckySearchEnabled}
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
