import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { combineLatest, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { HeroPage } from '../../components/HeroPage'
import { ExtensionsProps } from '../../extensions/ExtensionsClientCommonContext'
import { SettingsArea } from '../../settings/SettingsArea'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { createAggregateError, ErrorLike, isErrorLike } from '../../util/errors'
import { UserAccountArea, UserAccountAreaRoute } from '../account/UserAccountArea'
import { UserAccountSidebarItems } from '../account/UserAccountSidebar'
import { UserHeader } from './UserHeader'
import { UserOverviewPage } from './UserOverviewPage'

export const enableUserArea = localStorage.getItem('userArea') !== null

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

interface UserAreaProps extends RouteComponentProps<{ username: string }>, ExtensionsProps {
    sideBarItems: UserAccountSidebarItems
    routes: ReadonlyArray<UserAccountAreaRoute>

    /**
     * The currently authenticated user, NOT the user whose username is specified in the URL's "username" route
     * parameter.
     */
    user: GQL.IUser | null

    isLightTheme: boolean
}

interface UserAreaState {
    /**
     * The fetched user or an error if an error occurred; undefined while loading.
     */
    userOrError?: GQL.IUser | ErrorLike
}

/**
 * Properties passed to all page components in the user area.
 */
export interface UserAreaPageProps extends ExtensionsProps {
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

        const transferProps: UserAreaPageProps = {
            user: this.state.userOrError,
            onDidUpdateUser: this.onDidUpdateUser,
            authenticatedUser: this.props.user,
            extensions: this.props.extensions,
        }
        return (
            <div className="user-area area--vertical">
                <UserHeader className="area--vertical__header" {...this.props} {...transferProps} />
                <div className="area--vertical__content">
                    <div className="area--vertical__content-inner">
                        <Switch>
                            <Route
                                path={`${this.props.match.url}`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <UserOverviewPage {...routeComponentProps} {...transferProps} />
                                )}
                            />
                            <Route
                                path={`${this.props.match.url}/settings`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <SettingsArea
                                        {...routeComponentProps}
                                        {...transferProps}
                                        subject={transferProps.user}
                                        isLightTheme={this.props.isLightTheme}
                                        extraHeader={
                                            <>
                                                {transferProps.authenticatedUser &&
                                                    transferProps.user.id !== transferProps.authenticatedUser.id && (
                                                        <SiteAdminAlert className="sidebar__alert">
                                                            Viewing settings for{' '}
                                                            <strong>{transferProps.user.username}</strong>
                                                        </SiteAdminAlert>
                                                    )}
                                                <p>User settings override global and organization settings.</p>
                                            </>
                                        }
                                    />
                                )}
                            />
                            <Route
                                path={`${this.props.match.url}/account`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <UserAccountArea
                                        {...routeComponentProps}
                                        {...transferProps}
                                        routes={this.props.routes}
                                        sideBarItems={this.props.sideBarItems}
                                        isLightTheme={this.props.isLightTheme}
                                    />
                                )}
                            />
                            <Route key="hardcoded-key" component={NotFoundPage} />
                        </Switch>
                    </div>
                </div>
            </div>
        )
    }

    private onDidUpdateUser = () => this.refreshRequests.next()
}
