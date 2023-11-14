import { FC, HTMLAttributes, PropsWithChildren, useCallback, useRef } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'
import { Observable } from 'rxjs'

import { StreamingProgress, StreamingSearchResultsList } from '@sourcegraph/branded'
import { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import { FilePrefetcher } from '@sourcegraph/shared/src/components/PrefetchableFile'
import { HighlightResponseFormat, SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { QueryState, QueryStateUpdate, QueryUpdate, SearchMode } from '@sourcegraph/shared/src/search'
import { AggregateStreamingSearchResults, StreamSearchOptions } from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Icon, H2, H4, useScrollManager } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { fetchBlob } from '../../../../repo/blob/backend'
import { isSearchJobsEnabled } from '../../../../search-jobs/utility'
import { buildSearchURLQueryFromQueryState, setSearchMode } from '../../../../stores'
import { GettingStartedTour } from '../../../../tour/GettingStartedTour'
import { submitSearch } from '../../../helpers'
import { DidYouMean } from '../../../suggestion/DidYouMean'
import { SmartSearch } from '../../../suggestion/SmartSearch'
import { SearchFiltersSidebar } from '../../sidebar/SearchFiltersSidebar'
import { AggregationUIMode, SearchAggregationResult } from '../aggregation'
import { SearchResultsInfoBar } from '../search-results-info-bar/SearchResultsInfoBar'
import { SearchAlert } from '../SearchAlert'
import { UnownedResultsAlert } from '../UnownedResultsAlert'
import { isSmartSearchAlert } from '../utils'

import styles from './NewSearchContent.module.scss'

interface NewSearchContentProps
    extends TelemetryProps,
        SettingsCascadeProps,
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

export const NewSearchContent: FC<NewSearchContentProps> = props => {
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

    const [sidebarCollapsed, setSidebarCollapsed] = useTemporarySetting('search.sidebar.collapsed', false)

    const containerRef = useRef<HTMLDivElement>(null)
    useScrollManager('SearchResultsContainer', containerRef)

    const prefetchFile: FilePrefetcher = useCallback(
        params =>
            fetchBlob({
                ...params,
                format: HighlightResponseFormat.JSON_SCIP,
            }),
        []
    )

    return (
        <div className={styles.root}>
            {!sidebarCollapsed && (
                <SearchFiltersSidebar
                    as={NewSearchSidebarWrapper}
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
                    className={classNames(styles.filters)}
                    setSidebarCollapsed={setSidebarCollapsed}
                    onNavbarQueryChange={onNavbarQueryChange}
                    onSearchSubmit={onSearchSubmit}
                />
            )}

            <SearchResultsInfoBar
                query={submittedURLQuery}
                patternType={patternType}
                results={results}
                options={options}
                allExpanded={allExpanded}
                caseSensitive={caseSensitive}
                enableCodeMonitoring={codeMonitoringEnabled}
                sidebarCollapsed={!!sidebarCollapsed}
                setSidebarCollapsed={setSidebarCollapsed}
                authenticatedUser={authenticatedUser}
                sourcegraphURL={platformContext.sourcegraphURL}
                isSourcegraphDotCom={isSourcegraphDotCom}
                telemetryService={telemetryService}
                className={styles.infobar}
                onExpandAllResultsToggle={onExpandAllResultsToggle}
                onShowMobileFiltersChanged={setSidebarCollapsed}
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

            <div className={styles.content} ref={containerRef}>
                {aggregationUIMode === AggregationUIMode.SearchPage && (
                    <SearchAggregationResult
                        query={submittedURLQuery}
                        patternType={patternType}
                        caseSensitive={caseSensitive}
                        aria-label="Aggregation results panel"
                        className="mt-3"
                        onQuerySubmit={onQuerySubmit}
                        telemetryService={telemetryService}
                    />
                )}

                {aggregationUIMode !== AggregationUIMode.SearchPage && (
                    <>
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
                            prefetchFileEnabled={true}
                            prefetchFile={prefetchFile}
                            enableKeyboardNavigation={true}
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
                    </>
                )}
            </div>
        </div>
    )
}

interface NewSearchSidebarWrapper extends HTMLAttributes<HTMLElement> {
    onClose: () => void
}

const NewSearchSidebarWrapper: FC<PropsWithChildren<NewSearchSidebarWrapper>> = props => {
    const { children, className, onClose, ...attributes } = props

    return (
        <aside {...attributes} className={classNames(styles.filters, className)}>
            <header className={styles.filtersHeader}>
                <H4 as={H2} className="mb-0">
                    Filters
                </H4>
                <Button variant="icon" aria-label="Close" onClick={onClose}>
                    <Icon aria-hidden={true} svgPath={mdiClose} />
                </Button>
            </header>
            <div className={styles.filtersContent}>{children}</div>
        </aside>
    )
}
