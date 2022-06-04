import React, { Suspense, useCallback, useEffect, useMemo } from 'react'

import { Redirect, Route, RouteComponentProps, Switch, matchPath } from 'react-router'
import { Observable } from 'rxjs'

import { TabbedPanelContent } from '@sourcegraph/branded/src/components/panel/TabbedPanelContent'
import { isMacPlatform } from '@sourcegraph/common'
import { SearchContextProps } from '@sourcegraph/search'
import { FetchFileParameters } from '@sourcegraph/search-ui'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import {
    KeyboardShortcutsProps,
    KEYBOARD_SHORTCUT_SHOW_HELP,
} from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import { KeyboardShortcutsHelp } from '@sourcegraph/shared/src/keyboardShortcuts/KeyboardShortcutsHelp'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import * as GQL from '@sourcegraph/shared/src/schema'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { getGlobalSearchContextFilter } from '@sourcegraph/shared/src/search/query/query'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'
import { LoadingSpinner, Panel, useObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser, authRequired as authRequiredObservable } from './auth'
import { BatchChangesProps } from './batches'
import { CodeIntelligenceProps } from './codeintel'
import { communitySearchContextsRoutes } from './communitySearchContexts/routes'
import { AppRouterContainer } from './components/AppRouterContainer'
import { useBreadcrumbs } from './components/Breadcrumbs'
import { ErrorBoundary } from './components/ErrorBoundary'
import { useScrollToLocationHash } from './components/useScrollToLocationHash'
import { GlobalContributions } from './contributions'
import { ExtensionAreaRoute } from './extensions/extension/ExtensionArea'
import { ExtensionAreaHeaderNavItem } from './extensions/extension/ExtensionAreaHeader'
import { ExtensionsAreaRoute } from './extensions/ExtensionsArea'
import { ExtensionsAreaHeaderActionButton } from './extensions/ExtensionsAreaHeader'
import { GlobalAlerts } from './global/GlobalAlerts'
import { GlobalDebug } from './global/GlobalDebug'
import { SurveyToast } from './marketing/SurveyToast'
import { GlobalNavbar } from './nav/GlobalNavbar'
import { useExtensionAlertAnimation } from './nav/UserNavItem'
import { OrgAreaRoute } from './org/area/OrgArea'
import { OrgAreaHeaderNavItem } from './org/area/OrgHeader'
import { RepoContainerRoute } from './repo/RepoContainer'
import { RepoHeaderActionButton } from './repo/RepoHeader'
import { RepoRevisionContainerRoute } from './repo/RepoRevisionContainer'
import { RepoSettingsAreaRoute } from './repo/settings/RepoSettingsArea'
import { RepoSettingsSideBarGroup } from './repo/settings/RepoSettingsSidebar'
import { LayoutRouteProps, LayoutRouteComponentProps } from './routes'
import { PageRoutes, EnterprisePageRoutes } from './routes.constants'
import { parseSearchURLQuery, HomePanelsProps, SearchStreamingProps, parseSearchURL } from './search'
import { SiteAdminAreaRoute } from './site-admin/SiteAdminArea'
import { SiteAdminSideBarGroups } from './site-admin/SiteAdminSidebar'
import { setQueryStateFromURL } from './stores'
import { useThemeProps } from './theme'
import { UserAreaRoute } from './user/area/UserArea'
import { UserAreaHeaderNavItem } from './user/area/UserAreaHeader'
import { UserSettingsAreaRoute } from './user/settings/UserSettingsArea'
import { UserSettingsSidebarItems } from './user/settings/UserSettingsSidebar'
import { UserExternalServicesOrRepositoriesUpdateProps } from './util'
import { parseBrowserRepoURL } from './util/url'

import styles from './Layout.module.scss'

export interface LayoutProps
    extends RouteComponentProps<{}>,
        SettingsCascadeProps<Settings>,
        PlatformContextProps,
        ExtensionsControllerProps,
        KeyboardShortcutsProps,
        TelemetryProps,
        ActivationProps,
        SearchContextProps,
        HomePanelsProps,
        SearchStreamingProps,
        UserExternalServicesOrRepositoriesUpdateProps,
        CodeIntelligenceProps,
        BatchChangesProps {
    extensionAreaRoutes: readonly ExtensionAreaRoute[]
    extensionAreaHeaderNavItems: readonly ExtensionAreaHeaderNavItem[]
    extensionsAreaRoutes: readonly ExtensionsAreaRoute[]
    extensionsAreaHeaderActionButtons: readonly ExtensionsAreaHeaderActionButton[]
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

    authenticatedUser: AuthenticatedUser | null

    /**
     * The subject GraphQL node ID of the viewer, which is used to look up the viewer's settings. This is either
     * the site's GraphQL node ID (for anonymous users) or the authenticated user's GraphQL node ID.
     */
    viewerSubject: Pick<GQL.ISettingsSubject, 'id' | 'viewerCanAdminister'>

    // Search
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>

    globbing: boolean
    isSourcegraphDotCom: boolean
    children?: never
}

export const Layout: React.FunctionComponent<React.PropsWithChildren<LayoutProps>> = props => {
    const routeMatch = props.routes.find(({ path, exact }) => matchPath(props.location.pathname, { path, exact }))?.path
    const isSearchRelatedPage = (routeMatch === '/:repoRevAndRest+' || routeMatch?.startsWith('/search')) ?? false
    const minimalNavLinks = routeMatch === '/cncf'
    const isSearchHomepage = props.location.pathname === '/search' && !parseSearchURLQuery(props.location.search)
    const isSearchConsolePage = routeMatch?.startsWith('/search/console')
    const isSearchNotebooksPage = routeMatch?.startsWith(PageRoutes.Notebooks)
    const isRepositoryRelatedPage = routeMatch === '/:repoRevAndRest+' ?? false

    // Update patternType, caseSensitivity, and selectedSearchContextSpec based on current URL
    const { history, selectedSearchContextSpec, location, setSelectedSearchContextSpec } = props

    useEffect(() => setQueryStateFromURL(location.search), [location.search])

    const { query = '' } = useMemo(() => parseSearchURL(location.search), [location.search])

    const searchContextSpec = useMemo(() => getGlobalSearchContextFilter(query)?.spec, [query])

    useEffect(() => {
        // Only override filters from URL if there is a search query
        if (query) {
            if (searchContextSpec && searchContextSpec !== selectedSearchContextSpec) {
                setSelectedSearchContextSpec(searchContextSpec)
            }
        }
    }, [history, selectedSearchContextSpec, query, setSelectedSearchContextSpec, searchContextSpec])

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

    const themeProps = useThemeProps()

    const breadcrumbProps = useBreadcrumbs()

    // Control browser extension discoverability animation here.
    // `Layout` is the lowest common ancestor of `UserNavItem` (target) and `RepoContainer` (trigger)
    const { isExtensionAlertAnimating, startExtensionAlertAnimation } = useExtensionAlertAnimation()
    const onExtensionAlertDismissed = useCallback(() => {
        startExtensionAlertAnimation()
    }, [startExtensionAlertAnimation])

    useScrollToLocationHash(props.location)

    // Note: this was a poor UX and is disabled for now, see https://github.com/sourcegraph/sourcegraph/issues/30192
    // const [tosAccepted, setTosAccepted] = useState(true) // Assume TOS has been accepted so that we don't show the TOS modal on initial load
    // useEffect(() => setTosAccepted(!props.authenticatedUser || props.authenticatedUser.tosAccepted), [
    //     props.authenticatedUser,
    // ])
    // const afterTosAccepted = useCallback(() => {
    //     setTosAccepted(true)
    // }, [])

    // Remove trailing slash (which is never valid in any of our URLs).
    if (props.location.pathname !== '/' && props.location.pathname.endsWith('/')) {
        return <Redirect to={{ ...props.location, pathname: props.location.pathname.slice(0, -1) }} />
    }

    // Note: this was a poor UX and is disabled for now, see https://github.com/sourcegraph/sourcegraph/issues/30192
    // If a user has not accepted the Terms of Service yet, show the modal to force them to accept
    // before continuing to use Sourcegraph. This is only done on self-hosted Sourcegraph Server;
    // cloud users are all considered to have accepted regarless of the value of `tosAccepted`.
    // if (!props.isSourcegraphDotCom && !tosAccepted) {
    //     return <TosConsentModal afterTosAccepted={afterTosAccepted} />
    // }

    const context: LayoutRouteComponentProps<any> = {
        ...props,
        ...themeProps,
        ...breadcrumbProps,
        onExtensionAlertDismissed,
        isMacPlatform: isMacPlatform(),
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
                        !isSearchNotebooksPage
                    }
                    variant={
                        isSearchHomepage
                            ? 'low-profile'
                            : isCommunitySearchContextPage
                            ? 'low-profile-with-logo'
                            : 'default'
                    }
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
                            <LoadingSpinner className="m-2" />
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
                                            <AppRouterContainer>
                                                {render({ ...context, ...routeComponentProps })}
                                            </AppRouterContainer>
                                        )}
                                    />
                                )
                        )}
                    </Switch>
                </Suspense>
            </ErrorBoundary>
            {parseQueryAndHash(props.location.search, props.location.hash).viewState &&
                props.location.pathname !== PageRoutes.SignIn && (
                    <Panel className={styles.panel} position="bottom" defaultSize={350} storageKey="panel-size">
                        <TabbedPanelContent
                            {...props}
                            {...themeProps}
                            repoName={`git://${parseBrowserRepoURL(props.location.pathname).repoName}`}
                            fetchHighlightedFileLineRanges={props.fetchHighlightedFileLineRanges}
                        />
                    </Panel>
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
