import React, { useMemo, useEffect } from 'react'

import classNames from 'classnames'
import { startCase } from 'lodash'

import { useQuery } from '@sourcegraph/http-client'
import { Card, LoadingSpinner, H2, Text, H4, AnchorLink, LineChart, type Series } from '@sourcegraph/wildcard'

import type { NotebooksStatisticsResult, NotebooksStatisticsVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { ChartContainer } from '../components/ChartContainer'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { TimeSavedCalculator, type TimeSavedCalculatorProps } from '../components/TimeSavedCalculatorGroup'
import { ToggleSelect } from '../components/ToggleSelect'
import { ValueLegendList, type ValueLegendListProps } from '../components/ValueLegendList'
import { useChartFilters } from '../useChartFilters'
import type { StandardDatum } from '../utils'

import { NOTEBOOKS_STATISTICS } from './queries'

import styles from './index.module.scss'

export const AnalyticsNotebooksPage: React.FunctionComponent = () => {
    const { dateRange, aggregation, grouping } = useChartFilters({ name: 'Notebooks' })
    const { data, error, loading } = useQuery<NotebooksStatisticsResult, NotebooksStatisticsVariables>(
        NOTEBOOKS_STATISTICS,
        {
            variables: {
                dateRange: dateRange.value,
                grouping: grouping.value,
            },
        }
    )
    useEffect(() => {
        eventLogger.logPageView('AdminAnalyticsNotebooks')
    }, [])
    const [stats, legends, calculatorProps] = useMemo(() => {
        if (!data) {
            return []
        }
        const { creations, views, blockRuns } = data.site.analytics.notebooks
        const stats: Series<StandardDatum>[] = [
            {
                id: 'creations',
                name: aggregation.selected === 'count' ? 'Notebooks created' : 'Users created notebooks',
                color: 'var(--cyan)',
                data: creations.nodes.map(
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
                id: 'views',
                name: aggregation.selected === 'count' ? 'Notebook views' : 'Users viewed notebooks',
                color: 'var(--orange)',
                data: views.nodes.map(
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
                value: creations.summary[aggregation.selected === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: aggregation.selected === 'count' ? 'Notebooks created' : 'Users created notebooks',
                color: 'var(--cyan)',
                tooltip:
                    aggregation.selected === 'count'
                        ? 'The number of notebooks created in the timeframe.'
                        : 'The number of users who created notebooks in the timeframe.',
            },
            {
                value: views.summary[aggregation.selected === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: aggregation.selected === 'count' ? 'Notebook views' : 'Users viewed notebooks',
                color: 'var(--orange)',
                tooltip:
                    aggregation.selected === 'count'
                        ? 'The number of views of all notebooks in the timeframe.'
                        : 'The number of users who viewed notebooks in the timeframe.',
            },
            {
                value: blockRuns.summary[aggregation.selected === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: aggregation.selected === 'count' ? 'Block runs' : 'Users ran blocks',
                color: 'var(--body-color)',
                position: 'right',
                tooltip:
                    aggregation.selected === 'count'
                        ? 'The number of blocks within each notebook that have been run. Some blocks such as the search results block must be run for the user to see code.'
                        : 'The number of users who ran notebook blocks in the timeframe.',
            },
        ]

        const calculatorProps: TimeSavedCalculatorProps = {
            page: 'Notebooks',
            label: 'Views',
            color: 'var(--body-color)',
            dateRange: dateRange.value,
            value: views.summary.totalCount,
            defaultMinPerItem: 5,
            description:
                'Notebooks reduce the time it takes to create living documentation and share it. Each notebook view accounts for time saved by both creators and consumers of notebooks.',
            temporarySettingsKey: 'search.notebooks.minSavedPerView',
        }

        return [stats, legends, calculatorProps]
    }, [data, dateRange.value, aggregation.selected])

    if (error) {
        throw error
    }

    if (loading) {
        return <LoadingSpinner />
    }

    const groupingLabel = startCase(grouping.value.toLowerCase())

    return (
        <>
            <AnalyticsPageTitle>Notebooks</AnalyticsPageTitle>

            <Card className="p-3 position-relative">
                <div className="d-flex justify-content-end align-items-stretch mb-2 text-nowrap">
                    <HorizontalSelect<typeof dateRange.value> {...dateRange} />
                </div>
                {legends && <ValueLegendList className="mb-3" items={legends} />}
                {stats && (
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
                            {width => <LineChart width={width} height={300} series={stats} />}
                        </ChartContainer>
                        <div className="d-flex justify-content-end align-items-stretch mb-2 text-nowrap">
                            <HorizontalSelect<typeof grouping.value> {...grouping} className="mr-4" />
                            <ToggleSelect<typeof aggregation.selected> {...aggregation} />
                        </div>
                    </div>
                )}
                <H2 className="my-3">Total time saved</H2>
                {calculatorProps && <TimeSavedCalculator {...calculatorProps} />}
                <div className={styles.suggestionBox}>
                    <H4 className="my-3">Suggestions</H4>
                    <div className={classNames(styles.border, 'mb-3')} />
                    <ul className="mb-3 pl-3">
                        <Text as="li">
                            <AnchorLink to="https://sourcegraph.com/blog/notebooks-ci" target="_blank">
                                Learn more
                            </AnchorLink>{' '}
                            about how notebooks improves onboarding, code reuse and saves developers time.
                        </Text>
                    </ul>
                </div>
            </Card>
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table and are updated every 24 hours.
            </Text>
        </>
    )
}
