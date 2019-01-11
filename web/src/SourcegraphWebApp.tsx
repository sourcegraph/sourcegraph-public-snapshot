import { ShortcutProvider } from '@slimsag/react-shortcuts'
import ServerIcon from 'mdi-react/ServerIcon'
import * as React from 'react'
import { Route } from 'react-router'
import { BrowserRouter } from 'react-router-dom'
import { combineLatest, from, Subscription } from 'rxjs'
import { startWith } from 'rxjs/operators'
import { setLinkComponent } from '../../shared/src/components/Link'
import {
    createController as createExtensionsController,
    ExtensionsControllerProps,
} from '../../shared/src/extensions/controller'
import * as GQL from '../../shared/src/graphql/schema'
import { Notifications } from '../../shared/src/notifications/Notifications'
import { PlatformContextProps } from '../../shared/src/platform/context'
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeProps } from '../../shared/src/settings/settings'
import { TelemetryContext } from '../../shared/src/telemetry/telemetryContext'
import { isErrorLike } from '../../shared/src/util/errors'
import { authenticatedUser } from './auth'
import { ErrorBoundary } from './components/ErrorBoundary'
import { FeedbackText } from './components/FeedbackText'
import { HeroPage } from './components/HeroPage'
import { RouterLinkOrAnchor } from './components/RouterLinkOrAnchor'
import { Tooltip } from './components/tooltip/Tooltip'
import { ExploreSectionDescriptor } from './explore/ExploreArea'
import { ExtensionAreaRoute } from './extensions/extension/ExtensionArea'
import { ExtensionAreaHeaderNavItem } from './extensions/extension/ExtensionAreaHeader'
import { ExtensionsAreaRoute } from './extensions/ExtensionsArea'
import { ExtensionsAreaHeaderActionButton } from './extensions/ExtensionsAreaHeader'
import { KeybindingsProps } from './keybindings'
import { Layout, LayoutProps } from './Layout'
import { updateUserSessionStores } from './marketing/util'
import { createPlatformContext } from './platform/context'
import { RepoHeaderActionButton } from './repo/RepoHeader'
import { RepoRevContainerRoute } from './repo/RepoRevContainer'
import { LayoutRouteProps } from './routes'
import { SiteAdminAreaRoute } from './site-admin/SiteAdminArea'
import { SiteAdminSideBarGroups } from './site-admin/SiteAdminSidebar'
import { eventLogger } from './tracking/eventLogger'
import { UserAccountAreaRoute } from './user/account/UserAccountArea'
import { UserAccountSidebarItems } from './user/account/UserAccountSidebar'
import { UserAreaRoute } from './user/area/UserArea'
import { UserAreaHeaderNavItem } from './user/area/UserAreaHeader'

export interface SourcegraphWebAppProps extends KeybindingsProps {
    exploreSections: ReadonlyArray<ExploreSectionDescriptor>
    extensionAreaRoutes: ReadonlyArray<ExtensionAreaRoute>
    extensionAreaHeaderNavItems: ReadonlyArray<ExtensionAreaHeaderNavItem>
    extensionsAreaRoutes: ReadonlyArray<ExtensionsAreaRoute>
    extensionsAreaHeaderActionButtons: ReadonlyArray<ExtensionsAreaHeaderActionButton>
    siteAdminAreaRoutes: ReadonlyArray<SiteAdminAreaRoute>
    siteAdminSideBarGroups: SiteAdminSideBarGroups
    siteAdminOverviewComponents: ReadonlyArray<React.ComponentType>
    userAreaHeaderNavItems: ReadonlyArray<UserAreaHeaderNavItem>
    userAreaRoutes: ReadonlyArray<UserAreaRoute>
    userAccountSideBarItems: UserAccountSidebarItems
    userAccountAreaRoutes: ReadonlyArray<UserAccountAreaRoute>
    repoRevContainerRoutes: ReadonlyArray<RepoRevContainerRoute>
    repoHeaderActionButtons: ReadonlyArray<RepoHeaderActionButton>
    routes: ReadonlyArray<LayoutRouteProps>
}

interface SourcegraphWebAppState extends PlatformContextProps, SettingsCascadeProps, ExtensionsControllerProps {
    error?: Error

    /** The currently authenticated user (or null if the viewer is anonymous). */
    authenticatedUser?: GQL.IUser | null

    viewerSubject: LayoutProps['viewerSubject']

    /**
     * Whether the light theme is enabled or not
     */
    isLightTheme: boolean

    /**
     * Whether the user is on MainPage and therefore not logged in
     */
    isMainPage: boolean

    /**
     * The current search query in the navbar.
     */
    navbarSearchQuery: string
}

const LIGHT_THEME_LOCAL_STORAGE_KEY = 'light-theme'

/** A fallback settings subject that can be constructed synchronously at initialization time. */
const SITE_SUBJECT_NO_ADMIN: Pick<GQL.ISettingsSubject, 'id' | 'viewerCanAdminister'> = {
    id: window.context.siteGQLID,
    viewerCanAdminister: false,
}

setLinkComponent(RouterLinkOrAnchor)

/**
 * The root component
 */
export class SourcegraphWebApp extends React.Component<SourcegraphWebAppProps, SourcegraphWebAppState> {
    constructor(props: SourcegraphWebAppProps) {
        super(props)
        const platformContext = createPlatformContext()
        this.state = {
            isLightTheme: localStorage.getItem(LIGHT_THEME_LOCAL_STORAGE_KEY) !== 'false',
            navbarSearchQuery: '',
            platformContext,
            extensionsController: createExtensionsController(platformContext),
            settingsCascade: EMPTY_SETTINGS_CASCADE,
            viewerSubject: SITE_SUBJECT_NO_ADMIN,
            isMainPage: false,
        }
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        updateUserSessionStores()

        document.body.classList.add('theme')
        this.subscriptions.add(
            authenticatedUser.subscribe(
                authenticatedUser => this.setState({ authenticatedUser }),
                () => this.setState({ authenticatedUser: null })
            )
        )

        this.subscriptions.add(
            combineLatest(from(this.state.platformContext.settings), authenticatedUser.pipe(startWith(null))).subscribe(
                ([cascade, authenticatedUser]) => {
                    this.setState(() => {
                        if (authenticatedUser) {
                            return { viewerSubject: authenticatedUser }
                        } else if (
                            cascade &&
                            !isErrorLike(cascade) &&
                            cascade.subjects &&
                            cascade.subjects.length > 0
                        ) {
                            return { viewerSubject: cascade.subjects[0].subject }
                        } else {
                            return { viewerSubject: SITE_SUBJECT_NO_ADMIN }
                        }
                    })
                }
            )
        )

        this.subscriptions.add(this.state.extensionsController)

        this.subscriptions.add(
            from(this.state.platformContext.settings).subscribe(settingsCascade => this.setState({ settingsCascade }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
        document.body.classList.remove('theme')
        document.body.classList.remove('theme-light')
        document.body.classList.remove('theme-dark')
    }

    public componentDidUpdate(): void {
        // Always show MainPage in dark theme look
        if (this.state.isMainPage && this.state.isLightTheme) {
            document.body.classList.remove('theme-light')
            document.body.classList.add('theme-dark')
        } else {
            localStorage.setItem(LIGHT_THEME_LOCAL_STORAGE_KEY, this.state.isLightTheme + '')
            document.body.classList.toggle('theme-light', this.state.isLightTheme)
            document.body.classList.toggle('theme-dark', !this.state.isLightTheme)
        }
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
                        {subtitle && <hr />}
                        <pre>{errorMessage}</pre>
                    </div>
                )
            } else {
                subtitle = <div className="app__error">{subtitle}</div>
            }
            return <HeroPage icon={ServerIcon} title={`${statusCode}: ${statusText}`} subtitle={subtitle} />
        }

        const { authenticatedUser } = this.state
        if (authenticatedUser === undefined) {
            return null
        }

        const { children, ...props } = this.props

        return (
            <ErrorBoundary location={null}>
                <ShortcutProvider>
                    <TelemetryContext.Provider value={eventLogger}>
                        <BrowserRouter key={0}>
                            <Route
                                path="/"
                                // tslint:disable-next-line:jsx-no-lambda RouteProps.render is an exception
                                render={routeComponentProps => (
                                    <Layout
                                        {...props}
                                        {...routeComponentProps}
                                        authenticatedUser={authenticatedUser}
                                        viewerSubject={this.state.viewerSubject}
                                        settingsCascade={this.state.settingsCascade}
                                        // Theme
                                        isLightTheme={this.state.isLightTheme}
                                        onThemeChange={this.onThemeChange}
                                        isMainPage={this.state.isMainPage}
                                        onMainPage={this.onMainPage}
                                        // Search query
                                        navbarSearchQuery={this.state.navbarSearchQuery}
                                        onNavbarQueryChange={this.onNavbarQueryChange}
                                        // Extensions
                                        platformContext={this.state.platformContext}
                                        extensionsController={this.state.extensionsController}
                                    />
                                )}
                            />
                        </BrowserRouter>
                        <Tooltip key={1} />
                        <Notifications key={2} extensionsController={this.state.extensionsController} />
                    </TelemetryContext.Provider>
                </ShortcutProvider>
            </ErrorBoundary>
        )
    }

    private onThemeChange = () => {
        this.setState(state => ({ isLightTheme: !state.isLightTheme }))
    }

    private onMainPage = (mainPage: boolean) => {
        this.setState(() => ({ isMainPage: mainPage }))
    }

    private onNavbarQueryChange = (navbarSearchQuery: string) => {
        this.setState({ navbarSearchQuery })
    }
}
