import React, { useMemo, useState, useEffect } from 'react'

import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { Card, H3, Text, LoadingSpinner, AnchorLink, H4 } from '@sourcegraph/wildcard'

import { LineChart, Series } from '../../../charts'
import { AnalyticsDateRange, SearchStatisticsResult, SearchStatisticsVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { ChartContainer } from '../components/ChartContainer'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { TimeSavedCalculatorGroup } from '../components/TimeSavedCalculatorGroup'
import { ToggleSelect } from '../components/ToggleSelect'
import { ValueLegendList, ValueLegendListProps } from '../components/ValueLegendList'
import { StandardDatum, buildStandardDatum } from '../utils'

import { SEARCH_STATISTICS } from './queries'

import styles from './index.module.scss'

export const AnalyticsSearchPage: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    const [eventAggregation, setEventAggregation] = useState<'count' | 'uniqueUsers'>('count')
    const [dateRange, setDateRange] = useState<AnalyticsDateRange>(AnalyticsDateRange.LAST_MONTH)
    const { data, error, loading } = useQuery<SearchStatisticsResult, SearchStatisticsVariables>(SEARCH_STATISTICS, {
        variables: {
            dateRange,
        },
    })
    useEffect(() => {
        eventLogger.logPageView('AdminAnalyticsSearch')
    }, [])
    const [stats, legends] = useMemo(() => {
        if (!data) {
            return []
        }
        const { searches, fileViews, fileOpens, resultClicks } = data.site.analytics.search
        const stats: Series<StandardDatum>[] = [
            {
                id: 'searches',
                name: eventAggregation === 'count' ? 'Searches' : 'Users searched',
                color: 'var(--cyan)',
                data: buildStandardDatum(
                    searches.nodes.map(node => ({
                        date: new Date(node.date),
                        value: node[eventAggregation],
                    })),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
            {
                id: 'resultClicks',
                name: eventAggregation === 'count' ? 'Result clicks' : 'Users clicked results',
                color: 'var(--purple)',
                data: buildStandardDatum(
                    resultClicks.nodes.map(node => ({
                        date: new Date(node.date),
                        value: node[eventAggregation],
                    })),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
            {
                id: 'fileViews',
                name: eventAggregation === 'count' ? 'File views' : 'Users viewed files',
                color: 'var(--orange)',
                data: buildStandardDatum(
                    fileViews.nodes.map(node => ({
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
                value: searches.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'Searches' : 'Users searched',
                color: 'var(--cyan)',
            },
            {
                value: resultClicks.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'Result clicks' : 'Users clicked results',
                color: 'var(--purple)',
            },
            {
                value: fileViews.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'File views' : 'Users viewed files',
                color: 'var(--orange)',
            },
            {
                value: fileOpens.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'File opens' : 'Users opened files',
                color: 'var(--body-color)',
                position: 'right',
            },
        ]
        return [stats, legends]
    }, [data, eventAggregation, dateRange])

    const calculatorProps = useMemo(() => {
        if (!data) {
            return
        }
        const { searches, fileViews } = data.site.analytics.search

        const totalCount = searches.summary.totalCount + fileViews.summary.totalCount
        return {
            label: 'Searches &<br/>file views',
            color: 'var(--blue)',
            description:
                'The value of code search greatly varies by use case. We’ve calculated this total value with defaults from the primary use cases below.',
            value: totalCount,
            items: [
                {
                    label: 'Complex searches',
                    minPerItem: 5,
                    description:
                        'These searches that would require ad-hoc scripting to accomplish without Sourcegraph.  These searches often answer  specific and valuable questions such as finding occurrences of log4j at a specific version globally.',
                    percentage: 3,
                    value: totalCount,
                },
                {
                    label: 'Global searches',
                    minPerItem: 5,
                    description:
                        "Searches that leverage Sourcegraph's ability to quickly and confidently query all of your company's code across code hosts, without locally cloning repositories or complex scripting.",
                    percentage: 22,
                    value: totalCount,
                },
                {
                    label: 'Core workflow',
                    minPerItem: 5,
                    description:
                        'Common code search use cases are made more efficient through Sourcegraph’s advanced query language and features like syntax aware search patterns and the ability to search code, diffs, and commit messages at any revision.',
                    percentage: 75,
                    value: totalCount,
                },
            ],
        }
    }, [data])

    if (error) {
        throw error
    }

    if (loading) {
        return <LoadingSpinner />
    }

    return (
        <>
            <AnalyticsPageTitle>Analytics / Search</AnalyticsPageTitle>

            <Card className="p-3">
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
                <H3 className="my-3">Time saved</H3>
                {calculatorProps && <TimeSavedCalculatorGroup {...calculatorProps} />}
                <div className={styles.suggestionBox}>
                    <H4 className="my-3">Suggestions</H4>
                    <div className={classNames(styles.border, 'mb-3')} />
                    <ul className="mb-3 pl-3">
                        <Text as="li">
                            Promote the{' '}
                            <AnchorLink to="/help/integration/editor" target="_blank">
                                IDE extension
                            </AnchorLink>{' '}
                            and{' '}
                            <AnchorLink to="/help/cli" target="_blank">
                                SRC CLI
                            </AnchorLink>{' '}
                            to your users to allow them to search where they work.
                        </Text>
                    </ul>
                </div>
            </Card>
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table.
            </Text>
        </>
    )
}
