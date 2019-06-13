import { ShortcutProvider } from '@slimsag/react-shortcuts'
import 'focus-visible'
import ServerIcon from 'mdi-react/ServerIcon'
import * as React from 'react'
import { hot } from 'react-hot-loader/root'
import { Route } from 'react-router'
import { BrowserRouter } from 'react-router-dom'
import { combineLatest, from, fromEventPattern, Subscription } from 'rxjs'
import { startWith } from 'rxjs/operators'
import { setLinkComponent } from '../../shared/src/components/Link'
import {
    Controller as ExtensionsController,
    createController as createExtensionsController,
} from '../../shared/src/extensions/controller'
import * as GQL from '../../shared/src/graphql/schema'
import { Notifications } from '../../shared/src/notifications/Notifications'
import { PlatformContext } from '../../shared/src/platform/context'
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeProps } from '../../shared/src/settings/settings'
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
import { fetchHighlightedFileLines } from './repo/backend'
import { RepoContainerRoute } from './repo/RepoContainer'
import { RepoHeaderActionButton } from './repo/RepoHeader'
import { RepoRevContainerRoute } from './repo/RepoRevContainer'
import { LayoutRouteProps } from './routes'
import { search } from './search/backend'
import { SiteAdminAreaRoute } from './site-admin/SiteAdminArea'
import { SiteAdminSideBarGroups } from './site-admin/SiteAdminSidebar'
import { Theme, ThemePreference } from './theme'
import { eventLogger } from './tracking/eventLogger'
import { withActivation } from './tracking/withActivation'
import { UserAreaRoute } from './user/area/UserArea'
import { UserAreaHeaderNavItem } from './user/area/UserAreaHeader'
import { UserSettingsAreaRoute } from './user/settings/UserSettingsArea'
import { UserSettingsSidebarItems } from './user/settings/UserSettingsSidebar'

import './util/empty.css'

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
    userSettingsSideBarItems: UserSettingsSidebarItems
    userSettingsAreaRoutes: ReadonlyArray<UserSettingsAreaRoute>
    repoContainerRoutes: ReadonlyArray<RepoContainerRoute>
    repoRevContainerRoutes: ReadonlyArray<RepoRevContainerRoute>
    repoHeaderActionButtons: ReadonlyArray<RepoHeaderActionButton>
    routes: ReadonlyArray<LayoutRouteProps>
}

interface SourcegraphWebAppState extends SettingsCascadeProps {
    error?: Error

    /** The currently authenticated user (or null if the viewer is anonymous). */
    authenticatedUser?: GQL.IUser | null

    viewerSubject: LayoutProps['viewerSubject']

    /** The user's preference for the theme (light, dark or following system theme) */
    themePreference: ThemePreference

    /**
     * Whether the OS uses light theme, synced from a media query.
     * If the browser/OS does not this, will default to true.
     */
    systemIsLightTheme: boolean

    /**
     * The current search query in the navbar.
     */
    navbarSearchQuery: string
}

const LIGHT_THEME_LOCAL_STORAGE_KEY = 'light-theme'
/** Reads the stored theme preference from localStorage */
const readStoredThemePreference = (): ThemePreference => {
    const value = localStorage.getItem(LIGHT_THEME_LOCAL_STORAGE_KEY)
    // Handle both old and new preference values
    switch (value) {
        case 'true':
        case 'light':
            return ThemePreference.Light
        case 'false':
        case 'dark':
            return ThemePreference.Dark
        default:
            return ThemePreference.System
    }
}

/** A fallback settings subject that can be constructed synchronously at initialization time. */
const SITE_SUBJECT_NO_ADMIN: Pick<GQL.ISettingsSubject, 'id' | 'viewerCanAdminister'> = {
    id: window.context.siteGQLID,
    viewerCanAdminister: false,
}

setLinkComponent(RouterLinkOrAnchor)

const LayoutWithActivation = window.context.sourcegraphDotComMode ? Layout : withActivation(Layout)

/**
 * The root component.
 *
 * This is the non-hot-reload component. It is wrapped in `hot(...)` below to make it
 * hot-reloadable in development.
 */
class ColdSourcegraphWebApp extends React.Component<SourcegraphWebAppProps, SourcegraphWebAppState> {
    private readonly subscriptions = new Subscription()
    private readonly darkThemeMediaList = window.matchMedia('(prefers-color-scheme: dark)')
    private readonly highContrastThemeMediaList = window.matchMedia('(-ms-high-contrast: active)') // TODO!(sqs): implement
    private readonly platformContext: PlatformContext = createPlatformContext()
    private readonly extensionsController: ExtensionsController = createExtensionsController(this.platformContext)

    constructor(props: SourcegraphWebAppProps) {
        super(props)
        this.subscriptions.add(this.extensionsController)
        this.state = {
            themePreference: readStoredThemePreference(),
            systemIsLightTheme: !this.darkThemeMediaList.matches,
            navbarSearchQuery: '',
            settingsCascade: EMPTY_SETTINGS_CASCADE,
            viewerSubject: SITE_SUBJECT_NO_ADMIN,
        }
    }

    /** Returns the active theme. */
    private activeTheme(): Theme {
        if (this.state.themePreference === ThemePreference.System) {
            return this.state.systemIsLightTheme ? ThemePreference.Light : ThemePreference.Dark
        }
        return this.state.themePreference
    }

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
            combineLatest([from(this.platformContext.settings), authenticatedUser.pipe(startWith(null))]).subscribe(
                ([cascade, authenticatedUser]) => {
                    this.setState(() => {
                        if (authenticatedUser) {
                            return { viewerSubject: authenticatedUser }
                        }
                        if (cascade && !isErrorLike(cascade) && cascade.subjects && cascade.subjects.length > 0) {
                            return { viewerSubject: cascade.subjects[0].subject }
                        }
                        return { viewerSubject: SITE_SUBJECT_NO_ADMIN }
                    })
                }
            )
        )

        this.subscriptions.add(
            from(this.platformContext.settings).subscribe(settingsCascade => this.setState({ settingsCascade }))
        )

        // React to OS theme change
        this.subscriptions.add(
            fromEventPattern<MediaQueryListEvent>(
                // Need to use addListener() because addEventListener() is not supported yet in Safari
                handler => this.darkThemeMediaList.addListener(handler),
                handler => this.darkThemeMediaList.removeListener(handler)
            ).subscribe(event => {
                this.setState({ systemIsLightTheme: !event.matches })
            })
        )
    }

    private setThemeClass(theme: Theme | null): void {
        for (const className of document.body.classList) {
            if (className.startsWith('theme-')) {
                document.body.classList.remove(className)
            }
        }
        if (theme) {
            document.body.classList.add(`theme-${theme}`)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
        document.body.classList.remove('theme')
        this.setThemeClass(null)
    }

    public componentDidUpdate(): void {
        localStorage.setItem(LIGHT_THEME_LOCAL_STORAGE_KEY, this.state.themePreference)
        this.setThemeClass(this.activeTheme())
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

        const { authenticatedUser } = this.state
        if (authenticatedUser === undefined) {
            return null
        }

        const { children, ...props } = this.props

        return (
            <ErrorBoundary location={null}>
                <ShortcutProvider>
                    <BrowserRouter key={0}>
                        <Route
                            path="/"
                            // tslint:disable-next-line:jsx-no-lambda RouteProps.render is an exception
                            render={routeComponentProps => (
                                <LayoutWithActivation
                                    {...props}
                                    {...routeComponentProps}
                                    authenticatedUser={authenticatedUser}
                                    viewerSubject={this.state.viewerSubject}
                                    settingsCascade={this.state.settingsCascade}
                                    // Theme
                                    isLightTheme={this.activeTheme() === ThemePreference.Light}
                                    themePreference={this.state.themePreference}
                                    onThemePreferenceChange={this.onThemePreferenceChange}
                                    // Search query
                                    navbarSearchQuery={this.state.navbarSearchQuery}
                                    onNavbarQueryChange={this.onNavbarQueryChange}
                                    fetchHighlightedFileLines={fetchHighlightedFileLines}
                                    searchRequest={search}
                                    // Extensions
                                    platformContext={this.platformContext}
                                    extensionsController={this.extensionsController}
                                    telemetryService={eventLogger}
                                    isSourcegraphDotCom={window.context.sourcegraphDotComMode}
                                />
                            )}
                        />
                    </BrowserRouter>
                    <Tooltip key={1} />
                    <Notifications key={2} extensionsController={this.extensionsController} />
                </ShortcutProvider>
            </ErrorBoundary>
        )
    }

    private onThemePreferenceChange = (themePreference: ThemePreference) => {
        this.setState({ themePreference })
    }

    private onNavbarQueryChange = (navbarSearchQuery: string) => {
        this.setState({ navbarSearchQuery })
    }
}

export const SourcegraphWebApp = hot(ColdSourcegraphWebApp)
