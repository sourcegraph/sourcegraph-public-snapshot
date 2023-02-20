import 'focus-visible'

import * as React from 'react'

import { ApolloProvider } from '@apollo/client'
import ServerIcon from 'mdi-react/ServerIcon'
import { Router } from 'react-router'
import { CompatRouter, Routes, Route } from 'react-router-dom-v5-compat'
import { combineLatest, from, Subscription, fromEvent, of, Subject, Observable } from 'rxjs'
import { first, startWith, switchMap, map, distinctUntilChanged } from 'rxjs/operators'

import { logger } from '@sourcegraph/common'
import { GraphQLClient, HTTPStatusError } from '@sourcegraph/http-client'
import { SharedSpanName, TraceSpanProvider } from '@sourcegraph/observability-client'
import { NotificationType } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { FetchFileParameters, fetchHighlightedFileLineRanges } from '@sourcegraph/shared/src/backend/file'
import { setCodeIntelSearchContext } from '@sourcegraph/shared/src/codeintel/searchContext'
import { Controller as ExtensionsController } from '@sourcegraph/shared/src/extensions/controller'
import { createController as createExtensionsController } from '@sourcegraph/shared/src/extensions/createLazyLoadedController'
import { createNoopController } from '@sourcegraph/shared/src/extensions/createNoopLoadedController'
import { BrandedNotificationItemStyleProps } from '@sourcegraph/shared/src/notifications/NotificationItem'
import { Notifications } from '@sourcegraph/shared/src/notifications/Notifications'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
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
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TemporarySettingsProvider } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsProvider'
import { TemporarySettingsStorage } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsStorage'
import { globbingEnabledFromSettings } from '@sourcegraph/shared/src/util/globbing'
import { FeedbackText, setLinkComponent, RouterLink, WildcardThemeContext, WildcardTheme } from '@sourcegraph/wildcard'

import { authenticatedUser, AuthenticatedUser } from './auth'
import { getWebGraphQLClient } from './backend/graphql'
import { BatchChangesProps, isBatchChangesExecutionEnabled } from './batches'
import type { CodeIntelligenceProps } from './codeintel'
import { CodeMonitoringProps } from './codeMonitoring'
import { ComponentsComposer } from './components/ComponentsComposer'
import { ErrorBoundary } from './components/ErrorBoundary'
import { HeroPage } from './components/HeroPage'
import { FeatureFlagsProvider } from './featureFlags/FeatureFlagsProvider'
import type { CodeInsightsProps } from './insights/types'
import { LegacyLayout, LegacyLayoutProps } from './LegacyLayout'
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
import type { LayoutRouteProps } from './routes'
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

export interface LegacySourcegraphWebAppProps
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

interface LegacySourcegraphWebAppState extends SettingsCascadeProps {
    error?: Error

    /**
     * The currently authenticated user:
     * - `undefined` until `CurrentAuthState` query completion.
     * - `AuthenticatedUser` if the viewer is authenticated.
     * - `null` if the viewer is anonymous.
     */
    authenticatedUser?: AuthenticatedUser | null

    /** GraphQL client initialized asynchronously to restore persisted cache. */
    graphqlClient?: GraphQLClient

    temporarySettingsStorage?: TemporarySettingsStorage

    viewerSubject: LegacyLayoutProps['viewerSubject']

    selectedSearchContextSpec?: string

    /**
     * Whether globbing is enabled for filters.
     */
    globbing: boolean
}

const notificationStyles: BrandedNotificationItemStyleProps = {
    notificationItemVariants: {
        [NotificationType.Log]: 'secondary',
        [NotificationType.Success]: 'success',
        [NotificationType.Info]: 'info',
        [NotificationType.Warning]: 'warning',
        [NotificationType.Error]: 'danger',
    },
}

const WILDCARD_THEME: WildcardTheme = {
    isBranded: true,
}

const GLOBAL_SEARCH_CONTEXT_SPEC = 'global'

setLinkComponent(RouterLink)

/**
 * The root component.
 */
export class LegacySourcegraphWebApp extends React.Component<
    LegacySourcegraphWebAppProps,
    LegacySourcegraphWebAppState
> {
    private readonly subscriptions = new Subscription()
    private readonly userRepositoriesUpdates = new Subject<void>()
    private readonly platformContext: PlatformContext = createPlatformContext()
    private readonly extensionsController: ExtensionsController | null = window.context.enableLegacyExtensions
        ? createExtensionsController(this.platformContext)
        : createNoopController(this.platformContext)

    constructor(props: LegacySourcegraphWebAppProps) {
        super(props)

        if (this.extensionsController !== null) {
            this.subscriptions.add(this.extensionsController)
        }

        this.state = {
            settingsCascade: EMPTY_SETTINGS_CASCADE,
            viewerSubject: siteSubjectNoAdmin(),
            globbing: false,
        }
    }

    public componentDidMount(): void {
        const parsedSearchURL = parseSearchURL(window.location.search)
        const parsedSearchQuery = parsedSearchURL.query || ''

        document.documentElement.classList.add('theme')

        getWebGraphQLClient()
            .then(graphqlClient => {
                this.setState({
                    graphqlClient,
                    temporarySettingsStorage: new TemporarySettingsStorage(
                        graphqlClient,
                        window.context.isAuthenticatedUser
                    ),
                })
            })
            .catch(error => {
                logger.error('Error initializing GraphQL client', error)
            })

        this.subscriptions.add(
            combineLatest([
                from(this.platformContext.settings),
                // Start with `undefined` while we don't know if the viewer is authenticated or not.
                authenticatedUser.pipe(startWith(undefined)),
            ]).subscribe(
                ([settingsCascade, authenticatedUser]) => {
                    setExperimentalFeaturesFromSettings(settingsCascade)
                    setQueryStateFromSettings(settingsCascade)
                    this.setState({
                        settingsCascade,
                        authenticatedUser,
                        globbing: globbingEnabledFromSettings(settingsCascade),
                        viewerSubject: viewerSubjectFromSettings(settingsCascade, authenticatedUser),
                    })
                },
                () => this.setState({ authenticatedUser: null })
            )
        )

        /**
         * Listens for uncaught 401 errors when a user when a user was previously authenticated.
         *
         * Don't subscribe to this event when there wasn't an authenticated user,
         * as it could lead to an infinite loop of 401 -> reload -> 401
         */
        this.subscriptions.add(
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
            this.setSelectedSearchContextSpecWithNoChecks(GLOBAL_SEARCH_CONTEXT_SPEC)
        }
        if (!parsedSearchQuery) {
            // If no query is present (e.g. search page, settings page),
            // select the user's default search context.
            this.setSelectedSearchContextSpecToDefault()
        }

        this.setWorkspaceSearchContext(this.state.selectedSearchContextSpec).catch(error => {
            logger.error('Error sending search context to extensions!', error)
        })

        // Update search query state whenever the URL changes
        this.subscriptions.add(
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
                        this.props.searchContextsEnabled
                            ? isSearchContextSpecAvailable({
                                  spec: searchContext,
                                  platformContext: this.platformContext,
                              })
                                  .pipe(first())
                                  .toPromise()
                            : Promise.resolve(false),
                }),
            ]).subscribe(([enableExperimentalSearchInput, parsedSearchURLAndContext]) => {
                if (parsedSearchURLAndContext.query) {
                    // Only override filters and update query from URL if there
                    // is a search query.
                    if (!parsedSearchURLAndContext.searchContextSpec) {
                        // If no search context is present we have to fall back
                        // to the global search context to match the server
                        // behavior.
                        this.setSelectedSearchContextSpec(GLOBAL_SEARCH_CONTEXT_SPEC)
                    } else if (
                        parsedSearchURLAndContext.searchContextSpec.spec !== this.state.selectedSearchContextSpec
                    ) {
                        this.setSelectedSearchContextSpec(parsedSearchURLAndContext.searchContextSpec.spec)
                    }

                    const processedQuery =
                        !enableExperimentalSearchInput &&
                        parsedSearchURLAndContext.searchContextSpec &&
                        this.props.searchContextsEnabled
                            ? omitFilter(
                                  parsedSearchURLAndContext.query,
                                  parsedSearchURLAndContext.searchContextSpec.filter
                              )
                            : parsedSearchURLAndContext.query

                    setQueryStateFromURL(parsedSearchURLAndContext, processedQuery)
                }
            })
        )

        this.userRepositoriesUpdates.next()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
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

        const { authenticatedUser, graphqlClient, temporarySettingsStorage } = this.state

        if (authenticatedUser === undefined || graphqlClient === undefined || temporarySettingsStorage === undefined) {
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
                                    <LegacyLayout
                                        {...this.props}
                                        authenticatedUser={authenticatedUser}
                                        viewerSubject={this.state.viewerSubject}
                                        settingsCascade={this.state.settingsCascade}
                                        batchChangesEnabled={this.props.batchChangesEnabled}
                                        batchChangesExecutionEnabled={isBatchChangesExecutionEnabled(
                                            this.state.settingsCascade
                                        )}
                                        batchChangesWebhookLogsEnabled={window.context.batchChangesWebhookLogsEnabled}
                                        // Search query
                                        fetchHighlightedFileLineRanges={this.fetchHighlightedFileLineRanges}
                                        // Extensions
                                        platformContext={this.platformContext}
                                        extensionsController={this.extensionsController}
                                        telemetryService={eventLogger}
                                        isSourcegraphDotCom={window.context.sourcegraphDotComMode}
                                        searchContextsEnabled={this.props.searchContextsEnabled}
                                        selectedSearchContextSpec={this.getSelectedSearchContextSpec()}
                                        setSelectedSearchContextSpec={this.setSelectedSearchContextSpec}
                                        getUserSearchContextNamespaces={getUserSearchContextNamespaces}
                                        fetchSearchContexts={fetchSearchContexts}
                                        fetchSearchContextBySpec={fetchSearchContextBySpec}
                                        fetchSearchContext={fetchSearchContext}
                                        createSearchContext={createSearchContext}
                                        updateSearchContext={updateSearchContext}
                                        deleteSearchContext={deleteSearchContext}
                                        isSearchContextSpecAvailable={isSearchContextSpecAvailable}
                                        globbing={this.state.globbing}
                                        streamSearch={aggregateStreamingSearch}
                                    />
                                }
                            />
                        </Routes>
                    </CompatRouter>
                </Router>
                {this.extensionsController !== null && window.context.enableLegacyExtensions ? (
                    <Notifications
                        key={2}
                        extensionsController={this.extensionsController}
                        notificationItemStyleProps={notificationStyles}
                    />
                ) : null}
                <UserSessionStores />
            </ComponentsComposer>
        )
    }

    private getSelectedSearchContextSpec = (): string | undefined =>
        getExperimentalFeatures().showSearchContext ? this.state.selectedSearchContextSpec : undefined

    private setSelectedSearchContextSpecWithNoChecks = (spec: string): void => {
        this.setState({ selectedSearchContextSpec: spec })
        this.setWorkspaceSearchContext(spec).catch(error => {
            logger.error('Error sending search context to extensions', error)
        })
    }

    private setSelectedSearchContextSpec = (spec: string): void => {
        if (!this.props.searchContextsEnabled) {
            return
        }

        // The global search context is always available.
        if (spec === GLOBAL_SEARCH_CONTEXT_SPEC) {
            this.setSelectedSearchContextSpecWithNoChecks(spec)
        }

        // Check if the wanted search context is available.
        this.subscriptions.add(
            isSearchContextSpecAvailable({
                spec,
                platformContext: this.platformContext,
            }).subscribe(isAvailable => {
                if (isAvailable) {
                    this.setSelectedSearchContextSpecWithNoChecks(spec)
                } else if (!this.state.selectedSearchContextSpec) {
                    // If the wanted search context is not available and
                    // there is no currently selected search context,
                    // set the current selection to the default search context.
                    // Otherwise, keep the current selection.
                    this.setSelectedSearchContextSpecToDefault()
                }
            })
        )
    }

    private setSelectedSearchContextSpecToDefault = (): void => {
        if (!this.props.searchContextsEnabled) {
            return
        }

        this.subscriptions.add(
            getDefaultSearchContextSpec({ platformContext: this.platformContext }).subscribe(spec => {
                // Fall back to global if no default is returned.
                this.setSelectedSearchContextSpecWithNoChecks(spec || GLOBAL_SEARCH_CONTEXT_SPEC)
            })
        )
    }

    private async setWorkspaceSearchContext(spec: string | undefined): Promise<void> {
        // NOTE(2022-09-08) Inform the inlined code from
        // sourcegraph/code-intel-extensions about the change of search context.
        // The old extension code previously accessed this information from the
        // 'sourcegraph' npm package, and updating the context like this was the
        // simplest solution to mirror the old behavior while deprecating
        // extensions on a tight deadline. It would be nice to properly pass
        // around this via React state in the future.
        setCodeIntelSearchContext(spec)
        if (this.extensionsController === null) {
            return
        }
        const extensionHostAPI = await this.extensionsController.extHostAPI
        await extensionHostAPI.setSearchContext(spec)
    }

    private fetchHighlightedFileLineRanges = (
        parameters: FetchFileParameters,
        force?: boolean | undefined
    ): Observable<string[][]> =>
        fetchHighlightedFileLineRanges({ ...parameters, platformContext: this.platformContext }, force)
}
