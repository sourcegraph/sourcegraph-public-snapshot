import React, { useMemo, useEffect } from 'react'

import { startCase } from 'lodash'
import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { Card, LoadingSpinner, Text, LineChart, Series } from '@sourcegraph/wildcard'

import { InsightsStatisticsResult, InsightsStatisticsVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { ChartContainer } from '../components/ChartContainer'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { ToggleSelect } from '../components/ToggleSelect'
import { ValueLegendList, ValueLegendListProps } from '../components/ValueLegendList'
import { useChartFilters } from '../useChartFilters'
import { StandardDatum } from '../utils'

import { INSIGHTS_STATISTICS } from './queries'

export const AnalyticsCodeInsightsPage: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    const { dateRange, aggregation, grouping } = useChartFilters({ name: 'Insights', aggregation: 'count' })
    const { data, error, loading } = useQuery<InsightsStatisticsResult, InsightsStatisticsVariables>(
        INSIGHTS_STATISTICS,
        {
            variables: {
                dateRange: dateRange.value,
                grouping: grouping.value,
            },
        }
    )
    useEffect(() => {
        eventLogger.logPageView('AdminAnalyticsCodeInsights')
    }, [])

    const legends = useMemo(() => {
        if (!data) {
            return []
        }
        const { insightHovers, insightDataPointClicks } = data.site.analytics.codeInsights
        const legends: ValueLegendListProps['items'] = [
            {
                value:
                    aggregation.selected === 'count'
                        ? insightHovers.summary.totalCount
                        : insightHovers.summary.totalRegisteredUsers,
                description: aggregation.selected === 'count' ? 'Insight hovers' : 'Users hovering insights',
                color: 'var(--orange)',
                tooltip: 'TODO:',
            },
            {
                value:
                    aggregation.selected === 'count'
                        ? insightDataPointClicks.summary.totalCount
                        : insightDataPointClicks.summary.totalRegisteredUsers,
                description: aggregation.selected === 'count' ? 'Datapoint clicks' : 'Users clicking datapoints',
                color: 'var(--purple)',
                tooltip: 'TODO:',
            },
        ]

        return legends
    }, [aggregation.selected, data])

    const activities = useMemo(() => {
        if (!data) {
            return []
        }
        const { insightHovers, insightDataPointClicks } = data.site.analytics.codeInsights
        const activities: Series<StandardDatum>[] = [
            {
                id: 'activity',
                name: aggregation.selected === 'count' ? 'Insight hovers' : 'Users hovering insights',
                color: 'var(--orange)',
                data: insightHovers.nodes.map(
                    node => ({
                        date: new Date(node.date),
                        value: node[aggregation.selected],
                    }),
                    dateRange.value
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
            {
                id: 'activity',
                name: aggregation.selected === 'count' ? 'Datapoint clicks' : 'Users clicking datapoints',
                color: 'var(--purple)',
                data: insightDataPointClicks.nodes.map(
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

    if (error) {
        throw error
    }

    if (loading) {
        return <LoadingSpinner />
    }

    const groupingLabel = startCase(grouping.value.toLowerCase())

    return (
        <>
            <AnalyticsPageTitle>Insights</AnalyticsPageTitle>
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
            </Card>
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table and are updated every 24 hours..
            </Text>
        </>
    )
}
