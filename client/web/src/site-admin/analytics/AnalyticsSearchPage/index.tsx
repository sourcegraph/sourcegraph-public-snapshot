import React, { useMemo, useEffect } from 'react'

import classNames from 'classnames'
import { startCase } from 'lodash'
import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { Card, H2, Text, LoadingSpinner, AnchorLink, H4, LineChart, Series } from '@sourcegraph/wildcard'

import { SearchStatisticsResult, SearchStatisticsVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { ChartContainer } from '../components/ChartContainer'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { TimeSavedCalculatorGroup } from '../components/TimeSavedCalculatorGroup'
import { ToggleSelect } from '../components/ToggleSelect'
import { ValueLegendList, ValueLegendListProps } from '../components/ValueLegendList'
import { useChartFilters } from '../useChartFilters'
import { StandardDatum } from '../utils'

import { SEARCH_STATISTICS } from './queries'

import styles from './index.module.scss'

export const AnalyticsSearchPage: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    const { dateRange, aggregation, grouping } = useChartFilters({ name: 'Search' })
    const { data, error, loading } = useQuery<SearchStatisticsResult, SearchStatisticsVariables>(SEARCH_STATISTICS, {
        variables: {
            dateRange: dateRange.value,
            grouping: grouping.value,
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
                name: aggregation.selected === 'count' ? 'Searches' : 'Users searched',
                color: 'var(--cyan)',
                data: searches.nodes.map(
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
                id: 'resultClicks',
                name: aggregation.selected === 'count' ? 'Result clicks' : 'Users clicked results',
                color: 'var(--purple)',
                data: resultClicks.nodes.map(
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
                id: 'fileViews',
                name: aggregation.selected === 'count' ? 'File views' : 'Users viewed files',
                color: 'var(--orange)',
                data: fileViews.nodes.map(
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
                value: searches.summary[aggregation.selected === 'count' ? 'totalCount' : 'totalRegisteredUsers'],
                description: aggregation.selected === 'count' ? 'Searches' : 'Users searched',
                color: 'var(--cyan)',
                tooltip: 'Any search conducted via the UI, API, or browser or IDE extensions.',
            },
            {
                value: resultClicks.summary[aggregation.selected === 'count' ? 'totalCount' : 'totalRegisteredUsers'],
                description: aggregation.selected === 'count' ? 'Result clicks' : 'Users clicked results',
                color: 'var(--purple)',
                tooltip:
                    'This event is triggered when a user clicks a result, which may be a file, repository, diff, or commit. Note that at times, a user is able to find the answer to their query directly in search results, therefore fewer interactions may actually speak to higher relevancy and usefulness of search results.',
            },
            {
                value: fileViews.summary[aggregation.selected === 'count' ? 'totalCount' : 'totalRegisteredUsers'],
                description: aggregation.selected === 'count' ? 'File views' : 'Users viewed files',
                color: 'var(--orange)',
                tooltip: 'File views can be generated from a search result, or be linked to directly.',
            },
            {
                value: fileOpens.summary[aggregation.selected === 'count' ? 'totalCount' : 'totalRegisteredUsers'],
                description: aggregation.selected === 'count' ? 'File opens' : 'Users opened files',
                color: 'var(--body-color)',
                position: 'right',
                tooltip:
                    aggregation.selected === 'count'
                        ? 'The number of times a file is opened in the code host or IDE.'
                        : 'The number users who opened file in the code host or IDE.',
            },
        ]
        return [stats, legends]
    }, [data, aggregation.selected, dateRange.value])

    const calculatorProps = useMemo(() => {
        if (!data) {
            return
        }
        const { searches, fileViews } = data.site.analytics.search

        const totalCount = searches.summary.totalCount + fileViews.summary.totalCount
        return {
            page: 'Search',
            dateRange: dateRange.value,
            label: 'Searches &<br/>file views',
            color: 'var(--blue)',
            description:
                'The value of code search greatly varies by use case. We’ve calculated this total value with defaults from the primary use cases below.',
            value: totalCount,
            items: [
                {
                    label: 'Complex searches',
                    minPerItem: 120,
                    description:
                        'Without Sourcegraph, these searches would require complex scripting. They often answer specific and valuable questions such as finding occurrences of Log4j at a specific version globally.',
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
                    minPerItem: 0.5,
                    description:
                        'Common code search use cases are made more efficient through Sourcegraph’s advanced query language, syntax aware search patterns, and the ability to search code, diffs, and commit messages at any revision.',
                    percentage: 75,
                    value: totalCount,
                },
            ],
        }
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
            <AnalyticsPageTitle>Search</AnalyticsPageTitle>

            <Card className="p-3">
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
                All events are generated from entries in the event logs table and are updated every 24 hours.
            </Text>
        </>
    )
}
