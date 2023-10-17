import { useState, useMemo, useEffect, type FC } from 'react'

import classNames from 'classnames'
import { startCase } from 'lodash'

import { useQuery } from '@sourcegraph/http-client'
import { Card, LoadingSpinner, useMatchMedia, Text, LineChart, BarChart, type Series } from '@sourcegraph/wildcard'

import type { UsersStatisticsResult, UsersStatisticsVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { checkRequestAccessAllowed } from '../../../util/checkRequestAccessAllowed'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { ChartContainer } from '../components/ChartContainer'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { ToggleSelect } from '../components/ToggleSelect'
import { ValueLegendList, type ValueLegendListProps } from '../components/ValueLegendList'
import { useChartFilters } from '../useChartFilters'
import { type StandardDatum, type FrequencyDatum, buildFrequencyDatum } from '../utils'

import { USERS_STATISTICS } from './queries'

import styles from './AnalyticsUsersPage.module.scss'

export const AnalyticsUsersPage: FC = () => {
    const { dateRange, aggregation, grouping } = useChartFilters({ name: 'Users', aggregation: 'uniqueUsers' })
    const { data, error, loading } = useQuery<UsersStatisticsResult, UsersStatisticsVariables>(USERS_STATISTICS, {
        variables: {
            dateRange: dateRange.value,
            grouping: grouping.value,
        },
    })
    useEffect(() => {
        eventLogger.logPageView('AdminAnalyticsUsers')
    }, [])
    const [uniqueOrPercentage, setUniqueOrPercentage] = useState<'unique' | 'percentage'>('unique')

    const [frequencies, legends] = useMemo(() => {
        if (!data) {
            return []
        }
        const { users } = data.site.analytics
        let legends: ValueLegendListProps['items'] = [
            {
                value: users.activity.summary.totalUniqueUsers,
                description: 'Active users',
                color: 'var(--purple)',
                tooltip: 'The number of users using the application in the selected timeframe including deleted users.',
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

        const isRequestAccessAllowed = checkRequestAccessAllowed(window.context)
        if (isRequestAccessAllowed) {
            legends = [
                ...legends.slice(0, 1),
                {
                    value: data.pendingAccessRequests.totalCount,
                    description: 'Pending requests',
                    color: 'var(--cyan)',
                    position: 'right',
                    tooltip: 'The number of users who have requested access to your Sourcegraph instance.',
                },
                ...legends.slice(1),
            ]
        }

        const frequencies: FrequencyDatum[] = buildFrequencyDatum(users.frequencies, uniqueOrPercentage, 30)

        return [frequencies, legends]
    }, [data, uniqueOrPercentage])

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
                    {!!data?.site.analytics.users.monthlyActiveUsers && (
                        <ChartContainer
                            title="Monthly active users"
                            labelX="Months"
                            labelY="Unique users"
                            className={classNames(styles.barChart)}
                        >
                            {width => (
                                <BarChart
                                    width={isWideScreen ? 280 : width}
                                    height={300}
                                    data={data?.site.analytics.users.monthlyActiveUsers}
                                    getDatumName={datum => datum.date}
                                    getDatumValue={datum => datum.count}
                                    getDatumColor={() => 'var(--bar-color)'}
                                    getDatumFadeColor={() => 'var(--bar-fade-color)'}
                                />
                            )}
                        </ChartContainer>
                    )}
                    {frequencies && (
                        <ChartContainer
                            title="Frequency of use"
                            labelX="Minimum days used"
                            labelY={uniqueOrPercentage === 'unique' ? 'Unique users' : 'Percentage of active users'}
                            className={classNames(styles.barChart)}
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
                                    getDatumHoverValueLabel={datum =>
                                        `${datum.value}${uniqueOrPercentage !== 'unique' ? '%' : ''}`
                                    }
                                />
                            )}
                        </ChartContainer>
                    )}
                </div>
                <div className="d-flex justify-content-end align-items-stretch mb-4 text-nowrap">
                    {frequencies && (
                        <ToggleSelect<'unique' | 'percentage'>
                            className={styles.toggleSelect}
                            selected={uniqueOrPercentage}
                            onChange={setUniqueOrPercentage}
                            items={[
                                {
                                    value: 'unique',
                                    label: 'Total',
                                    tooltip: 'The number of users who used the platform atleast n days.',
                                },
                                {
                                    value: 'percentage',
                                    label: 'Percentage',
                                    tooltip:
                                        'Percentage of users out of total active users who used the platform atleast n days.',
                                },
                            ]}
                        />
                    )}
                </div>
            </Card>
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table and are updated every 24 hours.
            </Text>
        </>
    )
}
