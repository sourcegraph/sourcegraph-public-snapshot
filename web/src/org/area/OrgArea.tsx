import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Observable } from 'rxjs/Observable'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { gql, queryGraphQL } from '../../backend/graphql'
import { HeroPage } from '../../components/HeroPage'
import { createAggregateError, ErrorLike, isErrorLike } from '../../util/errors'
import { memoizeObservable } from '../../util/memoize'
import { OrgSettingsArea } from '../settings/OrgSettingsArea'
import { OrgHeader } from './OrgHeader'
import { OrgMembersPage } from './OrgMembersPage'
import { OrgOverviewPage } from './OrgOverviewPage'

const fetchOrg = memoizeObservable(
    (args: { name: string }): Observable<GQL.IOrg | null> =>
        queryGraphQL(
            gql`
                query Organization($name: String!) {
                    organization(name: $name) {
                        id
                        name
                        displayName
                        viewerIsMember
                        viewerCanAdminister
                        createdAt
                        tags {
                            name
                        }
                    }
                }
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.organization) {
                    throw createAggregateError(errors)
                }
                return data.organization
            })
        ),
    ({ name }) => name
)

const NotFoundPage = () => (
    <HeroPage
        icon={DirectionalSignIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested organization was not found."
    />
)

interface Props extends RouteComponentProps<{ name: string }> {
    /**
     * The currently authenticated user.
     */
    user: GQL.IUser | null

    isLightTheme: boolean
}

interface State {
    /**
     * The fetched org or an error if an error occurred; undefined while loading.
     */
    orgOrError?: GQL.IOrg | ErrorLike
}

/**
 * Properties passed to all page components in the org area.
 */
export interface OrgAreaPageProps {
    /** The org that is the subject of the page. */
    org: GQL.IOrg

    /** The currently authenticated user. */
    authenticatedUser: GQL.IUser | null
}

/**
 * An organization's public profile area.
 */
export class OrgArea extends React.Component<Props> {
    public state: State = {}

    private routeMatchChanges = new Subject<{ name: string }>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Fetch organization.
        this.subscriptions.add(
            this.routeMatchChanges
                .pipe(
                    map(({ name }) => name),
                    distinctUntilChanged(),
                    switchMap(name => {
                        type PartialStateUpdate = Pick<State, 'orgOrError'>
                        return fetchOrg({ name }).pipe(
                            catchError(error => [error]),
                            map(c => ({ orgOrError: c } as PartialStateUpdate)),
                            startWith<PartialStateUpdate>({ orgOrError: undefined })
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
        if (!this.state.orgOrError) {
            return null // loading
        }
        if (isErrorLike(this.state.orgOrError)) {
            return <HeroPage icon={ErrorIcon} title="Error" subtitle={upperFirst(this.state.orgOrError.message)} />
        }

        const transferProps: OrgAreaPageProps = {
            authenticatedUser: this.props.user,
            org: this.state.orgOrError,
        }
        return (
            <div className="org-area area--vertical">
                <OrgHeader className="area--vertical__header" {...this.props} {...transferProps} />
                <div className="area--vertical__content">
                    <div className="area--vertical__content-inner">
                        <Switch>
                            <Route
                                path={`${this.props.match.url}`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <OrgOverviewPage {...routeComponentProps} {...transferProps} />
                                )}
                            />
                            <Route
                                path={`${this.props.match.url}/members`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <OrgMembersPage {...routeComponentProps} {...transferProps} />
                                )}
                            />
                            <Route
                                path={`${this.props.match.url}/settings`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <OrgSettingsArea
                                        {...routeComponentProps}
                                        {...transferProps}
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
}
