import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Observable } from 'rxjs/Observable'
import { combineLatest } from 'rxjs/observable/combineLatest'
import { merge } from 'rxjs/observable/merge'
import { of } from 'rxjs/observable/of'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { publishReplay } from 'rxjs/operators/publishReplay'
import { refCount } from 'rxjs/operators/refCount'
import { switchMap } from 'rxjs/operators/switchMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { gql, queryGraphQL } from '../../backend/graphql'
import { HeroPage } from '../../components/HeroPage'
import { SettingsArea } from '../../settings/SettingsArea'
import { createAggregateError, ErrorLike, isErrorLike } from '../../util/errors'
import { memoizeObservable } from '../../util/memoize'
import { UserHeader } from './UserHeader'
import { UserOverviewPage } from './UserOverviewPage'

export const enableUserArea = localStorage.getItem('userArea') !== null

const fetchUser = memoizeObservable(
    (args: { username: string }): Observable<GQL.IUser | null> =>
        queryGraphQL(
            gql`
                query User($username: String!) {
                    user(username: $username) {
                        id
                        externalID
                        username
                        displayName
                        avatarURL
                        viewerCanAdminister
                        createdAt
                        emails {
                            email
                            verified
                        }
                        orgs {
                            id
                            displayName
                            name
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
        ),
    ({ username }) => username
)

const NotFoundPage = () => (
    <HeroPage icon={DirectionalSignIcon} title="404: Not Found" subtitle="Sorry, the requested user was not found." />
)

interface Props extends RouteComponentProps<{ username: string }> {
    /**
     * The currently authenticated user, NOT the user whose username is specified in the URL's "username" route
     * parameter.
     */
    user: GQL.IUser | null

    isLightTheme: boolean
    onThemeChange: () => void
}

interface State {
    /**
     * The fetched user or an error if an error occurred; undefined while loading.
     */
    userOrError?: GQL.IUser | ErrorLike
}

/**
 * Properties passed to all page components in the user area.
 */
export interface UserAreaPageProps {
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
export class UserArea extends React.Component<Props> {
    public state: State = {}

    private routeMatchChanges = new Subject<{ username: string }>()
    private refreshRequests = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Changes to the route-matched username.
        const usernameChanges = this.routeMatchChanges.pipe(map(({ username }) => username), distinctUntilChanged())

        // Fetch user.
        this.subscriptions.add(
            combineLatest(usernameChanges, merge(this.refreshRequests.pipe(map(() => true)), of(false)))
                .pipe(
                    switchMap(([username, forceRefresh]) => {
                        type PartialStateUpdate = Pick<State, 'userOrError'>
                        const result = fetchUser({ username }, forceRefresh).pipe(
                            catchError(error => [error]),
                            map(c => ({ userOrError: c } as PartialStateUpdate)),
                            publishReplay<PartialStateUpdate>(),
                            refCount()
                        )
                        return merge(
                            // Don't clear old user data while we reload, to avoid unmounting all components during
                            // loading.
                            of(forceRefresh ? {} : { userOrError: undefined }),

                            result
                        )
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        this.routeMatchChanges.next(this.props.match.params)
    }

    public componentWillReceiveProps(props: Props): void {
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
            return <HeroPage icon={ErrorIcon} title="Error" subtitle={upperFirst(this.state.userOrError.message)} />
        }

        const transferProps: UserAreaPageProps = {
            user: this.state.userOrError,
            onDidUpdateUser: this.onDidUpdateUser,
            authenticatedUser: this.props.user,
        }
        return (
            <div className="user-area area--vertical">
                <UserHeader className="area--vertical__header" {...this.props} {...transferProps} />
                <div className="area--vertical__content user-area__content">
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
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <SettingsArea
                                        {...routeComponentProps}
                                        {...transferProps}
                                        isLightTheme={this.props.isLightTheme}
                                        onThemeChange={this.props.onThemeChange}
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
