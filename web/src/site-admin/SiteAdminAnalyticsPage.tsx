import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { fetchUserAnalytics } from './backend'

interface Props extends RouteComponentProps<any> {}

export interface State {
    users?: GQL.IUser[]
}

/**
 * A page displaying usage analytics for the site.
 */
export class SiteAdminAnalyticsPage extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminAnalytics')

        this.subscriptions.add(fetchUserAnalytics().subscribe(users => this.setState({ users: users || undefined })))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-analytics-page">
                <PageTitle title="Analytics - Admin" />
                <h2>Analytics</h2>
                <table className="table table-hover">
                    <thead>
                        <tr>
                            <th>User</th>
                            <th>Page views</th>
                            <th>Search queries</th>
                        </tr>
                    </thead>
                    <tbody>
                        {this.state.users &&
                            this.state.users.map(user => (
                                <tr key={user.id}>
                                    <td>{user.username}</td>
                                    <td>{user.activity ? user.activity.pageViews : '?'}</td>
                                    <td>{user.activity ? user.activity.searchQueries : '?'}</td>
                                </tr>
                            ))}
                    </tbody>
                    {this.state.users && (
                        <tfoot>
                            <tr>
                                <th>Total</th>
                                <td>
                                    {this.state.users.reduce((c, v) => c + (v.activity ? v.activity.pageViews : 0), 0)}
                                </td>
                                <td>
                                    {this.state.users.reduce(
                                        (c, v) => c + (v.activity ? v.activity.searchQueries : 0),
                                        0
                                    )}
                                </td>
                            </tr>
                        </tfoot>
                    )}
                </table>
            </div>
        )
    }
}
