import DeleteIcon from 'mdi-react/DeleteIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../../shared/src/util/errors'
import { mutateGraphQL, queryGraphQL } from '../../../backend/graphql'
import { FilteredConnection } from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import { SiteFlags } from '../../../site'
import { siteFlags } from '../../../site/backend'
import { eventLogger } from '../../../tracking/eventLogger'
import { setUserEmailVerified } from '../backend'
import { AddUserEmailForm } from './AddUserEmailForm'
import { ErrorAlert } from '../../../components/alerts'

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
        let verifiedFragment: React.ReactFragment
        if (that.props.node.verified) {
            verifiedFragment = <span className="badge badge-success">Verified</span>
        } else if (that.props.node.verificationPending) {
            verifiedFragment = <span className="badge badge-info">Verification pending</span>
        } else {
            verifiedFragment = <span className="badge badge-secondary">Not verified</span>
        }

        return (
            <li className="list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        <strong>{that.props.node.email}</strong> &nbsp;{verifiedFragment}&nbsp;
                        {that.props.node.isPrimary && <span className="badge badge-primary">Primary</span>}
                    </div>
                    <div>
                        <button
                            type="button"
                            className="btn btn-sm btn-danger"
                            onClick={that.remove}
                            disabled={that.state.loading}
                            data-tooltip="Remove email address"
                        >
                            <DeleteIcon className="icon-inline" />
                        </button>{' '}
                        {that.props.node.viewerCanManuallyVerify && (
                            <button
                                type="button"
                                className="btn btn-sm btn-secondary"
                                onClick={that.props.node.verified ? that.setAsUnverified : that.setAsVerified}
                                disabled={that.state.loading}
                            >
                                {that.props.node.verified ? 'Mark as unverified' : 'Mark as verified'}
                            </button>
                        )}
                    </div>
                </div>
                {that.state.errorDescription && <ErrorAlert className="mt-2" error={that.state.errorDescription} />}
            </li>
        )
    }

    private remove = (): void => {
        if (!window.confirm(`Remove the email address ${that.props.node.email}?`)) {
            return
        }

        that.setState({
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
            { user: that.props.user.id, email: that.props.node.email }
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
                    that.setState({ loading: false })
                    eventLogger.log('UserEmailAddressDeleted')
                    if (that.props.onDidUpdate) {
                        that.props.onDidUpdate()
                    }
                },
                error => that.setState({ loading: false, errorDescription: error.message })
            )
    }

    private setAsVerified = (): void => that.setVerified(true)
    private setAsUnverified = (): void => that.setVerified(false)

    private setVerified(verified: boolean): void {
        that.setState({
            errorDescription: undefined,
            loading: true,
        })

        setUserEmailVerified(that.props.user.id, that.props.node.email, verified).subscribe(
            () => {
                that.setState({ loading: false })
                if (verified) {
                    eventLogger.log('UserEmailAddressMarkedVerified')
                } else {
                    eventLogger.log('UserEmailAddressMarkedUnverified')
                }
                if (that.props.onDidUpdate) {
                    that.props.onDidUpdate()
                }
            },
            error => that.setState({ loading: false, errorDescription: error.message })
        )
    }
}

interface Props extends RouteComponentProps<{}> {
    user: GQL.IUser
}

interface State {
    siteFlags?: SiteFlags
}

/** We fake a XyzConnection type because our GraphQL API doesn't have one (or need one) for user emails. */
interface UserEmailConnection {
    nodes: GQL.IUserEmail[]
    totalCount: number
}

export class UserSettingsEmailsPage extends React.Component<Props, State> {
    public state: State = {}

    private userEmailUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSettingsEmails')

        this.subscriptions.add(siteFlags.subscribe(siteFlags => this.setState({ siteFlags })))
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
            <div className="user-settings-emails-page">
                <PageTitle title="Emails" />
                <h2>Emails</h2>
                {this.state.siteFlags && !this.state.siteFlags.sendsEmailVerificationEmails && (
                    <div className="alert alert-warning mt-2">
                        Sourcegraph is not configured to send email verifications. Newly added email addresses must be
                        manually verified by a site admin.
                    </div>
                )}
                <FilteredConnection<GQL.IUserEmail, Pick<UserEmailNodeProps, 'user' | 'onDidUpdate'>>
                    className="list-group list-group-flush mt-3"
                    noun="email address"
                    pluralNoun="email addresses"
                    queryConnection={this.queryUserEmails}
                    nodeComponent={UserEmailNode}
                    nodeComponentProps={nodeProps}
                    updates={this.userEmailUpdates}
                    hideSearch={true}
                    noSummaryIfAllNodesVisible={true}
                    history={this.props.history}
                    location={this.props.location}
                />
                <AddUserEmailForm className="mt-4" user={this.props.user.id} onDidAdd={this.onDidUpdateUserEmail} />
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
                                isPrimary
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

    private onDidUpdateUserEmail = (): void => this.userEmailUpdates.next()
}
