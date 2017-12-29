import format from 'date-fns/format'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { pluralize } from '../util/strings'
import { fetchAllUsers, randomizeUserPasswordBySiteAdmin, setUserIsSiteAdmin } from './backend'
import { SettingsInfo } from './util/SettingsInfo'

interface UserListItemProps {
    className: string

    /**
     * The user to display in this list item.
     */
    user: GQL.IUser

    /**
     * The currently authenticated user.
     */
    currentUser: GQL.IUser

    /**
     * Called when the user is updated by an action in this list item.
     */
    onDidUpdate?: () => void
}

interface UserListItemState {
    loading: boolean
    errorDescription?: string
    resetPasswordURL?: string
}

class UserListItem extends React.PureComponent<UserListItemProps, UserListItemState> {
    public state: UserListItemState = {
        loading: false,
    }

    public render(): JSX.Element | null {
        const actions: JSX.Element[] = []
        if (this.props.user.authID !== this.props.currentUser.authID) {
            if (this.props.user.siteAdmin) {
                actions.push(
                    <button
                        key="demote"
                        className="btn btn-sm"
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
                        className="btn btn-primary btn-sm"
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
                    className="btn btn-link btn-sm"
                    onClick={this.randomizePassword}
                    disabled={this.state.loading || !!this.state.resetPasswordURL}
                >
                    Reset password
                </button>
            )
        }

        return (
            <li className={`site-admin-all-users-page__item-container ${this.props.className}`}>
                <div className="site-admin-all-users-page__item">
                    <div className="site-admin-detail-list__header">
                        <span className="site-admin-detail-list__name">{this.props.user.username}</span>
                        <br />
                        <span className="site-admin-detail-list__display-name">{this.props.user.displayName}</span>
                    </div>
                    <ul className="site-admin-detail-list__info">
                        {this.props.user.siteAdmin && (
                            <li>
                                <strong>Site admin</strong>
                            </li>
                        )}
                        {this.props.user.email && (
                            <li>
                                Email: <a href={`mailto:${this.props.user.email}`}>{this.props.user.email}</a>
                            </li>
                        )}
                        <li>ID: {this.props.user.id}</li>
                        {this.props.user.createdAt && (
                            <li>Created: {format(this.props.user.createdAt, 'YYYY-MM-DD')}</li>
                        )}
                        {this.props.user.orgs &&
                            this.props.user.orgs.length > 0 && (
                                <li>Orgs: {this.props.user.orgs.map(org => org.name).join(', ')}</li>
                            )}
                        {this.props.user.latestSettings && (
                            <li>
                                <SettingsInfo
                                    settings={this.props.user.latestSettings}
                                    filename={`user-settings-${this.props.user.id}.json`}
                                />
                            </li>
                        )}
                        {this.props.user.tags &&
                            this.props.user.tags.length > 0 && (
                                <li>Tags: {this.props.user.tags.map(tag => tag.name).join(', ')}</li>
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
                            Password was reset. You must manually send <strong>{this.props.user.username}</strong> this
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
                    ? `Really promote user ${this.props.user.username} to site admin?`
                    : `Really revoke site admin status from user ${this.props.user.username}?`
            )
        ) {
            return
        }

        this.setState({
            errorDescription: undefined,
            loading: true,
        })

        setUserIsSiteAdmin(this.props.user.id, siteAdmin)
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
                    this.props.user.username
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

        randomizeUserPasswordBySiteAdmin(this.props.user.id)
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
}

interface Props extends RouteComponentProps<any> {
    user: GQL.IUser
}

export interface State {
    users?: GQL.IUser[]
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

        this.subscriptions.add(
            this.userUpdates.pipe(mergeMap(fetchAllUsers)).subscribe(users => this.setState({ users }))
        )
        this.userUpdates.next()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-detail-list site-admin-all-users-page">
                <PageTitle title="Users - Admin" />
                <h2>Users</h2>
                <p>
                    See <a href="https://about.sourcegraph.com/docs/server/config/">Sourcegraph documentation</a> for
                    information about configuring user accounts and authentication.
                </p>
                <p>
                    <Link to="/site-admin/invite-user" className="btn btn-primary btn-sm">
                        Invite user
                    </Link>
                </p>
                <ul className="site-admin-detail-list__list">
                    {this.state.users &&
                        this.state.users.map(user => (
                            <UserListItem
                                key={user.id}
                                className="site-admin-detail-list__item"
                                user={user}
                                currentUser={this.props.user}
                                onDidUpdate={this.onDidUpdateUser}
                            />
                        ))}
                </ul>
                {this.state.users && (
                    <p>
                        <small>
                            {this.state.users.length} {pluralize('user', this.state.users.length)} total
                        </small>
                    </p>
                )}
            </div>
        )
    }

    private onDidUpdateUser = () => this.userUpdates.next()
}
