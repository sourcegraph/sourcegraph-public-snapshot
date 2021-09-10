import 'focus-visible'

import { ApolloClient, ApolloProvider, NormalizedCacheObject } from '@apollo/client'
import { ShortcutProvider } from '@slimsag/react-shortcuts'
import { createBrowserHistory } from 'history'
import ServerIcon from 'mdi-react/ServerIcon'
import * as React from 'react'
import { Route, Router } from 'react-router'
import { combineLatest, from, Subscription, fromEvent, of, Subject } from 'rxjs'
import { bufferCount, catchError, distinctUntilChanged, map, startWith, switchMap } from 'rxjs/operators'

import { Tooltip } from '@sourcegraph/branded/src/components/tooltip/Tooltip'
import { getEnabledExtensions } from '@sourcegraph/shared/src/api/client/enabledExtensions'
import { preloadExtensions } from '@sourcegraph/shared/src/api/client/preload'
import { NotificationType } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { HTTPStatusError } from '@sourcegraph/shared/src/backend/fetch'
import { setLinkComponent } from '@sourcegraph/shared/src/components/Link'
import {
    Controller as ExtensionsController,
    createController as createExtensionsController,
} from '@sourcegraph/shared/src/extensions/controller'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { Notifications } from '@sourcegraph/shared/src/notifications/Notifications'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'
import { aggregateStreamingSearch } from '@sourcegraph/shared/src/search/stream'
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { authenticatedUser, AuthenticatedUser } from './auth'
import { getWebGraphQLClient } from './backend/graphql'
import { BatchChangesProps } from './batches'
import { CodeIntelligenceProps } from './codeintel'
import { ErrorBoundary } from './components/ErrorBoundary'
import { queryExternalServices } from './components/externalServices/backend'
import { FeedbackText } from './components/FeedbackText'
import { HeroPage } from './components/HeroPage'
import { RouterLinkOrAnchor } from './components/RouterLinkOrAnchor'
import { ExtensionAreaRoute } from './extensions/extension/ExtensionArea'
import { ExtensionAreaHeaderNavItem } from './extensions/extension/ExtensionAreaHeader'
import { ExtensionsAreaRoute } from './extensions/ExtensionsArea'
import { ExtensionsAreaHeaderActionButton } from './extensions/ExtensionsAreaHeader'
import { FeatureFlagName, fetchFeatureFlags, FlagSet } from './featureFlags/featureFlags'
import { logInsightMetrics } from './insights/analytics'
import { CodeInsightsProps } from './insights/types'
import { KeyboardShortcutsProps } from './keyboardShortcuts/keyboardShortcuts'
import { Layout, LayoutProps } from './Layout'
import { updateUserSessionStores } from './marketing/util'
import { OrgAreaRoute } from './org/area/OrgArea'
import { OrgAreaHeaderNavItem } from './org/area/OrgHeader'
import { createPlatformContext } from './platform/context'
import { fetchHighlightedFileLineRanges } from './repo/backend'
import { RepoContainerRoute } from './repo/RepoContainer'
import { RepoHeaderActionButton } from './repo/RepoHeader'
import { RepoRevisionContainerRoute } from './repo/RepoRevisionContainer'
import { RepoSettingsAreaRoute } from './repo/settings/RepoSettingsArea'
import { RepoSettingsSideBarGroup } from './repo/settings/RepoSettingsSidebar'
import { LayoutRouteProps } from './routes'
import { VersionContext } from './schema/site.schema'
import {
    resolveVersionContext,
    parseSearchURL,
    getAvailableSearchContextSpecOrDefault,
    isSearchContextSpecAvailable,
} from './search'
import {
    fetchSavedSearches,
    fetchRecentSearches,
    fetchRecentFileViews,
    fetchAutoDefinedSearchContexts,
    fetchSearchContexts,
    convertVersionContextToSearchContext,
    fetchSearchContext,
    createSearchContext,
    updateSearchContext,
    deleteSearchContext,
    getUserSearchContextNamespaces,
    fetchSearchContextBySpec,
} from './search/backend'
import { QueryState } from './search/helpers'
import { SearchResultsCacheProvider } from './search/results/SearchResultsCacheProvider'
import { TemporarySettingsProvider } from './settings/temporary/TemporarySettingsProvider'
import { listUserRepositories } from './site-admin/backend'
import { SiteAdminAreaRoute } from './site-admin/SiteAdminArea'
import { SiteAdminSideBarGroups } from './site-admin/SiteAdminSidebar'
import { CodeHostScopeProvider } from './site/CodeHostScopeAlerts/CodeHostScopeProvider'
import { eventLogger } from './tracking/eventLogger'
import { withActivation } from './tracking/withActivation'
import { UserAreaRoute } from './user/area/UserArea'
import { UserAreaHeaderNavItem } from './user/area/UserAreaHeader'
import { UserSettingsAreaRoute } from './user/settings/UserSettingsArea'
import { UserSettingsSidebarItems } from './user/settings/UserSettingsSidebar'
import { globbingEnabledFromSettings } from './util/globbing'
import { observeLocation } from './util/location'
import {
    SITE_SUBJECT_NO_ADMIN,
    viewerSubjectFromSettings,
    defaultCaseSensitiveFromSettings,
    defaultPatternTypeFromSettings,
    experimentalFeaturesFromSettings,
} from './util/settings'

export interface SourcegraphWebAppProps
    extends CodeIntelligenceProps,
        BatchChangesProps,
        CodeInsightsProps,
        KeyboardShortcutsProps {
    extensionAreaRoutes: readonly ExtensionAreaRoute[]
    extensionAreaHeaderNavItems: readonly ExtensionAreaHeaderNavItem[]
    extensionsAreaRoutes: readonly ExtensionsAreaRoute[]
    extensionsAreaHeaderActionButtons: readonly ExtensionsAreaHeaderActionButton[]
    siteAdminAreaRoutes: readonly SiteAdminAreaRoute[]
    siteAdminSideBarGroups: SiteAdminSideBarGroups
    siteAdminOverviewComponents: readonly React.ComponentType[]
    userAreaHeaderNavItems: readonly UserAreaHeaderNavItem[]
    userAreaRoutes: readonly UserAreaRoute[]
    userSettingsSideBarItems: UserSettingsSidebarItems
    userSettingsAreaRoutes: readonly UserSettingsAreaRoute[]
    orgAreaHeaderNavItems: readonly OrgAreaHeaderNavItem[]
    orgAreaRoutes: readonly OrgAreaRoute[]
    repoContainerRoutes: readonly RepoContainerRoute[]
    repoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[]
    repoHeaderActionButtons: readonly RepoHeaderActionButton[]
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarGroups: readonly RepoSettingsSideBarGroup[]
    routes: readonly LayoutRouteProps<any>[]
}

interface SourcegraphWebAppState extends SettingsCascadeProps {
    error?: Error

    /** The currently authenticated user (or null if the viewer is anonymous). */
    authenticatedUser?: AuthenticatedUser | null

    /** GraphQL client initialized asynchronously to restore persisted cache. */
    graphqlClient?: ApolloClient<NormalizedCacheObject>

    viewerSubject: LayoutProps['viewerSubject']

    /**
     * The current search query in the navbar.
     */
    navbarSearchQueryState: QueryState

    /**
     * The current parsed search query, with all UI-configurable parameters
     * (eg. pattern type, case sensitivity, version context) removed
     */
    parsedSearchQuery: string

    /**
     * The current search pattern type.
     */
    searchPatternType: SearchPatternType

    /**
     * Whether the current search is case sensitive.
     */
    searchCaseSensitivity: boolean

    /*
     * The version context the instance is in. If undefined, it means no version context is selected.
     */
    versionContext?: string

    /**
     * Available version contexts defined in the site configuration.
     */
    availableVersionContexts?: VersionContext[]

    /**
     * The previously used version context, as specified in localStorage.
     */
    previousVersionContext: string | null

    showRepogroupHomepage: boolean

    showOnboardingTour: boolean

    showEnterpriseHomePanels: boolean

    showSearchContext: boolean
    showSearchContextManagement: boolean
    selectedSearchContextSpec?: string
    defaultSearchContextSpec: string
    hasUserAddedRepositories: boolean
    hasUserSyncedPublicRepositories: boolean
    hasUserAddedExternalServices: boolean

    /**
     * Whether globbing is enabled for filters.
     */
    globbing: boolean

    /**
     * Whether we show the mulitiline editor at /search/console
     */
    showMultilineSearchConsole: boolean

    /**
     * Whether we show the search notebook.
     */
    showSearchNotebook: boolean

    /**
     * Whether we show the multiline editor at /search/query-builder
     */
    showQueryBuilder: boolean

    /**
     * Whether the code monitoring feature flag is enabled.
     */
    enableCodeMonitoring: boolean

    /**
     * Whether the API docs feature flag is enabled.
     */
    enableAPIDocs: boolean

    /**
     * Evaluated feature flags for the current viewer
     */
    featureFlags: FlagSet
}

const notificationClassNames = {
    [NotificationType.Log]: 'alert alert-secondary',
    [NotificationType.Success]: 'alert alert-success',
    [NotificationType.Info]: 'alert alert-info',
    [NotificationType.Warning]: 'alert alert-warning',
    [NotificationType.Error]: 'alert alert-danger',
}

const LAST_VERSION_CONTEXT_KEY = 'sg-last-version-context'
const LAST_SEARCH_CONTEXT_KEY = 'sg-last-search-context'

setLinkComponent(RouterLinkOrAnchor)

const LayoutWithActivation = window.context.sourcegraphDotComMode ? Layout : withActivation(Layout)

const history = createBrowserHistory()

/**
 * The root component.
 */
export class SourcegraphWebApp extends React.Component<SourcegraphWebAppProps, SourcegraphWebAppState> {
    private readonly subscriptions = new Subscription()
    private readonly userRepositoriesUpdates = new Subject<void>()
    private readonly platformContext: PlatformContext = createPlatformContext()
    private readonly extensionsController: ExtensionsController = createExtensionsController(this.platformContext)

    constructor(props: SourcegraphWebAppProps) {
        super(props)
        this.subscriptions.add(this.extensionsController)

        // Preload extensions whenever user enabled extensions or the viewed language changes.
        this.subscriptions.add(
            combineLatest([
                getEnabledExtensions(this.platformContext),
                observeLocation(history).pipe(
                    startWith(location),
                    map(location => getModeFromPath(location.pathname)),
                    distinctUntilChanged()
                ),
            ]).subscribe(([extensions, languageID]) => {
                preloadExtensions({
                    extensions,
                    languages: new Set([languageID]),
                })
            })
        )

        const parsedSearchURL = parseSearchURL(window.location.search)
        // The patternType in the URL query parameter. If none is provided, default to literal.
        // This will be updated with the default in settings when the web app mounts.
        const urlPatternType = parsedSearchURL.patternType || SearchPatternType.literal
        const urlCase = parsedSearchURL.caseSensitive
        const availableVersionContexts = window.context.experimentalFeatures.versionContexts
        const previousVersionContext = localStorage.getItem(LAST_VERSION_CONTEXT_KEY)
        const resolvedVersionContext = availableVersionContexts
            ? resolveVersionContext(parsedSearchURL.versionContext || undefined, availableVersionContexts) ||
              resolveVersionContext(previousVersionContext || undefined, availableVersionContexts) ||
              undefined
            : undefined

        this.state = {
            navbarSearchQueryState: { query: '' },
            settingsCascade: EMPTY_SETTINGS_CASCADE,
            viewerSubject: SITE_SUBJECT_NO_ADMIN,
            parsedSearchQuery: parsedSearchURL.query || '',
            searchPatternType: urlPatternType,
            searchCaseSensitivity: urlCase,
            versionContext: resolvedVersionContext,
            availableVersionContexts,
            previousVersionContext,
            showRepogroupHomepage: false,
            showOnboardingTour: false,
            showSearchContext: false,
            showSearchContextManagement: false,
            defaultSearchContextSpec: 'global', // global is default for now, user will be able to change this at some point
            hasUserAddedRepositories: false,
            hasUserSyncedPublicRepositories: false,
            hasUserAddedExternalServices: false,
            showEnterpriseHomePanels: false,
            globbing: false,
            showMultilineSearchConsole: false,
            showSearchNotebook: false,
            showQueryBuilder: false,
            enableCodeMonitoring: false,
            // Disabling linter here as otherwise the application fails to compile. Bad lint?
            // See 7a137b201330eb2118c746f8cc5acddf63c1f039
            // eslint-disable-next-line react/no-unused-state
            enableAPIDocs: false,
            featureFlags: new Map<FeatureFlagName, boolean>(),
        }
    }

    public componentDidMount(): void {
        updateUserSessionStores()

        document.documentElement.classList.add('theme')

        getWebGraphQLClient()
            .then(graphqlClient => this.setState({ graphqlClient }))
            .catch(error => {
                console.error('Error initalizing GraphQL client', error)
            })

        this.subscriptions.add(
            combineLatest([from(this.platformContext.settings), authenticatedUser.pipe(startWith(null))]).subscribe(
                ([settingsCascade, authenticatedUser]) => {
                    this.setState(state => ({
                        settingsCascade,
                        authenticatedUser,
                        ...experimentalFeaturesFromSettings(settingsCascade),
                        globbing: globbingEnabledFromSettings(settingsCascade),
                        searchCaseSensitivity:
                            defaultCaseSensitiveFromSettings(settingsCascade) || state.searchCaseSensitivity,
                        searchPatternType: defaultPatternTypeFromSettings(settingsCascade) || state.searchPatternType,
                        viewerSubject: viewerSubjectFromSettings(settingsCascade, authenticatedUser),
                    }))
                },
                () => this.setState({ authenticatedUser: null })
            )
        )

        // Track static metrics fo code insights.
        // Insight count, insights settings, observe settings mutations for analytics
        // Track add delete and update events of code insights via
        this.subscriptions.add(
            combineLatest([from(this.platformContext.settings), authenticatedUser])
                .pipe(bufferCount(2, 1))
                .subscribe(([[oldSettings], [newSettings, authUser]]) => {
                    if (authUser) {
                        logInsightMetrics(oldSettings, newSettings, eventLogger)
                    }
                })
        )

        this.subscriptions.add(
            combineLatest([this.userRepositoriesUpdates, authenticatedUser])
                .pipe(
                    switchMap(([, authenticatedUser]) =>
                        authenticatedUser
                            ? combineLatest([
                                  listUserRepositories({ id: authenticatedUser.id, first: 1 }),
                                  queryExternalServices({ namespace: authenticatedUser.id, first: 1, after: null }),
                              ])
                            : of(null)
                    ),
                    catchError(error => [asError(error)])
                )
                .subscribe(result => {
                    if (!isErrorLike(result) && result !== null) {
                        const [userRepositoriesResult, externalServicesResult] = result
                        this.setState({
                            hasUserAddedRepositories: userRepositoriesResult.nodes.length > 0,
                            hasUserAddedExternalServices: externalServicesResult.nodes.length > 0,
                        })
                    }
                })
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

        this.subscriptions.add(
            fetchFeatureFlags().subscribe(event => {
                // Disabling linter here because this is not yet used anywhere.
                // This can be re-enabled as soon as feature flags are leveraged.
                // eslint-disable-next-line react/no-unused-state
                this.setState({ featureFlags: event })
            })
        )

        if (this.state.parsedSearchQuery && !filterExists(this.state.parsedSearchQuery, FilterType.context)) {
            // If a context filter does not exist in the query, we have to switch the selected context
            // to global to match the UI with the backend semantics (if no context is specified in the query,
            // the query is run in global context).
            this.setSelectedSearchContextSpec('global')
        }
        if (!this.state.parsedSearchQuery) {
            // If no query is present (e.g. search page, settings page), select the last saved
            // search context from localStorage as currently selected search context.
            const lastSelectedSearchContextSpec = localStorage.getItem(LAST_SEARCH_CONTEXT_KEY) || 'global'
            this.setSelectedSearchContextSpec(lastSelectedSearchContextSpec)
        }

        // Send initial versionContext to extensions
        this.setVersionContext(this.state.versionContext).catch(error => {
            console.error('Error sending initial version context to extensions', error)
        })

        this.setWorkspaceSearchContext(this.state.selectedSearchContextSpec).catch(error => {
            console.error('Error sending search context to extensions', error)
        })

        this.userRepositoriesUpdates.next()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactFragment | null {
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
                    <div className="app__error">
                        {subtitle}
                        {subtitle && <hr className="my-3" />}
                        <pre>{errorMessage}</pre>
                    </div>
                )
            } else {
                subtitle = <div className="app__error">{subtitle}</div>
            }
            return <HeroPage icon={ServerIcon} title={`${statusCode}: ${statusText}`} subtitle={subtitle} />
        }

        const { authenticatedUser, graphqlClient } = this.state
        if (authenticatedUser === undefined || graphqlClient === undefined) {
            return null
        }

        const { children, ...props } = this.props

        return (
            <ApolloProvider client={graphqlClient}>
                <ErrorBoundary location={null}>
                    <ShortcutProvider>
                        <TemporarySettingsProvider isAuthenticatedUser={window.context?.isAuthenticatedUser}>
                            <SearchResultsCacheProvider>
                                <Router history={history} key={0}>
                                    <Route
                                        path="/"
                                        render={routeComponentProps => (
                                            <CodeHostScopeProvider authenticatedUser={authenticatedUser}>
                                                <LayoutWithActivation
                                                    {...props}
                                                    {...routeComponentProps}
                                                    authenticatedUser={authenticatedUser}
                                                    viewerSubject={this.state.viewerSubject}
                                                    settingsCascade={this.state.settingsCascade}
                                                    batchChangesEnabled={this.props.batchChangesEnabled}
                                                    // Search query
                                                    navbarSearchQueryState={this.state.navbarSearchQueryState}
                                                    onNavbarQueryChange={this.onNavbarQueryChange}
                                                    fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                                                    parsedSearchQuery={this.state.parsedSearchQuery}
                                                    setParsedSearchQuery={this.setParsedSearchQuery}
                                                    patternType={this.state.searchPatternType}
                                                    setPatternType={this.setPatternType}
                                                    caseSensitive={this.state.searchCaseSensitivity}
                                                    setCaseSensitivity={this.setCaseSensitivity}
                                                    versionContext={this.state.versionContext}
                                                    setVersionContext={this.setVersionContext}
                                                    availableVersionContexts={this.state.availableVersionContexts}
                                                    previousVersionContext={this.state.previousVersionContext}
                                                    // Extensions
                                                    platformContext={this.platformContext}
                                                    extensionsController={this.extensionsController}
                                                    telemetryService={eventLogger}
                                                    isSourcegraphDotCom={window.context.sourcegraphDotComMode}
                                                    showRepogroupHomepage={this.state.showRepogroupHomepage}
                                                    showOnboardingTour={this.state.showOnboardingTour}
                                                    showSearchContext={this.state.showSearchContext}
                                                    hasUserAddedRepositories={this.hasUserAddedRepositories()}
                                                    hasUserAddedExternalServices={
                                                        this.state.hasUserAddedExternalServices
                                                    }
                                                    showSearchContextManagement={this.state.showSearchContextManagement}
                                                    selectedSearchContextSpec={this.getSelectedSearchContextSpec()}
                                                    setSelectedSearchContextSpec={this.setSelectedSearchContextSpec}
                                                    getUserSearchContextNamespaces={getUserSearchContextNamespaces}
                                                    fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                                                    fetchSearchContexts={fetchSearchContexts}
                                                    fetchSearchContextBySpec={fetchSearchContextBySpec}
                                                    fetchSearchContext={fetchSearchContext}
                                                    createSearchContext={createSearchContext}
                                                    updateSearchContext={updateSearchContext}
                                                    deleteSearchContext={deleteSearchContext}
                                                    convertVersionContextToSearchContext={
                                                        convertVersionContextToSearchContext
                                                    }
                                                    isSearchContextSpecAvailable={isSearchContextSpecAvailable}
                                                    defaultSearchContextSpec={this.state.defaultSearchContextSpec}
                                                    showEnterpriseHomePanels={this.state.showEnterpriseHomePanels}
                                                    globbing={this.state.globbing}
                                                    showMultilineSearchConsole={this.state.showMultilineSearchConsole}
                                                    showSearchNotebook={this.state.showSearchNotebook}
                                                    showQueryBuilder={this.state.showQueryBuilder}
                                                    enableCodeMonitoring={this.state.enableCodeMonitoring}
                                                    fetchSavedSearches={fetchSavedSearches}
                                                    fetchRecentSearches={fetchRecentSearches}
                                                    fetchRecentFileViews={fetchRecentFileViews}
                                                    streamSearch={aggregateStreamingSearch}
                                                    onUserExternalServicesOrRepositoriesUpdate={
                                                        this.onUserExternalServicesOrRepositoriesUpdate
                                                    }
                                                    onSyncedPublicRepositoriesUpdate={
                                                        this.onSyncedPublicRepositoriesUpdate
                                                    }
                                                    featureFlags={this.state.featureFlags}
                                                />
                                            </CodeHostScopeProvider>
                                        )}
                                    />
                                </Router>
                                <Tooltip key={1} />
                                <Notifications
                                    key={2}
                                    extensionsController={this.extensionsController}
                                    notificationClassNames={notificationClassNames}
                                />
                            </SearchResultsCacheProvider>
                        </TemporarySettingsProvider>
                    </ShortcutProvider>
                </ErrorBoundary>
            </ApolloProvider>
        )
    }

    private onNavbarQueryChange = (navbarSearchQueryState: QueryState): void => {
        this.setState({ navbarSearchQueryState })
    }

    private setParsedSearchQuery = (query: string): void => {
        this.setState({ parsedSearchQuery: query })
    }

    private setPatternType = (patternType: SearchPatternType): void => {
        this.setState({
            searchPatternType: patternType,
        })
    }

    private setCaseSensitivity = (caseSensitive: boolean): void => {
        this.setState({
            searchCaseSensitivity: caseSensitive,
        })
    }

    private setVersionContext = async (versionContext: string | undefined): Promise<void> => {
        const resolvedVersionContext = resolveVersionContext(versionContext, this.state.availableVersionContexts)
        if (!resolvedVersionContext) {
            localStorage.removeItem(LAST_VERSION_CONTEXT_KEY)
            this.setState({ versionContext: undefined, previousVersionContext: null })
        } else {
            localStorage.setItem(LAST_VERSION_CONTEXT_KEY, resolvedVersionContext)
            this.setState({ versionContext: resolvedVersionContext, previousVersionContext: resolvedVersionContext })
        }

        const extensionHostAPI = await this.extensionsController.extHostAPI
        // Note: `setVersionContext` is now asynchronous since the version context
        // is sent directly to extensions in the worker thread. This means that when the Promise
        // is in a fulfilled state, we know that extensions have received the latest version context
        await extensionHostAPI.setVersionContext(resolvedVersionContext)
    }

    private onUserExternalServicesOrRepositoriesUpdate = (
        externalServicesCount: number,
        userRepoCount: number
    ): void => {
        this.setState({
            hasUserAddedExternalServices: externalServicesCount > 0,
            hasUserAddedRepositories: userRepoCount > 0,
        })
    }

    private onSyncedPublicRepositoriesUpdate = (publicReposCount: number): void => {
        this.setState({
            hasUserSyncedPublicRepositories: publicReposCount > 0,
        })
    }

    private hasUserAddedRepositories = (): boolean =>
        this.state.hasUserAddedRepositories || this.state.hasUserSyncedPublicRepositories

    private getSelectedSearchContextSpec = (): string | undefined =>
        this.state.showSearchContext ? this.state.selectedSearchContextSpec : undefined

    private setSelectedSearchContextSpec = (spec: string): void => {
        const { defaultSearchContextSpec } = this.state
        this.subscriptions.add(
            getAvailableSearchContextSpecOrDefault({ spec, defaultSpec: defaultSearchContextSpec }).subscribe(
                availableSearchContextSpecOrDefault => {
                    this.setState({ selectedSearchContextSpec: availableSearchContextSpecOrDefault })
                    localStorage.setItem(LAST_SEARCH_CONTEXT_KEY, availableSearchContextSpecOrDefault)

                    this.setWorkspaceSearchContext(availableSearchContextSpecOrDefault).catch(error => {
                        console.error('Error sending search context to extensions', error)
                    })
                }
            )
        )
    }

    private async setWorkspaceSearchContext(spec: string | undefined): Promise<void> {
        const extensionHostAPI = await this.extensionsController.extHostAPI
        await extensionHostAPI.setSearchContext(spec)
    }
}
