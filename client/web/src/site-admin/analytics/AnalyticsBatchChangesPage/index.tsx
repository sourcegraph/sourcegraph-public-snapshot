import React, { useMemo, useEffect } from 'react'

import { startCase } from 'lodash'

import { useQuery } from '@sourcegraph/http-client'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { Card, LoadingSpinner, H2, Text, LineChart, type Series } from '@sourcegraph/wildcard'

import type { BatchChangesStatisticsResult, BatchChangesStatisticsVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { ChartContainer } from '../components/ChartContainer'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { TimeSavedCalculator, type TimeSavedCalculatorProps } from '../components/TimeSavedCalculatorGroup'
import { ValueLegendList, type ValueLegendListProps } from '../components/ValueLegendList'
import { useChartFilters } from '../useChartFilters'
import type { StandardDatum } from '../utils'

import { BATCHCHANGES_STATISTICS } from './queries'

export const DEFAULT_MINS_SAVED_PER_CHANGESET = 15

export const AnalyticsBatchChangesPage: React.FunctionComponent = () => {
    const { dateRange, grouping } = useChartFilters({ name: 'BatchChanges' })
    const { data, error, loading } = useQuery<BatchChangesStatisticsResult, BatchChangesStatisticsVariables>(
        BATCHCHANGES_STATISTICS,
        {
            variables: {
                dateRange: dateRange.value,
                grouping: grouping.value,
            },
        }
    )
    useEffect(() => {
        // telemetryRecorder.recordEvent('adminAnalyticsBatchChanges', 'viewed')
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
                data: changesetsCreated.nodes.map(
                    node => ({
                        date: new Date(node.date),
                        value: node.count,
                    }),
                    dateRange.value
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
            {
                id: 'changesets_merged',
                name: 'Changesets merged',
                color: 'var(--cyan)',
                data: changesetsMerged.nodes.map(
                    node => ({
                        date: new Date(node.date),
                        value: node.count,
                    }),
                    dateRange.value
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
                tooltip:
                    'The number of changesets created on a code host during the timeframe. This does not include changesets that are unpublished.',
            },
            {
                value: changesetsMerged.summary.totalCount,
                description: 'Changesets merged',
                color: 'var(--cyan)',
                tooltip: 'The number of changesets merged on the code host during the timeframe.',
            },
        ]

        const calculatorProps: TimeSavedCalculatorProps = {
            page: 'BatchChanges',
            dateRange: dateRange.value,
            label: 'Changesets merged',
            color: 'var(--cyan)',
            value: changesetsMerged.summary.totalCount,
            defaultMinPerItem: DEFAULT_MINS_SAVED_PER_CHANGESET,
            description:
                'Batch Changes automates opening changesets across many repositories and codehosts. It also significantly reduces the time required to manage cross-repository changes via tracking and management functions that are superior to custom solutions, spreadsheets and manually reaching out to developers.',
            temporarySettingsKey: 'batches.minSavedPerChangeset',
        }

        return [stats, legends, calculatorProps]
    }, [data, dateRange.value])

    if (error) {
        throw error
    }

    if (loading) {
        return <LoadingSpinner />
    }

    const groupingLabel = startCase(grouping.value.toLowerCase())

    return (
        <>
            <AnalyticsPageTitle>Batch Changes</AnalyticsPageTitle>

            <Card className="p-3 position-relative">
                <div className="d-flex justify-content-end align-items-stretch mb-2 text-nowrap">
                    <HorizontalSelect<typeof dateRange.value> {...dateRange} />
                </div>
                {legends && <ValueLegendList className="mb-3" items={legends} />}
                {stats && (
                    <div>
                        <ChartContainer title={`${groupingLabel} activity`} labelX="Time" labelY="Activity">
                            {width => <LineChart width={width} height={300} series={stats} />}
                        </ChartContainer>
                    </div>
                )}
                <div>
                    <div className="d-flex justify-content-end align-items-stretch mb-4 text-nowrap">
                        <HorizontalSelect<typeof grouping.value> {...grouping} />
                    </div>
                </div>
                <H2 className="my-3">Total time saved</H2>
                {calculatorProps && (
                    <TimeSavedCalculator {...calculatorProps} telemetryRecorder={noOpTelemetryRecorder} />
                )}
            </Card>
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table and are updated every 24 hours.
            </Text>
        </>
    )
}
