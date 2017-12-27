import format from 'date-fns/format'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { pluralize } from '../util/strings'
import { fetchAllUsers, setUserIsSiteAdmin } from './backend'
import { SettingsInfo } from './util/SettingsInfo'

interface Props extends RouteComponentProps<any> {
    user: GQL.IUser
}

export interface State {
    users?: GQL.IUser[]

    /**
     * Errors that occurred while performing an action on a user.
     */
    userErrorDescription: Map<GQLID, string>

    /**
     * Whether an action is currently being performed on a user.
     */
    userLoading: Set<GQLID>
}

/**
 * A page displaying the users on this site.
 */
export class AllUsersPage extends React.Component<Props, State> {
    public state: State = {
        userErrorDescription: new Map<GQLID, string>(),
        userLoading: new Set<GQLID>(),
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminAllUsers')

        this.subscriptions.add(fetchAllUsers().subscribe(users => this.setState({ users })))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const userActions = new Map<GQLID, JSX.Element[]>()
        if (this.state.users) {
            for (const user of this.state.users) {
                const loading = this.state.userLoading.has(user.id)
                const actions: JSX.Element[] = []
                if (user.auth0ID !== this.props.user.auth0ID) {
                    if (user.siteAdmin) {
                        actions.push(
                            <button
                                key={`revoke${user.id}`}
                                className="btn btn-sm"
                                // tslint:disable-next-line:jsx-no-lambda
                                onClick={() => this.setSiteAdmin(user, false)}
                                disabled={loading}
                            >
                                Revoke site admin
                            </button>
                        )
                    } else {
                        actions.push(
                            <button
                                key={`promote${user.id}`}
                                className="btn btn-primary btn-sm"
                                // tslint:disable-next-line:jsx-no-lambda
                                onClick={() => this.setSiteAdmin(user, true)}
                                disabled={loading}
                            >
                                Promote to site admin
                            </button>
                        )
                    }
                }
                if (actions.length > 0) {
                    userActions.set(user.id, actions)
                }
            }
        }

        return (
            <div className="site-admin-detail-list site-admin-all-users-page">
                <PageTitle title="Users - Admin" />
                <h2>Users</h2>
                <p>
                    See <a href="https://about.sourcegraph.com/docs/server/config/">Sourcegraph documentation</a> for
                    information about configuring user accounts and authentication.
                </p>
                <ul className="site-admin-detail-list__list">
                    {this.state.users &&
                        this.state.users.map(user => (
                            <li key={user.id} className="site-admin-detail-list__item">
                                <div className="site-admin-detail-list__header">
                                    <span className="site-admin-detail-list__name">{user.username}</span>
                                    <br />
                                    <span className="site-admin-detail-list__display-name">{user.displayName}</span>
                                </div>
                                <ul className="site-admin-detail-list__info">
                                    {user.siteAdmin && (
                                        <li>
                                            <strong>Site admin</strong>
                                        </li>
                                    )}
                                    {user.email && (
                                        <li>
                                            Email: <a href={`mailto:${user.email}`}>{user.email}</a>
                                        </li>
                                    )}
                                    <li>ID: {user.id}</li>
                                    {user.createdAt && <li>Created: {format(user.createdAt, 'YYYY-MM-DD')}</li>}
                                    {user.orgs &&
                                        user.orgs.length > 0 && (
                                            <li>Orgs: {user.orgs.map(org => org.name).join(', ')}</li>
                                        )}
                                    {user.latestSettings && (
                                        <li>
                                            <SettingsInfo
                                                settings={user.latestSettings}
                                                filename={`user-settings-${user.id}.json`}
                                            />
                                        </li>
                                    )}
                                    {user.tags &&
                                        user.tags.length > 0 && (
                                            <li>Tags: {user.tags.map(tag => tag.name).join(', ')}</li>
                                        )}
                                </ul>
                                <div>
                                    {userActions.get(user.id)}
                                    {this.state.userErrorDescription.has(user.id) && (
                                        <p className="site-admin-detail-list__error">
                                            {this.state.userErrorDescription.get(user.id)}
                                        </p>
                                    )}
                                </div>
                            </li>
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

    private setSiteAdmin(user: GQL.IUser, siteAdmin: boolean): void {
        if (
            !window.confirm(
                siteAdmin
                    ? `Really promote user ${user.username} to site admin?`
                    : `Really revoke site admin status from user ${user.username}?`
            )
        ) {
            return
        }

        this.state.userErrorDescription.delete(user.id)
        this.state.userLoading.add(user.id)
        setUserIsSiteAdmin(user.id, siteAdmin)
            .toPromise()
            .then(
                () => {
                    this.state.userLoading.delete(user.id)
                    // Patch state locally.
                    this.setState({
                        users: this.state.users!.map(u => {
                            if (u.id === user.id) {
                                return { ...u, siteAdmin }
                            }
                            return u
                        }),
                    })
                },
                err => {
                    this.state.userLoading.delete(user.id)
                    this.state.userErrorDescription.set(user.id, err.message)
                    this.forceUpdate()
                }
            )
    }
}
