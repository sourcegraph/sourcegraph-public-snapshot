import DeleteIcon from '@sourcegraph/icons/lib/Delete'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { gql, mutateGraphQL, queryGraphQL } from '../../backend/graphql'
import { FilteredConnection } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError } from '../../util/errors'
import { AddUserEmailForm } from './AddUserEmailForm'

interface UserEmailNodeProps {
    node: GQL.IUserEmail
    user: GQL.IUser

    onDidUpdate: () => void
}

interface UserEmailNodeState {
    loading: boolean
    errorDescription?: string
}

class UserEmailNode extends React.PureComponent<UserEmailNodeProps, UserEmailNodeState> {
    public state: UserEmailNodeState = {
        loading: false,
    }

    public render(): JSX.Element | null {
        let statusFragment: React.ReactFragment
        if (this.props.node.verified) {
            statusFragment = <span className="badge badge-success">Verified</span>
        } else if (this.props.node.verificationPending) {
            statusFragment = <span className="badge badge-info">Verification pending</span>
        } else {
            statusFragment = <span className="badge badge-secondary">Not verified</span>
        }

        return (
            <li key={this.props.node.email} className="site-admin-detail-list__item">
                <div className="site-admin-detail-list__header site-admin-detail-list__header--center">
                    <strong>{this.props.node.email}</strong> &nbsp;{statusFragment}
                </div>
                <div className="site-admin-detail-list__actions">
                    <button
                        className="btn btn-sm btn-outline-danger site-admin-detail-list__action"
                        onClick={this.remove}
                        disabled={this.state.loading}
                        data-tooltip="Remove email address"
                    >
                        <DeleteIcon className="icon-inline" />
                    </button>
                    {this.props.node.viewerCanManuallyVerify && (
                        <button
                            className="btn btn-sm btn-secondary site-admin-detail-list__action"
                            onClick={this.props.node.verified ? this.setAsUnverified : this.setAsVerified}
                            disabled={this.state.loading}
                        >
                            {this.props.node.verified ? 'Mark as unverified' : 'Mark as verified'}
                        </button>
                    )}
                    {this.state.errorDescription && (
                        <p className="site-admin-detail-list__error">{this.state.errorDescription}</p>
                    )}
                </div>
            </li>
        )
    }

    private remove = () => {
        if (!window.confirm(`Really remove the email address ${this.props.node.email}?`)) {
            return
        }

        this.setState({
            errorDescription: undefined,
            loading: true,
        })

        mutateGraphQL(
            gql`
                mutation RemoveUserEmail($user: ID!, $email: String!) {
                    removeUserEmail(user: $user, email: $email) {
                        alwaysNil
                    }
                }
            `,
            { user: this.props.user.id, email: this.props.node.email }
        )
            .pipe(
                map(({ data, errors }) => {
                    if (!data || (errors && errors.length > 0)) {
                        throw createAggregateError(errors)
                    }
                })
            )
            .subscribe(
                () => {
                    this.setState({ loading: false })
                    if (this.props.onDidUpdate) {
                        this.props.onDidUpdate()
                    }
                },
                error => this.setState({ loading: false, errorDescription: error.message })
            )
    }

    private setAsVerified = () => this.setVerified(true)
    private setAsUnverified = () => this.setVerified(false)

    private setVerified(verified: boolean): void {
        this.setState({
            errorDescription: undefined,
            loading: true,
        })

        mutateGraphQL(
            gql`
                mutation SetUserEmailVerified($user: ID!, $email: String!, $verified: Boolean!) {
                    setUserEmailVerified(user: $user, email: $email, verified: $verified) {
                        alwaysNil
                    }
                }
            `,
            { user: this.props.user.id, email: this.props.node.email, verified }
        )
            .pipe(
                map(({ data, errors }) => {
                    if (!data || (errors && errors.length > 0)) {
                        throw createAggregateError(errors)
                    }
                })
            )
            .subscribe(
                () => {
                    this.setState({ loading: false })
                    if (this.props.onDidUpdate) {
                        this.props.onDidUpdate()
                    }
                },
                error => this.setState({ loading: false, errorDescription: error.message })
            )
    }
}

interface Props extends RouteComponentProps<{}> {
    user: GQL.IUser
}

/** We fake a XyzConnection type because our GraphQL API doesn't have one (or need one) for user emails. */
interface UserEmailConnection {
    nodes: GQL.IUserEmail[]
    totalCount: number
}

class FilteredUserEmailConnection extends FilteredConnection<
    GQL.IUserEmail,
    Pick<UserEmailNodeProps, 'user' | 'onDidUpdate'>
> {}

// TODO(sqs): this is feature-flagged because it doesn't yet send verification emails, which is confusing. See TODO
// in graphqlbackend AddUserEmail method.
const enableAddUserEmail = localStorage.getItem('addUserEmail') !== null

export class UserSettingsEmailsPage extends React.Component<Props> {
    private userEmailUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSettingsEmails')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<UserEmailNodeProps, 'user' | 'onDidUpdate'> = {
            user: this.props.user,
            onDidUpdate: this.onDidUpdateUserEmail,
        }

        return (
            <div className="site-admin-detail-list user-settings-emails-page">
                <PageTitle title="Emails" />
                <h2 className="site-admin-page__header-title">Emails</h2>
                <FilteredUserEmailConnection
                    className="site-admin-page__filtered-connection"
                    noun="email address"
                    pluralNoun="email addresses"
                    queryConnection={this.queryUserEmails}
                    nodeComponent={UserEmailNode}
                    nodeComponentProps={nodeProps}
                    updates={this.userEmailUpdates}
                    hideFilter={true}
                    noSummaryIfAllNodesVisible={true}
                    history={this.props.history}
                    location={this.props.location}
                />
                {enableAddUserEmail && (
                    <AddUserEmailForm className="mt-4" user={this.props.user.id} onDidAdd={this.onDidUpdateUserEmail} />
                )}
            </div>
        )
    }

    private queryUserEmails = (args: {}): Observable<UserEmailConnection> =>
        queryGraphQL(
            gql`
                query UserEmails($user: ID!) {
                    node(id: $user) {
                        ... on User {
                            emails {
                                email
                                verified
                                verificationPending
                                viewerCanManuallyVerify
                            }
                        }
                    }
                }
            `,
            { user: this.props.user.id }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node) {
                    throw createAggregateError(errors)
                }
                const user = data.node as GQL.IUser
                if (!user.emails) {
                    throw createAggregateError(errors)
                }
                return {
                    nodes: user.emails,
                    totalCount: user.emails.length,
                }
            })
        )

    private onDidUpdateUserEmail = () => this.userEmailUpdates.next()
}
