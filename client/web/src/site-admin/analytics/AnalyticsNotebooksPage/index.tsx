import React, { useMemo, useState, useEffect } from 'react'

import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { Card, LoadingSpinner } from '@sourcegraph/wildcard'

import { LineChart, Series } from '../../../charts'
import {
    AnalyticsDateRange,
    NotebooksStatisticsResult,
    NotebooksStatisticsVariables,
} from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { ChartContainer } from '../components/ChartContainer'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { ToggleSelect } from '../components/ToggleSelect'
import { ValueLegendList, ValueLegendListProps } from '../components/ValueLegendList'
import { StandardDatum, buildStandardDatum } from '../utils'

import { NOTEBOOKS_STATISTICS } from './queries'

export const AnalyticsNotebooksPage: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    const [eventAggregation, setEventAggregation] = useState<'count' | 'uniqueUsers'>('count')
    const [dateRange, setDateRange] = useState<AnalyticsDateRange>(AnalyticsDateRange.LAST_MONTH)
    const { data, error, loading } = useQuery<NotebooksStatisticsResult, NotebooksStatisticsVariables>(
        NOTEBOOKS_STATISTICS,
        {
            variables: {
                dateRange,
            },
        }
    )
    useEffect(() => {
        eventLogger.logPageView('AdminAnalyticsNotebooks')
    }, [])
    const [stats, legends] = useMemo(() => {
        if (!data) {
            return []
        }
        const { creations, views, blockRuns } = data.site.analytics.notebooks
        const stats: Series<StandardDatum>[] = [
            {
                id: 'creations',
                name: eventAggregation === 'count' ? 'Notebooks created' : 'Users created notebooks',
                color: 'var(--cyan)',
                data: buildStandardDatum(
                    creations.nodes.map(node => ({
                        date: new Date(node.date),
                        value: node[eventAggregation],
                    })),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
            {
                id: 'views',
                name: eventAggregation === 'count' ? 'Notebook views' : 'Users viewed notebooks',
                color: 'var(--orange)',
                data: buildStandardDatum(
                    views.nodes.map(node => ({
                        date: new Date(node.date),
                        value: node[eventAggregation],
                    })),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
        ]
        const legends: ValueLegendListProps['items'] = [
            {
                value: creations.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'Notebooks created' : 'Users created notebooks',
                color: 'var(--cyan)',
            },
            {
                value: views.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'Notebook views' : 'Users viewed notebooks',
                color: 'var(--orange)',
            },
            {
                value: blockRuns.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'Block runs' : 'Users ran blocks',
                color: 'var(--body-color)',
                position: 'right',
            },
        ]

        return [stats, legends]
    }, [data, dateRange, eventAggregation])

    if (error) {
        throw error
    }

    if (loading) {
        return <LoadingSpinner />
    }

    return (
        <>
            <AnalyticsPageTitle>Analytics / Notebooks</AnalyticsPageTitle>

            <Card className="p-3 position-relative">
                <div className="d-flex justify-content-end align-items-stretch mb-2">
                    <HorizontalSelect<AnalyticsDateRange>
                        value={dateRange}
                        label="Date&nbsp;range"
                        onChange={setDateRange}
                        items={[
                            { value: AnalyticsDateRange.LAST_WEEK, label: 'Last week' },
                            { value: AnalyticsDateRange.LAST_MONTH, label: 'Last month' },
                            { value: AnalyticsDateRange.LAST_THREE_MONTHS, label: 'Last 3 months' },
                            { value: AnalyticsDateRange.CUSTOM, label: 'Custom (coming soon)', disabled: true },
                        ]}
                    />
                </div>
                {legends && <ValueLegendList className="mb-3" items={legends} />}
                {stats && (
                    <div>
                        <ChartContainer
                            title={eventAggregation === 'count' ? 'Activity by day' : 'Unique users by day'}
                            labelX="Time"
                            labelY={eventAggregation === 'count' ? 'Activity' : 'Unique users'}
                        >
                            {width => <LineChart width={width} height={300} series={stats} />}
                        </ChartContainer>
                        <div className="d-flex justify-content-end align-items-stretch mb-2">
                            <ToggleSelect<typeof eventAggregation>
                                selected={eventAggregation}
                                onChange={setEventAggregation}
                                items={[
                                    {
                                        tooltip: 'total # of actions triggered',
                                        label: 'Totals',
                                        value: 'count',
                                    },
                                    {
                                        tooltip: 'unique # of users triggered',
                                        label: 'Uniques',
                                        value: 'uniqueUsers',
                                    },
                                ]}
                            />
                        </div>
                    </div>
                )}
            </Card>
        </>
    )
}
