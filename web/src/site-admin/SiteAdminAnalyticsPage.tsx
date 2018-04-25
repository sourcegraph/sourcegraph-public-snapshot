import Loader from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs'
import * as GQL from '../backend/graphqlschema'
import { PageTitle } from '../components/PageTitle'
import { Timestamp } from '../components/time/Timestamp'
import { eventLogger } from '../tracking/eventLogger'
import { fetchUserAndSiteAnalytics } from './backend'

interface Props extends RouteComponentProps<any> {}

export interface State {
    users?: GQL.IUser[]
    siteActivity?: GQL.ISiteActivity
    error?: Error
}

const showExpandedAnalytics = localStorage.getItem('showExpandedAnalytics') !== null

/**
 * A page displaying usage analytics for the site.
 */
export class SiteAdminAnalyticsPage extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminAnalytics')

        this.subscriptions.add(
            fetchUserAndSiteAnalytics().subscribe(
                ({ users, siteActivity }) => this.setState({ users: users || undefined, siteActivity }),
                error => this.setState({ error })
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-analytics-page">
                <PageTitle title="Analytics - Admin" />
                <h2>Analytics</h2>
                {this.state.error && <p className="site-admin-analytics-page__error">{this.state.error.message}</p>}
                <table className="table table-hover">
                    <thead>
                        <tr>
                            <th>User</th>
                            <th>Page views</th>
                            <th>Search queries</th>
                            <th>Code intelligence actions</th>
                            <th className="site-admin-analytics-page__date-column">Last active</th>
                            {showExpandedAnalytics && (
                                <th className="site-admin-analytics-page__date-column">
                                    Last active in code host or code review
                                </th>
                            )}
                        </tr>
                    </thead>
                    <tbody>
                        {!this.state.users && (
                            <tr>
                                <td colSpan={5}>
                                    <Loader className="icon-inline" />
                                </td>
                            </tr>
                        )}
                        {this.state.users &&
                            this.state.users.map(user => (
                                <tr key={user.id}>
                                    <td>{user.username}</td>
                                    <td>{user.activity ? user.activity.pageViews : '?'}</td>
                                    <td>{user.activity ? user.activity.searchQueries : '?'}</td>
                                    <td>{user.activity ? user.activity.codeIntelligenceActions : '?'}</td>
                                    <td className="site-admin-analytics-page__date-column">
                                        {user.activity && user.activity.lastActiveTime ? (
                                            <Timestamp date={user.activity.lastActiveTime} />
                                        ) : (
                                            '?'
                                        )}
                                    </td>
                                    {showExpandedAnalytics && (
                                        <td className="site-admin-analytics-page__date-column">
                                            {user.activity && user.activity.lastActiveCodeHostIntegrationTime ? (
                                                <Timestamp date={user.activity.lastActiveCodeHostIntegrationTime} />
                                            ) : (
                                                '?'
                                            )}
                                        </td>
                                    )}
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
                                <td>
                                    {this.state.users.reduce(
                                        (c, v) => c + (v.activity ? v.activity.codeIntelligenceActions : 0),
                                        0
                                    )}
                                </td>
                                <td className="site-admin-analytics-page__date-column" />
                                {showExpandedAnalytics && <td className="site-admin-analytics-page__date-column" />}
                            </tr>
                        </tfoot>
                    )}
                </table>
            </div>
        )
    }
}
