import * as React from 'react'
import { useCallback, useMemo, useState, useEffect } from 'react'

import { endOfDay, endOfMonth, endOfWeek } from 'date-fns'
import FileDownloadIcon from 'mdi-react/FileDownloadIcon'
import { RouteComponentProps } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { UserActivePeriod } from '@sourcegraph/shared/src/graphql-operations'
import * as GQL from '@sourcegraph/shared/src/schema'
import {
    Button,
    Icon,
    H2,
    H3,
    Card,
    useObservableWithStatus,
    Tooltip,
    Tabs,
    TabList,
    Tab,
    TabPanels,
    TabPanel,
} from '@sourcegraph/wildcard'

import { getLineColor, LegendItem, LegendList, LineChart, ParentSize, Series } from '../../charts'
import { FilteredConnection, FilteredConnectionFilter } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { RadioButtons } from '../../components/RadioButtons'
import { Timestamp } from '../../components/time/Timestamp'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchSiteUsageStatistics, fetchUserUsageStatistics } from '../backend'

import styles from '../SiteAdminUsageStatisticsPage.module.scss'

interface ChartData {
    label: string
    getXValue: (datum: StandardDatum) => Date
    getYValue: (datum: StandardDatum) => number | null
}

type ChartOptions = Record<'daus' | 'waus' | 'maus', ChartData>

const chartGeneratorOptions: ChartOptions = {
    daus: {
        label: 'Daily unique users',
        getXValue: (datum: StandardDatum): Date => endOfDay(new Date(datum.x)),
        getYValue: (datum: StandardDatum): number | null => datum.value,
    },
    waus: {
        label: 'Weekly unique users',
        getXValue: (datum: StandardDatum): Date => endOfWeek(new Date(datum.x)),
        getYValue: (datum: StandardDatum): number | null => datum.value,
    },
    maus: {
        label: 'Monthly unique users',
        getXValue: (datum: StandardDatum): Date => endOfMonth(new Date(datum.x)),
        getYValue: (datum: StandardDatum): number | null => datum.value,
    },
}

const CHART_ID_KEY = 'latest-usage-statistics-chart-id'

interface UsageChartPageProps {
    isLightTheme: boolean
    stats: GQL.ISiteUsageStatistics
    chartID: keyof ChartOptions
    header?: JSX.Element
    showLegend?: boolean
}
interface StandardDatum {
    value: number | null
    x: string
}

const UsageChart: React.FunctionComponent<UsageChartPageProps> = props => {
    const series = useMemo(
        (): Series<StandardDatum>[] =>
            [
                {
                    name: 'Deleted or anonymous',
                    data: props.stats[props.chartID].map(({ startTime: x, anonymousUserCount: value }) => ({
                        x,
                        value,
                    })),
                    color: 'var(--blue)',
                },
                {
                    name: 'Registered',
                    data: props.stats[props.chartID].map(({ startTime: x, registeredUserCount: value }) => ({
                        x,
                        value,
                    })),
                    color: 'var(--green)',
                },
            ].map(({ name, data, color }) => ({
                id: name,
                name,
                color,
                getXValue: chartGeneratorOptions[props.chartID].getXValue,
                getYValue: chartGeneratorOptions[props.chartID].getYValue,
                data,
            })),
        [props.chartID, props.stats]
    )
    return (
        <div>
            <Card className="p-2">
                <div className="d-flex justify-content-between align-items-center">
                    {props.header || <H3>{chartGeneratorOptions[props.chartID].label}</H3>}
                    <small>
                        <i>GMT/UTC time</i>
                    </small>
                </div>
                <ParentSize>
                    {({ width }) => (
                        <LineChart
                            width={width}
                            height={400}
                            series={series}
                            stacked={true}
                            isSeriesSelected={() => true}
                            isSeriesHovered={() => true}
                        />
                    )}
                </ParentSize>
                <LegendList className="d-flex justify-content-center my-3">
                    {series.map(line => (
                        <LegendItem key={line.id} color={getLineColor(line)} name={line.name} />
                    ))}
                </LegendList>
            </Card>
        </div>
    )
}

interface UserUsageStatisticsHeaderFooterProps {
    nodes: GQL.IUser[]
}

const UserUsageStatisticsHeader = React.memo(function UserUsageStatisticsHeader() {
    return (
        <thead>
            <tr>
                <th className={styles.headerColumn}>User</th>
                <th className={styles.headerColumn}>Page views</th>
                <th className={styles.headerColumn}>Search queries</th>
                <th className={styles.headerColumn}>Code&#160;intelligence actions</th>
                <th className={styles.dateColumn}>Last active</th>
                <th className={styles.dateColumn}>Last active in code host or code review</th>
            </tr>
        </thead>
    )
})

const UserUsageStatisticsFooter = React.memo(function UserUsageStatisticsFooter(
    props: UserUsageStatisticsHeaderFooterProps
) {
    return (
        <tfoot>
            <tr>
                <th>Total</th>
                <td>
                    {props.nodes.reduce(
                        (count, node) => count + (node.usageStatistics ? node.usageStatistics.pageViews : 0),
                        0
                    )}
                </td>
                <td>
                    {props.nodes.reduce(
                        (count, node) => count + (node.usageStatistics ? node.usageStatistics.searchQueries : 0),
                        0
                    )}
                </td>
                <td>
                    {props.nodes.reduce(
                        (count, node) =>
                            count + (node.usageStatistics ? node.usageStatistics.codeIntelligenceActions : 0),
                        0
                    )}
                </td>
                <td className={styles.dateColumn} />
                <td className={styles.dateColumn} />
            </tr>
        </tfoot>
    )
})

interface UserUsageStatisticsNodeProps {
    /**
     * The user to display in this list item.
     */
    node: GQL.IUser
}

const UserUsageStatisticsNode = React.memo(function UserUsageStatisticsNode(props: UserUsageStatisticsNodeProps) {
    return (
        <tr>
            <td>{props.node.username}</td>
            <td>{props.node.usageStatistics ? props.node.usageStatistics.pageViews : 'n/a'}</td>
            <td>{props.node.usageStatistics ? props.node.usageStatistics.searchQueries : 'n/a'}</td>
            <td>{props.node.usageStatistics ? props.node.usageStatistics.codeIntelligenceActions : 'n/a'}</td>
            <td className={styles.dateColumn}>
                {props.node.usageStatistics?.lastActiveTime ? (
                    <Timestamp date={props.node.usageStatistics.lastActiveTime} />
                ) : (
                    'never'
                )}
            </td>
            <td className={styles.dateColumn}>
                {props.node.usageStatistics?.lastActiveCodeHostIntegrationTime ? (
                    <Timestamp date={props.node.usageStatistics.lastActiveCodeHostIntegrationTime} />
                ) : (
                    'never'
                )}
            </td>
        </tr>
    )
})

export const USER_ACTIVITY_FILTERS: FilteredConnectionFilter[] = [
    {
        label: '',
        type: 'select',
        id: 'user-activity-filters',
        values: [
            {
                label: 'All users',
                value: 'all',
                tooltip: 'Show all users',
                args: { activePeriod: UserActivePeriod.ALL_TIME },
            },
            {
                label: 'Active today',
                value: 'today',
                tooltip: 'Show users active since this morning at 00:00 UTC',
                args: { activePeriod: UserActivePeriod.TODAY },
            },
            {
                label: 'Active this week',
                value: 'week',
                tooltip: 'Show users active since Monday at 00:00 UTC',
                args: { activePeriod: UserActivePeriod.THIS_WEEK },
            },
            {
                label: 'Active this month',
                value: 'month',
                tooltip: 'Show users active since the first day of the month at 00:00 UTC',
                args: { activePeriod: UserActivePeriod.THIS_MONTH },
            },
        ],
    },
]

const TabSummary: React.FunctionComponent<AdvancedStatisticsPageProps> = props => {
    const [chartID, setChartID] = useState<keyof ChartOptions>(
        useMemo(() => {
            const latest = localStorage.getItem(CHART_ID_KEY)
            return latest && latest in chartGeneratorOptions ? (latest as keyof ChartOptions) : 'daus'
        }, [])
    )
    useEffect(() => {
        eventLogger.logViewEvent('SiteAdminUsageStatistics')
    }, [])
    useEffect(() => {
        localStorage.setItem(CHART_ID_KEY, chartID)
    }, [chartID])

    const [stats, , error] = useObservableWithStatus(useMemo(() => fetchSiteUsageStatistics(), []))

    const onChartIndexChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        switch (event.target.value as keyof ChartOptions) {
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
        setChartID(event.target.value as keyof ChartOptions)
    }, [])

    return (
        <div>
            <PageTitle title="Usage statistics - Admin" />
            <H2>Usage statistics</H2>
            {error && <ErrorAlert className="mb-3" error={error} />}

            <Tooltip content="Download usage stats archive">
                <Button href="/site-admin/usage-statistics/archive" download="true" variant="secondary" as="a">
                    <Icon as={FileDownloadIcon} aria-hidden={true} /> Download usage stats archive
                </Button>
            </Tooltip>

            {stats && (
                <>
                    <RadioButtons
                        nodes={Object.entries(chartGeneratorOptions).map(([key, { label }]) => ({
                            label,
                            id: key,
                        }))}
                        name="chart-options"
                        onChange={onChartIndexChange}
                        selected={chartID}
                    />
                    <UsageChart {...props} chartID={chartID} stats={stats} />
                </>
            )}
            <H3 className="mt-4">All registered users</H3>
            {!error && (
                <FilteredConnection
                    listComponent="table"
                    className="table"
                    hideSearch={false}
                    filters={USER_ACTIVITY_FILTERS}
                    noShowMore={false}
                    noun="user"
                    pluralNoun="users"
                    queryConnection={fetchUserUsageStatistics}
                    nodeComponent={UserUsageStatisticsNode}
                    headComponent={UserUsageStatisticsHeader}
                    footComponent={UserUsageStatisticsFooter}
                    history={props.history}
                    location={props.location}
                />
            )}
        </div>
    )
}

interface AdvancedStatisticsPageProps extends RouteComponentProps<{}> {
    isLightTheme: boolean
}

/**
 * A page displaying usage statistics for the site.
 */
export const AdvancedStatisticsPage: React.FunctionComponent<AdvancedStatisticsPageProps> = props => (
    <Tabs lazy={true} behavior="memoize" size="large">
        <TabList>
            <Tab>Summary</Tab>
            <Tab>Batch Changes</Tab>
            <Tab>Insights</Tab>
            <Tab>Notebooks</Tab>
            <Tab>Monitors</Tab>
        </TabList>
        <TabPanels>
            <TabPanel>
                <TabSummary {...props} />
            </TabPanel>
            <TabPanel>TODO:</TabPanel>
            <TabPanel>TODO:</TabPanel>
            <TabPanel>TODO:</TabPanel>
            <TabPanel>TODO:</TabPanel>
        </TabPanels>
    </Tabs>
)
