import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { combineLatest, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { HeroPage } from '../../components/HeroPage'
import { createAggregateError, ErrorLike, isErrorLike } from '../../util/errors'
import { OrgSettingsArea } from '../settings/OrgSettingsArea'
import { OrgExtensionsPage } from './OrgExtensionsPage'
import { OrgHeader } from './OrgHeader'
import { OrgInvitationPage } from './OrgInvitationPage'
import { OrgMembersPage } from './OrgMembersPage'
import { OrgOverviewPage } from './OrgOverviewPage'

function queryOrganization(args: { name: string }): Observable<GQL.IOrg | null> {
    return queryGraphQL(
        gql`
            query Organization($name: String!) {
                organization(name: $name) {
                    __typename
                    id
                    name
                    displayName
                    url
                    settingsURL
                    viewerPendingInvitation {
                        id
                        sender {
                            username
                            displayName
                            avatarURL
                            createdAt
                        }
                        respondURL
                    }
                    viewerIsMember
                    viewerCanAdminister
                    createdAt
                    tags {
                        name
                    }
                    registryExtensions {
                        url
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
    )
}

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

    /** Called when the organization is updated and must be reloaded. */
    onOrganizationUpdate: () => void

    /** The currently authenticated user. */
    authenticatedUser: GQL.IUser | null
}

/**
 * An organization's public profile area.
 */
export class OrgArea extends React.Component<Props> {
    public state: State = {}

    private routeMatchChanges = new Subject<{ name: string }>()
    private refreshRequests = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Changes to the route-matched org name.
        const nameChanges = this.routeMatchChanges.pipe(map(({ name }) => name), distinctUntilChanged())

        // Fetch organization.
        this.subscriptions.add(
            combineLatest(nameChanges, merge(this.refreshRequests.pipe(mapTo(false)), of(true)))
                .pipe(
                    switchMap(([name, forceRefresh]) => {
                        type PartialStateUpdate = Pick<State, 'orgOrError'>
                        return queryOrganization({ name }).pipe(
                            catchError(error => [error]),
                            map(c => ({ orgOrError: c } as PartialStateUpdate)),

                            // Don't clear old org data while we reload, to avoid unmounting all components during
                            // loading.
                            startWith<PartialStateUpdate>(forceRefresh ? { orgOrError: undefined } : {})
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
            onOrganizationUpdate: this.onDidUpdateOrganization,
        }

        if (this.props.location.pathname === `${this.props.match.url}/invitation`) {
            // The OrgInvitationPage is displayed without the OrgHeader because it is modal-like.
            return <OrgInvitationPage {...transferProps} onDidRespondToInvitation={this.onDidRespondToInvitation} />
        }

        return (
            <div
                className="org-area" // OrgsArea applies .area--vertical class, don't apply here or else v-scroll breaks
            >
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
                            {window.context.platformEnabled && (
                                <Route
                                    path={`${this.props.match.url}/extensions`}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    // tslint:disable-next-line:jsx-no-lambda
                                    render={routeComponentProps => (
                                        <OrgExtensionsPage {...routeComponentProps} {...transferProps} />
                                    )}
                                />
                            )}
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

    private onDidRespondToInvitation = () => this.refreshRequests.next()

    private onDidUpdateOrganization = () => this.refreshRequests.next()
}
