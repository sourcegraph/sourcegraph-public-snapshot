import { isEqual } from 'lodash'
import AddIcon from 'mdi-react/AddIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { merge, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import * as GQL from '../backend/graphqlschema'
import { CopyableText } from '../components/CopyableText'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { userURL } from '../user'
import { setUserEmailVerified } from '../user/account/backend'
import { asError } from '../util/errors'
import { deleteUser, fetchAllUsers, randomizeUserPassword, setUserIsSiteAdmin } from './backend'

interface UserNodeProps {
    /**
     * The user to display in this list item.
     */
    node: GQL.IUser

    /**
     * The currently authenticated user.
     */
    currentUser: GQL.IUser

    /**
     * Called when the user is updated by an action in this list item.
     */
    onDidUpdate?: () => void
}

interface UserNodeState {
    loading: boolean
    errorDescription?: string
    resetPasswordURL?: string | null
}

class UserNode extends React.PureComponent<UserNodeProps, UserNodeState> {
    public state: UserNodeState = {
        loading: false,
    }

    private emailVerificationClicks = new Subject<{ email: string; verified: boolean }>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.emailVerificationClicks
                .pipe(
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    switchMap(({ email, verified }) =>
                        merge(
                            of({
                                errorDescription: undefined,
                                resetPasswordURL: undefined,
                                loading: true,
                            }),
                            setUserEmailVerified(this.props.node.id, email, verified).pipe(
                                map(() => ({ loading: false })),
                                catchError(error => [{ loading: false, errorDescription: asError(error).message }])
                            )
                        )
                    )
                )
                .subscribe(
                    stateUpdate => {
                        this.setState(stateUpdate)
                        if (this.props.onDidUpdate) {
                            this.props.onDidUpdate()
                        }
                    },
                    err => console.error(err)
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <li className="list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        <Link to={`/users/${this.props.node.username}`}>
                            <strong>{this.props.node.username}</strong>
                        </Link>
                        <br />
                        <span className="text-muted">{this.props.node.displayName}</span>
                    </div>
                    <div>
                        <Link className="btn btn-sm btn-secondary" to={`${userURL(this.props.node.username)}/settings`}>
                            <SettingsIcon className="icon-inline" /> Settings
                        </Link>{' '}
                        {window.context.resetPasswordEnabled && (
                            <button
                                className="btn btn-sm btn-secondary"
                                onClick={this.randomizePassword}
                                disabled={this.state.loading || !!this.state.resetPasswordURL}
                            >
                                Reset password
                            </button>
                        )}{' '}
                        {this.props.node.id !== this.props.currentUser.id &&
                            (this.props.node.siteAdmin ? (
                                <button
                                    className="btn btn-sm btn-secondary"
                                    onClick={this.demoteFromSiteAdmin}
                                    disabled={this.state.loading}
                                >
                                    Revoke site admin
                                </button>
                            ) : (
                                <button
                                    key="promote"
                                    className="btn btn-sm btn-secondary"
                                    onClick={this.promoteToSiteAdmin}
                                    disabled={this.state.loading}
                                >
                                    Promote to site admin
                                </button>
                            ))}{' '}
                        {this.props.node.id !== this.props.currentUser.id && (
                            <button
                                className="btn btn-sm btn-danger"
                                onClick={this.deleteUser}
                                disabled={this.state.loading}
                                data-tooltip="Delete user"
                            >
                                <DeleteIcon className="icon-inline" />
                            </button>
                        )}
                    </div>
                </div>
                {this.state.errorDescription && (
                    <div className="alert alert-danger mt-2">{this.state.errorDescription}</div>
                )}
                {this.state.resetPasswordURL && (
                    <div className="alert alert-success mt-2">
                        <p>
                            Password was reset. You must manually send <strong>{this.props.node.username}</strong> this
                            reset link:
                        </p>
                        <CopyableText text={this.state.resetPasswordURL} size={40} />
                    </div>
                )}
            </li>
        )
    }

    private promoteToSiteAdmin = () => this.setSiteAdmin(true)
    private demoteFromSiteAdmin = () => this.setSiteAdmin(false)

    private setSiteAdmin(siteAdmin: boolean): void {
        if (
            !window.confirm(
                siteAdmin
                    ? `Really promote user ${this.props.node.username} to site admin?`
                    : `Really revoke site admin status from user ${this.props.node.username}?`
            )
        ) {
            return
        }

        this.setState({
            errorDescription: undefined,
            loading: true,
        })

        setUserIsSiteAdmin(this.props.node.id, siteAdmin)
            .toPromise()
            .then(
                () => {
                    this.setState({ loading: false })
                    if (this.props.onDidUpdate) {
                        this.props.onDidUpdate()
                    }
                },
                err => this.setState({ loading: false, errorDescription: err.message })
            )
    }

    private randomizePassword = () => {
        if (
            !window.confirm(
                `Really reset the password for ${
                    this.props.node.username
                } to a random password? The user must reset their password to sign in again.`
            )
        ) {
            return
        }

        this.setState({
            errorDescription: undefined,
            resetPasswordURL: undefined,
            loading: true,
        })

        randomizeUserPassword(this.props.node.id)
            .toPromise()
            .then(
                ({ resetPasswordURL }) => {
                    this.setState({
                        loading: false,
                        resetPasswordURL,
                    })
                },
                err => this.setState({ loading: false, errorDescription: err.message })
            )
    }

    private deleteUser = () => {
        if (!window.confirm(`Really delete the user ${this.props.node.username}?`)) {
            return
        }

        this.setState({
            errorDescription: undefined,
            resetPasswordURL: undefined,
            loading: true,
        })

        deleteUser(this.props.node.id)
            .toPromise()
            .then(
                () => {
                    this.setState({ loading: false })
                    if (this.props.onDidUpdate) {
                        this.props.onDidUpdate()
                    }
                },
                err => this.setState({ loading: false, errorDescription: err.message })
            )
    }
}

interface Props extends RouteComponentProps<any> {
    user: GQL.IUser
}

interface State {
    users?: GQL.IUser[]
    totalCount?: number
}

class FilteredUserConnection extends FilteredConnection<
    GQL.IUser,
    Pick<UserNodeProps, 'currentUser' | 'onDidUpdate'>
> {}

/**
 * A page displaying the users on this site.
 */
export class SiteAdminAllUsersPage extends React.Component<Props, State> {
    public state: State = {}

    private userUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminAllUsers')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<UserNodeProps, 'currentUser' | 'onDidUpdate'> = {
            currentUser: this.props.user,
            onDidUpdate: this.onDidUpdateUser,
        }

        return (
            <div className="site-admin-all-users-page">
                <PageTitle title="Users - Admin" />
                <h2>Users</h2>
                <div>
                    <Link to="/site-admin/users/new" className="btn btn-primary">
                        <AddIcon className="icon-inline" /> Create user account
                    </Link>
                    &nbsp;
                    <Link to="/site-admin/configuration" className="btn btn-secondary">
                        <SettingsIcon className="icon-inline" /> Configure SSO
                    </Link>
                </div>
                <FilteredUserConnection
                    className="mt-3"
                    listClassName="list-group list-group-flush"
                    noun="user"
                    pluralNoun="users"
                    queryConnection={fetchAllUsers}
                    nodeComponent={UserNode}
                    nodeComponentProps={nodeProps}
                    updates={this.userUpdates}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private onDidUpdateUser = () => this.userUpdates.next()
}
