import React, { Suspense, useCallback, useEffect, useState } from 'react'

import { Shortcut } from '@slimsag/react-shortcuts'
import classNames from 'classnames'
import { Redirect, Route, RouteComponentProps, Switch, matchPath } from 'react-router'
import { Observable } from 'rxjs'

import { TabbedPanelContent } from '@sourcegraph/branded/src/components/panel/TabbedPanelContent'
import { isMacPlatform } from '@sourcegraph/common'
import { SearchContextProps } from '@sourcegraph/search'
import { FetchFileParameters } from '@sourcegraph/search-ui'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { useKeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts/useKeyboardShortcut'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import * as GQL from '@sourcegraph/shared/src/schema'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { useCoreWorkflowImprovementsEnabled } from '@sourcegraph/shared/src/settings/useCoreWorkflowImprovementsEnabled'
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
import { FuzzyFinder } from './components/fuzzyFinder/FuzzyFinder'
import { KeyboardShortcutsHelp } from './components/KeyboardShortcutsHelp/KeyboardShortcutsHelp'
import { useScrollToLocationHash } from './components/useScrollToLocationHash'
import { GlobalContributions } from './contributions'
import { ExtensionAreaRoute } from './extensions/extension/ExtensionArea'
import { ExtensionAreaHeaderNavItem } from './extensions/extension/ExtensionAreaHeader'
import { ExtensionsAreaRoute } from './extensions/ExtensionsArea'
import { ExtensionsAreaHeaderActionButton } from './extensions/ExtensionsAreaHeader'
import { useFeatureFlag } from './featureFlags/useFeatureFlag'
import { GlobalAlerts } from './global/GlobalAlerts'
import { GlobalDebug } from './global/GlobalDebug'
import { SurveyToast } from './marketing/toast'
import { GlobalNavbar } from './nav/GlobalNavbar'
import type { BlockInput } from './notebooks'
import { OrgAreaRoute } from './org/area/OrgArea'
import { OrgAreaHeaderNavItem } from './org/area/OrgHeader'
import { RepoContainerRoute } from './repo/RepoContainer'
import { RepoHeaderActionButton } from './repo/RepoHeader'
import { RepoRevisionContainerRoute } from './repo/RepoRevisionContainer'
import { RepoSettingsAreaRoute } from './repo/settings/RepoSettingsArea'
import { RepoSettingsSideBarGroup } from './repo/settings/RepoSettingsSidebar'
import { LayoutRouteProps, LayoutRouteComponentProps } from './routes'
import { PageRoutes, EnterprisePageRoutes } from './routes.constants'
import { parseSearchURLQuery, HomePanelsProps, SearchStreamingProps } from './search'
import { NotepadContainer } from './search/Notepad'
import { SiteAdminAreaRoute } from './site-admin/SiteAdminArea'
import { SiteAdminSideBarGroups } from './site-admin/SiteAdminSidebar'
import { useThemeProps } from './theme'
import { UserAreaRoute } from './user/area/UserArea'
import { UserAreaHeaderNavItem } from './user/area/UserAreaHeader'
import { UserSettingsAreaRoute } from './user/settings/UserSettingsArea'
import { UserSettingsSidebarItems } from './user/settings/UserSettingsSidebar'
import { getExperimentalFeatures } from './util/get-experimental-features'
import { parseBrowserRepoURL } from './util/url'

import styles from './Layout.module.scss'

export interface LayoutProps
    extends RouteComponentProps<{}>,
        SettingsCascadeProps<Settings>,
        PlatformContextProps,
        ExtensionsControllerProps,
        TelemetryProps,
        ActivationProps,
        SearchContextProps,
        HomePanelsProps,
        SearchStreamingProps,
        CodeIntelligenceProps,
        BatchChangesProps {
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

    authenticatedUser: AuthenticatedUser | null

    /**
     * The subject GraphQL node ID of the viewer, which is used to look up the viewer's settings. This is either
     * the site's GraphQL node ID (for anonymous users) or the authenticated user's GraphQL node ID.
     */
    viewerSubject: Pick<GQL.ISettingsSubject, 'id' | 'viewerCanAdminister'>

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
    const isSearchRelatedPage = (routeMatch === '/:repoRevAndRest+' || routeMatch?.startsWith('/search')) ?? false
    const minimalNavLinks = routeMatch === '/cncf'
    const isSearchHomepage = props.location.pathname === '/search' && !parseSearchURLQuery(props.location.search)
    const isSearchConsolePage = routeMatch?.startsWith('/search/console')
    const isSearchNotebooksPage = routeMatch?.startsWith(PageRoutes.Notebooks)
    const isSearchNotebookListPage = props.location.pathname === PageRoutes.Notebooks
    const isRepositoryRelatedPage = routeMatch === '/:repoRevAndRest+' ?? false

    const [isFuzzyFinderVisible, setIsFuzzyFinderVisible] = useState(false)
    const fuzzyFinderShortcut = useKeyboardShortcut('fuzzyFinder')
    const [retainFuzzyFinderCache, setRetainFuzzyFinderCache] = useState(true)

    let { fuzzyFinder } = getExperimentalFeatures(props.settingsCascade.final)
    if (fuzzyFinder === undefined) {
        // Happens even when `"default": true` is defined in
        // settings.schema.json.
        fuzzyFinder = true
    }

    useEffect(() => {
        if (!isRepositoryRelatedPage && isFuzzyFinderVisible) {
            setIsFuzzyFinderVisible(false)
        }
    }, [isRepositoryRelatedPage, isFuzzyFinderVisible])

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
    const [enableContrastCompliantSyntaxHighlighting] = useFeatureFlag('contrast-compliant-syntax-highlighting')
    const [coreWorkflowImprovementsEnabled] = useCoreWorkflowImprovementsEnabled()

    const breadcrumbProps = useBreadcrumbs()

    useScrollToLocationHash(props.location)

    const showHelpShortcut = useKeyboardShortcut('keyboardShortcutsHelp')
    const [keyboardShortcutsHelpOpen, setKeyboardShortcutsHelpOpen] = useState(false)
    const showKeyboardShortcutsHelp = useCallback(() => setKeyboardShortcutsHelpOpen(true), [])
    const hideKeyboardShortcutsHelp = useCallback(() => setKeyboardShortcutsHelpOpen(false), [])

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
        isMacPlatform: isMacPlatform(),
        onHandleFuzzyFinder: setIsFuzzyFinderVisible,
    }

    return (
        <div
            className={classNames(
                styles.layout,
                enableContrastCompliantSyntaxHighlighting && CONTRAST_COMPLIANT_CLASSNAME,
                coreWorkflowImprovementsEnabled && 'core-workflow-improvements-enabled'
            )}
        >
            {showHelpShortcut?.keybindings.map((keybinding, index) => (
                <Shortcut key={index} {...keybinding} onMatch={showKeyboardShortcutsHelp} />
            ))}
            <KeyboardShortcutsHelp isOpen={keyboardShortcutsHelpOpen} onDismiss={hideKeyboardShortcutsHelp} />
            <GlobalAlerts
                authenticatedUser={props.authenticatedUser}
                settingsCascade={props.settingsCascade}
                isSourcegraphDotCom={props.isSourcegraphDotCom}
            />
            {!isSiteInit && <SurveyToast authenticatedUser={props.authenticatedUser} />}
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
                    isRepositoryRelatedPage={isRepositoryRelatedPage}
                    showKeyboardShortcutsHelp={showKeyboardShortcutsHelp}
                    onHandleFuzzyFinder={setIsFuzzyFinderVisible}
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
                    <Panel
                        className={styles.panel}
                        position="bottom"
                        defaultSize={350}
                        storageKey="panel-size"
                        ariaLabel="References panel"
                    >
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
            {props.extensionsController !== null ? (
                <GlobalDebug {...props} extensionsController={props.extensionsController} />
            ) : null}
            {(isSearchNotebookListPage || (isSearchRelatedPage && !isSearchHomepage)) && (
                <NotepadContainer onCreateNotebook={props.onCreateNotebookFromNotepad} />
            )}
            {fuzzyFinderShortcut?.keybindings.map((keybinding, index) => (
                <Shortcut
                    key={index}
                    {...keybinding}
                    onMatch={() => {
                        setIsFuzzyFinderVisible(true)
                        setRetainFuzzyFinderCache(true)
                        const input = document.querySelector<HTMLInputElement>('#fuzzy-modal-input')
                        input?.focus()
                        input?.select()
                    }}
                />
            ))}
            {isRepositoryRelatedPage && retainFuzzyFinderCache && fuzzyFinder && (
                <FuzzyFinder
                    setIsVisible={bool => setIsFuzzyFinderVisible(bool)}
                    isVisible={isFuzzyFinderVisible}
                    telemetryService={props.telemetryService}
                    location={props.location}
                    setCacheRetention={bool => setRetainFuzzyFinderCache(bool)}
                />
            )}
        </div>
    )
}
