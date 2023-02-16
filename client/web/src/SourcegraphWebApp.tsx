import 'focus-visible'

import * as React from 'react'
import { useCallback, useEffect, useRef, useState } from 'react'

import { ApolloProvider } from '@apollo/client'
import ServerIcon from 'mdi-react/ServerIcon'
import { Router } from 'react-router'
import { CompatRouter, Routes, Route } from 'react-router-dom-v5-compat'
import { combineLatest, from, Subscription, fromEvent, of, Subject, Observable } from 'rxjs'
import { distinctUntilChanged, first, map, startWith, switchMap } from 'rxjs/operators'

import { isMacPlatform, logger } from '@sourcegraph/common'
import { GraphQLClient, HTTPStatusError } from '@sourcegraph/http-client'
import { SharedSpanName, TraceSpanProvider } from '@sourcegraph/observability-client'
import { FetchFileParameters, fetchHighlightedFileLineRanges } from '@sourcegraph/shared/src/backend/file'
import { setCodeIntelSearchContext } from '@sourcegraph/shared/src/codeintel/searchContext'
import { ShortcutProvider } from '@sourcegraph/shared/src/react-shortcuts'
import {
    getUserSearchContextNamespaces,
    SearchContextProps,
    fetchSearchContexts,
    fetchSearchContext,
    fetchSearchContextBySpec,
    createSearchContext,
    updateSearchContext,
    deleteSearchContext,
    isSearchContextSpecAvailable,
    SearchQueryStateStoreProvider,
    getDefaultSearchContextSpec,
} from '@sourcegraph/shared/src/search'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'
import { aggregateStreamingSearch } from '@sourcegraph/shared/src/search/stream'
import {
    EMPTY_SETTINGS_CASCADE,
    Settings,
    SettingsCascadeOrError,
    SettingsSubjectCommonFields,
} from '@sourcegraph/shared/src/settings/settings'
import { TemporarySettingsProvider } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsProvider'
import { TemporarySettingsStorage } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsStorage'
import { globbingEnabledFromSettings } from '@sourcegraph/shared/src/util/globbing'
import { FeedbackText, setLinkComponent, RouterLink, WildcardThemeContext, WildcardTheme } from '@sourcegraph/wildcard'

import { authenticatedUser, AuthenticatedUser } from './auth'
import { getWebGraphQLClient } from './backend/graphql'
import { BatchChangesProps, isBatchChangesExecutionEnabled } from './batches'
import type { CodeIntelligenceProps } from './codeintel'
import { CodeMonitoringProps } from './codeMonitoring'
import { useBreadcrumbs } from './components/Breadcrumbs'
import { ComponentsComposer } from './components/ComponentsComposer'
import { ErrorBoundary } from './components/ErrorBoundary'
import { HeroPage } from './components/HeroPage'
import { FeatureFlagsProvider } from './featureFlags/FeatureFlagsProvider'
import type { CodeInsightsProps } from './insights/types'
import { Layout } from './Layout'
import { NotebookProps } from './notebooks'
import type { OrgAreaRoute } from './org/area/OrgArea'
import type { OrgAreaHeaderNavItem } from './org/area/OrgHeader'
import type { OrgSettingsAreaRoute } from './org/settings/OrgSettingsArea'
import type { OrgSettingsSidebarItems } from './org/settings/OrgSettingsSidebar'
import { createPlatformContext } from './platform/context'
import type { RepoContainerRoute } from './repo/RepoContainer'
import type { RepoHeaderActionButton } from './repo/RepoHeader'
import type { RepoRevisionContainerRoute } from './repo/RepoRevisionContainer'
import type { RepoSettingsAreaRoute } from './repo/settings/RepoSettingsArea'
import type { RepoSettingsSideBarGroup } from './repo/settings/RepoSettingsSidebar'
import type { LayoutRouteProps, LegacyLayoutRouteComponentProps } from './routes'
import { parseSearchURL, getQueryStateFromLocation, SearchAggregationProps } from './search'
import { SearchResultsCacheProvider } from './search/results/SearchResultsCacheProvider'
import type { SiteAdminAreaRoute } from './site-admin/SiteAdminArea'
import type { SiteAdminSideBarGroups } from './site-admin/SiteAdminSidebar'
import {
    setQueryStateFromSettings,
    setExperimentalFeaturesFromSettings,
    getExperimentalFeatures,
    useNavbarQueryState,
    observeStore,
    useExperimentalFeatures,
} from './stores'
import { setQueryStateFromURL } from './stores/navbarSearchQueryState'
import { useThemeProps } from './theme'
import { eventLogger } from './tracking/eventLogger'
import type { UserAreaRoute } from './user/area/UserArea'
import type { UserAreaHeaderNavItem } from './user/area/UserAreaHeader'
import type { UserSettingsAreaRoute } from './user/settings/UserSettingsArea'
import type { UserSettingsSidebarItems } from './user/settings/UserSettingsSidebar'
import { UserSessionStores } from './UserSessionStores'
import { globalHistory } from './util/globalHistory'
import { observeLocation } from './util/location'
import { siteSubjectNoAdmin, viewerSubjectFromSettings } from './util/settings'

import styles from './LegacySourcegraphWebApp.module.scss'

interface SourcegraphWebAppProps
    extends CodeIntelligenceProps,
        CodeInsightsProps,
        Pick<BatchChangesProps, 'batchChangesEnabled'>,
        Pick<SearchContextProps, 'searchContextsEnabled'>,
        NotebookProps,
        CodeMonitoringProps,
        SearchAggregationProps {
    siteAdminAreaRoutes: readonly SiteAdminAreaRoute[]
    siteAdminSideBarGroups: SiteAdminSideBarGroups
    siteAdminOverviewComponents: readonly React.ComponentType<React.PropsWithChildren<unknown>>[]
    userAreaHeaderNavItems: readonly UserAreaHeaderNavItem[]
    userAreaRoutes: readonly UserAreaRoute[]
    userSettingsSideBarItems: UserSettingsSidebarItems
    userSettingsAreaRoutes: readonly UserSettingsAreaRoute[]
    orgSettingsSideBarItems: OrgSettingsSidebarItems
    orgSettingsAreaRoutes: readonly OrgSettingsAreaRoute[]
    orgAreaHeaderNavItems: readonly OrgAreaHeaderNavItem[]
    orgAreaRoutes: readonly OrgAreaRoute[]
    repoContainerRoutes: readonly RepoContainerRoute[]
    repoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[]
    repoHeaderActionButtons: readonly RepoHeaderActionButton[]
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarGroups: readonly RepoSettingsSideBarGroup[]
    routes: readonly LayoutRouteProps[]
}

const WILDCARD_THEME: WildcardTheme = {
    isBranded: true,
}

const GLOBAL_SEARCH_CONTEXT_SPEC = 'global'

setLinkComponent(RouterLink)

export const SourcegraphWebApp: React.FC<SourcegraphWebAppProps> = props => {
    const [subscriptions] = useState(() => new Subscription())
    const [userRepositoriesUpdates] = useState(() => new Subject<void>())
    const [platformContext] = useState(() => createPlatformContext())

    const [resolvedAuthenticatedUser, setResolvedAuthenticatedUser] = useState<AuthenticatedUser | null>(null)
    const [settingsCascade, setSettingsCascade] = useState<SettingsCascadeOrError<Settings>>(EMPTY_SETTINGS_CASCADE)
    const [viewerSubject, setViewerSubject] = useState<SettingsSubjectCommonFields>(() => siteSubjectNoAdmin())
    const [globbing, setGlobbing] = useState(false)

    const [graphqlClient, setGraphqlClient] = useState<GraphQLClient | null>(null)
    const [temporarySettingsStorage, setTemporarySettingsStorage] = useState<TemporarySettingsStorage | null>(null)

    const [selectedSearchContextSpec, _setSelectedSearchContextSpec] = useState<string | null>(null)

    // NOTE(2022-09-08) Inform the inlined code from
    // sourcegraph/code-intel-extensions about the change of search context.
    // The old extension code previously accessed this information from the
    // 'sourcegraph' npm package, and updating the context like this was the
    // simplest solution to mirror the old behavior while deprecating
    // extensions on a tight deadline. It would be nice to properly pass
    // around this via React state in the future.
    const setWorkspaceSearchContext = useCallback((spec: string | null): void => {
        setCodeIntelSearchContext(spec ?? undefined)
    }, [])
    const setSelectedSearchContextSpecWithNoChecks = useCallback(
        (spec: string): void => {
            _setSelectedSearchContextSpec(spec)
            setWorkspaceSearchContext(spec)
        },
        [setWorkspaceSearchContext]
    )
    const setSelectedSearchContextSpecToDefault = useCallback((): void => {
        if (!props.searchContextsEnabled) {
            return
        }
        subscriptions.add(
            getDefaultSearchContextSpec({ platformContext }).subscribe(spec => {
                // Fall back to global if no default is returned.
                setSelectedSearchContextSpecWithNoChecks(spec || GLOBAL_SEARCH_CONTEXT_SPEC)
            })
        )
    }, [platformContext, props.searchContextsEnabled, setSelectedSearchContextSpecWithNoChecks, subscriptions])
    const setSelectedSearchContextSpec = useCallback(
        (spec: string): void => {
            if (!props.searchContextsEnabled) {
                return
            }

            // The global search context is always available.
            if (spec === GLOBAL_SEARCH_CONTEXT_SPEC) {
                setSelectedSearchContextSpecWithNoChecks(spec)
            }

            // Check if the wanted search context is available.
            subscriptions.add(
                isSearchContextSpecAvailable({
                    spec,
                    platformContext,
                }).subscribe(isAvailable => {
                    if (isAvailable) {
                        setSelectedSearchContextSpecWithNoChecks(spec)
                    } else if (!selectedSearchContextSpec) {
                        // If the wanted search context is not available and
                        // there is no currently selected search context,
                        // set the current selection to the default search context.
                        // Otherwise, keep the current selection.
                        setSelectedSearchContextSpecToDefault()
                    }
                })
            )
        },
        [
            platformContext,
            props.searchContextsEnabled,
            selectedSearchContextSpec,
            setSelectedSearchContextSpecToDefault,
            setSelectedSearchContextSpecWithNoChecks,
            subscriptions,
        ]
    )

    const _fetchHighlightedFileLineRanges = useCallback(
        (parameters: FetchFileParameters, force?: boolean | undefined): Observable<string[][]> =>
            fetchHighlightedFileLineRanges({ ...parameters, platformContext }, force),
        [platformContext]
    )

    const selectedSearchContextSpecRef = useRef(selectedSearchContextSpec)
    selectedSearchContextSpecRef.current = selectedSearchContextSpec

    const getSelectedSearchContextSpec = useCallback(
        (): string | undefined =>
            getExperimentalFeatures().showSearchContext ? selectedSearchContextSpecRef.current ?? undefined : undefined,
        []
    )

    // TODO: Move all of this initialization outside React so we don't need to
    // handle the optional states everywhere
    useEffect(() => {
        const parsedSearchURL = parseSearchURL(window.location.search)
        const parsedSearchQuery = parsedSearchURL.query || ''

        document.documentElement.classList.add('theme')

        getWebGraphQLClient()
            .then(graphqlClient => {
                setGraphqlClient(graphqlClient)
                setTemporarySettingsStorage(
                    new TemporarySettingsStorage(graphqlClient, window.context.isAuthenticatedUser)
                )
            })
            .catch(error => {
                logger.error('Error initializing GraphQL client', error)
            })

        subscriptions.add(
            combineLatest([
                from(platformContext.settings),
                // Start with `undefined` while we don't know if the viewer is authenticated or not.
                authenticatedUser.pipe(startWith(undefined)),
            ]).subscribe(
                ([settingsCascade, authenticatedUser]) => {
                    setExperimentalFeaturesFromSettings(settingsCascade)
                    setQueryStateFromSettings(settingsCascade)
                    setSettingsCascade(settingsCascade)
                    setResolvedAuthenticatedUser(authenticatedUser ?? null)
                    setGlobbing(globbingEnabledFromSettings(settingsCascade))
                    setViewerSubject(viewerSubjectFromSettings(settingsCascade, authenticatedUser))
                },
                () => setResolvedAuthenticatedUser(null)
            )
        )

        /**
         * Listens for uncaught 401 errors when a user when a user was previously authenticated.
         *
         * Don't subscribe to this event when there wasn't an authenticated user,
         * as it could lead to an infinite loop of 401 -> reload -> 401
         */
        subscriptions.add(
            authenticatedUser
                .pipe(
                    switchMap(authenticatedUser =>
                        authenticatedUser ? fromEvent<ErrorEvent>(window, 'error') : of(null)
                    )
                )
                .subscribe(event => {
                    if (event?.error instanceof HTTPStatusError && event.error.status === 401) {
                        location.reload()
                    }
                })
        )

        if (parsedSearchQuery && !filterExists(parsedSearchQuery, FilterType.context)) {
            // If a context filter does not exist in the query, we have to switch the selected context
            // to global to match the UI with the backend semantics (if no context is specified in the query,
            // the query is run in global context).
            setSelectedSearchContextSpecWithNoChecks(GLOBAL_SEARCH_CONTEXT_SPEC)
        }
        if (!parsedSearchQuery) {
            // If no query is present (e.g. search page, settings page),
            // select the user's default search context.
            setSelectedSearchContextSpecToDefault()
        }

        setWorkspaceSearchContext(selectedSearchContextSpec)

        // Update search query state whenever the URL changes
        subscriptions.add(
            combineLatest([
                observeStore(useExperimentalFeatures).pipe(
                    map(([features]) => features.searchQueryInput === 'experimental'),
                    // This ensures that the query stays unmodified until we know
                    // whether the feature flag is set or not.
                    startWith(true),
                    distinctUntilChanged()
                ),
                getQueryStateFromLocation({
                    location: observeLocation(globalHistory).pipe(startWith(globalHistory.location)),
                    isSearchContextAvailable: (searchContext: string) =>
                        props.searchContextsEnabled
                            ? isSearchContextSpecAvailable({ spec: searchContext, platformContext })
                                  .pipe(first())
                                  .toPromise()
                            : Promise.resolve(false),
                }),
            ]).subscribe(([enableExperimentalQueryInput, parsedSearchURLAndContext]) => {
                if (parsedSearchURLAndContext.query) {
                    // Only override filters and update query from URL if there
                    // is a search query.
                    if (!parsedSearchURLAndContext.searchContextSpec) {
                        // If no search context is present we have to fall back
                        // to the global search context to match the server
                        // behavior.
                        setSelectedSearchContextSpec(GLOBAL_SEARCH_CONTEXT_SPEC)
                    } else if (
                        parsedSearchURLAndContext.searchContextSpec.spec !== selectedSearchContextSpecRef.current
                    ) {
                        setSelectedSearchContextSpec(parsedSearchURLAndContext.searchContextSpec.spec)
                    }

                    const processedQuery =
                        !enableExperimentalQueryInput &&
                        parsedSearchURLAndContext.searchContextSpec &&
                        props.searchContextsEnabled
                            ? omitFilter(
                                  parsedSearchURLAndContext.query,
                                  parsedSearchURLAndContext.searchContextSpec.filter
                              )
                            : parsedSearchURLAndContext.query

                    setQueryStateFromURL(parsedSearchURLAndContext, processedQuery)
                }
            })
        )

        userRepositoriesUpdates.next()

        return () => subscriptions.unsubscribe()

        // We only ever want to run this hook once when the component mounts for
        // parity with the old behavior.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    const breadcrumbProps = useBreadcrumbs()
    const themeProps = useThemeProps()

    const context = {
        ...props,
        ...themeProps,
        ...breadcrumbProps,
        isMacPlatform: isMacPlatform(),
        telemetryService: eventLogger,
        isSourcegraphDotCom: window.context.sourcegraphDotComMode,
        selectedSearchContextSpec: getSelectedSearchContextSpec(),
        setSelectedSearchContextSpec,
        getUserSearchContextNamespaces,
        fetchSearchContexts,
        fetchSearchContextBySpec,
        fetchSearchContext,
        createSearchContext,
        updateSearchContext,
        deleteSearchContext,
        isSearchContextSpecAvailable,
        globbing,
        streamSearch: aggregateStreamingSearch,
        codeIntelligenceEnabled: !!props.codeInsightsEnabled,
        notebooksEnabled: props.notebooksEnabled,
        codeMonitoringEnabled: props.codeMonitoringEnabled,
        searchAggregationEnabled: props.searchAggregationEnabled,
        batchChangesExecutionEnabled: isBatchChangesExecutionEnabled(settingsCascade),
        platformContext,
        authenticatedUser: resolvedAuthenticatedUser,
        viewerSubject,
        fetchHighlightedFileLineRanges: _fetchHighlightedFileLineRanges,
        settingsCascade,
        extensionsController: null,
        batchChangesWebhookLogsEnabled: window.context.batchChangesWebhookLogsEnabled,
    } satisfies Omit<LegacyLayoutRouteComponentProps, 'location' | 'history' | 'match' | 'staticContext'>

    if (window.pageError && window.pageError.statusCode !== 404) {
        const statusCode = window.pageError.statusCode
        const statusText = window.pageError.statusText
        const errorMessage = window.pageError.error
        const errorID = window.pageError.errorID

        let subtitle: JSX.Element | undefined
        if (errorID) {
            subtitle = <FeedbackText headerText="Sorry, there's been a problem." />
        }
        if (errorMessage) {
            subtitle = (
                <div className={styles.error}>
                    {subtitle}
                    {subtitle && <hr className="my-3" />}
                    <pre>{errorMessage}</pre>
                </div>
            )
        } else {
            subtitle = <div className={styles.error}>{subtitle}</div>
        }

        return <HeroPage icon={ServerIcon} title={`${statusCode}: ${statusText}`} subtitle={subtitle} />
    }

    if (authenticatedUser === null || graphqlClient === null || temporarySettingsStorage === null) {
        return null
    }

    return (
        <ComponentsComposer
            components={[
                // `ComponentsComposer` provides children via `React.cloneElement`.
                /* eslint-disable react/no-children-prop, react/jsx-key */
                <ApolloProvider client={graphqlClient} children={undefined} />,
                <WildcardThemeContext.Provider value={WILDCARD_THEME} />,
                <ErrorBoundary location={null} />,
                <TraceSpanProvider name={SharedSpanName.AppMount} />,
                <FeatureFlagsProvider />,
                <ShortcutProvider />,
                <TemporarySettingsProvider temporarySettingsStorage={temporarySettingsStorage} />,
                <SearchResultsCacheProvider />,
                <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState} />,
                /* eslint-enable react/no-children-prop, react/jsx-key */
            ]}
        >
            <Router history={globalHistory}>
                <CompatRouter>
                    <Routes>
                        <Route
                            path="*"
                            element={
                                <Layout
                                    authenticatedUser={resolvedAuthenticatedUser}
                                    viewerSubject={viewerSubject}
                                    settingsCascade={settingsCascade}
                                    batchChangesEnabled={props.batchChangesEnabled}
                                    batchChangesExecutionEnabled={isBatchChangesExecutionEnabled(settingsCascade)}
                                    batchChangesWebhookLogsEnabled={window.context.batchChangesWebhookLogsEnabled}
                                    // Search query
                                    fetchHighlightedFileLineRanges={_fetchHighlightedFileLineRanges}
                                    // Extensions
                                    platformContext={platformContext}
                                    extensionsController={null}
                                    telemetryService={eventLogger}
                                    isSourcegraphDotCom={window.context.sourcegraphDotComMode}
                                    searchContextsEnabled={props.searchContextsEnabled}
                                    selectedSearchContextSpec={getSelectedSearchContextSpec()}
                                    setSelectedSearchContextSpec={setSelectedSearchContextSpec}
                                    getUserSearchContextNamespaces={getUserSearchContextNamespaces}
                                    fetchSearchContexts={fetchSearchContexts}
                                    fetchSearchContextBySpec={fetchSearchContextBySpec}
                                    fetchSearchContext={fetchSearchContext}
                                    createSearchContext={createSearchContext}
                                    updateSearchContext={updateSearchContext}
                                    deleteSearchContext={deleteSearchContext}
                                    isSearchContextSpecAvailable={isSearchContextSpecAvailable}
                                    globbing={globbing}
                                    streamSearch={aggregateStreamingSearch}
                                    codeIntelligenceEnabled={!!props.codeInsightsEnabled}
                                    notebooksEnabled={props.notebooksEnabled}
                                    codeMonitoringEnabled={props.codeMonitoringEnabled}
                                    searchAggregationEnabled={props.searchAggregationEnabled}
                                    themeProps={themeProps}
                                />
                            }
                        >
                            {props.routes.map(
                                ({ condition = () => true, ...route }) =>
                                    condition(context) && (
                                        <Route
                                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                            path={route.path.slice(1)} // remove leading slash
                                            element={route.render(context)}
                                        />
                                    )
                            )}
                        </Route>
                    </Routes>
                </CompatRouter>
            </Router>
            <UserSessionStores />
        </ComponentsComposer>
    )
}
