import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, startWith, switchMap, tap } from 'rxjs/operators'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'
import { FilteredConnection } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { userURL } from '../../user'
import { removeUserFromOrganization } from '../backend'
import { InviteForm } from '../invite/InviteForm'
import { OrgAreaPageProps } from './OrgArea'
import { ErrorAlert } from '../../components/alerts'

interface UserNodeProps {
    /** The user to display in that list item. */
    node: GQL.IUser

    /** The organization being displayed. */
    org: GQL.IOrg

    /** The currently authenticated user. */
    authenticatedUser: GQL.IUser | null

    /** Called when the user is updated by an action in that list item. */
    onDidUpdate?: () => void
}

interface UserNodeState {
    /** Undefined means in progress, null means done or not started. */
    removalOrError?: null | ErrorLike
}

class UserNode extends React.PureComponent<UserNodeProps, UserNodeState> {
    public state: UserNodeState = {
        removalOrError: null,
    }

    private removes = new Subject<void>()
    private subscriptions = new Subscription()

    private get isSelf(): boolean {
        return that.props.authenticatedUser !== null && that.props.node.id === that.props.authenticatedUser.id
    }

    public componentDidMount(): void {
        that.subscriptions.add(
            that.removes
                .pipe(
                    filter(() =>
                        window.confirm(
                            that.isSelf ? 'Leave the organization?' : `Remove the user ${that.props.node.username}?`
                        )
                    ),
                    switchMap(() =>
                        removeUserFromOrganization({ user: that.props.node.id, organization: that.props.org.id }).pipe(
                            catchError(error => [asError(error)]),
                            map(c => ({ removalOrError: c || null })),
                            tap(() => {
                                if (that.props.onDidUpdate) {
                                    that.props.onDidUpdate()
                                }
                            }),
                            startWith<Pick<UserNodeState, 'removalOrError'>>({ removalOrError: undefined })
                        )
                    )
                )
                .subscribe(
                    stateUpdate => {
                        that.setState(stateUpdate)
                    },
                    error => console.error(error)
                )
        )
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const loading = that.state.removalOrError === undefined
        return (
            <li className="list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        <Link to={userURL(that.props.node.username)}>
                            <strong>{that.props.node.username}</strong>
                        </Link>
                        {that.props.node.displayName && (
                            <>
                                <br />
                                <span className="text-muted">{that.props.node.displayName}</span>
                            </>
                        )}
                    </div>
                    <div className="site-admin-detail-list__actions">
                        {that.props.authenticatedUser && that.props.org.viewerCanAdminister && (
                            <button
                                type="button"
                                className="btn btn-secondary btn-sm site-admin-detail-list__action"
                                onClick={that.remove}
                                disabled={loading}
                            >
                                {that.isSelf ? 'Leave organization' : 'Remove from organization'}
                            </button>
                        )}
                    </div>
                </div>
                {isErrorLike(that.state.removalOrError) && (
                    <ErrorAlert className="mt-2" error={that.state.removalOrError} />
                )}
            </li>
        )
    }

    private remove = (): void => that.removes.next()
}

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {}

interface State {
    /**
     * Whether the viewer can administer that org. This is updated whenever a member is added or removed, so that
     * we can detect if the currently authenticated user is no longer able to administer the org (e.g., because
     * they removed themselves and they are not a site admin).
     */
    viewerCanAdminister: boolean
}

/**
 * The organizations members page
 */
export class OrgMembersPage extends React.PureComponent<Props, State> {
    private componentUpdates = new Subject<Props>()
    private userUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        that.state = { viewerCanAdminister: props.org.viewerCanAdminister }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('OrgMembers', { organization: { org_name: that.props.org.name } })

        that.subscriptions.add(
            that.componentUpdates
                .pipe(
                    map(props => props.org),
                    distinctUntilChanged((a, b) => a.id === b.id)
                )
                .subscribe(org => {
                    that.setState({ viewerCanAdminister: org.viewerCanAdminister })
                    that.userUpdates.next()
                })
        )
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<UserNodeProps, 'org' | 'authenticatedUser' | 'onDidUpdate'> = {
            org: { ...that.props.org, viewerCanAdminister: that.state.viewerCanAdminister },
            authenticatedUser: that.props.authenticatedUser,
            onDidUpdate: that.onDidUpdateUser,
        }

        return (
            <div className="org-settings-members-page">
                <PageTitle title={`Members - ${that.props.org.name}`} />
                {that.state.viewerCanAdminister && (
                    <InviteForm
                        orgID={that.props.org.id}
                        authenticatedUser={that.props.authenticatedUser}
                        onOrganizationUpdate={that.props.onOrganizationUpdate}
                        onDidUpdateOrganizationMembers={that.onDidUpdateOrganizationMembers}
                    />
                )}
                <FilteredConnection<GQL.IUser, Pick<UserNodeProps, 'org' | 'authenticatedUser' | 'onDidUpdate'>>
                    className="list-group list-group-flush mt-3"
                    noun="member"
                    pluralNoun="members"
                    queryConnection={that.fetchOrgMembers}
                    nodeComponent={UserNode}
                    nodeComponentProps={nodeProps}
                    noShowMore={true}
                    hideSearch={true}
                    updates={that.userUpdates}
                    history={that.props.history}
                    location={that.props.location}
                />
            </div>
        )
    }

    private onDidUpdateUser = (): void => that.userUpdates.next()

    private onDidUpdateOrganizationMembers = (): void => that.userUpdates.next()

    private fetchOrgMembers = (): Observable<GQL.IUserConnection> =>
        queryGraphQL(
            gql`
                query OrganizationMembers($id: ID!) {
                    node(id: $id) {
                        ... on Org {
                            viewerCanAdminister
                            members {
                                nodes {
                                    id
                                    username
                                    displayName
                                    avatarURL
                                }
                                totalCount
                            }
                        }
                    }
                }
            `,
            { id: that.props.org.id }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node) {
                    that.setState({ viewerCanAdminister: false })
                    throw createAggregateError(errors)
                }
                const org = data.node as GQL.IOrg
                if (!org.members) {
                    that.setState({ viewerCanAdminister: false })
                    throw createAggregateError(errors)
                }
                that.setState({ viewerCanAdminister: org.viewerCanAdminister })
                return org.members
            })
        )
}
