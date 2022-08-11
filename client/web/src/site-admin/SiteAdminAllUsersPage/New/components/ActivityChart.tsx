import React, { useMemo } from 'react'

import { startCase } from 'lodash'

import { useQuery } from '@sourcegraph/http-client'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { LineChart, Series } from '../../../../charts'
import { UsersManagementChartResult, UsersManagementChartVariables } from '../../../../graphql-operations'
import { ChartContainer } from '../../../analytics/components/ChartContainer'
import { HorizontalSelect } from '../../../analytics/components/HorizontalSelect'
import { ToggleSelect } from '../../../analytics/components/ToggleSelect'
import { ValueLegendList, ValueLegendListProps } from '../../../analytics/components/ValueLegendList'
import { useChartFilters } from '../../../analytics/useChartFilters'
import { StandardDatum } from '../../../analytics/utils'
import { USERS_MANAGEMENT_CHART } from '../queries'

export const ActivityChart: React.FunctionComponent = () => {
    const { dateRange, aggregation, grouping } = useChartFilters({ name: 'Users', aggregation: 'uniqueUsers' })

    const { data, error, loading } = useQuery<UsersManagementChartResult, UsersManagementChartVariables>(
        USERS_MANAGEMENT_CHART,
        {
            variables: {
                dateRange: dateRange.value,
                grouping: grouping.value,
            },
        }
    )

    const [activities, legends] = useMemo(() => {
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

        const legends: ValueLegendListProps['items'] = [
            {
                value: users.activity.summary.totalUniqueUsers,
                description: 'Active users',
                color: 'var(--purple)',
                tooltip: 'Users using the application in the selected timeframe.',
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
            {
                value: data.site.adminUsers?.totalCount ?? 0,
                description: 'Administrators',
                color: 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of users with site admin permissions.',
            },
        ]

        return [activities, legends]
    }, [data, aggregation.selected, dateRange.value])

    if (error) {
        throw error
    }

    if (loading || !data) {
        return <LoadingSpinner />
    }

    const groupingLabel = startCase(grouping.value.toLowerCase())

    return (
        <>
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
        </>
    )
}
