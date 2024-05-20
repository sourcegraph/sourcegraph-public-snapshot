import React, { useMemo, useEffect } from 'react'

import { startCase } from 'lodash'

import { useQuery } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import { Card, LoadingSpinner, Text, LineChart, type Series, H2 } from '@sourcegraph/wildcard'

import type { InsightsStatisticsResult, InsightsStatisticsVariables } from '../../../graphql-operations'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { ChartContainer } from '../components/ChartContainer'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { ToggleSelect } from '../components/ToggleSelect'
import { ValueLegendList, type ValueLegendListProps } from '../components/ValueLegendList'
import { useChartFilters } from '../useChartFilters'
import type { StandardDatum } from '../utils'

import { INSIGHTS_STATISTICS } from './queries'

/**
 * Minutes saved constants for code insights.
 */
const MinutesSaved = {
    SearchSeries: 150,
    LanguageSeries: 3,
    ComputeSeries: 1,
}

/**
 * Calculates the total time saved in minutes for a given series.
 *
 * This is used to in "Analytics / Overview" page.
 */
export const calculateMinutesSaved = (data: typeof MinutesSaved): number =>
    data.SearchSeries * MinutesSaved.SearchSeries +
    data.LanguageSeries * MinutesSaved.LanguageSeries +
    data.ComputeSeries * MinutesSaved.ComputeSeries

interface Props extends TelemetryV2Props {}

export const AnalyticsCodeInsightsPage: React.FunctionComponent<Props> = ({ telemetryRecorder }) => {
    const { dateRange, aggregation, grouping } = useChartFilters({
        name: 'Insights',
        aggregation: 'count',
        telemetryRecorder,
    })
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
        EVENT_LOGGER.logPageView('AdminAnalyticsCodeInsights')
        telemetryRecorder.recordEvent('admin.analytics.codeInsights', 'view')
    }, [telemetryRecorder])

    const legends = useMemo(() => {
        if (!data) {
            return []
        }
        const { insightHovers, insightDataPointClicks } = data.site.analytics.codeInsights
        const { insightViews, insightsDashboards } = data

        const legends: ValueLegendListProps['items'] = [
            {
                value:
                    aggregation.selected === 'count'
                        ? insightHovers.summary.totalCount
                        : insightHovers.summary.totalUniqueUsers,
                description: aggregation.selected === 'count' ? 'Insight hovers' : 'Users hovering insights',
                color: 'var(--orange)',
                tooltip:
                    aggregation.selected === 'count'
                        ? 'The number of insight datapoint hovers during the timeframe.'
                        : 'The number of users hovering over insight data points during the timeframe.',
            },
            {
                value:
                    aggregation.selected === 'count'
                        ? insightDataPointClicks.summary.totalCount
                        : insightDataPointClicks.summary.totalUniqueUsers,
                description: aggregation.selected === 'count' ? 'Datapoint clicks' : 'Users clicking datapoints',
                color: 'var(--purple)',
                tooltip:
                    aggregation.selected === 'count'
                        ? 'The number of insight datapoint clicks during the timeframe.'
                        : 'The number of users clicking on insight data points during the timeframe.',
            },
            {
                value: insightViews.nodes.length,
                description: 'Total insights',
                position: 'right',
                tooltip: 'The number of currently existing insights.',
            },
            {
                value: insightsDashboards.nodes.length,
                description: 'Total dashboards',
                position: 'right',
                tooltip: 'The number of currently existing insight dashboards.',
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
                id: 'insight-hovers',
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
                id: 'datapoint-clicks',
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

                <H2 className="my-3">Total time saved</H2>
                <Text>Coming soon...</Text>
            </Card>
            <Text className="font-italic text-center mt-2">
                Some metrics are generated from entries in the event logs table and are updated every 24 hours.
            </Text>
        </>
    )
}
