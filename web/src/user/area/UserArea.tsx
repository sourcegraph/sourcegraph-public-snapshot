import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { matchPath, Route, RouteComponentProps, Switch } from 'react-router'
import { combineLatest, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap, filter } from 'rxjs/operators'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { gql, dataOrThrowErrors } from '../../../../shared/src/graphql/graphql'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { isErrorLike, asError } from '../../../../shared/src/util/errors'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { NamespaceProps } from '../../namespaces'
import { ThemeProps } from '../../../../shared/src/theme'
import { RouteDescriptor } from '../../util/contributions'
import { UserSettingsAreaRoute } from '../settings/UserSettingsArea'
import { UserSettingsSidebarItems } from '../settings/UserSettingsSidebar'
import { UserAreaHeader, UserAreaHeaderNavItem } from './UserAreaHeader'
import { PatternTypeProps, OnboardingTourProps } from '../../search'
import { ErrorMessage } from '../../components/alerts'
import { isDefined } from '../../../../shared/src/util/types'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { AuthenticatedUser } from '../../auth'
import { UserAreaUserFields } from '../../graphql-operations'
import { BreadcrumbsProps, BreadcrumbSetters } from '../../components/Breadcrumbs'
import { queryGraphQL } from '../../backend/graphql'

const fetchUser = (args: { username: string; siteAdmin: boolean }): Observable<UserAreaUserFields> =>
    queryGraphQL(
        gql`
            query User($username: String!, $siteAdmin: Boolean!) {
                user(username: $username) {
                    ...UserAreaUserFields
                }
            }

            fragment UserAreaUserFields on User {
                __typename
                id
                username
                displayName
                url
                settingsURL
                avatarURL
                viewerCanAdminister
                siteAdmin @include(if: $siteAdmin)
                builtinAuth
                createdAt
                emails @include(if: $siteAdmin) {
                    email
                    verified
                }
                organizations {
                    nodes {
                        id
                        displayName
                        name
                    }
                }
                permissionsInfo @include(if: $siteAdmin) {
                    syncedAt
                    updatedAt
                }
                tags @include(if: $siteAdmin)
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.user) {
                throw new Error(`User not found: ${JSON.stringify(args.username)}`)
            }
            return data.user as UserAreaUserFields
        })
    )

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested user page was not found." />
)

export interface UserAreaRoute extends RouteDescriptor<UserAreaRouteContext> {}

interface UserAreaProps
    extends RouteComponentProps<{ username: string }>,
        ExtensionsControllerProps,
        PlatformContextProps,
        SettingsCascadeProps,
        ThemeProps,
        TelemetryProps,
        ActivationProps,
        OnboardingTourProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        Omit<PatternTypeProps, 'setPatternType'> {
    userAreaRoutes: readonly UserAreaRoute[]
    userAreaHeaderNavItems: readonly UserAreaHeaderNavItem[]
    userSettingsSideBarItems: UserSettingsSidebarItems
    userSettingsAreaRoutes: readonly UserSettingsAreaRoute[]

    /**
     * The currently authenticated user, NOT the user whose username is specified in the URL's "username" route
     * parameter.
     */
    authenticatedUser: AuthenticatedUser | null

    isSourcegraphDotCom: boolean
}

interface UserAreaState extends BreadcrumbSetters {
    /**
     * The fetched user (who is the subject of the page), or an error if an error occurred; undefined while
     * loading.
     */
    userOrError?: UserAreaUserFields | Error
}

/**
 * Properties passed to all page components in the user area.
 */
export interface UserAreaRouteContext
    extends ExtensionsControllerProps,
        PlatformContextProps,
        SettingsCascadeProps,
        ThemeProps,
        TelemetryProps,
        ActivationProps,
        NamespaceProps,
        OnboardingTourProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        Omit<PatternTypeProps, 'setPatternType'> {
    /** The user area main URL. */
    url: string

    /**
     * The user who is the subject of the page.
     */
    user: UserAreaUserFields

    /** Called when the user is updated and must be reloaded. */
    onDidUpdateUser: () => void

    /**
     * The currently authenticated user, NOT (necessarily) the user who is the subject of the page.
     *
     * For example, if Alice is viewing a user area page about Bob, then the authenticatedUser is Alice and the
     * user is Bob.
     */
    authenticatedUser: AuthenticatedUser | null
    userSettingsSideBarItems: UserSettingsSidebarItems
    userSettingsAreaRoutes: readonly UserSettingsAreaRoute[]

    isSourcegraphDotCom: boolean
}

/**
 * A user's public profile area.
 */
export class UserArea extends React.Component<UserAreaProps, UserAreaState> {
    public state: UserAreaState

    private componentUpdates = new Subject<UserAreaProps>()
    private refreshRequests = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: UserAreaProps) {
        super(props)
        this.state = {
            setBreadcrumb: props.setBreadcrumb,
            useBreadcrumb: props.useBreadcrumb,
        }
    }

    public componentDidMount(): void {
        // Changes to the route-matched username.
        const usernameChanges = this.componentUpdates.pipe(
            map(props => props.match.params.username),
            distinctUntilChanged()
        )

        // Fetch user.
        this.subscriptions.add(
            combineLatest([usernameChanges, merge(this.refreshRequests.pipe(mapTo(false)), of(true))])
                .pipe(
                    switchMap(([username, forceRefresh]) => {
                        type PartialStateUpdate = Pick<UserAreaState, 'userOrError'>
                        return fetchUser({
                            username,
                            siteAdmin: !!this.props.authenticatedUser?.siteAdmin,
                        }).pipe(
                            filter(isDefined),
                            catchError(error => [asError(error)]),
                            map((userOrError): PartialStateUpdate => ({ userOrError })),

                            // Don't clear old user data while we reload, to avoid unmounting all components during
                            // loading.
                            startWith<PartialStateUpdate>(forceRefresh ? { userOrError: undefined } : {})
                        )
                    })
                )
                .subscribe(
                    stateUpdate => {
                        if (stateUpdate.userOrError && !isErrorLike(stateUpdate.userOrError)) {
                            const childBreadcrumbSetters = this.props.setBreadcrumb({
                                key: 'UserArea',
                                link: { to: stateUpdate.userOrError.url, label: stateUpdate.userOrError.username },
                            })
                            this.subscriptions.add(childBreadcrumbSetters)
                            this.setState({
                                ...stateUpdate,
                                setBreadcrumb: childBreadcrumbSetters.setBreadcrumb,
                                useBreadcrumb: childBreadcrumbSetters.useBreadcrumb,
                            })
                        } else {
                            this.setState(stateUpdate)
                        }
                    },
                    error => console.error(error)
                )
        )

        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.userOrError) {
            return null // loading
        }
        if (isErrorLike(this.state.userOrError)) {
            return (
                <HeroPage
                    icon={AlertCircleIcon}
                    title="Error"
                    subtitle={<ErrorMessage error={this.state.userOrError} history={this.props.history} />}
                />
            )
        }

        const context: UserAreaRouteContext = {
            authenticatedUser: this.props.authenticatedUser,
            extensionsController: this.props.extensionsController,
            isLightTheme: this.props.isLightTheme,
            isSourcegraphDotCom: this.props.isSourcegraphDotCom,
            patternType: this.props.patternType,
            platformContext: this.props.platformContext,
            settingsCascade: this.props.settingsCascade,
            showOnboardingTour: this.props.showOnboardingTour,
            telemetryService: this.props.telemetryService,
            userSettingsAreaRoutes: this.props.userSettingsAreaRoutes,
            userSettingsSideBarItems: this.props.userSettingsSideBarItems,
            activation: this.props.activation,
            url: this.props.match.url,
            user: this.state.userOrError,
            onDidUpdateUser: this.onDidUpdateUser,
            namespace: this.state.userOrError,
            breadcrumbs: this.props.breadcrumbs,
            useBreadcrumb: this.state.useBreadcrumb,
            setBreadcrumb: this.state.setBreadcrumb,
        }

        const routeMatch = this.props.userAreaRoutes.find(({ path, exact }) =>
            matchPath(this.props.location.pathname, { path: this.props.match.url + path, exact })
        )?.path

        // Hide header and use full-width container for campaigns pages.
        const isCampaigns = routeMatch === '/campaigns'
        const hideHeader = isCampaigns
        const containerClassName = isCampaigns ? '' : 'container'

        return (
            <div className="user-area w-100">
                {!hideHeader && (
                    <UserAreaHeader
                        {...this.props}
                        {...context}
                        navItems={this.props.userAreaHeaderNavItems}
                        className="border-bottom mt-4 mb-3"
                    />
                )}
                <div className={containerClassName}>
                    <ErrorBoundary location={this.props.location}>
                        <React.Suspense fallback={<LoadingSpinner className="icon-inline m-2" />}>
                            <Switch>
                                {this.props.userAreaRoutes.map(
                                    ({ path, exact, render, condition = () => true }) =>
                                        condition(context) && (
                                            <Route
                                                // eslint-disable-next-line react/jsx-no-bind
                                                render={routeComponentProps =>
                                                    render({ ...context, ...routeComponentProps })
                                                }
                                                path={this.props.match.url + path}
                                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                                exact={exact}
                                            />
                                        )
                                )}
                                <Route key="hardcoded-key" component={NotFoundPage} />
                            </Switch>
                        </React.Suspense>
                    </ErrorBoundary>
                </div>
            </div>
        )
    }

    private onDidUpdateUser = (): void => {
        this.refreshRequests.next()
    }
}
