import { useMemo, useEffect, FC } from 'react'

import classNames from 'classnames'
import { startCase } from 'lodash'
import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { AlertType } from '@sourcegraph/shared/src/graphql-operations'
import { Card, LoadingSpinner, useMatchMedia, Text, LineChart, BarChart, Series } from '@sourcegraph/wildcard'

import { GlobalAlert } from '../../../global/GlobalAlert'
import { UsersStatisticsResult, UsersStatisticsVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { ChartContainer } from '../components/ChartContainer'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { ToggleSelect } from '../components/ToggleSelect'
import { ValueLegendList, ValueLegendListProps } from '../components/ValueLegendList'
import { useChartFilters } from '../useChartFilters'
import { StandardDatum, FrequencyDatum, buildFrequencyDatum } from '../utils'

import { USERS_STATISTICS } from './queries'

import styles from './AnalyticsUsersPage.module.scss'

export const AnalyticsUsersPage: FC<RouteComponentProps> = () => {
    const { dateRange, aggregation, grouping } = useChartFilters({ name: 'Users', aggregation: 'registeredUsers' })
    const { data, error, loading } = useQuery<UsersStatisticsResult, UsersStatisticsVariables>(USERS_STATISTICS, {
        variables: {
            dateRange: dateRange.value,
            grouping: grouping.value,
        },
    })
    useEffect(() => {
        eventLogger.logPageView('AdminAnalyticsUsers')
    }, [])
    const [frequencies, legends] = useMemo(() => {
        if (!data) {
            return []
        }
        const { users } = data.site.analytics
        const legends: ValueLegendListProps['items'] = [
            {
                value: users.activity.summary.totalRegisteredUsers,
                description: 'Active users',
                color: 'var(--purple)',
                tooltip: 'Currently registered users using the application in the selected timeframe.',
            },
            {
                value: data.users.totalCount,
                description: 'Registered Users',
                color: 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of users who have created an account.',
            },
            {
                value: data.site.productSubscription.license?.userCount ?? 0,
                description: 'User licenses',
                color: 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of user licenses your current account is provisioned for.',
            },
        ]

        const frequencies: FrequencyDatum[] = buildFrequencyDatum(users.frequencies, 1, 30)

        return [frequencies, legends]
    }, [data])

    const activities = useMemo(() => {
        if (!data) {
            return []
        }
        const { users } = data.site.analytics
        const activities: Series<StandardDatum>[] = [
            {
                id: 'activity',
                name: aggregation.selected === 'count' ? 'Activities' : 'Active users',
                color: aggregation.selected === 'count' ? 'var(--cyan)' : 'var(--purple)',
                data: users.activity.nodes.map(
                    node => ({
                        date: new Date(node.date),
                        value: node[aggregation.selected],
                    }),
                    dateRange.value
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
        ]

        return activities
    }, [data, aggregation.selected, dateRange.value])

    const summary = useMemo(() => {
        if (!data) {
            return []
        }
        const { avgDAU, avgWAU, avgMAU } = data.site.analytics.users.summary
        return [
            {
                value: avgDAU,
                label: 'DAU',
            },
            {
                value: avgWAU,
                label: 'WAU',
            },
            {
                value: avgMAU,
                label: 'MAU',
            },
        ]
    }, [data])

    const isWideScreen = useMatchMedia('(min-width: 992px)', false)

    if (error) {
        throw error
    }

    if (loading) {
        return <LoadingSpinner />
    }

    const groupingLabel = startCase(grouping.value.toLowerCase())

    return (
        <>
            <AnalyticsPageTitle>Users</AnalyticsPageTitle>
            <Card className="p-3">
                <div className="d-flex justify-content-end align-items-stretch mb-2 text-nowrap">
                    <HorizontalSelect<typeof dateRange.value> {...dateRange} />
                </div>
                <GlobalAlert
                    alert={{
                        message:
                            'Note these charts are experimental. For billing information, use [usage stats](/site-admin/usage-statistics).',
                        type: AlertType.INFO,
                        isDismissibleWithKey: '',
                    }}
                    className="my-3"
                />
                {legends && <ValueLegendList className="mb-3" items={legends} />}
                {activities && (
                    <div>
                        <ChartContainer
                            title={
                                aggregation.selected === 'count'
                                    ? `${groupingLabel} activity`
                                    : `${groupingLabel} unique users`
                            }
                            labelX="Time"
                            labelY={aggregation.selected === 'count' ? 'Activity' : 'Unique users'}
                        >
                            {width => <LineChart width={width} height={300} series={activities} />}
                        </ChartContainer>
                        <div className="d-flex justify-content-end align-items-stretch mb-4 text-nowrap">
                            <HorizontalSelect<typeof grouping.value> {...grouping} className="mr-4" />
                            <ToggleSelect<typeof aggregation.selected> {...aggregation} />
                        </div>
                    </div>
                )}
                <div className={classNames(isWideScreen && 'd-flex')}>
                    {summary && (
                        <ChartContainer
                            title="Average user activity by period"
                            labelX="Average DAU/WAU/MAU"
                            labelY="Unique users"
                            className={classNames(styles.barChart, 'mb-5')}
                        >
                            {width => (
                                <BarChart
                                    width={isWideScreen ? 280 : width}
                                    height={300}
                                    data={summary}
                                    getDatumName={datum => datum.label}
                                    getDatumValue={datum => datum.value}
                                    getDatumColor={() => 'var(--bar-color)'}
                                    getDatumFadeColor={() => 'var(--bar-fade-color)'}
                                />
                            )}
                        </ChartContainer>
                    )}
                    {frequencies && (
                        <ChartContainer
                            title="Frequency of use"
                            labelX="Days used"
                            labelY="Unique users"
                            className={classNames(styles.barChart, 'mb-5')}
                        >
                            {width => (
                                <BarChart
                                    width={isWideScreen ? 540 : width}
                                    height={300}
                                    data={frequencies}
                                    pixelsPerXTick={20}
                                    getDatumName={datum => datum.label}
                                    getDatumValue={datum => datum.value}
                                    getDatumColor={() => 'var(--bar-color)'}
                                    getDatumFadeColor={() => 'var(--bar-fade-color)'}
                                />
                            )}
                        </ChartContainer>
                    )}
                </div>
            </Card>
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table and are updated every 24 hours..
            </Text>
        </>
    )
}
