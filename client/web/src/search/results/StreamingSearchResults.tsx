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
} from '@sourcegraph/search-ui'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { CtaAlert } from '@sourcegraph/shared/src/components/CtaAlert'
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
import { buildGetStartedURL } from '@sourcegraph/shared/src/util/url'
import { useLocalStorage } from '@sourcegraph/wildcard'

import { SearchStreamingProps } from '..'
import { AuthenticatedUser } from '../../auth'
import { SearchBetaIcon } from '../../components/CtaIcons'
import { PageTitle } from '../../components/PageTitle'
import { usePersistentCadence } from '../../hooks'
import { useIsActiveIdeIntegrationUser } from '../../IdeExtensionTracker'
import { CodeInsightsProps } from '../../insights/types'
import { isCodeInsightsEnabled } from '../../insights/utils/is-code-insights-enabled'
import { BrowserExtensionAlert } from '../../repo/actions/BrowserExtensionAlert'
import { IDEExtensionAlert } from '../../repo/actions/IdeExtensionAlert'
import { SavedSearchModal } from '../../savedSearches/SavedSearchModal'
import {
    useExperimentalFeatures,
    useNavbarQueryState,
    useNotepad,
    buildSearchURLQueryFromQueryState,
} from '../../stores'
import { useTourQueryParameters } from '../../tour/components/Tour/TourAgent'
import { GettingStartedTour } from '../../tour/GettingStartedTour'
import { useIsBrowserExtensionActiveUser } from '../../tracking/BrowserExtensionTracker'
import { SearchUserNeedsCodeHost } from '../../user/settings/codeHosts/OrgUserNeedsCodeHost'
import { submitSearch } from '../helpers'

import { DidYouMean } from './DidYouMean'
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

const CTA_ALERTS_CADENCE_KEY = 'SearchResultCtaAlerts.pageViews'
const CTA_ALERT_DISPLAY_CADENCE = 6
const IDE_CTA_CADENCE_SHIFT = 3

type CtaToDisplay = 'signup' | 'browser' | 'ide'

function useCtaAlert(
    isAuthenticated: boolean,
    areResultsFound: boolean
): {
    ctaToDisplay?: CtaToDisplay
    onCtaAlertDismissed: () => void
} {
    const [hasDismissedSignupAlert, setHasDismissedSignupAlert] = useLocalStorage<boolean>(
        'StreamingSearchResults.hasDismissedSignupAlert',
        false
    )
    const [hasDismissedBrowserExtensionAlert, setHasDismissedBrowserExtensionAlert] = useTemporarySetting(
        'cta.browserExtensionAlertDismissed',
        false
    )
    const [hasDismissedIDEExtensionAlert, setHasDismissedIDEExtensionAlert] = useTemporarySetting(
        'cta.ideExtensionAlertDismissed',
        false
    )
    const isBrowserExtensionActiveUser = useIsBrowserExtensionActiveUser()
    const isUsingIdeIntegration = useIsActiveIdeIntegrationUser()

    const displaySignupAndBrowserExtensionCTAsBasedOnCadence = usePersistentCadence(
        CTA_ALERTS_CADENCE_KEY,
        CTA_ALERT_DISPLAY_CADENCE
    )
    const displayIDEExtensionCTABasedOnCadence = usePersistentCadence(
        CTA_ALERTS_CADENCE_KEY,
        CTA_ALERT_DISPLAY_CADENCE,
        IDE_CTA_CADENCE_SHIFT
    )

    const tourQueryParameters = useTourQueryParameters()

    const ctaToDisplay = useMemo<CtaToDisplay | undefined>((): CtaToDisplay | undefined => {
        if (!areResultsFound) {
            return
        }
        if (tourQueryParameters.isTour) {
            return
        }

        if (!hasDismissedSignupAlert && !isAuthenticated && displaySignupAndBrowserExtensionCTAsBasedOnCadence) {
            return 'signup'
        }

        if (
            hasDismissedBrowserExtensionAlert === false &&
            isAuthenticated &&
            isBrowserExtensionActiveUser === false &&
            displaySignupAndBrowserExtensionCTAsBasedOnCadence
        ) {
            return 'browser'
        }

        if (
            isUsingIdeIntegration === false &&
            displayIDEExtensionCTABasedOnCadence &&
            hasDismissedIDEExtensionAlert === false
        ) {
            return 'ide'
        }

        return
    }, [
        areResultsFound,
        tourQueryParameters?.isTour,
        hasDismissedSignupAlert,
        isAuthenticated,
        displaySignupAndBrowserExtensionCTAsBasedOnCadence,
        hasDismissedBrowserExtensionAlert,
        isBrowserExtensionActiveUser,
        isUsingIdeIntegration,
        displayIDEExtensionCTABasedOnCadence,
        hasDismissedIDEExtensionAlert,
    ])

    const onCtaAlertDismissed = useCallback((): void => {
        if (ctaToDisplay === 'signup') {
            setHasDismissedSignupAlert(true)
        } else if (ctaToDisplay === 'browser') {
            setHasDismissedBrowserExtensionAlert(true)
        } else if (ctaToDisplay === 'ide') {
            setHasDismissedIDEExtensionAlert(true)
        }
    }, [
        ctaToDisplay,
        setHasDismissedBrowserExtensionAlert,
        setHasDismissedIDEExtensionAlert,
        setHasDismissedSignupAlert,
    ])

    return {
        ctaToDisplay,
        onCtaAlertDismissed,
    }
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
            console.error(results.error)
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
    const [showSidebar, setShowSidebar] = useState(false)

    const onSignUpClick = useCallback((): void => {
        const args = { page: 'search' }
        telemetryService.log('SignUpPLGSearchCTA_1_Search', args, args)
    }, [telemetryService])

    const resultsFound = useMemo<boolean>(() => (results ? results.results.length > 0 : false), [results])
    const { ctaToDisplay, onCtaAlertDismissed } = useCtaAlert(!!authenticatedUser, resultsFound)

    // Log view event when signup CTA is shown
    useEffect(() => {
        if (ctaToDisplay === 'signup') {
            telemetryService.log('SearchResultResultsCTAShown')
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [ctaToDisplay])

    return (
        <div className={styles.streamingSearchResults}>
            <PageTitle key="page-title" title={query} />

            <SearchSidebar
                activation={props.activation}
                caseSensitive={caseSensitive}
                patternType={patternType}
                settingsCascade={props.settingsCascade}
                telemetryService={props.telemetryService}
                selectedSearchContextSpec={props.selectedSearchContextSpec}
                className={classNames(
                    styles.streamingSearchResultsSidebar,
                    showSidebar && styles.streamingSearchResultsSidebarShow
                )}
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
                className={classNames('flex-grow-1', styles.streamingSearchResultsInfobar)}
                allExpanded={allExpanded}
                onExpandAllResultsToggle={onExpandAllResultsToggle}
                onSaveQueryClick={onSaveQueryClick}
                onShowFiltersChanged={show => setShowSidebar(show)}
                stats={
                    <StreamingProgress
                        progress={results?.progress || { durationMs: 0, matchCount: 0, skipped: [] }}
                        state={results?.state || 'loading'}
                        onSearchAgain={onSearchAgain}
                        showTrace={!!trace}
                    />
                }
            />

            <DidYouMean
                telemetryService={props.telemetryService}
                query={query}
                patternType={patternType}
                caseSensitive={caseSensitive}
                selectedSearchContextSpec={props.selectedSearchContextSpec}
            />

            <div className={styles.streamingSearchResultsContainer}>
                <GettingStartedTour.Info className="mt-2 mr-3 mb-3" isSourcegraphDotCom={props.isSourcegraphDotCom} />
                {showSavedSearchModal && (
                    <SavedSearchModal
                        {...props}
                        patternType={patternType}
                        query={query}
                        authenticatedUser={authenticatedUser}
                        onDidCancel={onSaveQueryModalClose}
                    />
                )}
                {results?.alert && (
                    <div className={classNames(styles.streamingSearchResultsContentCentered, 'mt-4')}>
                        <SearchAlert alert={results.alert} caseSensitive={caseSensitive} patternType={patternType} />
                    </div>
                )}
                {ctaToDisplay === 'signup' && (
                    <CtaAlert
                        title="Sign up to add your public and private repositories and unlock search flow"
                        description="Do all the things editors can’t: search multiple repos & commit history, monitor, save
                searches and more."
                        cta={{
                            label: 'Get started',
                            href: buildGetStartedURL('search-cta', '/user/settings/repositories'),
                            onClick: onSignUpClick,
                        }}
                        icon={<SearchBetaIcon />}
                        className="mr-3 percy-display-none"
                        onClose={onCtaAlertDismissed}
                    />
                )}
                {ctaToDisplay === 'browser' && (
                    <BrowserExtensionAlert
                        className="mr-3 percy-display-none"
                        onAlertDismissed={onCtaAlertDismissed}
                        page="search"
                    />
                )}
                {ctaToDisplay === 'ide' && (
                    <IDEExtensionAlert
                        className="mr-3 percy-display-none"
                        onAlertDismissed={onCtaAlertDismissed}
                        page="search"
                    />
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
