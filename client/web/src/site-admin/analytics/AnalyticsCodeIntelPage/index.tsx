import React, { useMemo, useState, useEffect } from 'react'

import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { Card, H3, Text, LoadingSpinner, AnchorLink, H4 } from '@sourcegraph/wildcard'

import { LineChart, Series } from '../../../charts'
import {
    AnalyticsDateRange,
    CodeIntelStatisticsResult,
    CodeIntelStatisticsVariables,
} from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { ChartContainer } from '../components/ChartContainer'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { TimeSavedCalculatorGroup } from '../components/TimeSavedCalculatorGroup'
import { ToggleSelect } from '../components/ToggleSelect'
import { ValueLegendList, ValueLegendListProps } from '../components/ValueLegendList'
import { StandardDatum, buildStandardDatum } from '../utils'

import { CODEINTEL_STATISTICS } from './queries'

import styles from './index.module.scss'

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
    useEffect(() => {
        eventLogger.logPageView('AdminAnalyticsCodeIntel')
    }, [])
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
    const totalUsers = data?.users?.totalCount || 0
    const browserExtensionInstalls =
        data?.site.analytics.codeIntel.browserExtensionInstalls.summary.totalRegisteredUsers || 0
    const browserExtensionInstallPercentage = totalUsers ? (browserExtensionInstalls * 100) / totalUsers : 0

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
                <div className={styles.suggestionBox}>
                    <H4 className="my-3">Suggestions</H4>
                    <div className={classNames(styles.border, 'mb-3')} />
                    <ul className="mb-3 pl-3">
                        <Text as="li">
                            <b>{browserExtensionInstallPercentage}%</b> of users have installed the browser extension.{' '}
                            <AnchorLink to="/help/integration/browser_extension" target="_blank">
                                Promote installation of the browser extesion to increase value.
                            </AnchorLink>
                        </Text>
                        {repos && (
                            <Text as="li">
                                <b>{repos.preciseCodeIntelCount}</b> of your <b>{repos.count}</b> repositories have
                                precise code intel.{' '}
                                <AnchorLink
                                    to="/help/code_intelligence/explanations/precise_code_intelligence"
                                    target="_blank"
                                >
                                    Learn how to improve precise code intel coverage.
                                </AnchorLink>
                            </Text>
                        )}
                    </ul>
                </div>
            </Card>
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table.
            </Text>
        </>
    )
}
