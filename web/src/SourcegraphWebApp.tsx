import { ShortcutProvider } from '@slimsag/react-shortcuts'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ServerIcon from 'mdi-react/ServerIcon'
import * as React from 'react'
import { Route } from 'react-router'
import { BrowserRouter } from 'react-router-dom'
import { combineLatest, from, Subscription } from 'rxjs'
import { startWith } from 'rxjs/operators'
import { EMPTY_ENVIRONMENT as EXTENSIONS_EMPTY_ENVIRONMENT } from '../../shared/src/api/client/environment'
import { TextDocumentItem } from '../../shared/src/api/client/types/textDocument'
import { WorkspaceRoot } from '../../shared/src/api/protocol/plainTypes'
import {
    createController as createExtensionsController,
    ExtensionsControllerProps,
} from '../../shared/src/extensions/controller'
import { ConfiguredExtension } from '../../shared/src/extensions/extension'
import { viewerConfiguredExtensions } from '../../shared/src/extensions/helpers'
import * as GQL from '../../shared/src/graphql/schema'
import { Notifications } from '../../shared/src/notifications/Notifications'
import { PlatformContextProps } from '../../shared/src/platform/context'
import {
    ConfiguredSubject,
    isSettingsValid,
    SettingsCascadeOrError,
    SettingsCascadeProps,
} from '../../shared/src/settings/settings'
import { isErrorLike } from '../../shared/src/util/errors'
import { authenticatedUser } from './auth'
import { FeedbackText } from './components/FeedbackText'
import { HeroPage } from './components/HeroPage'
import { Tooltip } from './components/tooltip/Tooltip'
import { ExploreSectionDescriptor } from './explore/ExploreArea'
import { ExtensionsEnvironmentProps } from './extensions/environment/ExtensionsEnvironment'
import { ExtensionAreaRoute } from './extensions/extension/ExtensionArea'
import { ExtensionAreaHeaderNavItem } from './extensions/extension/ExtensionAreaHeader'
import { ExtensionsAreaRoute } from './extensions/ExtensionsArea'
import { ExtensionsAreaHeaderActionButton } from './extensions/ExtensionsAreaHeader'
import { createPlatformContext } from './extensions/ExtensionsClientCommonContext'
import { KeybindingsProps } from './keybindings'
import { Layout, LayoutProps } from './Layout'
import { updateUserSessionStores } from './marketing/util'
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

interface SourcegraphWebAppState
    extends SettingsCascadeProps,
        PlatformContextProps,
        ExtensionsEnvironmentProps,
        ExtensionsControllerProps {
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
            settingsCascade: { subjects: null, final: null },
            platformContext,
            extensionsEnvironment: {
                ...EXTENSIONS_EMPTY_ENVIRONMENT,
                context: {
                    'clientApplication.isSourcegraph': true,
                },
            },
            extensionsController: createExtensionsController(platformContext),
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
            combineLatest(
                from(this.state.platformContext.settingsCascade).pipe(startWith(null)),
                authenticatedUser.pipe(startWith(null))
            ).subscribe(([cascade, authenticatedUser]) => {
                this.setState(() => {
                    if (authenticatedUser) {
                        return { viewerSubject: authenticatedUser }
                    } else if (
                        cascade &&
                        !isErrorLike(cascade) &&
                        cascade.subjects &&
                        !isErrorLike(cascade.subjects) &&
                        cascade.subjects.length > 0
                    ) {
                        return { viewerSubject: cascade.subjects[0].subject }
                    } else {
                        return { viewerSubject: SITE_SUBJECT_NO_ADMIN }
                    }
                })
            })
        )

        this.subscriptions.add(this.state.extensionsController)

        this.subscriptions.add(
            this.state.platformContext.settingsCascade.subscribe(
                v => this.onSettingsCascadeChange(v),
                err => console.error(err)
            )
        )

        // Keep the Sourcegraph extensions controller's extensions up-to-date.
        //
        // TODO(sqs): handle loading and errors
        this.subscriptions.add(
            viewerConfiguredExtensions(this.state.platformContext).subscribe(
                extensions => this.onViewerConfiguredExtensionsChange(extensions),
                err => console.error(err)
            )
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
        if (this.state.error) {
            return <HeroPage icon={AlertCircleIcon} title={'Something happened'} subtitle={this.state.error.message} />
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
            <ShortcutProvider>
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
                                extensionsEnvironment={this.state.extensionsEnvironment}
                                extensionsOnRootsChange={this.extensionsOnRootsChange}
                                extensionsOnVisibleTextDocumentsChange={this.extensionsOnVisibleTextDocumentsChange}
                                extensionsController={this.state.extensionsController}
                            />
                        )}
                    />
                </BrowserRouter>
                <Tooltip key={1} />
                <Notifications key={2} extensionsController={this.state.extensionsController} />
            </ShortcutProvider>
        )
    }

    private onThemeChange = () => {
        this.setState(
            state => ({ isLightTheme: !state.isLightTheme }),
            () => {
                eventLogger.log(this.state.isLightTheme ? 'LightThemeClicked' : 'DarkThemeClicked')
            }
        )
    }

    private onMainPage = (mainPage: boolean) => {
        this.setState(state => ({ isMainPage: mainPage }))
    }

    private onNavbarQueryChange = (navbarSearchQuery: string) => {
        this.setState({ navbarSearchQuery })
    }

    private onSettingsCascadeChange(settingsCascade: SettingsCascadeOrError): void {
        this.setState(
            prevState => {
                const update: Pick<SourcegraphWebAppState, 'settingsCascade' | 'extensionsEnvironment'> = {
                    settingsCascade,
                    extensionsEnvironment: prevState.extensionsEnvironment,
                }
                if (isSettingsValid(settingsCascade)) {
                    // Only update Sourcegraph extensions environment configuration if the configuration was
                    // successfully parsed.
                    update.extensionsEnvironment = {
                        ...prevState.extensionsEnvironment,
                        configuration: {
                            subjects: settingsCascade.subjects.filter(
                                (subject): subject is ConfiguredSubject =>
                                    subject.settings !== null && !isErrorLike(subject.settings)
                            ),
                            final: settingsCascade.final,
                        },
                    }
                }
                return update
            },
            () => this.state.extensionsController.setEnvironment(this.state.extensionsEnvironment)
        )
    }

    private extensionsOnRootsChange = (roots: WorkspaceRoot[] | null): void => {
        this.setState(
            prevState => ({ extensionsEnvironment: { ...prevState.extensionsEnvironment, roots } }),
            () => this.state.extensionsController.setEnvironment(this.state.extensionsEnvironment)
        )
    }

    private onViewerConfiguredExtensionsChange(viewerConfiguredExtensions: ConfiguredExtension[]): void {
        this.setState(
            prevState => ({
                extensionsEnvironment: {
                    ...prevState.extensionsEnvironment,
                    extensions: viewerConfiguredExtensions,
                },
            }),
            () => this.state.extensionsController.setEnvironment(this.state.extensionsEnvironment)
        )
    }

    private extensionsOnVisibleTextDocumentsChange = (visibleTextDocuments: TextDocumentItem[] | null): void => {
        this.setState(
            prevState => ({ extensionsEnvironment: { ...prevState.extensionsEnvironment, visibleTextDocuments } }),
            () => this.state.extensionsController.setEnvironment(this.state.extensionsEnvironment)
        )
    }
}
