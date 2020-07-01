import format from 'date-fns/format'
import React, { useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../shared/src/graphql/schema'
import { FilteredConnection, FilteredConnectionFilter } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { fetchSiteUsageStatistics, fetchUserUsageStatistics } from './backend'
import { ErrorAlert } from '../components/alerts'
import FileDownloadIcon from 'mdi-react/FileDownloadIcon'
import { useObservable } from '../../../shared/src/util/useObservable'
import { catchError, startWith } from 'rxjs/operators'
import { asError, isErrorLike } from '../../../shared/src/util/errors'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { snakeCase } from 'lodash'
import { Observable } from 'rxjs'
import { ChartViewContent } from '../views/ChartViewContent'
import { LineChartContent, BarChartContent } from 'sourcegraph'
import { Tab, TabsWithLocalStorageViewStatePersistence } from '../../../shared/src/components/Tabs'
import { Timestamp } from '../components/time/Timestamp'

interface UserUsageStatisticsHeaderFooterProps {
    nodes: GQL.IUser[]
}

class UserUsageStatisticsHeader extends React.PureComponent<UserUsageStatisticsHeaderFooterProps> {
    public render(): JSX.Element | null {
        return (
            <thead>
                <tr>
                    <th>User</th>
                    <th>Page views</th>
                    <th>Search queries</th>
                    <th>Code intelligence actions</th>
                    <th className="site-admin-usage-statistics-page__date-column">Last active</th>
                    <th className="site-admin-usage-statistics-page__date-column">
                        Last active in code host or code review
                    </th>
                </tr>
            </thead>
        )
    }
}

class UserUsageStatisticsFooter extends React.PureComponent<UserUsageStatisticsHeaderFooterProps> {
    public render(): JSX.Element | null {
        return (
            <tfoot>
                <tr>
                    <th>Total</th>
                    <td>
                        {this.props.nodes.reduce(
                            (count, node) => count + (node.usageStatistics ? node.usageStatistics.pageViews : 0),
                            0
                        )}
                    </td>
                    <td>
                        {this.props.nodes.reduce(
                            (count, node) => count + (node.usageStatistics ? node.usageStatistics.searchQueries : 0),
                            0
                        )}
                    </td>
                    <td>
                        {this.props.nodes.reduce(
                            (count, node) =>
                                count + (node.usageStatistics ? node.usageStatistics.codeIntelligenceActions : 0),
                            0
                        )}
                    </td>
                    <td className="site-admin-usage-statistics-page__date-column" />
                    <td className="site-admin-usage-statistics-page__date-column" />
                </tr>
            </tfoot>
        )
    }
}

interface UserUsageStatisticsNodeProps {
    /**
     * The user to display in this list item.
     */
    node: GQL.IUser
}

class UserUsageStatisticsNode extends React.PureComponent<UserUsageStatisticsNodeProps> {
    public render(): JSX.Element | null {
        return (
            <tr>
                <td>{this.props.node.username}</td>
                <td>{this.props.node.usageStatistics ? this.props.node.usageStatistics.pageViews : 'n/a'}</td>
                <td>{this.props.node.usageStatistics ? this.props.node.usageStatistics.searchQueries : 'n/a'}</td>
                <td>
                    {this.props.node.usageStatistics ? this.props.node.usageStatistics.codeIntelligenceActions : 'n/a'}
                </td>
                <td className="site-admin-usage-statistics-page__date-column">
                    {this.props.node.usageStatistics?.lastActiveTime ? (
                        <Timestamp date={this.props.node.usageStatistics.lastActiveTime} />
                    ) : (
                        'never'
                    )}
                </td>
                <td className="site-admin-usage-statistics-page__date-column">
                    {this.props.node.usageStatistics?.lastActiveCodeHostIntegrationTime ? (
                        <Timestamp date={this.props.node.usageStatistics.lastActiveCodeHostIntegrationTime} />
                    ) : (
                        'never'
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

export interface UsageStatistics extends GQL.ISiteUsageStatistics {
    userCount: number
    repositoryCount: number
}

interface SiteAdminUsageStatisticsPageProps extends RouteComponentProps<{}> {
    isLightTheme: boolean
    fetchSiteStatistics?: () => Observable<UsageStatistics>
    fetchUserStatistics?: () => Observable<GQL.IUserConnection>
}

interface KPI {
    label: string
    value: string
}

const HighlightedKPIs: React.FunctionComponent<{ highlights: KPI[] }> = ({ highlights }) => (
    <div className="site-admin-usage-statistics-page__kpi-container">
        {highlights.map(({ label, value }) => (
            <div className="site-admin-usage-statistics-page__kpi" key={snakeCase(label)}>
                <span className="site-admin-usage-statistics-page__kpi-value">{value}</span>
                <br />
                <span className="site-admin-usage-statistics-page__kpi-label">{label}</span>
            </div>
        ))}
    </div>
)

const getUsageKPIs = ({ waus, daus }: UsageStatistics): KPI[] => [
    {
        label: 'Total searches (last 90 days)',
        value: waus.reduce((totalCount, { searchActionCount }) => totalCount + searchActionCount, 0).toLocaleString(),
    },
    {
        label: 'Seaches per weekday',
        value: Math.floor(
            daus.reduce((totalCount, { searchActionCount }) => totalCount + searchActionCount, 0) / daus.length
        ).toLocaleString(),
    },
    {
        label: 'Code intel events per weekday',
        value: Math.floor(
            daus.reduce((totalCount, { codeIntelligenceActionCount }) => totalCount + codeIntelligenceActionCount, 0) /
                daus.length
        ).toLocaleString(),
    },
]

const getUsageChartContent = ({ waus }: UsageStatistics): BarChartContent<GQL.ISiteUsagePeriod, 'startTime'> => ({
    chart: 'bar',
    data: waus.reverse().map(({ startTime, ...rest }) => ({
        startTime: format(Date.parse(startTime) + 1000 * 60 * 60 * 24, 'MM/dd/yy'),
        ...rest,
    })),
    series: [
        {
            dataKey: 'searchActionCount',
            name: 'Search',
            fill: 'var(--primary)',
        },
        {
            dataKey: 'codeIntelligenceActionCount',
            name: 'Code Intelligence',
            fill: 'var(--secondary)',
        },
    ],
    xAxis: {
        dataKey: 'startTime',
        type: 'category',
    },
})

const getFeatureKPIs = ({ userCount, mergedCampaignChangesets, repositoryCount }: UsageStatistics): KPI[] => [
    {
        label: 'Repositories',
        value: repositoryCount.toLocaleString(),
    },
    {
        label: 'Campaign changesets merged',
        value: mergedCampaignChangesets.toLocaleString(),
    },
    {
        label: 'User accounts',
        value: userCount.toLocaleString(),
    },
]

const getFeatureChartContent = ({ waus }: UsageStatistics): LineChartContent<GQL.ISiteUsagePeriod, 'startTime'> => ({
    chart: 'line',
    data: waus.reverse().map(({ startTime, ...rest }) => ({
        startTime: format(Date.parse(startTime) + 1000 * 60 * 60 * 24, 'E, MMM d'),
        ...rest,
    })),
    series: [
        {
            dataKey: 'userCount',
            name: 'Total',
            stroke: 'var(--info)',
        },
        {
            dataKey: 'integrationUserCount',
            name: 'Code host integrations',
            stroke: 'var(--success)',
        },
        {
            dataKey: 'codeIntelligenceUserCount',
            name: 'Code intelligence',
            stroke: 'var(--warning)',
        },
        {
            dataKey: 'searchUserCount',
            name: 'Search',
            stroke: 'var(--primary)',
        },
    ],
    xAxis: {
        dataKey: 'startTime',
        type: 'category',
    },
})

/**
 * A page displaying usage statistics for the site.
 */
export const SiteAdminUsageStatisticsPage: React.FunctionComponent<SiteAdminUsageStatisticsPageProps> = ({
    fetchSiteStatistics = fetchSiteUsageStatistics,
    fetchUserStatistics = fetchUserUsageStatistics,
    ...props
}) => {
    useEffect(() => {
        eventLogger.logViewEvent('SiteAdminUsageStatistics')
    })
    const statsOrError = useObservable(
        useMemo(
            () =>
                fetchSiteStatistics().pipe(
                    startWith('loading' as const),
                    catchError(error => [asError(error)])
                ),
            [fetchSiteStatistics]
        )
    )
    if (!statsOrError) {
        return null
    }

    const tabs = ['Usage', 'Features', 'Activity log'].map((label): Tab<string> => ({ label, id: snakeCase(label) }))
    return (
        <div className="site-admin-usage-statistics-page">
            <PageTitle title="Usage statistics - Admin" />

            {statsOrError === 'loading' ? (
                <LoadingSpinner />
            ) : isErrorLike(statsOrError) ? (
                <ErrorAlert className="mb-3" error={statsOrError} history={props.history} />
            ) : (
                <>
                    <TabsWithLocalStorageViewStatePersistence tabs={tabs} storageKey="activeUsageStatisticsTab">
                        <div key="usage">
                            <HighlightedKPIs highlights={getUsageKPIs(statsOrError)} />
                            <h3>Total actions per week</h3>
                            <div className="site-admin-usage-statistics-page__usage-chart">
                                <ChartViewContent {...props} content={getUsageChartContent(statsOrError)} />
                            </div>
                        </div>
                        <div key="features">
                            <HighlightedKPIs highlights={getFeatureKPIs(statsOrError)} />
                            <h3>Active users by feature</h3>
                            <div className="site-admin-usage-statistics-page__usage-chart">
                                <ChartViewContent {...props} content={getFeatureChartContent(statsOrError)} />
                            </div>
                        </div>
                        <div key="activity_log">
                            <h3 className="mt-4">All registered users</h3>
                            <FilteredUserConnection
                                {...props}
                                listComponent="table"
                                className="table"
                                hideSearch={false}
                                filters={USER_ACTIVITY_FILTERS}
                                noShowMore={false}
                                noun="user"
                                pluralNoun="users"
                                queryConnection={fetchUserStatistics}
                                nodeComponent={UserUsageStatisticsNode}
                                headComponent={UserUsageStatisticsHeader}
                                footComponent={UserUsageStatisticsFooter}
                            />
                        </div>
                    </TabsWithLocalStorageViewStatePersistence>
                    <a
                        href="/site-admin/usage-statistics/archive"
                        data-tooltip="Download usage stats archive"
                        download="true"
                    >
                        <FileDownloadIcon className="icon-inline" /> Download usage stats archive
                    </a>
                </>
            )}
        </div>
    )
}
