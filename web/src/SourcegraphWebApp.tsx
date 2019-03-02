import { ShortcutProvider } from '@slimsag/react-shortcuts'
import H from 'history'
import ServerIcon from 'mdi-react/ServerIcon'
import * as React from 'react'
import { Route } from 'react-router'
import { BrowserRouter } from 'react-router-dom'
import { combineLatest, from, fromEventPattern, Observable, Subscription } from 'rxjs'
import { first, map, startWith } from 'rxjs/operators'
import { ActivationCompleted, ActivationStep } from '../../shared/src/components/activation/Activation'
import { setLinkComponent } from '../../shared/src/components/Link'
import {
    createController as createExtensionsController,
    ExtensionsControllerProps,
} from '../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../shared/src/graphql/graphql'
import * as GQL from '../../shared/src/graphql/schema'
import { Notifications } from '../../shared/src/notifications/Notifications'
import { PlatformContextProps } from '../../shared/src/platform/context'
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeProps } from '../../shared/src/settings/settings'
import { TelemetryContext } from '../../shared/src/telemetry/telemetryContext'
import { isErrorLike } from '../../shared/src/util/errors'
import { authenticatedUser } from './auth'
import { queryGraphQL } from './backend/graphql'
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
import { ThemePreference } from './theme'
import { eventLogger } from './tracking/eventLogger'
import { logUserEvent } from './user/account/backend'
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

    /**
     * Defines the activation steps that a user must complete and where
     * this is fetched from.
     */
    activation?: ActivationParams

    /**
     * Specifies which activation steps have been completed.
     */
    activationCompleted?: ActivationCompleted
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

/**
 * The root component
 */
export class SourcegraphWebApp extends React.Component<SourcegraphWebAppProps, SourcegraphWebAppState> {
    private subscriptions = new Subscription()
    private darkThemeMediaList = window.matchMedia('(prefers-color-scheme: dark)')

    constructor(props: SourcegraphWebAppProps) {
        super(props)
        const platformContext = createPlatformContext()
        this.state = {
            themePreference: readStoredThemePreference(),
            systemIsLightTheme: !this.darkThemeMediaList.matches,
            navbarSearchQuery: '',
            platformContext,
            extensionsController: createExtensionsController(platformContext),
            settingsCascade: EMPTY_SETTINGS_CASCADE,
            viewerSubject: SITE_SUBJECT_NO_ADMIN,
        }
    }

    /** Returns whether Sourcegraph should be in light theme */
    private isLightTheme(): boolean {
        return this.state.themePreference === 'system'
            ? this.state.systemIsLightTheme
            : this.state.themePreference === 'light'
    }

    public componentDidMount(): void {
        updateUserSessionStores()

        document.body.classList.add('theme')
        this.subscriptions.add(
            authenticatedUser.subscribe(
                authenticatedUser => {
                    this.setState({
                        authenticatedUser,
                        activation: this.getActivationState(authenticatedUser),
                    })
                },
                () => this.setState({ authenticatedUser: null, activation: undefined, activationCompleted: undefined })
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

    private getActivationState = (authenticatedUser: GQL.IUser | null): ActivationParams | undefined => {
        if (window.context.sourcegraphDotComMode || !authenticatedUser) {
            return
        }
        const fetcher = fetchActivationStatus(authenticatedUser.siteAdmin)
        const firstFetched = new Promise<void>(resolve => {
            fetcher()
                .pipe(first())
                .subscribe(completed => {
                    this.setState({ activationCompleted: completed })
                    resolve()
                })
        })
        return {
            steps: getActivationSteps(authenticatedUser.siteAdmin),
            fetcher,
            firstFetched,
        }
    }

    private refetchActivation = () => {
        if (!this.state.activation) {
            return
        }
        this.state.activation
            .fetcher()
            .pipe(first())
            .subscribe(c => this.setState({ activationCompleted: c }))
    }

    private updateActivation = (update: ActivationCompleted) => {
        if (!this.state.activation) {
            return
        }
        this.state.activation.firstFetched.then(() => {
            if (!this.state.activation) {
                return
            }

            // Send update to server for events that don't themselves trigger
            // an update.
            if (update.FoundReferences) {
                logUserEvent(GQL.UserEvent.CODEINTELREFS)
            }

            const newVal: ActivationCompleted = {}
            Object.assign(newVal, this.state.activationCompleted)
            for (const step of this.state.activation.steps) {
                if (update[step.id] !== undefined) {
                    newVal[step.id] = update[step.id]
                }
            }
            this.setState({ activationCompleted: newVal })
        })
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
        document.body.classList.remove('theme')
        document.body.classList.remove('theme-light')
        document.body.classList.remove('theme-dark')
    }

    public componentDidUpdate(): void {
        localStorage.setItem(LIGHT_THEME_LOCAL_STORAGE_KEY, this.state.themePreference)
        document.body.classList.toggle('theme-light', this.isLightTheme())
        document.body.classList.toggle('theme-dark', !this.isLightTheme())
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
                                        isLightTheme={this.isLightTheme()}
                                        themePreference={this.state.themePreference}
                                        onThemePreferenceChange={this.onThemePreferenceChange}
                                        // Search query
                                        navbarSearchQuery={this.state.navbarSearchQuery}
                                        onNavbarQueryChange={this.onNavbarQueryChange}
                                        // Extensions
                                        platformContext={this.state.platformContext}
                                        extensionsController={this.state.extensionsController}
                                        // Activation
                                        activation={
                                            this.state.activation &&
                                            this.state.activationCompleted && {
                                                steps: this.state.activation.steps,
                                                update: this.updateActivation,
                                                refetch: this.refetchActivation,
                                                completed: this.state.activationCompleted,
                                            }
                                        }
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

    private onThemePreferenceChange = (themePreference: ThemePreference) => {
        this.setState({ themePreference })
    }

    private onNavbarQueryChange = (navbarSearchQuery: string) => {
        this.setState({ navbarSearchQuery })
    }
}

interface ActivationParams {
    fetcher: () => Observable<ActivationCompleted>
    firstFetched: Promise<void>
    steps: ActivationStep[]
}

const fetchActivationStatus = (isSiteAdmin: boolean) => () =>
    queryGraphQL(
        isSiteAdmin
            ? gql`
                  query {
                      externalServices {
                          totalCount
                      }
                      repositories(enabled: true) {
                          totalCount
                      }
                      viewerSettings {
                          final
                      }
                      users {
                          totalCount
                      }
                      currentUser {
                          usageStatistics {
                              searchQueries
                              findReferencesActions
                          }
                      }
                  }
              `
            : gql`
                  query {
                      currentUser {
                          usageStatistics {
                              searchQueries
                              findReferencesActions
                          }
                      }
                  }
              `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            const authProviders = window.context.authProviders
            const completed: ActivationCompleted = {
                DidSearch: !!data.currentUser && data.currentUser.usageStatistics.searchQueries > 0,
                FoundReferences: !!data.currentUser && data.currentUser.usageStatistics.findReferencesActions > 0,
            }
            if (isSiteAdmin) {
                completed.ConnectedCodeHost = data.externalServices && data.externalServices.totalCount > 0
                completed.EnabledRepository =
                    data.repositories && data.repositories.totalCount !== null && data.repositories.totalCount > 0
                if (authProviders) {
                    completed.EnabledSharing =
                        data.users.totalCount > 1 || authProviders.filter(p => !p.isBuiltin).length > 0
                }
            }
            return completed
        })
    )

const fetchReferencesLink = (): Observable<string | null> =>
    queryGraphQL(gql`
        query {
            repositories(enabled: true, cloned: true, first: 100, indexed: true) {
                nodes {
                    url
                    gitRefs {
                        totalCount
                    }
                }
            }
        }
    `).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.repositories.nodes) {
                return null
            }
            const repositoryURLs = data.repositories.nodes
                .filter(r => r.gitRefs && r.gitRefs.totalCount > 0)
                .sort((r1, r2) => r2.gitRefs!.totalCount! - r1.gitRefs!.totalCount)
                .map(r => r.url)
            if (repositoryURLs.length === 0) {
                return null
            }
            return repositoryURLs[0]
        })
    )

const getActivationSteps = (isSiteAdmin: boolean): ActivationStep[] => {
    const sources: (ActivationStep & { siteAdminOnly?: boolean })[] = [
        {
            id: 'ConnectedCodeHost',
            title: 'Connect your code host',
            detail: 'Configure Sourcegraph to talk to your code host and fetch a list of your repositories.',
            link: { to: '/site-admin/external-services' },
            siteAdminOnly: true,
        },
        {
            id: 'EnabledRepository',
            title: 'Enable repositories',
            detail: 'Select which repositories Sourcegraph should pull and index from your code host(s).',
            link: { to: '/site-admin/repositories' },
            siteAdminOnly: true,
        },
        {
            id: 'DidSearch',
            title: 'Search your code',
            detail: 'Perform a search query on your code.',
            link: { to: '/search' },
        },
        {
            id: 'FoundReferences',
            title: 'Find some references',
            detail:
                'To find references of a token, navigate to a code file in one of your repositories, hover over a token to activate the tooltip, and then click "Find references".',
            onClick: (event: React.MouseEvent<HTMLElement>, history: H.History) =>
                fetchReferencesLink()
                    .pipe(first())
                    .subscribe(r => {
                        if (r) {
                            history.push(r)
                        } else {
                            alert('Must add repositories before finding references')
                        }
                    }),
        },
        {
            id: 'EnabledSharing',
            title: 'Configure SSO or share with teammates',
            detail: 'Configure a single-sign on (SSO) provider or have at least one other teammate sign up.',
            link: { to: 'https://docs.sourcegraph.com/admin/auth', target: '_blank' },
            siteAdminOnly: true,
        },
    ]
    return sources.filter(e => true || !e.siteAdminOnly).map(({ siteAdminOnly, ...step }) => step)
}
