import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { combineLatest, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { createAggregateError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { NamespaceProps } from '../../namespaces'
import { ThemeProps } from '../../theme'
import { RouteDescriptor } from '../../util/contributions'
import { UserSettingsAreaRoute } from '../settings/UserSettingsArea'
import { UserSettingsSidebarItems } from '../settings/UserSettingsSidebar'
import { UserAreaHeader, UserAreaHeaderNavItem } from './UserAreaHeader'

const fetchUser = (args: { username: string }): Observable<GQL.IUser | null> =>
    queryGraphQL(
        gql`
            query User($username: String!) {
                user(username: $username) {
                    __typename
                    id
                    username
                    displayName
                    url
                    settingsURL
                    avatarURL
                    viewerCanAdminister
                    siteAdmin
                    createdAt
                    emails {
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
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.user) {
                throw createAggregateError(errors)
            }
            return data.user
        })
    )

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested user page was not found." />
)

export interface UserAreaRoute extends RouteDescriptor<UserAreaRouteContext> {}

interface UserAreaProps
    extends RouteComponentProps<{ username: string }>,
        ExtensionsControllerProps,
        PlatformContextProps,
        SettingsCascadeProps,
        ThemeProps,
        ActivationProps {
    userAreaRoutes: ReadonlyArray<UserAreaRoute>
    userAreaHeaderNavItems: ReadonlyArray<UserAreaHeaderNavItem>
    userSettingsSideBarItems: UserSettingsSidebarItems
    userSettingsAreaRoutes: ReadonlyArray<UserSettingsAreaRoute>

    /**
     * The currently authenticated user, NOT the user whose username is specified in the URL's "username" route
     * parameter.
     */
    authenticatedUser: GQL.IUser | null
}

interface UserAreaState {
    /**
     * The fetched user (who is the subject of the page), or an error if an error occurred; undefined while
     * loading.
     */
    userOrError?: GQL.IUser | ErrorLike
}

/**
 * Properties passed to all page components in the user area.
 */
export interface UserAreaRouteContext
    extends PlatformContextProps,
        SettingsCascadeProps,
        ThemeProps,
        ActivationProps,
        NamespaceProps,
        ExtensionsControllerProps {
    /** The user area main URL. */
    url: string

    /**
     * The user who is the subject of the page.
     */
    user: GQL.IUser

    /** Called when the user is updated and must be reloaded. */
    onDidUpdateUser: () => void

    /**
     * The currently authenticated user, NOT (necessarily) the user who is the subject of the page.
     *
     * For example, if Alice is viewing a user area page about Bob, then the authenticatedUser is Alice and the
     * user is Bob.
     */
    authenticatedUser: GQL.IUser | null
    userSettingsSideBarItems: UserSettingsSidebarItems
    userSettingsAreaRoutes: ReadonlyArray<UserSettingsAreaRoute>
}

/**
 * A user's public profile area.
 */
export class UserArea extends React.Component<UserAreaProps, UserAreaState> {
    public state: UserAreaState = {}

    private routeMatchChanges = new Subject<{ username: string }>()
    private refreshRequests = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Changes to the route-matched username.
        const usernameChanges = this.routeMatchChanges.pipe(
            map(({ username }) => username),
            distinctUntilChanged()
        )

        // Fetch user.
        this.subscriptions.add(
            combineLatest(usernameChanges, merge(this.refreshRequests.pipe(mapTo(false)), of(true)))
                .pipe(
                    switchMap(([username, forceRefresh]) => {
                        type PartialStateUpdate = Pick<UserAreaState, 'userOrError'>
                        return fetchUser({ username }).pipe(
                            catchError(error => [error]),
                            map((c): PartialStateUpdate => ({ userOrError: c })),

                            // Don't clear old user data while we reload, to avoid unmounting all components during
                            // loading.
                            startWith<PartialStateUpdate>(forceRefresh ? { userOrError: undefined } : {})
                        )
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        this.routeMatchChanges.next(this.props.match.params)
    }

    public componentWillReceiveProps(props: UserAreaProps): void {
        if (props.match.params !== this.props.match.params) {
            this.routeMatchChanges.next(props.match.params)
        }
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
                <HeroPage icon={AlertCircleIcon} title="Error" subtitle={upperFirst(this.state.userOrError.message)} />
            )
        }

        const context: UserAreaRouteContext = {
            url: this.props.match.url,
            user: this.state.userOrError,
            onDidUpdateUser: this.onDidUpdateUser,
            authenticatedUser: this.props.authenticatedUser,
            platformContext: this.props.platformContext,
            settingsCascade: this.props.settingsCascade,
            isLightTheme: this.props.isLightTheme,
            activation: this.props.activation,
            userSettingsAreaRoutes: this.props.userSettingsAreaRoutes,
            userSettingsSideBarItems: this.props.userSettingsSideBarItems,
            extensionsController: this.props.extensionsController,
            namespace: this.state.userOrError,
        }
        return (
            <div className="user-area w-100">
                <UserAreaHeader
                    {...this.props}
                    {...context}
                    navItems={this.props.userAreaHeaderNavItems}
                    className="border-bottom mt-4"
                />
                <div className="container mt-3">
                    <ErrorBoundary location={this.props.location}>
                        <React.Suspense fallback={<LoadingSpinner className="icon-inline m-2" />}>
                            <Switch>
                                {this.props.userAreaRoutes.map(
                                    ({ path, exact, render, condition = () => true }) =>
                                        condition(context) && (
                                            <Route
                                                path={this.props.match.url + path}
                                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                                exact={exact}
                                                // tslint:disable-next-line:jsx-no-lambda
                                                render={routeComponentProps =>
                                                    render({ ...context, ...routeComponentProps })
                                                }
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

    private onDidUpdateUser = () => this.refreshRequests.next()
}
