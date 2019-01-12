import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { combineLatest, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { createAggregateError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { RouteDescriptor } from '../../util/contributions'
import { UserAccountAreaRoute } from '../account/UserAccountArea'
import { UserAccountSidebarItems } from '../account/UserAccountSidebar'
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
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested user was not found." />
)

export interface UserAreaRoute extends RouteDescriptor<UserAreaRouteContext> {}

interface UserAreaProps extends RouteComponentProps<{ username: string }>, PlatformContextProps, SettingsCascadeProps {
    userAreaRoutes: ReadonlyArray<UserAreaRoute>
    userAreaHeaderNavItems: ReadonlyArray<UserAreaHeaderNavItem>
    userAccountSideBarItems: UserAccountSidebarItems
    userAccountAreaRoutes: ReadonlyArray<UserAccountAreaRoute>

    /**
     * The currently authenticated user, NOT the user whose username is specified in the URL's "username" route
     * parameter.
     */
    authenticatedUser: GQL.IUser | null

    isLightTheme: boolean
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
export interface UserAreaRouteContext extends PlatformContextProps, SettingsCascadeProps {
    /** The extension registry area main URL. */
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

    isLightTheme: boolean
    userAccountSideBarItems: UserAccountSidebarItems
    userAccountAreaRoutes: ReadonlyArray<UserAccountAreaRoute>
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
                            map(c => ({ userOrError: c } as PartialStateUpdate)),

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
            userAccountAreaRoutes: this.props.userAccountAreaRoutes,
            userAccountSideBarItems: this.props.userAccountSideBarItems,
        }
        return (
            <div className="user-area area--vertical">
                <UserAreaHeader
                    className="area--vertical__header"
                    {...this.props}
                    {...context}
                    navItems={this.props.userAreaHeaderNavItems}
                />
                <div className="area--vertical__content">
                    <div className="area--vertical__content-inner">
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
            </div>
        )
    }

    private onDidUpdateUser = () => this.refreshRequests.next()
}
