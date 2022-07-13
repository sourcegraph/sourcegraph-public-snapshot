import React, { useMemo, useState, useEffect } from 'react'

import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { Card, LoadingSpinner, H3, Text } from '@sourcegraph/wildcard'

import { LineChart, Series } from '../../../charts'
import {
    AnalyticsDateRange,
    BatchChangesStatisticsResult,
    BatchChangesStatisticsVariables,
} from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { ChartContainer } from '../components/ChartContainer'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { TimeSavedCalculator } from '../components/TimeSavedCalculatorGroup'
import { ValueLegendList, ValueLegendListProps } from '../components/ValueLegendList'
import { StandardDatum, buildStandardDatum } from '../utils'

import { BATCHCHANGES_STATISTICS } from './queries'

export const AnalyticsBatchChangesPage: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    const [dateRange, setDateRange] = useState<AnalyticsDateRange>(AnalyticsDateRange.LAST_MONTH)
    const { data, error, loading } = useQuery<BatchChangesStatisticsResult, BatchChangesStatisticsVariables>(
        BATCHCHANGES_STATISTICS,
        {
            variables: {
                dateRange,
            },
        }
    )
    useEffect(() => {
        eventLogger.logPageView('AdminAnalyticsBatchChanges')
    }, [])
    const [stats, legends, calculatorProps] = useMemo(() => {
        if (!data) {
            return []
        }
        const { changesetsCreated, changesetsMerged } = data.site.analytics.batchChanges
        const stats: Series<StandardDatum>[] = [
            {
                id: 'changesets_created',
                name: 'Changesets created',
                color: 'var(--orange)',
                data: buildStandardDatum(
                    changesetsCreated.nodes.map(node => ({
                        date: new Date(node.date),
                        value: node.count,
                    })),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
            {
                id: 'changesets_merged',
                name: 'Changesets merged',
                color: 'var(--cyan)',
                data: buildStandardDatum(
                    changesetsMerged.nodes.map(node => ({
                        date: new Date(node.date),
                        value: node.count,
                    })),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
        ]
        const legends: ValueLegendListProps['items'] = [
            {
                value: changesetsCreated.summary.totalCount,
                description: 'Changesets created',
                color: 'var(--orange)',
            },
            {
                value: changesetsMerged.summary.totalCount,
                description: 'Changesets merged',
                color: 'var(--cyan)',
            },
        ]

        const calculatorProps = {
            label: 'Changesets merged',
            color: 'var(--cyan)',
            value: changesetsMerged.summary.totalCount,
            minPerItem: 15,
            description:
                'Batch Changes automates opening changesets across many repositories and codehosts. It also significantly reduces the time required to manage cross-repository changes via tracking and management functions that are superior to custom solutions, spreadsheets and manually reaching out to developers.',
        }

        return [stats, legends, calculatorProps]
    }, [data, dateRange])

    if (error) {
        throw error
    }

    if (loading) {
        return <LoadingSpinner />
    }

    return (
        <>
            <AnalyticsPageTitle>Analytics / Batch Changes</AnalyticsPageTitle>

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
                        <ChartContainer title="Activity by day" labelX="Time" labelY="Activity">
                            {width => <LineChart width={width} height={300} series={stats} />}
                        </ChartContainer>
                    </div>
                )}
                <H3 className="my-3">Time saved</H3>
                {calculatorProps && <TimeSavedCalculator {...calculatorProps} />}
            </Card>
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table.
            </Text>
        </>
    )
}
