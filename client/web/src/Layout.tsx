import React, { Suspense, useCallback, useRef, useState } from 'react'

import classNames from 'classnames'
import { matchPath, useLocation, Route, Routes, Navigate } from 'react-router-dom-v5-compat'
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
import { FeedbackPrompt, LoadingSpinner, Panel } from '@sourcegraph/wildcard'

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
import type { NotebookProps } from './notebooks'
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
import { SetupWizard } from './setup-wizard'
import type { SiteAdminAreaRoute } from './site-admin/SiteAdminArea'
import type { SiteAdminSideBarGroups } from './site-admin/SiteAdminSidebar'
import { useExperimentalFeatures } from './stores'
import { useTheme, useThemeProps } from './theme'
import type { UserAreaRoute } from './user/area/UserArea'
import type { UserAreaHeaderNavItem } from './user/area/UserAreaHeader'
import type { UserSettingsAreaRoute } from './user/settings/UserSettingsArea'
import type { UserSettingsSidebarItems } from './user/settings/UserSettingsSidebar'
import { getExperimentalFeatures } from './util/get-experimental-features'
import { parseBrowserRepoURL } from './util/url'

import styles from './Layout.module.scss'

export interface LayoutProps
    extends SettingsCascadeProps<Settings>,
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
    routes: readonly LayoutRouteProps[]

    authenticatedUser: AuthenticatedUser | null

    /**
     * The subject GraphQL node ID of the viewer, which is used to look up the viewer's settings. This is either
     * the site's GraphQL node ID (for anonymous users) or the authenticated user's GraphQL node ID.
     */
    viewerSubject: SettingsSubjectCommonFields

    // Search
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>

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
    const location = useLocation()

    // TODO: Replace with useMatches once top-level <Router/> is V6
    const routeMatch = props.routes.find(
        route =>
            matchPath(route.path, location.pathname) || matchPath(route.path.replace(/\/\*$/, ''), location.pathname)
    )?.path

    const isSearchRelatedPage = (routeMatch === PageRoutes.RepoContainer || routeMatch?.startsWith('/search')) ?? false
    const isSearchHomepage = location.pathname === '/search' && !parseSearchURLQuery(location.search)
    const isSearchConsolePage = routeMatch?.startsWith('/search/console')
    const isSearchNotebooksPage = routeMatch?.startsWith(EnterprisePageRoutes.Notebooks)
    const isSearchNotebookListPage = location.pathname === EnterprisePageRoutes.Notebooks
    const isRepositoryRelatedPage = routeMatch === PageRoutes.RepoContainer ?? false

    const { setupWizard } = useExperimentalFeatures()
    const isSetupWizardPage = setupWizard && location.pathname.startsWith(PageRoutes.SetupWizard)

    // enable fuzzy finder by default unless it's explicitly disabled in settings
    const fuzzyFinder = getExperimentalFeatures(props.settingsCascade.final).fuzzyFinder ?? true
    const [isFuzzyFinderVisible, setFuzzyFinderVisible] = useState(false)
    const userHistory = useUserHistory(isRepositoryRelatedPage)

    const communitySearchContextPaths = communitySearchContextsRoutes.map(route => route.path)
    const isCommunitySearchContextPage = communitySearchContextPaths.includes(location.pathname)

    // TODO add a component layer as the parent of the Layout component rendering "top-level" routes that do not render the navbar,
    // so that Layout can always render the navbar.
    const needsSiteInit = window.context?.needsSiteInit
    const disableFeedbackSurvey = window.context?.disableFeedbackSurvey
    const isSiteInit = location.pathname === PageRoutes.SiteAdminInit
    const isSignInOrUp =
        location.pathname === PageRoutes.SignIn ||
        location.pathname === PageRoutes.SignUp ||
        location.pathname === PageRoutes.PasswordReset ||
        location.pathname === PageRoutes.Welcome

    const themeProps = useThemeProps()
    const themeState = useTheme()
    const themeStateRef = useRef(themeState)
    themeStateRef.current = themeState
    const [enableContrastCompliantSyntaxHighlighting] = useFeatureFlag('contrast-compliant-syntax-highlighting')

    const breadcrumbProps = useBreadcrumbs()

    useScrollToLocationHash(location)

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
    if (location.pathname !== '/' && location.pathname.endsWith('/')) {
        return <Navigate replace={true} to={{ ...location, pathname: location.pathname.slice(0, -1) }} />
    }

    const context = {
        ...props,
        ...themeProps,
        ...breadcrumbProps,
        isMacPlatform: isMacPlatform(),
    } satisfies Omit<LayoutRouteComponentProps, 'location' | 'history' | 'match' | 'staticContext'>

    if (isSetupWizardPage) {
        return <SetupWizard />
    }

    return (
        <div
            className={classNames(
                styles.layout,
                enableContrastCompliantSyntaxHighlighting && CONTRAST_COMPLIANT_CLASSNAME
            )}
        >
            {showHelpShortcut?.keybindings.map((keybinding, index) => (
                <Shortcut key={index} {...keybinding} onMatch={showKeyboardShortcutsHelp} />
            ))}
            <KeyboardShortcutsHelp isOpen={keyboardShortcutsHelpOpen} onDismiss={hideKeyboardShortcutsHelp} />

            {feedbackModalOpen && (
                <FeedbackPrompt
                    onSubmit={handleSubmitFeedback}
                    modal={true}
                    openByDefault={true}
                    authenticatedUser={
                        props.authenticatedUser
                            ? {
                                  username: props.authenticatedUser.username || '',
                                  email: props.authenticatedUser.emails.find(email => email.isPrimary)?.email || '',
                              }
                            : null
                    }
                    onClose={() => setFeedbackModalOpen(false)}
                />
            )}

            <GlobalAlerts
                authenticatedUser={props.authenticatedUser}
                settingsCascade={props.settingsCascade}
                isSourcegraphDotCom={props.isSourcegraphDotCom}
            />
            {!isSiteInit && !isSignInOrUp && !props.isSourcegraphDotCom && !disableFeedbackSurvey && (
                <SurveyToast authenticatedUser={props.authenticatedUser} />
            )}
            {!isSiteInit && !isSignInOrUp && (
                <GlobalNavbar
                    {...props}
                    {...themeProps}
                    showSearchBox={
                        isSearchRelatedPage &&
                        !isSearchHomepage &&
                        !isCommunitySearchContextPage &&
                        !isSearchConsolePage &&
                        !isSearchNotebooksPage
                    }
                    setFuzzyFinderIsVisible={setFuzzyFinderVisible}
                    isRepositoryRelatedPage={isRepositoryRelatedPage}
                    showKeyboardShortcutsHelp={showKeyboardShortcutsHelp}
                    showFeedbackModal={showFeedbackModal}
                    enableLegacyExtensions={window.context.enableLegacyExtensions}
                />
            )}
            {needsSiteInit && !isSiteInit && <Navigate replace={true} to="/site-admin/init" />}
            <ErrorBoundary location={location}>
                <Suspense
                    fallback={
                        <div className="flex flex-1">
                            <LoadingSpinner className="m-2" />
                        </div>
                    }
                >
                    <AppRouterContainer>
                        <Routes>
                            {props.routes.map(
                                ({ condition = () => true, ...route }) =>
                                    condition(context) && (
                                        <Route
                                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                            path={route.path}
                                            element={route.render(context)}
                                        />
                                    )
                            )}
                        </Routes>
                    </AppRouterContainer>
                </Suspense>
            </ErrorBoundary>
            {parseQueryAndHash(location.search, location.hash).viewState && location.pathname !== PageRoutes.SignIn && (
                <Panel
                    className={styles.panel}
                    position="bottom"
                    defaultSize={350}
                    storageKey="panel-size"
                    ariaLabel="References panel"
                    id="references-panel"
                >
                    <TabbedPanelContent
                        {...props}
                        {...themeProps}
                        repoName={`git://${parseBrowserRepoURL(location.pathname).repoName}`}
                        fetchHighlightedFileLineRanges={props.fetchHighlightedFileLineRanges}
                    />
                </Panel>
            )}
            <GlobalContributions
                key={3}
                extensionsController={props.extensionsController}
                platformContext={props.platformContext}
            />
            {(isSearchNotebookListPage || (isSearchRelatedPage && !isSearchHomepage)) && (
                <NotepadContainer userId={props.authenticatedUser?.id} />
            )}
            {fuzzyFinder && (
                <LazyFuzzyFinder
                    isVisible={isFuzzyFinderVisible}
                    setIsVisible={setFuzzyFinderVisible}
                    themeState={themeStateRef}
                    isRepositoryRelatedPage={isRepositoryRelatedPage}
                    settingsCascade={props.settingsCascade}
                    telemetryService={props.telemetryService}
                    location={location}
                    userHistory={userHistory}
                />
            )}
        </div>
    )
}
