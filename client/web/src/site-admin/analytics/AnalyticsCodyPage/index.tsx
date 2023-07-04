import React, { useMemo, useEffect } from 'react'

import { startCase } from 'lodash'

import { useQuery } from '@sourcegraph/http-client'
import { Alert, Card, H2, LineChart, LoadingSpinner, Series, Text } from '@sourcegraph/wildcard'

import { CodyStatisticsResult, CodyStatisticsVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { ChartContainer } from '../components/ChartContainer'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { ToggleSelect } from '../components/ToggleSelect'
import { useChartFilters } from '../useChartFilters'
import { StandardDatum } from '../utils'

import { CODY_STATISTICS } from './queries'

interface Props {}

export const AnalyticsCodyPage: React.FunctionComponent<Props> = () => {
    const { dateRange, aggregation, grouping } = useChartFilters({ name: 'Cody', aggregation: 'count' })
    const { data, error, loading } = useQuery<CodyStatisticsResult, CodyStatisticsVariables>(CODY_STATISTICS, {
        variables: {
            dateRange: dateRange.value,
            grouping: grouping.value,
        },
    })
    useEffect(() => {
        eventLogger.logPageView('AdminAnalyticsCody')
    }, [])

    const activities = useMemo(() => {
        if (!data) {
            return []
        }
        const { prompts, completionsAccepted, completionsSuggested } = data.site.analytics.cody
        const activities: Series<StandardDatum>[] = [
            {
                id: 'cody-prompts',
                name:
                    aggregation.selected === 'count'
                        ? 'Cody prompts and recipes'
                        : 'Users using Cody prompts and recipes',
                color: 'var(--orange)',
                data: prompts.nodes.map(
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
                id: 'cody-completions-suggested',
                name:
                    aggregation.selected === 'count'
                        ? 'Cody completions suggested'
                        : 'Users seeing Cody completion suggestions',
                color: 'var(--purple)',
                data: completionsSuggested.nodes.map(
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
                id: 'cody-completions-accepted',
                name:
                    aggregation.selected === 'count' ? 'Cody completions accepted' : 'Users accepting Cody completions',
                color: 'var(--purple)',
                data: completionsAccepted.nodes.map(
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
            <AnalyticsPageTitle>Cody</AnalyticsPageTitle>
            <Card className="p-3">
                <div className="d-flex justify-content-end align-items-stretch mb-2 text-nowrap">
                    <HorizontalSelect<typeof dateRange.value> {...dateRange} />
                </div>
                {activities && (
                    <div>
                        <Alert variant="warning" className="d-flex align-center">
                            Cody analytics are a work in progress, and subject to change in future updates
                        </Alert>
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
        </>
    )
}
