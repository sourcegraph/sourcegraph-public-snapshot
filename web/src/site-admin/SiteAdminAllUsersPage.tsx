import AddIcon from '@sourcegraph/icons/lib/Add'
import GearIcon from '@sourcegraph/icons/lib/Gear'
import format from 'date-fns/format'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { deleteUser, fetchAllUsers, randomizeUserPasswordBySiteAdmin, setUserIsSiteAdmin } from './backend'
import { SettingsInfo } from './util/SettingsInfo'

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
    resetPasswordURL?: string
}

class UserNode extends React.PureComponent<UserNodeProps, UserNodeState> {
    public state: UserNodeState = {
        loading: false,
    }

    public render(): JSX.Element | null {
        const actions: JSX.Element[] = []
        if (this.props.node.id !== this.props.currentUser.id) {
            if (this.props.node.siteAdmin) {
                actions.push(
                    <button
                        key="demote"
                        className="btn btn-secondary btn-sm site-admin-detail-list__action"
                        onClick={this.demoteFromSiteAdmin}
                        disabled={this.state.loading}
                    >
                        Revoke site admin
                    </button>
                )
            } else {
                actions.push(
                    <button
                        key="promote"
                        className="btn btn-secondary btn-sm site-admin-detail-list__action"
                        onClick={this.promoteToSiteAdmin}
                        disabled={this.state.loading}
                    >
                        Promote to site admin
                    </button>
                )
            }
            actions.push(
                <button
                    key="randomizePassword"
                    className="btn btn-secondary btn-sm site-admin-detail-list__action"
                    onClick={this.randomizePassword}
                    disabled={this.state.loading || !!this.state.resetPasswordURL}
                >
                    Reset password
                </button>
            )
            actions.push(
                <button
                    key="deleteUser"
                    className="btn btn-secondary btn-sm site-admin-detail-list__action"
                    onClick={this.deleteUser}
                    disabled={this.state.loading}
                >
                    Delete user
                </button>
            )
        }

        return (
            <li className="site-admin-detail-list__item site-admin-all-users-page__item-container">
                <div className="site-admin-all-users-page__item">
                    <div className="site-admin-detail-list__header">
                        <span className="site-admin-detail-list__name">{this.props.node.username}</span>
                        <br />
                        <span className="site-admin-detail-list__display-name">{this.props.node.displayName}</span>
                    </div>
                    <ul className="site-admin-detail-list__info">
                        {this.props.node.siteAdmin && (
                            <li>
                                <strong>Site admin</strong>
                            </li>
                        )}
                        {this.props.node.email && (
                            <li>
                                Email: <a href={`mailto:${this.props.node.email}`}>{this.props.node.email}</a>
                            </li>
                        )}
                        {this.props.node.createdAt && (
                            <li>Created: {format(this.props.node.createdAt, 'YYYY-MM-DD')}</li>
                        )}
                        {this.props.node.orgs &&
                            this.props.node.orgs.length > 0 && (
                                <li>Orgs: {this.props.node.orgs.map(org => org.name).join(', ')}</li>
                            )}
                        {this.props.node.latestSettings && (
                            <li>
                                <SettingsInfo
                                    settings={this.props.node.latestSettings}
                                    filename={`user-settings-${this.props.node.id}.json`}
                                />
                            </li>
                        )}
                        {this.props.node.tags &&
                            this.props.node.tags.length > 0 && (
                                <li>Tags: {this.props.node.tags.map(tag => tag.name).join(', ')}</li>
                            )}
                    </ul>
                    <div className="site-admin-detail-list__actions">
                        {actions}
                        {this.state.errorDescription && (
                            <p className="site-admin-detail-list__error">{this.state.errorDescription}</p>
                        )}
                    </div>
                </div>
                {this.state.resetPasswordURL && (
                    <div className="alert alert-success site-admin-all-users-page__item-alert">
                        <p>
                            Password was reset. You must manually send <strong>{this.props.node.username}</strong> this
                            reset link:
                        </p>
                        <div>
                            <code className="site-admin-all-users-page__url">{this.state.resetPasswordURL}</code>
                        </div>
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

        randomizeUserPasswordBySiteAdmin(this.props.node.id)
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

export interface State {
    users?: GQL.IUser[]
    totalCount?: number
}

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
            <div className="site-admin-detail-list site-admin-all-users-page">
                <PageTitle title="Users - Admin" />
                <h2>Users</h2>
                <div className="site-admin-page__actions">
                    <Link to="/site-admin/invite-user" className="btn btn-primary btn-sm site-admin-page__actions-btn">
                        <AddIcon className="icon-inline" /> Invite user
                    </Link>
                    &nbsp;
                    <Link
                        to="/site-admin/configuration"
                        className="btn btn-secondary btn-sm site-admin-page__actions-btn"
                    >
                        <GearIcon className="icon-inline" /> Configure SSO
                    </Link>
                </div>
                <FilteredConnection
                    className="site-admin-page__filtered-connection"
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
