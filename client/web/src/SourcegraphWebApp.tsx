import 'focus-visible'

import { ApolloProvider } from '@apollo/client'
import { ShortcutProvider } from '@slimsag/react-shortcuts'
import { createBrowserHistory } from 'history'
import ServerIcon from 'mdi-react/ServerIcon'
import React, { useMemo, useEffect, useState, useCallback } from 'react'
import { Route, Router } from 'react-router'
import { combineLatest, from, fromEvent, of, Subject } from 'rxjs'
import { bufferCount, catchError, distinctUntilChanged, map, startWith, switchMap, tap } from 'rxjs/operators'

import { Tooltip } from '@sourcegraph/branded/src/components/tooltip/Tooltip'
import { getEnabledExtensions } from '@sourcegraph/shared/src/api/client/enabledExtensions'
import { preloadExtensions } from '@sourcegraph/shared/src/api/client/preload'
import { NotificationType } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { HTTPStatusError } from '@sourcegraph/shared/src/backend/fetch'
import { setLinkComponent } from '@sourcegraph/shared/src/components/Link'
import { createController as createExtensionsController } from '@sourcegraph/shared/src/extensions/controller'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { Notifications } from '@sourcegraph/shared/src/notifications/Notifications'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'
import { aggregateStreamingSearch } from '@sourcegraph/shared/src/search/stream'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { authenticatedUser } from './auth'
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
import { FeatureFlagName, fetchFeatureFlags } from './featureFlags/featureFlags'
import { logInsightMetrics } from './insights/analytics'
import { CodeInsightsProps } from './insights/types'
import { KeyboardShortcutsProps } from './keyboardShortcuts/keyboardShortcuts'
import { Layout } from './Layout'
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
export const SourcegraphWebApp: React.FunctionComponent<SourcegraphWebAppProps> = ({ children, ...props }) => {
    const location = history.location

    const platformContext = useMemo(() => createPlatformContext(), [])
    const extensionsController = useMemo(() => createExtensionsController(platformContext), [platformContext])
    useEffect(() => () => extensionsController.unsubscribe(), [extensionsController])

    // Preload extensions whenever user enabled extensions or the viewed language changes.
    useObservable(
        useMemo(
            () =>
                combineLatest([
                    getEnabledExtensions(platformContext),
                    observeLocation(history).pipe(
                        startWith(location),
                        map(location => getModeFromPath(location.pathname)),
                        distinctUntilChanged()
                    ),
                ]).pipe(
                    tap(([extensions, languageID]) => {
                        preloadExtensions({
                            extensions,
                            languages: new Set([languageID]),
                        })
                    })
                ),
            [location, platformContext]
        )
    )

    const graphqlClient = useObservable(
        useMemo(
            () =>
                from(getWebGraphQLClient()).pipe(
                    catchError(error => {
                        console.error('Error initalizing GraphQL client', error)
                        return of(undefined)
                    })
                ),
            []
        )
    )

    useEffect(() => {
        updateUserSessionStores()
    }, [])

    const parsedSearchURL = useMemo(() => parseSearchURL(location.search), [location.search])
    const [parsedSearchQuery, setParsedSearchQuery] = useState(parsedSearchURL.query || '')
    // The search patternType, and case in the URL query parameter. If none is provided, default to
    // literal, and these will be updated with the defaults in settings when the web app mounts.
    const [searchPatternType, setPatternType] = useState(parsedSearchURL.patternType || SearchPatternType.literal)
    const [searchCaseSensitivity, setCaseSensitivity] = useState(parsedSearchURL.caseSensitive)

    const userAndSettingsProps = useObservable(
        useMemo(
            () =>
                combineLatest([from(platformContext.settings), authenticatedUser.pipe(startWith(null))]).pipe(
                    map(([settingsCascade, authenticatedUser]) => ({
                        settingsCascade,
                        authenticatedUser,
                        ...experimentalFeaturesFromSettings(settingsCascade),
                        globbing: globbingEnabledFromSettings(settingsCascade),
                        caseSensitive: defaultCaseSensitiveFromSettings(settingsCascade) ?? searchCaseSensitivity,
                        patternType: defaultPatternTypeFromSettings(settingsCascade) ?? searchPatternType,
                        viewerSubject: viewerSubjectFromSettings(settingsCascade, authenticatedUser),
                    }))
                ),
            [platformContext.settings, searchCaseSensitivity, searchPatternType]
        )
    )

    // Track static metrics for code insights.
    // Insight count, insights settings, observe settings mutations for analytics
    // Track add delete and update events of code insights via
    useObservable(
        useMemo(
            () =>
                combineLatest([from(platformContext.settings), authenticatedUser]).pipe(
                    bufferCount(2, 1),
                    tap(([[oldSettings], [newSettings, authUser]]) => {
                        if (authUser) {
                            logInsightMetrics(oldSettings, newSettings, eventLogger)
                        }
                    })
                ),
            [platformContext.settings]
        )
    )

    const userRepositoriesUpdates = useMemo(() => new Subject<void>(), [])
    interface HasUserAddedProps {
        hasUserAddedRepositories: boolean
        hasUserAddedExternalServices: boolean
    }
    const hasUserAddedProps: HasUserAddedProps | undefined = useObservable(
        useMemo(
            () =>
                combineLatest([userRepositoriesUpdates, authenticatedUser]).pipe(
                    switchMap(([, authenticatedUser]) =>
                        authenticatedUser
                            ? combineLatest([
                                  listUserRepositories({ id: authenticatedUser.id, first: 1 }),
                                  queryExternalServices({ namespace: authenticatedUser.id, first: 1, after: null }),
                              ])
                            : of(null)
                    ),
                    catchError(error => [asError(error)]),
                    map(result => {
                        if (!isErrorLike(result) && result !== null) {
                            const [userRepositoriesResult, externalServicesResult] = result
                            return {
                                hasUserAddedRepositories: userRepositoriesResult.nodes.length > 0,
                                hasUserAddedExternalServices: externalServicesResult.nodes.length > 0,
                            }
                        }
                        return undefined
                    })
                ),
            [userRepositoriesUpdates]
        )
    )
    useEffect(() => userRepositoriesUpdates.next(), [userRepositoriesUpdates])
    const [hasUserAddedProps2, setHasUserAddedProps2] = useState<HasUserAddedProps>()
    const onUserExternalServicesOrRepositoriesUpdate = useCallback(
        (externalServicesCount: number, userRepoCount: number): void => {
            setHasUserAddedProps2({
                hasUserAddedExternalServices: externalServicesCount > 0,
                hasUserAddedRepositories: userRepoCount > 0,
            })
        },
        []
    )
    const [hasUserSyncedPublicRepositories, setHasUserSyncedPublicRepositories] = useState(false)
    const onSyncedPublicRepositoriesUpdate = useCallback(
        (publicReposCount: number): void => setHasUserSyncedPublicRepositories(publicReposCount > 0),
        []
    )
    const hasUserAddedRepositories =
        hasUserAddedProps2?.hasUserAddedRepositories ||
        hasUserAddedProps?.hasUserAddedRepositories ||
        hasUserSyncedPublicRepositories ||
        false
    const hasUserAddedExternalServices =
        hasUserAddedProps2?.hasUserAddedExternalServices || hasUserAddedProps?.hasUserAddedExternalServices || false

    useEffect(() => document.documentElement.classList.add('theme'), [])

    const featureFlags = useObservable(useMemo(() => fetchFeatureFlags(), [])) ?? new Map<FeatureFlagName, boolean>()

    const [navbarSearchQueryState, onNavbarQueryChange] = useState<QueryState>({ query: '' })

    /**
     * Listens for uncaught 401 errors when a user when a user was previously authenticated.
     *
     * Don't subscribe to this event when there wasn't an authenticated user,
     * as it could lead to an infinite loop of 401 -> reload -> 401
     */
    useObservable(
        useMemo(
            () =>
                authenticatedUser.pipe(
                    switchMap(authenticatedUser =>
                        authenticatedUser ? fromEvent<ErrorEvent>(window, 'error') : of(null)
                    ),
                    tap(event => {
                        if (event?.error instanceof HTTPStatusError && event.error.status === 401) {
                            window.location.reload()
                        }
                    })
                ),
            []
        )
    )

    const setWorkspaceSearchContext = useCallback(
        async (spec: string | undefined): Promise<void> => {
            const extensionHostAPI = await extensionsController.extHostAPI
            await extensionHostAPI.setSearchContext(spec)
        },
        [extensionsController.extHostAPI]
    )
    const availableVersionContexts = window.context.experimentalFeatures.versionContexts
    const [previousVersionContext, setPreviousVersionContext] = useState(localStorage.getItem(LAST_VERSION_CONTEXT_KEY))
    const resolvedVersionContext = availableVersionContexts
        ? resolveVersionContext(parsedSearchURL.versionContext || undefined, availableVersionContexts) ||
          resolveVersionContext(previousVersionContext || undefined, availableVersionContexts) ||
          undefined
        : undefined
    const [versionContext, setVersionContext] = useState(resolvedVersionContext)
    const [selectedSearchContextSpec, setSelectedSearchContextSpec] = useState<string>()
    const defaultSearchContextSpec = 'global' // global is default for now, user will be able to change this at some point
    const getSelectedSearchContextSpec = useCallback(
        (): string | undefined => (userAndSettingsProps?.showSearchContext ? selectedSearchContextSpec : undefined),
        [selectedSearchContextSpec, userAndSettingsProps?.showSearchContext]
    )
    const setSelectedSearchContextSpec2 = useCallback(
        async (spec: string): Promise<void> => {
            const availableSearchContextSpecOrDefault = await getAvailableSearchContextSpecOrDefault({
                spec,
                defaultSpec: defaultSearchContextSpec,
            }).toPromise()
            setSelectedSearchContextSpec(availableSearchContextSpecOrDefault)
            localStorage.setItem(LAST_SEARCH_CONTEXT_KEY, availableSearchContextSpecOrDefault)

            setWorkspaceSearchContext(availableSearchContextSpecOrDefault).catch(error => {
                console.error('Error sending search context to extensions', error)
            })
        },
        [setWorkspaceSearchContext]
    )
    const setVersionContext2 = useCallback(
        async (versionContext: string | undefined): Promise<void> => {
            const resolvedVersionContext = resolveVersionContext(versionContext, availableVersionContexts)
            if (!resolvedVersionContext) {
                localStorage.removeItem(LAST_VERSION_CONTEXT_KEY)
                setVersionContext(undefined)
                setPreviousVersionContext(null)
            } else {
                localStorage.setItem(LAST_VERSION_CONTEXT_KEY, resolvedVersionContext)
                setVersionContext(resolvedVersionContext)
                setPreviousVersionContext(resolvedVersionContext)
            }

            const extensionHostAPI = await extensionsController.extHostAPI
            // Note: `setVersionContext` is now asynchronous since the version context
            // is sent directly to extensions in the worker thread. This means that when the Promise
            // is in a fulfilled state, we know that extensions have received the latest version context
            await extensionHostAPI.setVersionContext(resolvedVersionContext)
        },
        [availableVersionContexts, extensionsController.extHostAPI]
    )
    useEffect(() => {
        if (parsedSearchQuery && !filterExists(parsedSearchQuery, FilterType.context)) {
            // If a context filter does not exist in the query, we have to switch the selected context
            // to global to match the UI with the backend semantics (if no context is specified in the query,
            // the query is run in global context).
            setSelectedSearchContextSpec2('global').catch(error => {
                console.error('Error setting selected search context', error)
            })
        }
        if (!parsedSearchQuery) {
            // If no query is present (e.g. search page, settings page), select the last saved
            // search context from localStorage as currently selected search context.
            const lastSelectedSearchContextSpec = localStorage.getItem(LAST_SEARCH_CONTEXT_KEY) || 'global'
            setSelectedSearchContextSpec2(lastSelectedSearchContextSpec).catch(error => {
                console.error('Error setting selected search context', error)
            })
        }

        // Send initial versionContext to extensions
        setVersionContext2(versionContext).catch(error => {
            console.error('Error sending initial version context to extensions', error)
        })

        setWorkspaceSearchContext(selectedSearchContextSpec).catch(error => {
            console.error('Error sending search context to extensions', error)
        })

        // Only run on the initial render.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    if (!userAndSettingsProps) {
        return null
    }
    if (!graphqlClient) {
        return null
    }

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

    return (
        <ApolloProvider client={graphqlClient}>
            <ErrorBoundary location={null}>
                <ShortcutProvider>
                    <TemporarySettingsProvider authenticatedUser={userAndSettingsProps.authenticatedUser}>
                        <SearchResultsCacheProvider>
                            <Router history={history} key={0}>
                                <Route
                                    path="/"
                                    render={routeComponentProps => (
                                        <CodeHostScopeProvider
                                            authenticatedUser={userAndSettingsProps.authenticatedUser}
                                        >
                                            <LayoutWithActivation
                                                {...props}
                                                {...routeComponentProps}
                                                history={history}
                                                location={location}
                                                platformContext={platformContext}
                                                extensionsController={extensionsController}
                                                {...userAndSettingsProps}
                                                parsedSearchQuery={parsedSearchQuery}
                                                setParsedSearchQuery={setParsedSearchQuery}
                                                setPatternType={setPatternType}
                                                setCaseSensitivity={setCaseSensitivity}
                                                hasUserAddedRepositories={hasUserAddedRepositories}
                                                hasUserAddedExternalServices={hasUserAddedExternalServices}
                                                onUserExternalServicesOrRepositoriesUpdate={
                                                    onUserExternalServicesOrRepositoriesUpdate
                                                }
                                                onSyncedPublicRepositoriesUpdate={onSyncedPublicRepositoriesUpdate}
                                                featureFlags={featureFlags}
                                                navbarSearchQueryState={navbarSearchQueryState}
                                                onNavbarQueryChange={onNavbarQueryChange}
                                                versionContext={versionContext}
                                                setVersionContext={setVersionContext2}
                                                availableVersionContexts={availableVersionContexts}
                                                previousVersionContext={previousVersionContext}
                                                defaultSearchContextSpec={defaultSearchContextSpec}
                                                selectedSearchContextSpec={getSelectedSearchContextSpec()}
                                                setSelectedSearchContextSpec={setSelectedSearchContextSpec2}
                                                authenticatedUser={userAndSettingsProps.authenticatedUser}
                                                // Search query
                                                fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                                                // Extensions
                                                telemetryService={eventLogger}
                                                isSourcegraphDotCom={window.context.sourcegraphDotComMode}
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
                                                fetchSavedSearches={fetchSavedSearches}
                                                fetchRecentSearches={fetchRecentSearches}
                                                fetchRecentFileViews={fetchRecentFileViews}
                                                streamSearch={aggregateStreamingSearch}
                                            />
                                        </CodeHostScopeProvider>
                                    )}
                                />
                            </Router>
                            <Tooltip key={1} />
                            <Notifications
                                key={2}
                                extensionsController={extensionsController}
                                notificationClassNames={notificationClassNames}
                            />
                        </SearchResultsCacheProvider>
                    </TemporarySettingsProvider>
                </ShortcutProvider>
            </ErrorBoundary>
        </ApolloProvider>
    )
}
