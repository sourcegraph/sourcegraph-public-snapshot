/* eslint-disable import/order */
import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import { AuthenticatedUser, authRequired as authRequiredObservable } from './auth'
import {
    CaseSensitivityProps,
    HomePanelsProps,
    OnboardingTourProps,
    ParsedSearchQueryProps,
    PatternTypeProps,
    SearchContextProps,
    SearchStreamingProps,
    getGlobalSearchContextFilter,
    parseSearchURL,
    parseSearchURLQuery,
} from './search'
import { EnterprisePageRoutes, PageRoutes } from './routes.constants'
import { KEYBOARD_SHORTCUT_SHOW_HELP, KeyboardShortcutsProps } from './keyboardShortcuts/keyboardShortcuts'
import { LayoutRouteComponentProps, LayoutRouteProps } from './routes'
import React, { Suspense, useCallback, useEffect, useMemo } from 'react'
import { Redirect, Route, RouteComponentProps, Switch, matchPath } from 'react-router'
import { UserExternalServicesOrRepositoriesUpdateProps, isMacPlatform } from './util'

import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { BatchChangesProps } from './batches'
import { CodeInsightsProps } from './insights/types'
import { CodeIntelligenceProps } from './codeintel'
import { CodeMonitoringProps } from './code-monitoring'
import { ErrorBoundary } from './components/ErrorBoundary'
import { ExtensionAreaHeaderNavItem } from './extensions/extension/ExtensionAreaHeader'
import { ExtensionAreaRoute } from './extensions/extension/ExtensionArea'
import { ExtensionsAreaHeaderActionButton } from './extensions/ExtensionsAreaHeader'
import { ExtensionsAreaRoute } from './extensions/ExtensionsArea'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { FeatureFlagProps } from './featureFlags/featureFlags'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { GlobalAlerts } from './global/GlobalAlerts'
import { GlobalContributions } from './contributions'
import { GlobalDebug } from './global/GlobalDebug'
import { GlobalNavbar } from './nav/GlobalNavbar'
import { KeyboardShortcutsHelp } from './keyboardShortcuts/KeyboardShortcutsHelp'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Observable } from 'rxjs'
import { OrgAreaHeaderNavItem } from './org/area/OrgHeader'
import { OrgAreaRoute } from './org/area/OrgArea'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { RepoContainerRoute } from './repo/RepoContainer'
import { RepoHeaderActionButton } from './repo/RepoHeader'
import { RepoRevisionContainerRoute } from './repo/RepoRevisionContainer'
import { RepoSettingsAreaRoute } from './repo/settings/RepoSettingsArea'
import { RepoSettingsSideBarGroup } from './repo/settings/RepoSettingsSidebar'
import { ResizablePanel } from '@sourcegraph/branded/src/components/panel/Panel'
import { Settings } from './schema/settings.schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { SiteAdminAreaRoute } from './site-admin/SiteAdminArea'
import { SiteAdminSideBarGroups } from './site-admin/SiteAdminSidebar'
import { SurveyToast } from './marketing/SurveyToast'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { UserAreaHeaderNavItem } from './user/area/UserAreaHeader'
import { UserAreaRoute } from './user/area/UserArea'
import { UserSettingsAreaRoute } from './user/settings/UserSettingsArea'
import { UserSettingsSidebarItems } from './user/settings/UserSettingsSidebar'
import { communitySearchContextsRoutes } from './communitySearchContexts/routes'
import { fetchHighlightedFileLineRanges } from './repo/backend'
import { parseBrowserRepoURL } from './util/url'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'
import styles from './Layout.module.scss'
import { useBreadcrumbs } from './components/Breadcrumbs'
import { useExtensionAlertAnimation } from './nav/UserNavItem'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { useScrollToLocationHash } from './components/useScrollToLocationHash'
import { useTemporarySetting } from './settings/temporary/useTemporarySetting'
import { useTheme } from './theme'

export interface LayoutProps
    extends RouteComponentProps<{}>,
        SettingsCascadeProps<Settings>,
        PlatformContextProps,
        ExtensionsControllerProps,
        KeyboardShortcutsProps,
        TelemetryProps,
        ActivationProps,
        ParsedSearchQueryProps,
        PatternTypeProps,
        CaseSensitivityProps,
        OnboardingTourProps,
        SearchContextProps,
        HomePanelsProps,
        SearchStreamingProps,
        CodeMonitoringProps,
        SearchContextProps,
        UserExternalServicesOrRepositoriesUpdateProps,
        CodeIntelligenceProps,
        BatchChangesProps,
        CodeInsightsProps,
        FeatureFlagProps {
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

    authenticatedUser: AuthenticatedUser | null

    /**
     * The subject GraphQL node ID of the viewer, which is used to look up the viewer's settings. This is either
     * the site's GraphQL node ID (for anonymous users) or the authenticated user's GraphQL node ID.
     */
    viewerSubject: Pick<GQL.ISettingsSubject, 'id' | 'viewerCanAdminister'>

    // Search
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>

    globbing: boolean
    showMultilineSearchConsole: boolean
    showSearchNotebook: boolean
    isSourcegraphDotCom: boolean
    fetchSavedSearches: () => Observable<GQL.ISavedSearch[]>
    children?: never
}

export const Layout: React.FunctionComponent<LayoutProps> = props => {
    const routeMatch = props.routes.find(({ path, exact }) => matchPath(props.location.pathname, { path, exact }))?.path
    const isSearchRelatedPage = (routeMatch === '/:repoRevAndRest+' || routeMatch?.startsWith('/search')) ?? false
    const minimalNavLinks = routeMatch === '/cncf'
    const isSearchHomepage = props.location.pathname === '/search' && !parseSearchURLQuery(props.location.search)
    const isSearchConsolePage = routeMatch?.startsWith('/search/console')
    const isSearchNotebookPage = routeMatch?.startsWith('/search/notebook')
    const isRepositoryRelatedPage = routeMatch === '/:repoRevAndRest+' ?? false

    // Update parsedSearchQuery, patternType, caseSensitivity, and selectedSearchContextSpec based on current URL
    const {
        history,
        parsedSearchQuery: currentQuery,
        patternType: currentPatternType,
        caseSensitive: currentCaseSensitive,
        selectedSearchContextSpec,
        location,
        setParsedSearchQuery,
        setPatternType,
        setCaseSensitivity,
        setSelectedSearchContextSpec,
    } = props

    const { query = '', patternType, caseSensitive } = useMemo(() => parseSearchURL(location.search), [location.search])

    const searchContextSpec = useMemo(() => getGlobalSearchContextFilter(query)?.spec, [query])

    useEffect(() => {
        if (query !== currentQuery) {
            setParsedSearchQuery(query)
        }

        // Only override filters from URL if there is a search query
        if (query) {
            if (patternType && patternType !== currentPatternType) {
                setPatternType(patternType)
            }

            if (caseSensitive !== currentCaseSensitive) {
                setCaseSensitivity(caseSensitive)
            }

            if (searchContextSpec && searchContextSpec !== selectedSearchContextSpec) {
                setSelectedSearchContextSpec(searchContextSpec)
            }
        }
    }, [
        history,
        caseSensitive,
        currentCaseSensitive,
        currentPatternType,
        currentQuery,
        selectedSearchContextSpec,
        patternType,
        query,
        setCaseSensitivity,
        setParsedSearchQuery,
        setPatternType,
        setSelectedSearchContextSpec,
        searchContextSpec,
    ])

    const [hasUsedNonGlobalContext, setHasUsedNonGlobalContext] = useTemporarySetting('search.usedNonGlobalContext')
    useEffect(() => {
        if (selectedSearchContextSpec && selectedSearchContextSpec !== 'global' && !hasUsedNonGlobalContext) {
            setHasUsedNonGlobalContext(true)
        }
    }, [selectedSearchContextSpec, setHasUsedNonGlobalContext, hasUsedNonGlobalContext])

    const communitySearchContextPaths = communitySearchContextsRoutes.map(route => route.path)
    const isCommunitySearchContextPage = communitySearchContextPaths.includes(props.location.pathname)

    // TODO add a component layer as the parent of the Layout component rendering "top-level" routes that do not render the navbar,
    // so that Layout can always render the navbar.
    const needsSiteInit = window.context?.needsSiteInit
    const isSiteInit = props.location.pathname === PageRoutes.SiteAdminInit
    const isSignInOrUp =
        props.location.pathname === PageRoutes.SignIn ||
        props.location.pathname === PageRoutes.SignUp ||
        props.location.pathname === PageRoutes.PasswordReset ||
        props.location.pathname === PageRoutes.Welcome

    // TODO Change this behavior when we have global focus management system
    // Need to know this for disable autofocus on nav search input
    // and preserve autofocus for first textarea at survey page, creation UI etc.
    const isSearchAutoFocusRequired = routeMatch === PageRoutes.Survey || routeMatch === EnterprisePageRoutes.Insights

    const authRequired = useObservable(authRequiredObservable)

    const themeProps = useTheme()

    const breadcrumbProps = useBreadcrumbs()

    // Control browser extension discoverability animation here.
    // `Layout` is the lowest common ancestor of `UserNavItem` (target) and `RepoContainer` (trigger)
    const { isExtensionAlertAnimating, startExtensionAlertAnimation } = useExtensionAlertAnimation()
    const onExtensionAlertDismissed = useCallback(() => {
        startExtensionAlertAnimation()
    }, [startExtensionAlertAnimation])

    useScrollToLocationHash(props.location)
    // Remove trailing slash (which is never valid in any of our URLs).
    if (props.location.pathname !== '/' && props.location.pathname.endsWith('/')) {
        return <Redirect to={{ ...props.location, pathname: props.location.pathname.slice(0, -1) }} />
    }

    const context: LayoutRouteComponentProps<any> = {
        ...props,
        ...themeProps,
        ...breadcrumbProps,
        onExtensionAlertDismissed,
        isMacPlatform,
    }

    return (
        <div className={styles.layout}>
            <KeyboardShortcutsHelp
                keyboardShortcutForShow={KEYBOARD_SHORTCUT_SHOW_HELP}
                keyboardShortcuts={props.keyboardShortcuts}
            />
            <GlobalAlerts authenticatedUser={props.authenticatedUser} settingsCascade={props.settingsCascade} />
            {!isSiteInit && <SurveyToast />}
            {!isSiteInit && !isSignInOrUp && (
                <GlobalNavbar
                    {...props}
                    {...themeProps}
                    authRequired={!!authRequired}
                    showSearchBox={
                        isSearchRelatedPage &&
                        !isSearchHomepage &&
                        !isCommunitySearchContextPage &&
                        !isSearchConsolePage &&
                        !isSearchNotebookPage
                    }
                    variant={
                        isSearchHomepage
                            ? 'low-profile'
                            : isCommunitySearchContextPage
                            ? 'low-profile-with-logo'
                            : 'default'
                    }
                    hideNavLinks={false}
                    minimalNavLinks={minimalNavLinks}
                    isSearchAutoFocusRequired={!isSearchAutoFocusRequired}
                    isExtensionAlertAnimating={isExtensionAlertAnimating}
                    isRepositoryRelatedPage={isRepositoryRelatedPage}
                />
            )}
            {needsSiteInit && !isSiteInit && <Redirect to="/site-admin/init" />}
            <ErrorBoundary location={props.location}>
                <Suspense
                    fallback={
                        <div className="flex flex-1">
                            <LoadingSpinner className="icon-inline m-2" />
                        </div>
                    }
                >
                    <Switch>
                        {props.routes.map(
                            ({ render, condition = () => true, ...route }) =>
                                condition(context) && (
                                    <Route
                                        {...route}
                                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                        component={undefined}
                                        render={routeComponentProps => (
                                            <div className={styles.appRouterContainer}>
                                                {render({ ...context, ...routeComponentProps })}
                                            </div>
                                        )}
                                    />
                                )
                        )}
                    </Switch>
                </Suspense>
            </ErrorBoundary>
            {parseQueryAndHash(props.location.search, props.location.hash).viewState &&
                props.location.pathname !== PageRoutes.SignIn && (
                    <ResizablePanel
                        {...props}
                        {...themeProps}
                        repoName={`git://${parseBrowserRepoURL(props.location.pathname).repoName}`}
                        fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                    />
                )}
            <GlobalContributions
                key={3}
                extensionsController={props.extensionsController}
                platformContext={props.platformContext}
                history={props.history}
            />
            <GlobalDebug {...props} />
        </div>
    )
}
