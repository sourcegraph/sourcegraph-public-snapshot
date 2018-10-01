import format from 'date-fns/format'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs'
import * as GQL from '../backend/graphqlschema'
import { BarChart } from '../components/d3/BarChart'
import { FilteredConnection, FilteredConnectionFilter } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { RadioButtons } from '../components/RadioButtons'
import { Timestamp } from '../components/time/Timestamp'
import { eventLogger } from '../tracking/eventLogger'
import { fetchSiteAnalytics, fetchUserAnalytics } from './backend'

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

const CHART_ID_KEY = 'latest-analytics-chart-id'

interface UserActivityHeaderFooterProps {
    nodes: GQL.IUser[]
}

class UserActivityHeader extends React.PureComponent<UserActivityHeaderFooterProps> {
    public render(): JSX.Element | null {
        return (
            <thead>
                <tr>
                    <th>User</th>
                    <th>Page views</th>
                    <th>Search queries</th>
                    <th>Code intelligence actions</th>
                    <th className="site-admin-analytics-page__date-column">Last active</th>
                    <th className="site-admin-analytics-page__date-column">Last active in code host or code review</th>
                </tr>
            </thead>
        )
    }
}

class UserActivityFooter extends React.PureComponent<UserActivityHeaderFooterProps> {
    public render(): JSX.Element | null {
        return (
            <tfoot>
                <tr>
                    <th>Total</th>
                    <td>{this.props.nodes.reduce((c, v) => c + (v.activity ? v.activity.pageViews : 0), 0)}</td>
                    <td>{this.props.nodes.reduce((c, v) => c + (v.activity ? v.activity.searchQueries : 0), 0)}</td>
                    <td>
                        {this.props.nodes.reduce(
                            (c, v) => c + (v.activity ? v.activity.codeIntelligenceActions : 0),
                            0
                        )}
                    </td>
                    <td className="site-admin-analytics-page__date-column" />
                    <td className="site-admin-analytics-page__date-column" />
                </tr>
            </tfoot>
        )
    }
}

interface UserActivityNodeProps {
    /**
     * The user to display in this list item.
     */
    node: GQL.IUser
}

class UserActivityNode extends React.PureComponent<UserActivityNodeProps> {
    public render(): JSX.Element | null {
        return (
            <tr>
                <td>{this.props.node.username}</td>
                <td>{this.props.node.activity ? this.props.node.activity.pageViews : 'n/a'}</td>
                <td>{this.props.node.activity ? this.props.node.activity.searchQueries : 'n/a'}</td>
                <td>{this.props.node.activity ? this.props.node.activity.codeIntelligenceActions : 'n/a'}</td>
                <td className="site-admin-analytics-page__date-column">
                    {this.props.node.activity && this.props.node.activity.lastActiveTime ? (
                        <Timestamp date={this.props.node.activity.lastActiveTime} />
                    ) : (
                        'n/a'
                    )}
                </td>
                <td className="site-admin-analytics-page__date-column">
                    {this.props.node.activity && this.props.node.activity.lastActiveCodeHostIntegrationTime ? (
                        <Timestamp date={this.props.node.activity.lastActiveCodeHostIntegrationTime} />
                    ) : (
                        'n/a'
                    )}
                </td>
            </tr>
        )
    }
}

class FilteredUserConnection extends FilteredConnection<GQL.IUser, {}> {}
export const USER_ACTIVITY_FILTERS: FilteredConnectionFilter[] = [
    {
        label: 'All users',
        id: 'all',
        tooltip: 'Show all users',
        args: { activePeriod: GQL.UserActivePeriod.ALL_TIME },
    },
    {
        label: 'Active today',
        id: 'today',
        tooltip: 'Show users active since this morning at 00:00 UTC',
        args: { activePeriod: GQL.UserActivePeriod.TODAY },
    },
    {
        label: 'Active this week',
        id: 'week',
        tooltip: 'Show users active since Monday at 00:00 UTC',
        args: { activePeriod: GQL.UserActivePeriod.THIS_WEEK },
    },
    {
        label: 'Active this month',
        id: 'month',
        tooltip: 'Show users active since the first day of the month at 00:00 UTC',
        args: { activePeriod: GQL.UserActivePeriod.THIS_MONTH },
    },
]

interface SiteAdminAnalyticsPageProps extends RouteComponentProps<any> {
    isLightTheme: boolean
}

interface SiteAdminAnalyticsPageState {
    users?: GQL.IUserConnection
    siteActivity?: GQL.ISiteActivity
    error?: Error
    chartID: keyof ChartOptions
}

/**
 * A page displaying usage analytics for the site.
 */
export class SiteAdminAnalyticsPage extends React.Component<SiteAdminAnalyticsPageProps, SiteAdminAnalyticsPageState> {
    public state: SiteAdminAnalyticsPageState = {
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
            fetchSiteAnalytics().subscribe(
                siteActivity => this.setState({ siteActivity }),
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
                {this.state.error && <p className="alert alert-danger">{upperFirst(this.state.error.message)}</p>}
                {this.state.siteActivity && (
                    <>
                        <RadioButtons
                            nodes={Object.entries(chartGeneratorOptions).map(([key, opt]) => ({
                                label: opt.label,
                                id: key,
                            }))}
                            onChange={this.onChartIndexChange}
                            selected={this.state.chartID}
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
                {!this.state.error && (
                    <FilteredUserConnection
                        listComponent="table"
                        className="table"
                        hideSearch={false}
                        filters={USER_ACTIVITY_FILTERS}
                        noShowMore={false}
                        noun="user"
                        pluralNoun="users"
                        queryConnection={fetchUserAnalytics}
                        nodeComponent={UserActivityNode}
                        headComponent={UserActivityHeader}
                        footComponent={UserActivityFooter}
                        history={this.props.history}
                        location={this.props.location}
                    />
                )}
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
}
