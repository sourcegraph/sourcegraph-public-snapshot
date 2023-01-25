import React, { Suspense, useCallback, useRef, useState } from 'react'

import classNames from 'classnames'
import { matchPath, Redirect, RouteComponentProps, Switch } from 'react-router'
import { CompatRoute } from 'react-router-dom-v5-compat'
import { Observable } from 'rxjs'

import { TabbedPanelContent } from '@sourcegraph/branded/src/components/panel/TabbedPanelContent'
import { isMacPlatform } from '@sourcegraph/common'
import { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { useKeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts/useKeyboardShortcut'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Shortcut } from '@sourcegraph/shared/src/react-shortcuts'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SearchContextProps } from '@sourcegraph/shared/src/search'
import { SettingsCascadeProps, SettingsSubjectCommonFields } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'
import { FeedbackPrompt, H1, Link, LoadingSpinner, Panel } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from './auth'
import type { BatchChangesProps } from './batches'
import type { CodeIntelligenceProps } from './codeintel'
import { CodeMonitoringProps } from './codeMonitoring'
import { communitySearchContextsRoutes } from './communitySearchContexts/routes'
import { AppRouterContainer } from './components/AppRouterContainer'
import { useBreadcrumbs } from './components/Breadcrumbs'
import { ErrorBoundary } from './components/ErrorBoundary'
import { LazyFuzzyFinder } from './components/fuzzyFinder/LazyFuzzyFinder'
import { KeyboardShortcutsHelp } from './components/KeyboardShortcutsHelp/KeyboardShortcutsHelp'
import { useScrollToLocationHash } from './components/useScrollToLocationHash'
import { useUserHistory } from './components/useUserHistory'
import { GlobalContributions } from './contributions'
import { useFeatureFlag } from './featureFlags/useFeatureFlag'
import { GlobalAlerts } from './global/GlobalAlerts'
import { useHandleSubmitFeedback } from './hooks'
import { SurveyToast } from './marketing/toast'
import { GlobalNavbar } from './nav/GlobalNavbar'
import type { BlockInput, NotebookProps } from './notebooks'
import { OrgAreaRoute } from './org/area/OrgArea'
import type { OrgAreaHeaderNavItem } from './org/area/OrgHeader'
import type { OrgSettingsAreaRoute } from './org/settings/OrgSettingsArea'
import type { OrgSettingsSidebarItems } from './org/settings/OrgSettingsSidebar'
import type { RepoContainerRoute } from './repo/RepoContainer'
import { RepoHeaderActionButton } from './repo/RepoHeader'
import type { RepoRevisionContainerRoute } from './repo/RepoRevisionContainer'
import type { RepoSettingsAreaRoute } from './repo/settings/RepoSettingsArea'
import type { RepoSettingsSideBarGroup } from './repo/settings/RepoSettingsSidebar'
import type { LayoutRouteComponentProps, LayoutRouteProps } from './routes'
import { EnterprisePageRoutes, PageRoutes } from './routes.constants'
import { parseSearchURLQuery, SearchAggregationProps, SearchStreamingProps } from './search'
import { NotepadContainer } from './search/Notepad'
import type { SiteAdminAreaRoute } from './site-admin/SiteAdminArea'
import type { SiteAdminSideBarGroups } from './site-admin/SiteAdminSidebar'
import { useTheme, useThemeProps } from './theme'
import type { UserAreaRoute } from './user/area/UserArea'
import type { UserAreaHeaderNavItem } from './user/area/UserAreaHeader'
import type { UserSettingsAreaRoute } from './user/settings/UserSettingsArea'
import type { UserSettingsSidebarItems } from './user/settings/UserSettingsSidebar'
import { getExperimentalFeatures } from './util/get-experimental-features'
import { parseBrowserRepoURL } from './util/url'

import styles from './Layout.module.scss'
import { XCompatRoute } from './XCompatRoute'

export interface LayoutProps
    extends RouteComponentProps<{}>,
        SettingsCascadeProps<Settings>,
        PlatformContextProps,
        ExtensionsControllerProps,
        TelemetryProps,
        SearchContextProps,
        SearchStreamingProps,
        CodeIntelligenceProps,
        BatchChangesProps,
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
    routes: readonly LayoutRouteProps<any>[]

    authenticatedUser: AuthenticatedUser | null

    /**
     * The subject GraphQL node ID of the viewer, which is used to look up the viewer's settings. This is either
     * the site's GraphQL node ID (for anonymous users) or the authenticated user's GraphQL node ID.
     */
    viewerSubject: SettingsSubjectCommonFields

    // Search
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    onCreateNotebookFromNotepad: (blocks: BlockInput[]) => void

    globbing: boolean
    isSourcegraphDotCom: boolean
    children?: never
}
/**
 * Syntax highlighting changes for WCAG 2.1 contrast compliance (currently behind feature flag)
 * https://github.com/sourcegraph/sourcegraph/issues/36251
 */
const CONTRAST_COMPLIANT_CLASSNAME = 'theme-contrast-compliant-syntax-highlighting'

export const Layout: React.FunctionComponent<React.PropsWithChildren<LayoutProps>> = props => {
    const routeMatch = props.routes.find(({ path, exact }) => matchPath(props.location.pathname, { path, exact }))?.path
    const isSearchRelatedPage = (routeMatch === PageRoutes.RepoContainer || routeMatch?.startsWith('/search')) ?? false
    const isSearchHomepage = props.location.pathname === '/search' && !parseSearchURLQuery(props.location.search)
    const isSearchConsolePage = routeMatch?.startsWith('/search/console')
    const isSearchNotebooksPage = routeMatch?.startsWith(EnterprisePageRoutes.Notebooks)
    const isSearchNotebookListPage = props.location.pathname === EnterprisePageRoutes.Notebooks
    const isRepositoryRelatedPage = routeMatch === PageRoutes.RepoContainer ?? false

    // enable fuzzy finder by default unless it's explicitly disabled in settings
    const fuzzyFinder = getExperimentalFeatures(props.settingsCascade.final).fuzzyFinder ?? true
    const [isFuzzyFinderVisible, setFuzzyFinderVisible] = useState(false)
    const userHistory = useUserHistory(props.history, isRepositoryRelatedPage)

    const communitySearchContextPaths = communitySearchContextsRoutes.map(route => route.path)
    const isCommunitySearchContextPage = communitySearchContextPaths.includes(props.location.pathname)

    // TODO add a component layer as the parent of the Layout component rendering "top-level" routes that do not render the navbar,
    // so that Layout can always render the navbar.
    const needsSiteInit = window.context?.needsSiteInit
    const disableFeedbackSurvey = window.context?.disableFeedbackSurvey
    const isSiteInit = props.location.pathname === PageRoutes.SiteAdminInit
    const isSignInOrUp =
        props.location.pathname === PageRoutes.SignIn ||
        props.location.pathname === PageRoutes.SignUp ||
        props.location.pathname === PageRoutes.PasswordReset ||
        props.location.pathname === PageRoutes.Welcome

    const themeProps = useThemeProps()
    const themeState = useTheme()
    const themeStateRef = useRef(themeState)
    themeStateRef.current = themeState
    const [enableContrastCompliantSyntaxHighlighting] = useFeatureFlag('contrast-compliant-syntax-highlighting')

    const breadcrumbProps = useBreadcrumbs()

    useScrollToLocationHash(props.location)

    const showHelpShortcut = useKeyboardShortcut('keyboardShortcutsHelp')
    const [keyboardShortcutsHelpOpen, setKeyboardShortcutsHelpOpen] = useState(false)
    const [feedbackModalOpen, setFeedbackModalOpen] = useState(false)
    const showKeyboardShortcutsHelp = useCallback(() => setKeyboardShortcutsHelpOpen(true), [])
    const hideKeyboardShortcutsHelp = useCallback(() => setKeyboardShortcutsHelpOpen(false), [])
    const showFeedbackModal = useCallback(() => setFeedbackModalOpen(true), [])

    const { handleSubmitFeedback } = useHandleSubmitFeedback({
        routeMatch,
    })

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

    const context: LayoutRouteComponentProps<any> = {
        ...props,
        ...themeProps,
        ...breadcrumbProps,
        isMacPlatform: isMacPlatform(),
    }

    const limitedRoutes = props.routes
    //.filter(r => [EnterprisePageRoutes.CodeMonitoring].includes(r.path))

    console.log(limitedRoutes.map(r => r.path))

    // return (
    //     <div className={classNames(styles.layout)}>
    //         <Switch>
    //             <CompatRoute path="code-monitoring" render={() => <div>code monitoring</div>} />
    //             <XCompatRoute pathV6="code-monitoring" render={() => <div>code monitoring X</div>} />
    //         </Switch>
    //     </div>
    // )

    /**
     * There're two issues with the CompatRoute:
     * 1. "*" doesn't work properly because of https://github.com/remix-run/react-router/blame/4d915e3305df5b01f51abdeb1c01bf442453522e/packages/react-router-dom-v5-compat/lib/components.tsx#L18
     * 2. Nested routes in Switch doesn't work properly because RR6 and RR5 paths
     * should be different before the migration is completed. It seems that the CompatRoute
     * still renders the element if RR5 path is not specified. See implementation here:
     * https://github.com/remix-run/react-router/blob/4d915e3305df5b01f51abdeb1c01bf442453522e/packages/react-router-dom-v5-compat/lib/components.tsx#L21
     *
     * More context:
     * 1. https://github.com/remix-run/react-router/issues/9558#issuecomment-1315168000
     * 2. https://github.com/remix-run/react-router/discussions/9509
     *
     * We can implement an ugly migration keeping both RR5 and RR6 path props.
     * On migration completion we can remove `path` and renamed pathV6` or whatever seems
     * the best combination.
     */
    return (
        <div className={classNames(styles.layout)}>
            <header className="align-self-start m-3">
                <H1>Layout content</H1>
                <Link className="mr-3" to="/">
                    Home
                </Link>
                <Link className="mr-3" to="/code-monitoring">
                    Code monitoring
                </Link>
                <Link className="mr-3" to="/code-monitoring/new">
                    New code monitor
                </Link>
                <Link to="/batch-changes">Batch changes</Link>
            </header>

            <Suspense fallback={<LoadingSpinner className="m-2" />}>
                <Switch>
                    {limitedRoutes.map(({ render, ...route }) => (
                        <XCompatRoute
                            key="hardcoded-key"
                            path={route.path}
                            pathV6={`${route.path}/*`}
                            render={(routeComponentProps: RouteComponentProps<{}>) => {
                                console.log('render route', route.path)

                                return render({ ...context, ...routeComponentProps })
                            }}
                        />
                    ))}
                </Switch>
            </Suspense>
        </div>
    )
}
