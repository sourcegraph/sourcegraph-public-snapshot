import Loader from '@sourcegraph/icons/lib/Loader'
import format from 'date-fns/format'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs'
import * as GQL from '../backend/graphqlschema'
import { BarChart } from '../components/d3/BarChart'
import { PageTitle } from '../components/PageTitle'
import { RadioButtonNode, RadioButtons } from '../components/RadioButtons'
import { Timestamp } from '../components/time/Timestamp'
import { eventLogger } from '../tracking/eventLogger'
import { fetchUserAndSiteAnalytics } from './backend'

interface Props extends RouteComponentProps<any> {
    isLightTheme: boolean
}

interface ChartData {
    label: string
    dateFormat: string
}

interface ChartOptions {
    daus: ChartData
    waus: ChartData
    maus: ChartData
}

const chartGeneratorOptions: ChartOptions = {
    daus: { label: 'Daily unique users', dateFormat: 'ddd, MMM D' },
    waus: { label: 'Weekly unique users', dateFormat: 'ddd, MMM D' },
    maus: { label: 'Monthly unique users', dateFormat: 'MMMM YYYY' },
}

export interface State {
    users?: GQL.IUser[]
    siteActivity?: GQL.ISiteActivity
    error?: Error
    chartID: keyof ChartOptions
}

const CHART_ID_KEY = 'latest-analytics-chart-id'

const showExpandedAnalytics = localStorage.getItem('showExpandedAnalytics') !== null

/**
 * A page displaying usage analytics for the site.
 */
export class SiteAdminAnalyticsPage extends React.Component<Props, State> {
    public state: State = {
        chartID: this.loadLatestChartFromStorage(),
    }

    private subscriptions = new Subscription()

    private loadLatestChartFromStorage(): keyof ChartOptions {
        const latest = localStorage.getItem(CHART_ID_KEY)
        return latest && latest in chartGeneratorOptions ? (latest as keyof ChartOptions) : 'daus'
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminAnalytics')

        this.subscriptions.add(
            fetchUserAndSiteAnalytics().subscribe(
                ({ users, siteActivity }) => this.setState({ users: users || undefined, siteActivity }),
                error => this.setState({ error })
            )
        )
    }

    public componentDidUpdate(): void {
        localStorage.setItem(CHART_ID_KEY, this.state.chartID)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const chart = chartGeneratorOptions[this.state.chartID]
        return (
            <div className="site-admin-analytics-page">
                <PageTitle title="Analytics - Admin" />
                <h2>Analytics</h2>
                {this.state.error && <p className="site-admin-analytics-page__error">{this.state.error.message}</p>}
                {this.state.siteActivity && (
                    <>
                        <RadioButtons
                            nodes={Object.entries(chartGeneratorOptions).map(([key, opt]) => ({
                                label: opt.label,
                                id: key,
                            }))}
                            onChange={this.onChartIndexChange}
                            checked={this.radioChecked}
                        />
                        {
                            <>
                                <h3>{chart.label}</h3>
                                <BarChart
                                    showLabels={true}
                                    showLegend={true}
                                    width={500}
                                    height={200}
                                    isLightTheme={this.props.isLightTheme}
                                    data={this.state.siteActivity[this.state.chartID].map(p => ({
                                        xLabel: format(Date.parse(p.startTime) + 1000 * 60 * 60 * 24, chart.dateFormat),
                                        yValues: {
                                            Registered: p.registeredUserCount,
                                            Anonymous: p.anonymousUserCount,
                                        },
                                    }))}
                                />
                                <small className="site-admin-analytics-page__tz-note">
                                    <i>GMT/UTC time</i>
                                </small>
                            </>
                        }
                    </>
                )}
                <h3 className="mt-4">All registered users</h3>
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

    private onChartIndexChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        switch (e.target.value as keyof ChartOptions) {
            case 'daus':
                eventLogger.log('DAUsChartSelected')
                break
            case 'waus':
                eventLogger.log('WAUsChartSelected')
                break
            case 'maus':
                eventLogger.log('MAUsChartSelected')
                break
        }
        this.setState({ chartID: e.target.value as keyof ChartOptions })
    }
    private radioChecked = (n: RadioButtonNode) => this.state.chartID === n.id
}
