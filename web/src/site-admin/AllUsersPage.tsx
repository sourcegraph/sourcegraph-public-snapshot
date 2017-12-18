import format from 'date-fns/format'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { pluralize } from '../util/strings'
import { fetchAllUsers } from './backend'
import { SettingsInfo } from './util/SettingsInfo'

interface Props extends RouteComponentProps<any> {}

export interface State {
    users?: GQL.IUser[]
}

/**
 * A page displaying the users on this site.
 */
export class AllUsersPage extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminAllUsers')

        this.subscriptions.add(fetchAllUsers().subscribe(users => this.setState({ users: users || undefined })))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-detail-list site-admin-all-users-page">
                <PageTitle title="Users - Admin" />
                <h2>Users</h2>
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
                                    {user.email && (
                                        <li>
                                            Email: <a href={`mailto:${user.email}`}>{user.email}</a>
                                        </li>
                                    )}
                                    {user.id && <li>ID: {user.id}</li>}
                                    {user.createdAt && <li>Created: {format(user.createdAt, 'YYYY-MM-DD')}</li>}
                                    {user.orgs && user.orgs.length ? (
                                        <li>Orgs: {user.orgs.map(org => org.name).join(', ')}</li>
                                    ) : (
                                        undefined
                                    )}
                                    {user.latestSettings && (
                                        <li>
                                            <SettingsInfo
                                                settings={user.latestSettings}
                                                filename={`user-settings-${user.id}.json`}
                                            />
                                        </li>
                                    )}
                                    {user.tags && user.tags.length ? (
                                        <li>Tags: {user.tags.map(tag => tag.name).join(', ')}</li>
                                    ) : (
                                        undefined
                                    )}
                                </ul>
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
}
