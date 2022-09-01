import 'focus-visible'

import * as React from 'react'

import { ApolloProvider } from '@apollo/client'
import { ShortcutProvider } from '@slimsag/react-shortcuts'
import { createBrowserHistory } from 'history'
import ServerIcon from 'mdi-react/ServerIcon'
import { Route, Router } from 'react-router'
import { CompatRouter } from 'react-router-dom-v5-compat'
import { ScrollManager } from 'react-scroll-manager'
import { combineLatest, from, Subscription, fromEvent, of, Subject, Observable } from 'rxjs'
import { distinctUntilChanged, map, startWith, switchMap } from 'rxjs/operators'
import * as uuid from 'uuid'

import { GraphQLClient, HTTPStatusError } from '@sourcegraph/http-client'
import {
    fetchAutoDefinedSearchContexts,
    getUserSearchContextNamespaces,
    SearchContextProps,
    fetchSearchContexts,
    fetchSearchContext,
    fetchSearchContextBySpec,
    createSearchContext,
    updateSearchContext,
    deleteSearchContext,
    isSearchContextSpecAvailable,
    getAvailableSearchContextSpecOrDefault,
    SearchQueryStateStoreProvider,
} from '@sourcegraph/search'
import { FetchFileParameters } from '@sourcegraph/search-ui'
import { getEnabledExtensions } from '@sourcegraph/shared/src/api/client/enabledExtensions'
import { preloadExtensions } from '@sourcegraph/shared/src/api/client/preload'
import { NotificationType } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { fetchHighlightedFileLineRanges } from '@sourcegraph/shared/src/backend/file'
import { Controller as ExtensionsController } from '@sourcegraph/shared/src/extensions/controller'
import { createController as createExtensionsController } from '@sourcegraph/shared/src/extensions/createLazyLoadedController'
import { createNoopController } from '@sourcegraph/shared/src/extensions/createNoopLoadedController.ts'
import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { BrandedNotificationItemStyleProps } from '@sourcegraph/shared/src/notifications/NotificationItem'
import { Notifications } from '@sourcegraph/shared/src/notifications/Notifications'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'
import { aggregateStreamingSearch } from '@sourcegraph/shared/src/search/stream'
import { CoreWorkflowImprovementsEnabledProvider } from '@sourcegraph/shared/src/settings/CoreWorkflowImprovementsEnabledProvider'
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TemporarySettingsProvider } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsProvider'
import { TemporarySettingsStorage } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsStorage'
import { globbingEnabledFromSettings } from '@sourcegraph/shared/src/util/globbing'
import { FeedbackText, setLinkComponent, RouterLink, WildcardThemeContext, WildcardTheme } from '@sourcegraph/wildcard'

import { authenticatedUser, AuthenticatedUser } from './auth'
import { getWebGraphQLClient } from './backend/graphql'
import { BatchChangesProps, isBatchChangesExecutionEnabled } from './batches'
import { CodeIntelligenceProps } from './codeintel'
import { ErrorBoundary } from './components/ErrorBoundary'
import { HeroPage } from './components/HeroPage'
import { ExtensionAreaRoute } from './extensions/extension/ExtensionArea'
import { ExtensionAreaHeaderNavItem } from './extensions/extension/ExtensionAreaHeader'
import { ExtensionsAreaRoute } from './extensions/ExtensionsArea'
import { ExtensionsAreaHeaderActionButton } from './extensions/ExtensionsAreaHeader'
import { FeatureFlagsProvider } from './featureFlags/FeatureFlagsProvider'
import { CodeInsightsProps } from './insights/types'
import { Layout, LayoutProps } from './Layout'
import { BlockInput } from './notebooks'
import { createNotebook } from './notebooks/backend'
import { blockToGQLInput } from './notebooks/serialize'
import { OrgAreaRoute } from './org/area/OrgArea'
import { OrgAreaHeaderNavItem } from './org/area/OrgHeader'
import { createPlatformContext } from './platform/context'
import { RepoContainerRoute } from './repo/RepoContainer'
import { RepoHeaderActionButton } from './repo/RepoHeader'
import { RepoRevisionContainerRoute } from './repo/RepoRevisionContainer'
import { RepoSettingsAreaRoute } from './repo/settings/RepoSettingsArea'
import { RepoSettingsSideBarGroup } from './repo/settings/RepoSettingsSidebar'
import { LayoutRouteProps } from './routes'
import { PageRoutes } from './routes.constants'
import { parseSearchURL } from './search'
import { SearchResultsCacheProvider } from './search/results/SearchResultsCacheProvider'
import { SiteAdminAreaRoute } from './site-admin/SiteAdminArea'
import { SiteAdminSideBarGroups } from './site-admin/SiteAdminSidebar'
import { CodeHostScopeProvider } from './site/CodeHostScopeAlerts/CodeHostScopeProvider'
import {
    setQueryStateFromSettings,
    setQueryStateFromURL,
    setExperimentalFeaturesFromSettings,
    getExperimentalFeatures,
    useNavbarQueryState,
} from './stores'
import { eventLogger } from './tracking/eventLogger'
import { withActivation } from './tracking/withActivation'
import { UserAreaRoute } from './user/area/UserArea'
import { UserAreaHeaderNavItem } from './user/area/UserAreaHeader'
import { UserSettingsAreaRoute } from './user/settings/UserSettingsArea'
import { UserSettingsSidebarItems } from './user/settings/UserSettingsSidebar'
import { UserSessionStores } from './UserSessionStores'
import { observeLocation } from './util/location'
import { siteSubjectNoAdmin, viewerSubjectFromSettings } from './util/settings'

import styles from './SourcegraphWebApp.module.scss'

export interface SourcegraphWebAppProps
    extends CodeIntelligenceProps,
        CodeInsightsProps,
        Pick<BatchChangesProps, 'batchChangesEnabled'>,
        Pick<SearchContextProps, 'searchContextsEnabled'> {
    extensionAreaRoutes: readonly ExtensionAreaRoute[]
    extensionAreaHeaderNavItems: readonly ExtensionAreaHeaderNavItem[]
    extensionsAreaRoutes?: readonly ExtensionsAreaRoute[]
    extensionsAreaHeaderActionButtons?: readonly ExtensionsAreaHeaderActionButton[]
    siteAdminAreaRoutes: readonly SiteAdminAreaRoute[]
    siteAdminSideBarGroups: SiteAdminSideBarGroups
    siteAdminOverviewComponents: readonly React.ComponentType<React.PropsWithChildren<unknown>>[]
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

    viewerSubject: LayoutProps['viewerSubject']

    selectedSearchContextSpec?: string
    defaultSearchContextSpec: string

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

const LAST_SEARCH_CONTEXT_KEY = 'sg-last-search-context'
const WILDCARD_THEME: WildcardTheme = {
    isBranded: true,
}

setLinkComponent(RouterLink)

const LayoutWithActivation = window.context.sourcegraphDotComMode ? Layout : withActivation(Layout)

const history = createBrowserHistory()

/**
 * The root component.
 */
export class SourcegraphWebApp extends React.Component<
    React.PropsWithChildren<SourcegraphWebAppProps>,
    SourcegraphWebAppState
> {
    private readonly subscriptions = new Subscription()
    private readonly userRepositoriesUpdates = new Subject<void>()
    private readonly platformContext: PlatformContext = createPlatformContext()
    private readonly extensionsController: ExtensionsController | null = window.context.enableLegacyExtensions
        ? createExtensionsController(this.platformContext)
        : createNoopController(this.platformContext)

    constructor(props: SourcegraphWebAppProps) {
        super(props)

        if (this.extensionsController !== null) {
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
        }

        setQueryStateFromURL(window.location.search)

        this.state = {
            settingsCascade: EMPTY_SETTINGS_CASCADE,
            viewerSubject: siteSubjectNoAdmin(),
            defaultSearchContextSpec: 'global', // global is default for now, user will be able to change this at some point
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
                console.error('Error initializing GraphQL client', error)
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
            this.setSelectedSearchContextSpec('global')
        }
        if (!parsedSearchQuery) {
            // If no query is present (e.g. search page, settings page), select the last saved
            // search context from localStorage as currently selected search context.
            const lastSelectedSearchContextSpec = localStorage.getItem(LAST_SEARCH_CONTEXT_KEY) || 'global'
            this.setSelectedSearchContextSpec(lastSelectedSearchContextSpec)
        }

        this.setWorkspaceSearchContext(this.state.selectedSearchContextSpec).catch(error => {
            console.error('Error sending search context to extensions!', error)
        })

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

        const { children, ...props } = this.props

        return (
            <ApolloProvider client={graphqlClient}>
                <ErrorBoundary location={null}>
                    <FeatureFlagsProvider>
                        <ShortcutProvider>
                            <WildcardThemeContext.Provider value={WILDCARD_THEME}>
                                <TemporarySettingsProvider temporarySettingsStorage={temporarySettingsStorage}>
                                    <CoreWorkflowImprovementsEnabledProvider>
                                        <SearchResultsCacheProvider>
                                            <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState}>
                                                <ScrollManager history={history}>
                                                    <Router history={history} key={0}>
                                                        <CompatRouter>
                                                            <Route
                                                                path="/"
                                                                render={routeComponentProps => (
                                                                    <CodeHostScopeProvider
                                                                        authenticatedUser={authenticatedUser}
                                                                    >
                                                                        <LayoutWithActivation
                                                                            {...props}
                                                                            {...routeComponentProps}
                                                                            authenticatedUser={authenticatedUser}
                                                                            viewerSubject={this.state.viewerSubject}
                                                                            settingsCascade={this.state.settingsCascade}
                                                                            batchChangesEnabled={
                                                                                this.props.batchChangesEnabled
                                                                            }
                                                                            batchChangesExecutionEnabled={isBatchChangesExecutionEnabled(
                                                                                this.state.settingsCascade
                                                                            )}
                                                                            batchChangesWebhookLogsEnabled={
                                                                                window.context
                                                                                    .batchChangesWebhookLogsEnabled
                                                                            }
                                                                            // Search query
                                                                            fetchHighlightedFileLineRanges={
                                                                                this.fetchHighlightedFileLineRanges
                                                                            }
                                                                            // Extensions
                                                                            platformContext={this.platformContext}
                                                                            extensionsController={
                                                                                this.extensionsController
                                                                            }
                                                                            telemetryService={eventLogger}
                                                                            isSourcegraphDotCom={
                                                                                window.context.sourcegraphDotComMode
                                                                            }
                                                                            searchContextsEnabled={
                                                                                this.props.searchContextsEnabled
                                                                            }
                                                                            selectedSearchContextSpec={this.getSelectedSearchContextSpec()}
                                                                            setSelectedSearchContextSpec={
                                                                                this.setSelectedSearchContextSpec
                                                                            }
                                                                            getUserSearchContextNamespaces={
                                                                                getUserSearchContextNamespaces
                                                                            }
                                                                            fetchAutoDefinedSearchContexts={
                                                                                fetchAutoDefinedSearchContexts
                                                                            }
                                                                            fetchSearchContexts={fetchSearchContexts}
                                                                            fetchSearchContextBySpec={
                                                                                fetchSearchContextBySpec
                                                                            }
                                                                            fetchSearchContext={fetchSearchContext}
                                                                            createSearchContext={createSearchContext}
                                                                            updateSearchContext={updateSearchContext}
                                                                            deleteSearchContext={deleteSearchContext}
                                                                            isSearchContextSpecAvailable={
                                                                                isSearchContextSpecAvailable
                                                                            }
                                                                            defaultSearchContextSpec={
                                                                                this.state.defaultSearchContextSpec
                                                                            }
                                                                            globbing={this.state.globbing}
                                                                            streamSearch={aggregateStreamingSearch}
                                                                            onCreateNotebookFromNotepad={
                                                                                this.onCreateNotebook
                                                                            }
                                                                        />
                                                                    </CodeHostScopeProvider>
                                                                )}
                                                            />
                                                        </CompatRouter>
                                                    </Router>
                                                </ScrollManager>
                                                {this.extensionsController !== null &&
                                                window.context.enableLegacyExtensions ? (
                                                    <Notifications
                                                        key={2}
                                                        extensionsController={this.extensionsController}
                                                        notificationItemStyleProps={notificationStyles}
                                                    />
                                                ) : null}
                                                <UserSessionStores />
                                            </SearchQueryStateStoreProvider>
                                        </SearchResultsCacheProvider>
                                    </CoreWorkflowImprovementsEnabledProvider>
                                </TemporarySettingsProvider>
                            </WildcardThemeContext.Provider>
                        </ShortcutProvider>
                    </FeatureFlagsProvider>
                </ErrorBoundary>
            </ApolloProvider>
        )
    }

    private getSelectedSearchContextSpec = (): string | undefined =>
        getExperimentalFeatures().showSearchContext ? this.state.selectedSearchContextSpec : undefined

    private setSelectedSearchContextSpec = (spec: string): void => {
        if (!this.props.searchContextsEnabled) {
            return
        }

        const { defaultSearchContextSpec } = this.state
        this.subscriptions.add(
            getAvailableSearchContextSpecOrDefault({
                spec,
                defaultSpec: defaultSearchContextSpec,
                platformContext: this.platformContext,
            }).subscribe(availableSearchContextSpecOrDefault => {
                this.setState({ selectedSearchContextSpec: availableSearchContextSpecOrDefault })
                localStorage.setItem(LAST_SEARCH_CONTEXT_KEY, availableSearchContextSpecOrDefault)

                this.setWorkspaceSearchContext(availableSearchContextSpecOrDefault).catch(error => {
                    console.error('Error sending search context to extensions', error)
                })
            })
        )
    }

    private async setWorkspaceSearchContext(spec: string | undefined): Promise<void> {
        if (this.extensionsController === null) {
            return
        }
        const extensionHostAPI = await this.extensionsController.extHostAPI
        await extensionHostAPI.setSearchContext(spec)
    }

    private onCreateNotebook = (blocks: BlockInput[]): void => {
        if (!this.state.authenticatedUser) {
            return
        }

        this.subscriptions.add(
            createNotebook({
                notebook: {
                    title: 'New Notebook',
                    blocks: blocks.map(block => blockToGQLInput({ id: uuid.v4(), ...block })),
                    public: false,
                    namespace: this.state.authenticatedUser.id,
                },
            }).subscribe(createdNotebook => {
                history.push(PageRoutes.Notebook.replace(':id', createdNotebook.id))
            })
        )
    }
    private fetchHighlightedFileLineRanges = (
        parameters: FetchFileParameters,
        force?: boolean | undefined
    ): Observable<string[][]> =>
        fetchHighlightedFileLineRanges({ ...parameters, platformContext: this.platformContext }, force)
}
