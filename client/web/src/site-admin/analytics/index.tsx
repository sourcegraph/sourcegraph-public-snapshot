/* eslint-disable react/forbid-dom-props */
import React, { useMemo, useState } from 'react'

import classNames from 'classnames'
import { addDays, getDayOfYear, startOfDay, startOfWeek, sub } from 'date-fns'
import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { Card, H3, Text, LoadingSpinner, AnchorLink, H4, useMatchMedia } from '@sourcegraph/wildcard'

import { LineChart, Series } from '../../charts'
import { BarChart } from '../../charts/components/bar-chart/BarChart'
import {
    AnalyticsDateRange,
    SearchStatisticsResult,
    SearchStatisticsVariables,
    NotebooksStatisticsResult,
    NotebooksStatisticsVariables,
    UsersStatisticsResult,
    UsersStatisticsVariables,
    CodeIntelStatisticsResult,
    CodeIntelStatisticsVariables,
} from '../../graphql-operations'

import { AnalyticsPageTitle } from './components/AnalyticsPageTitle'
import { ChartContainer } from './components/ChartContainer'
import { HorizontalSelect } from './components/HorizontalSelect'
import { TimeSavedCalculatorGroup } from './components/TimeSavedCalculatorGroup'
import { ToggleSelect } from './components/ToggleSelect'
import { ValueLegendList, ValueLegendListProps } from './components/ValueLegendList'
import { SEARCH_STATISTICS, NOTEBOOKS_STATISTICS, USERS_STATISTICS, CODEINTEL_STATISTICS } from './queries'

import styles from './index.module.scss'

interface FrequencyDatum {
    label: string
    value: number
}

interface StandardDatum {
    date: Date
    value: number
}

function buildFrequencyDatum(
    datums: { daysUsed: number; frequency: number }[],
    min: number,
    max: number,
    isGradual = true
): FrequencyDatum[] {
    console.log('isGradual', isGradual)
    const result: FrequencyDatum[] = []
    for (let days = min; days <= max; ++days) {
        const index = datums.findIndex(datum => datum.daysUsed >= days)
        if (isGradual || days === max) {
            result.push({
                label: `${days} days`,
                value: index >= 0 ? datums.slice(index).reduce((sum, datum) => sum + datum.frequency, 0) : 0,
            })
        } else if (index >= 0 && datums[index].daysUsed === days) {
            result.push({
                label: `${days} days`,
                value: datums[index].frequency,
            })
        } else {
            result.push({
                label: `${days}+ days`,
                value: 0,
            })
        }
    }

    return result
}

function buildStandardDatum(datums: StandardDatum[], dateRange: AnalyticsDateRange): StandardDatum[] {
    // Generates 0 value series for dates that don't exist in the original data
    const [to, daysOffset] =
        dateRange === AnalyticsDateRange.LAST_THREE_MONTHS
            ? [startOfWeek(new Date(), { weekStartsOn: 1 }), 7]
            : [startOfDay(new Date()), 1]
    const from =
        dateRange === AnalyticsDateRange.LAST_THREE_MONTHS
            ? sub(to, { months: 3 })
            : dateRange === AnalyticsDateRange.LAST_MONTH
            ? sub(to, { months: 1 })
            : sub(to, { weeks: 1 })
    const newDatums: StandardDatum[] = []
    let date = to
    while (date >= from) {
        const datum = datums?.find(datum => getDayOfYear(datum.date) === getDayOfYear(date))
        newDatums.push(datum ? { ...datum, date } : { date, value: 0 })
        date = addDays(date, -daysOffset)
    }

    return newDatums
}

export const AnalyticsSearchPage: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    const [eventAggregation, setEventAggregation] = useState<'count' | 'uniqueUsers'>('count')
    const [dateRange, setDateRange] = useState<AnalyticsDateRange>(AnalyticsDateRange.LAST_MONTH)
    const { data, error, loading } = useQuery<SearchStatisticsResult, SearchStatisticsVariables>(SEARCH_STATISTICS, {
        variables: {
            dateRange,
        },
    })
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
            {
                id: 'fileOpens',
                name: eventAggregation === 'count' ? 'File opens' : 'Users opened files',
                color: 'var(--body-color)',
                data: buildStandardDatum(
                    fileOpens.nodes.map(node => ({
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
                value: fileViews.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'File views' : 'Users viewed files',
                color: 'var(--orange)',
            },
            {
                value: resultClicks.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'Result clicks' : 'Users clicked results',
                color: 'var(--purple)',
            },
            {
                value: fileOpens.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'File opens' : 'Users opened files',
                color: 'var(--body-color)',
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
                'The value of code search greatly varies by use case. We’ve calculated this total value with defaults from primary use cases below.',
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
                <H4 className="my-3">Suggestions</H4>
                <div className={classNames(styles.border, 'mb-3')} />
                <ul className="mb-3 pl-3">
                    <Text as="li">
                        Promote the{' '}
                        <AnchorLink to="https://docs.sourcegraph.com/integration/editor" target="_blank">
                            IDE extension
                        </AnchorLink>{' '}
                        and{' '}
                        <AnchorLink to="https://docs.sourcegraph.com/cli" target="_blank">
                            SRC CLI
                        </AnchorLink>{' '}
                        to your users to allow them to search where they work.
                    </Text>
                </ul>
            </Card>
        </>
    )
}

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

    const timeSavedStats = useMemo(() => {
        if (!data) {
            return []
        }
        const timeSavedStats = [
            {
                label: 'Views',
                color: 'var(--body-color)',
                minPerItem: 5,
                description:
                    'Notebooks reduce the time it takes to create living documentation and share it. Each notebook view accounts for time saved by both creators and consumers of notebooks.',
                value: data.site.analytics.notebooks.views.summary.totalCount,
            },
        ]
        return timeSavedStats
    }, [data])

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
                <H3 className="my-3">Time saved</H3>
            </Card>
        </>
    )
}

export const AnalyticsUsersPage: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    const [eventAggregation, setEventAggregation] = useState<'count' | 'uniqueUsers'>('uniqueUsers')
    const [dateRange, setDateRange] = useState<AnalyticsDateRange>(AnalyticsDateRange.LAST_MONTH)
    const { data, error, loading } = useQuery<UsersStatisticsResult, UsersStatisticsVariables>(USERS_STATISTICS, {
        variables: {
            dateRange,
        },
    })
    const [frequencies, legends] = useMemo(() => {
        if (!data) {
            return []
        }
        const { users } = data.site.analytics
        const legends: ValueLegendListProps['items'] = [
            {
                value: users.activity.summary.totalUniqueUsers,
                description: 'Active users',
                color: 'var(--purple)',
            },
            {
                value: data.users.totalCount,
                description: 'Registered Users',
                color: 'var(--body-color)',
                position: 'right',
            },
            {
                value: data.site.productSubscription.license?.userCount ?? 0,
                description: 'Users licenses',
                color: 'var(--body-color)',
                position: 'right',
            },
        ]

        const frequencies: FrequencyDatum[] = buildFrequencyDatum(users.frequencies, 1, 30)

        return [frequencies, legends]
    }, [data])

    const activities = useMemo(() => {
        if (!data) {
            return []
        }
        const { users } = data.site.analytics
        const activities: Series<StandardDatum>[] = [
            {
                id: 'activity',
                name: eventAggregation === 'count' ? 'Activities' : 'Active users',
                color: eventAggregation === 'count' ? 'var(--cyan)' : 'var(--purple)',
                data: buildStandardDatum(
                    users.activity.nodes.map(node => ({
                        date: new Date(node.date),
                        value: node[eventAggregation],
                    })),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
        ]

        return activities
    }, [data, eventAggregation, dateRange])

    const summary = useMemo(() => {
        if (!data) {
            return []
        }
        const { avgDAU, avgWAU, avgMAU } = data.site.analytics.users.summary
        return [
            {
                value: avgDAU.totalUniqueUsers,
                label: 'DAU',
            },
            {
                value: avgWAU.totalUniqueUsers,
                label: 'WAU',
            },
            {
                value: avgMAU.totalUniqueUsers,
                label: 'MAU',
            },
        ]
    }, [data])

    const isWideScreen = useMatchMedia('(min-width: 992px)', false)

    if (error) {
        throw error
    }

    if (loading) {
        return <LoadingSpinner />
    }

    return (
        <>
            <AnalyticsPageTitle>Analytics / Users</AnalyticsPageTitle>
            <Card className="p-3">
                <div className="d-flex justify-content-end align-items-stretch mb-2">
                    <HorizontalSelect<AnalyticsDateRange>
                        label="Date&nbsp;range"
                        value={dateRange}
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
                {activities && (
                    <div>
                        <ChartContainer
                            title={eventAggregation === 'count' ? 'Activity by day' : 'Unique users by day'}
                            labelX="Time"
                            labelY={eventAggregation === 'count' ? 'Activity' : 'Unique users'}
                        >
                            {width => <LineChart width={width} height={300} series={activities} />}
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
                <div className={classNames(isWideScreen && 'd-flex')}>
                    {summary && (
                        <ChartContainer
                            title="Average user activity by period"
                            className="mb-5"
                            labelX="Average DAU/WAU/MAU"
                            labelY="Unique users"
                        >
                            {width => (
                                <BarChart
                                    width={isWideScreen ? 280 : width}
                                    height={300}
                                    data={summary}
                                    getDatumName={datum => datum.label}
                                    getDatumValue={datum => datum.value}
                                    getDatumColor={() => 'var(--oc-blue-2)'}
                                />
                            )}
                        </ChartContainer>
                    )}
                    {frequencies && (
                        <ChartContainer
                            className="mb-5"
                            title="Frequency of use"
                            labelX="Days used"
                            labelY="Unique users"
                        >
                            {width => (
                                <BarChart
                                    width={isWideScreen ? 540 : width}
                                    height={300}
                                    data={frequencies}
                                    getDatumName={datum => datum.label}
                                    getDatumValue={datum => datum.value}
                                    getDatumColor={() => 'var(--oc-blue-2)'}
                                />
                            )}
                        </ChartContainer>
                    )}
                </div>
            </Card>
        </>
    )
}

export const AnalyticsCodeIntelPage: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    const [eventAggregation, setEventAggregation] = useState<'count' | 'uniqueUsers'>('count')
    const [dateRange, setDateRange] = useState<AnalyticsDateRange>(AnalyticsDateRange.LAST_MONTH)
    const { data, error, loading } = useQuery<CodeIntelStatisticsResult, CodeIntelStatisticsVariables>(
        CODEINTEL_STATISTICS,
        {
            variables: {
                dateRange,
            },
        }
    )
    const [stats, legends, calculatorProps] = useMemo(() => {
        if (!data) {
            return []
        }
        const {
            referenceClicks,
            definitionClicks,
            searchBasedEvents,
            preciseEvents,
            crossRepoEvents,
        } = data.site.analytics.codeIntel

        const totalEvents = definitionClicks.summary.totalCount + referenceClicks.summary.totalCount
        const totalHoverEvents = searchBasedEvents.summary.totalCount + preciseEvents.summary.totalCount

        const stats: Series<StandardDatum>[] = [
            {
                id: 'references',
                name:
                    eventAggregation === 'count' ? '"Find references" clicked' : 'Users who clicked "Find references"',
                color: 'var(--cyan)',
                data: buildStandardDatum(
                    referenceClicks.nodes.map(node => ({
                        date: new Date(node.date),
                        value: node[eventAggregation],
                    })),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
            {
                id: 'definitions',
                name:
                    eventAggregation === 'count'
                        ? '"Go to definition" clicked'
                        : 'Users who clicked "Go to definition"',
                color: 'var(--orange)',
                data: buildStandardDatum(
                    definitionClicks.nodes.map(node => ({
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
                value: referenceClicks.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'References' : 'Users using references',
                color: 'var(--cyan)',
            },
            {
                value: definitionClicks.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'Definitions' : 'Users using definitions',
                color: 'var(--orange)',
            },
            {
                value: Math.floor((crossRepoEvents.summary.totalCount * totalEvents) / totalHoverEvents || 0),
                description: 'Cross repo events',
                position: 'right',
                color: 'var(--black)',
            },
        ]

        const calculatorProps = {
            label: 'Intel Events',
            color: 'var(--purple)',
            description:
                'Code navigation helps users quickly understand a codebase, identify dependencies, reuse code, and perform more efficient and accurate code reviews.<br/><br/>We’ve broken this caculation down into use cases and types of code intel to be able to independantly value important product capabilities.',
            value: totalEvents,
            items: [
                {
                    label: 'Search based',
                    minPerItem: 0.5,
                    value: Math.floor((searchBasedEvents.summary.totalCount * totalEvents) / totalHoverEvents || 0),
                    description:
                        'Searched based code intel reconizes symbols and is supported across all languages. Search intel events are not exact, thus their time savings is not as high as precise events. ',
                },
                {
                    label: 'Precise events',
                    minPerItem: 1,
                    value: Math.floor((preciseEvents.summary.totalCount * totalEvents) / totalHoverEvents || 0),
                    description:
                        'Precise code intel takes users to the correct result as defined by SCIP, and does so cross repository. The reduction in false positives produced by other search engines represents significant time savings.',
                },
                {
                    label: 'Cross repository <br/> code intel events',
                    minPerItem: 3,
                    value: Math.floor((crossRepoEvents.summary.totalCount * totalEvents) / totalHoverEvents || 0),
                    description:
                        'Cross repository code intel identifies the correct symbol in code throughout your entire code base in a single click, without locating and downloading a repository.',
                },
            ],
        }

        return [stats, legends, calculatorProps]
    }, [data, dateRange, eventAggregation])

    if (error) {
        throw error
    }

    if (loading) {
        return <LoadingSpinner />
    }

    const repos = data?.site.analytics.repos
    const orgMembersCount = data?.currentUser?.organizationMemberships?.totalCount || 0
    const browserExtensionInstalls =
        data?.site.analytics.codeIntel.browserExtensionInstalls.summary.registeredUsers || 0
    const browserExtensionInstallPercentage = orgMembersCount ? (browserExtensionInstalls * 100) / orgMembersCount : 0

    return (
        <>
            <AnalyticsPageTitle>Analytics / Code intel</AnalyticsPageTitle>

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
                <H3 className="my-3">Time saved</H3>
                {calculatorProps && <TimeSavedCalculatorGroup {...calculatorProps} />}
                <H4 className="my-3">Suggestions</H4>
                <div className={classNames(styles.border, 'mb-3')} />
                <ul className="mb-3 pl-3">
                    <Text as="li">
                        <b>{browserExtensionInstallPercentage}%</b> of users have installed the browser extension.{' '}
                        <AnchorLink to="https://docs.sourcegraph.com/integration/browser_extension" target="_blank">
                            Promote installation of the browser extesion to increase value.
                        </AnchorLink>
                    </Text>
                    {repos && (
                        <Text as="li">
                            <b>{repos.preciseCodeIntelCount}</b> of your <b>{repos.count}</b> repositories have precise
                            code intel.{' '}
                            <AnchorLink
                                to="https://docs.sourcegraph.com/code_intelligence/explanations/precise_code_intelligence"
                                target="_blank"
                            >
                                Learn how to improve precise code intel coverage.
                            </AnchorLink>
                        </Text>
                    )}
                </ul>
                <Text className="font-italic text-center">
                    * All events are actually entries from this instance's event_logs table.{' '}
                </Text>
            </Card>
        </>
    )
}
