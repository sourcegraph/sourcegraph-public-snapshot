import { FC, useCallback, useState } from 'react'

import classNames from 'classnames'
import { useLocation } from 'react-router-dom'
import { Observable } from 'rxjs'

import { StreamingProgress, StreamingSearchResultsList } from '@sourcegraph/branded'
import { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import { FilePrefetcher } from '@sourcegraph/shared/src/components/PrefetchableFile'
import { HighlightResponseFormat, SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { QueryState, QueryStateUpdate, QueryUpdate, SearchMode } from '@sourcegraph/shared/src/search'
import {
    AggregateStreamingSearchResults,
    AlertKind,
    SmartSearchAlertKind,
    StreamSearchOptions,
} from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps, useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../../../../auth'
import { PageTitle } from '../../../../components/PageTitle'
import { useFeatureFlag } from '../../../../featureFlags/useFeatureFlag'
import { fetchBlob } from '../../../../repo/blob/backend'
import { isSearchJobsEnabled } from '../../../../search-jobs/utility'
import { buildSearchURLQueryFromQueryState, setSearchMode } from '../../../../stores'
import { GettingStartedTour } from '../../../../tour/GettingStartedTour'
import { useShowOnboardingTour } from '../../../../tour/hooks'
import { submitSearch } from '../../../helpers'
import { DidYouMean } from '../../../suggestion/DidYouMean'
import { SmartSearch } from '../../../suggestion/SmartSearch'
import { SearchFiltersSidebar } from '../../sidebar/SearchFiltersSidebar'
import { AggregationUIMode, SearchAggregationResult } from '../aggregation'
import { SearchResultsInfoBar } from '../search-results-info-bar/SearchResultsInfoBar'
import { SearchAlert } from '../SearchAlert'
import { UnownedResultsAlert } from '../UnownedResultsAlert'

import styles from './SearchContent.module.scss'

interface SearchContentProps
    extends SettingsCascadeProps,
        TelemetryProps,
        PlatformContextProps<'settings' | 'requestGraphQL' | 'sourcegraphURL'> {
    submittedURLQuery: string
    queryState: QueryState
    liveQuery: string
    allExpanded: boolean
    searchMode: SearchMode
    trace: boolean
    searchContextsEnabled: boolean
    patternType: SearchPatternType
    results: AggregateStreamingSearchResults | undefined
    showAggregationPanel: boolean
    selectedSearchContextSpec: string | undefined
    aggregationUIMode: AggregationUIMode
    caseSensitive: boolean
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    enableRepositoryMetadata: boolean
    options: StreamSearchOptions
    codeMonitoringEnabled: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    onNavbarQueryChange: (queryState: QueryStateUpdate) => void
    onSearchSubmit: (updates: QueryUpdate[], updatedSearchQuery?: string) => void
    onQuerySubmit: (newQuery: string, updatedQuerySearch: string) => void
    onExpandAllResultsToggle: () => void
    onSearchAgain: (additionalFilters: string[]) => void
    onDisableSmartSearch: () => void
    onLogSearchResultClick: (index: number, type: string, resultsLength: number) => void
}

export const SearchContent: FC<SearchContentProps> = props => {
    const {
        searchMode,
        submittedURLQuery,
        liveQuery,
        queryState,
        allExpanded,
        trace,
        patternType,
        searchContextsEnabled,
        results,
        showAggregationPanel,
        selectedSearchContextSpec,
        aggregationUIMode,
        settingsCascade,
        telemetryService,
        fetchHighlightedFileLineRanges,
        caseSensitive,
        authenticatedUser,
        isSourcegraphDotCom,
        enableRepositoryMetadata,
        codeMonitoringEnabled,
        options,
        platformContext,
        onNavbarQueryChange,
        onSearchSubmit,
        onQuerySubmit,
        onExpandAllResultsToggle,
        onSearchAgain,
        onDisableSmartSearch,
        onLogSearchResultClick,
    } = props

    const location = useLocation()
    const prefetchFileEnabled = useExperimentalFeatures(features => features.enableSearchFilePrefetch ?? false)
    const [enableSearchResultsKeyboardNavigation] = useFeatureFlag('search-results-keyboard-navigation', true)
    const [sidebarCollapsed, setSidebarCollapsed] = useTemporarySetting('search.sidebar.collapsed', false)
    const showOnboardingTour = useShowOnboardingTour({ authenticatedUser, isSourcegraphDotCom })

    const [showMobileSidebar, setShowMobileSidebar] = useState(false)

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
                selectedSearchContextSpec={selectedSearchContextSpec}
                aggregationUIMode={aggregationUIMode}
                settingsCascade={settingsCascade}
                telemetryService={telemetryService}
                caseSensitive={caseSensitive}
                className={classNames(styles.sidebar, showMobileSidebar && styles.sidebarShowMobile)}
                setSidebarCollapsed={setSidebarCollapsed}
                onNavbarQueryChange={onNavbarQueryChange}
                onSearchSubmit={onSearchSubmit}
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
                    onQuerySubmit={onQuerySubmit}
                    telemetryService={telemetryService}
                />
            )}

            {aggregationUIMode !== AggregationUIMode.SearchPage && (
                <>
                    <SearchResultsInfoBar
                        telemetryService={telemetryService}
                        isSourcegraphDotCom={isSourcegraphDotCom}
                        sourcegraphURL={platformContext.sourcegraphURL}
                        patternType={patternType}
                        authenticatedUser={authenticatedUser}
                        caseSensitive={caseSensitive}
                        query={submittedURLQuery}
                        results={results}
                        options={options}
                        enableCodeMonitoring={codeMonitoringEnabled}
                        className={styles.infobar}
                        allExpanded={allExpanded}
                        onExpandAllResultsToggle={onExpandAllResultsToggle}
                        onShowMobileFiltersChanged={show => setShowMobileSidebar(show)}
                        sidebarCollapsed={!!sidebarCollapsed}
                        setSidebarCollapsed={setSidebarCollapsed}
                        stats={
                            <StreamingProgress
                                showTrace={trace}
                                query={`${submittedURLQuery} patterntype:${patternType}`}
                                progress={results?.progress || { durationMs: 0, matchCount: 0, skipped: [] }}
                                state={results?.state || 'loading'}
                                onSearchAgain={onSearchAgain}
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
                            telemetryService={telemetryService}
                            platformContext={platformContext}
                            settingsCascade={settingsCascade}
                            searchContextsEnabled={searchContextsEnabled}
                            fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                            isSourcegraphDotCom={isSourcegraphDotCom}
                            enableRepositoryMetadata={enableRepositoryMetadata}
                            results={results}
                            allExpanded={allExpanded}
                            executedQuery={location.search}
                            prefetchFileEnabled={prefetchFileEnabled}
                            prefetchFile={prefetchFile}
                            enableKeyboardNavigation={enableSearchResultsKeyboardNavigation}
                            showQueryExamplesOnNoResultsPage={true}
                            queryState={queryState}
                            buildSearchURLQueryFromQueryState={buildSearchURLQueryFromQueryState}
                            searchMode={searchMode}
                            setSearchMode={setSearchMode}
                            submitSearch={submitSearch}
                            caseSensitive={caseSensitive}
                            searchQueryFromURL={submittedURLQuery}
                            selectedSearchContextSpec={selectedSearchContextSpec}
                            logSearchResultClicked={onLogSearchResultClick}
                        />
                    </div>
                </>
            )}
        </div>
    )
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
