import React, { useCallback, useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { Observable } from 'rxjs'

import { asError } from '@sourcegraph/common'
import { SearchContextProps } from '@sourcegraph/search'
import {
    SearchSidebar,
    StreamingProgress,
    StreamingSearchResultsList,
    FetchFileParameters,
    SidebarButtonStrip,
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
import { CodeInsightsProps } from '../../insights/types'
import { isCodeInsightsEnabled } from '../../insights/utils/is-code-insights-enabled'
import { SavedSearchModal } from '../../savedSearches/SavedSearchModal'
import {
    useExperimentalFeatures,
    useNavbarQueryState,
    useNotepad,
    buildSearchURLQueryFromQueryState,
} from '../../stores'
import { GettingStartedTour } from '../../tour/GettingStartedTour'
import { SearchUserNeedsCodeHost } from '../../user/settings/codeHosts/OrgUserNeedsCodeHost'
import { submitSearch } from '../helpers'
import { DidYouMean } from '../suggestion/DidYouMean'
import { LuckySearch } from '../suggestion/LuckySearch'

import { SearchAlert } from './SearchAlert'
import { useCachedSearchResults } from './SearchResultsCacheProvider'
import { SearchResultsInfoBar } from './SearchResultsInfoBar'
import { getRevisions } from './sidebar/Revisions'

import styles from './StreamingSearchResults.module.scss'

export interface StreamingSearchResultsProps
    extends SearchStreamingProps,
        Pick<ActivationProps, 'activation'>,
        Pick<SearchContextProps, 'selectedSearchContextSpec' | 'searchContextsEnabled'>,
        SettingsCascadeProps,
        ExtensionsControllerProps<'executeCommand' | 'extHostAPI'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'requestGraphQL'>,
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
        extensionsController: { extHostAPI: extensionHostAPI },
    } = props

    const enableCodeMonitoring = useExperimentalFeatures(features => features.codeMonitoring ?? false)
    const showSearchContext = useExperimentalFeatures(features => features.showSearchContext ?? false)
    const caseSensitive = useNavbarQueryState(state => state.searchCaseSensitivity)
    const patternType = useNavbarQueryState(state => state.searchPatternType)
    const query = useNavbarQueryState(state => state.searchQueryFromURL)

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
        const metrics = query ? collectMetrics(query) : undefined

        telemetryService.log(
            'SearchResultsQueried',
            {
                code_search: {
                    query_data: {
                        query: metrics,
                        combined: query,
                        empty: !query,
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
                            !authenticatedUser && isSourcegraphDotCom ? sanitizeQueryForTelemetry(query) : undefined,
                        empty: !query,
                    },
                },
            }
        )
        // Only log when the query changes
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [query])

    const trace = useMemo(() => new URLSearchParams(location.search).get('trace') ?? undefined, [location.search])

    const options: StreamSearchOptions = useMemo(
        () => ({
            version: LATEST_VERSION,
            patternType: patternType ?? SearchPatternType.literal,
            caseSensitive,
            trace,
        }),
        [caseSensitive, patternType, trace]
    )

    const results = useCachedSearchResults(streamSearch, query, options, extensionHostAPI, telemetryService)

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
        } else if (results?.state === 'error') {
            telemetryService.log('SearchResultsFetchFailed', {
                code_search: { error_message: asError(results.error).message },
            })
        }
    }, [results, telemetryService])

    useNotepad(
        useMemo(
            () =>
                results?.state === 'complete'
                    ? {
                          type: 'search',
                          query,
                          caseSensitive,
                          patternType,
                          searchContext: props.selectedSearchContextSpec,
                      }
                    : null,
            [results, query, patternType, caseSensitive, props.selectedSearchContextSpec]
        )
    )

    const [allExpanded, setAllExpanded] = useState(false)
    const onExpandAllResultsToggle = useCallback(() => {
        setAllExpanded(oldValue => !oldValue)
        telemetryService.log(allExpanded ? 'allResultsExpanded' : 'allResultsCollapsed')
    }, [allExpanded, telemetryService])

    const [showSavedSearchModal, setShowSavedSearchModal] = useState(false)
    const onSaveQueryClick = useCallback(() => setShowSavedSearchModal(true), [])
    const onSaveQueryModalClose = useCallback(() => {
        setShowSavedSearchModal(false)
        telemetryService.log('SavedQueriesToggleCreating', { queries: { creating: false } })
    }, [telemetryService])

    // Reset expanded state when new search is started
    useEffect(() => {
        setAllExpanded(false)
    }, [location.search])

    const onSearchAgain = useCallback(
        (additionalFilters: string[]) => {
            telemetryService.log('SearchSkippedResultsAgainClicked')
            submitSearch({
                ...props,
                caseSensitive,
                patternType,
                query: applyAdditionalFilters(query, additionalFilters),
                source: 'excludedResults',
            })
        },
        [query, telemetryService, patternType, caseSensitive, props]
    )
    const [showMobileSidebar, setShowMobileSidebar] = useState(false)
    const [selectedTab] = useTemporarySetting('search.sidebar.selectedTab', 'filters')

    const resultsFound = useMemo<boolean>(() => (results ? results.results.length > 0 : false), [results])

    return (
        <div className={classNames(styles.container, selectedTab !== 'filters' && styles.containerWithSidebarHidden)}>
            <PageTitle key="page-title" title={query} />

            <SidebarButtonStrip className={styles.sidebarButtonStrip} />

            <SearchSidebar
                activation={props.activation}
                caseSensitive={caseSensitive}
                patternType={patternType}
                settingsCascade={props.settingsCascade}
                telemetryService={props.telemetryService}
                selectedSearchContextSpec={props.selectedSearchContextSpec}
                className={classNames(styles.sidebar, showMobileSidebar && styles.sidebarShowMobile)}
                filters={results?.filters}
                getRevisions={getRevisions}
                prefixContent={
                    <GettingStartedTour
                        className="mb-1"
                        isSourcegraphDotCom={props.isSourcegraphDotCom}
                        telemetryService={props.telemetryService}
                        isAuthenticated={!!props.authenticatedUser}
                    />
                }
                buildSearchURLQueryFromQueryState={buildSearchURLQueryFromQueryState}
            />

            <SearchResultsInfoBar
                {...props}
                patternType={patternType}
                caseSensitive={caseSensitive}
                query={query}
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
                    query={query}
                    patternType={patternType}
                    caseSensitive={caseSensitive}
                    selectedSearchContextSpec={props.selectedSearchContextSpec}
                />

                {results?.alert?.kind && <LuckySearch alert={results?.alert} />}

                <GettingStartedTour.Info className="mt-2 mb-3" isSourcegraphDotCom={props.isSourcegraphDotCom} />

                {showSavedSearchModal && (
                    <SavedSearchModal
                        {...props}
                        patternType={patternType}
                        query={query}
                        authenticatedUser={authenticatedUser}
                        onDidCancel={onSaveQueryModalClose}
                    />
                )}
                {results?.alert && !results?.alert.kind && (
                    <div className={classNames(styles.alertArea, 'mt-4')}>
                        <SearchAlert alert={results.alert} caseSensitive={caseSensitive} patternType={patternType} />
                    </div>
                )}

                <StreamingSearchResultsList
                    {...props}
                    results={results}
                    allExpanded={allExpanded}
                    showSearchContext={showSearchContext}
                    assetsRoot={window.context?.assetsRoot || ''}
                    renderSearchUserNeedsCodeHost={user => (
                        <SearchUserNeedsCodeHost user={user} orgSearchContext={props.selectedSearchContextSpec} />
                    )}
                    executedQuery={location.search}
                />
            </div>
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
