import React, { useMemo, useState, useEffect } from 'react'

import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { Card, LoadingSpinner, H3, Text, H4, AnchorLink } from '@sourcegraph/wildcard'

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
import { TimeSavedCalculator } from '../components/TimeSavedCalculatorGroup'
import { ToggleSelect } from '../components/ToggleSelect'
import { ValueLegendList, ValueLegendListProps } from '../components/ValueLegendList'
import { StandardDatum } from '../utils'

import { NOTEBOOKS_STATISTICS } from './queries'

import styles from './index.module.scss'

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
    const [stats, legends, calculatorProps] = useMemo(() => {
        if (!data) {
            return []
        }
        const { creations, views, blockRuns } = data.site.analytics.notebooks
        const stats: Series<StandardDatum>[] = [
            {
                id: 'creations',
                name: eventAggregation === 'count' ? 'Notebooks created' : 'Users created notebooks',
                color: 'var(--cyan)',
                data: creations.nodes.map(
                    node => ({
                        date: new Date(node.date),
                        value: node[eventAggregation],
                    }),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
            {
                id: 'views',
                name: eventAggregation === 'count' ? 'Notebook views' : 'Users viewed notebooks',
                color: 'var(--orange)',
                data: views.nodes.map(
                    node => ({
                        date: new Date(node.date),
                        value: node[eventAggregation],
                    }),
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
                tooltip:
                    eventAggregation === 'count'
                        ? 'The number of notebooks created in the timeframe.'
                        : 'The number of users who created notebooks in the timeframe.',
            },
            {
                value: views.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'Notebook views' : 'Users viewed notebooks',
                color: 'var(--orange)',
                tooltip:
                    eventAggregation === 'count'
                        ? 'The number of views of all notebooks in the timeframe.'
                        : 'The number of users who viewed notebooks in the timeframe.',
            },
            {
                value: blockRuns.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'Block runs' : 'Users ran blocks',
                color: 'var(--body-color)',
                position: 'right',
                tooltip:
                    eventAggregation === 'count'
                        ? 'The number of of blocks within each notebook that are run. Some blocks such as the search results block must be run for the user to see code.'
                        : 'The number of users who ran blocks within each notebook in the timeframe.',
            },
        ]

        const calculatorProps = {
            page: 'Notebooks',
            label: 'Views',
            color: 'var(--body-color)',
            value: views.summary.totalCount,
            minPerItem: 5,
            description:
                'Notebooks reduce the time it takes to create living documentation and share it. Each notebook view accounts for time saved by both creators and consumers of notebooks.',
        }

        return [stats, legends, calculatorProps]
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
                        onChange={value => {
                            setDateRange(value)
                            eventLogger.log(`AdminAnalyticsNotebooksDateRange${value}Selected`)
                        }}
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
                                onChange={value => {
                                    setEventAggregation(value)
                                    eventLogger.log(
                                        `AdminAnalyticsNotebooksAgg${value === 'count' ? 'Totals' : 'Uniques'}Clicked`
                                    )
                                }}
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
                <H3 className="my-3">Time saved</H3>
                {calculatorProps && <TimeSavedCalculator {...calculatorProps} />}
                <div className={styles.suggestionBox}>
                    <H4 className="my-3">Suggestions</H4>
                    <div className={classNames(styles.border, 'mb-3')} />
                    <ul className="mb-3 pl-3">
                        <Text as="li">
                            <AnchorLink to="https://about.sourcegraph.com/blog/notebooks-ci" target="_blank">
                                Learn more
                            </AnchorLink>{' '}
                            about how notebooks improves onbaording, code reuse and saves developers time.
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
