import React, { useMemo, useEffect } from 'react'

import classNames from 'classnames'
import { startCase } from 'lodash'

import { useQuery } from '@sourcegraph/http-client'
import { Card, H2, Text, LoadingSpinner, AnchorLink, H4, LineChart, type Series } from '@sourcegraph/wildcard'

import type { ExtensionsStatisticsResult, ExtensionsStatisticsVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { ChartContainer } from '../components/ChartContainer'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { TimeSavedCalculatorGroup } from '../components/TimeSavedCalculatorGroup'
import { ToggleSelect } from '../components/ToggleSelect'
import { ValueLegendList, type ValueLegendListProps } from '../components/ValueLegendList'
import { useChartFilters } from '../useChartFilters'
import type { StandardDatum } from '../utils'

import { EXTENSIONS_STATISTICS } from './queries'

import styles from './index.module.scss'

export const AnalyticsExtensionsPage: React.FunctionComponent = () => {
    const { dateRange, aggregation, grouping } = useChartFilters({ name: 'Extensions' })
    const { data, error, loading } = useQuery<ExtensionsStatisticsResult, ExtensionsStatisticsVariables>(
        EXTENSIONS_STATISTICS,
        {
            variables: {
                dateRange: dateRange.value,
                grouping: grouping.value,
            },
        }
    )
    useEffect(() => {
        eventLogger.logPageView('AdminAnalyticsExtensions')
    }, [])
    const [stats, legends, calculatorProps, installationStats] = useMemo(() => {
        if (!data) {
            return []
        }
        const { jetbrains, vscode, browser } = data.site.analytics.extensions

        const totalEvents = vscode.summary.totalCount + jetbrains.summary.totalCount + browser.summary.totalCount

        const stats: Series<StandardDatum>[] = [
            {
                id: 'jetbrains',
                name: aggregation.selected === 'count' ? 'JetBrains IDE plugin' : 'Users using JetBrains IDE plugin',
                color: 'var(--cyan)',
                data: jetbrains.nodes.map(
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
                id: 'vscode',
                name: aggregation.selected === 'count' ? 'VS Code IDE extension' : 'Users using VS Code IDE extension',
                color: 'var(--purple)',
                data: vscode.nodes.map(
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
                id: 'browser',
                name: aggregation.selected === 'count' ? 'Browser extension' : 'Users using browser extension',
                color: 'var(--orange)',
                data: browser.nodes.map(
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
                value: jetbrains.summary[aggregation.selected === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description:
                    aggregation.selected === 'count' ? 'JetBrains\nIDE plugin' : 'Users using\nJetBrains IDE plugin',
                color: 'var(--cyan)',
                tooltip:
                    aggregation.selected === 'count'
                        ? 'The number searches in JetBrains IDE plugin.'
                        : 'The number of users searched in JetBrains IDE plugin.',
            },
            {
                value: vscode.summary[aggregation.selected === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description:
                    aggregation.selected === 'count' ? 'VS Code\nIDE extension' : 'Users using\nVS Code IDE extension',
                color: 'var(--purple)',
                tooltip:
                    aggregation.selected === 'count'
                        ? 'The number searches in VS Code IDE extension.'
                        : 'The number of users searched in IDE extension.',
            },
            {
                value: browser.summary[aggregation.selected === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: aggregation.selected === 'count' ? 'Browser\nextension' : 'Users using\nbrowser extension',
                color: 'var(--orange)',
                tooltip:
                    aggregation.selected === 'count'
                        ? 'The number of code navigation events in code hosts via Browser extensions.'
                        : 'The number of users using code navigation  events in code hosts via Browser extensions',
            },
        ]

        const calculatorProps = {
            page: 'Extensions',
            label: 'Total events',
            dateRange: dateRange.value,
            color: 'var(--purple)',
            description:
                'Our extensions allow users to complete their goals without switching tools and context. We’ve calculated the time saved by reducing context switching between tools.',
            value: totalEvents,
            items: [
                {
                    label: 'Browser extension',
                    minPerItem: 0.5,
                    value: browser.summary.totalCount,
                    description:
                        'Code navigation events (e.g. go to definition, find references) that have occurred on your code host. These make it easier to understand code and provide feedback during code reviews.',
                },
                {
                    label: 'JetBrains IDE plugin',
                    minPerItem: 5,
                    value: jetbrains.summary.totalCount,
                    description:
                        "Searches from JetBrains IDEs across all of your company's code without locally cloning repositories or complex scripting.",
                },
                {
                    label: 'VS Code IDE extension',
                    minPerItem: 5,
                    value: vscode.summary.totalCount,
                    description:
                        "Searches from VS Code across all of your company's code without locally cloning repositories or complex scripting.",
                },
            ],
        }
        const totalUsersCount = data?.site.users.totalCount
        const installationStats =
            totalUsersCount > 0
                ? {
                      vscode: Math.floor((vscode.summary.totalUniqueUsers * 100) / totalUsersCount),
                      jetbrains: Math.floor((jetbrains.summary.totalUniqueUsers * 100) / totalUsersCount),
                      browser: Math.floor((browser.summary.totalUniqueUsers * 100) / totalUsersCount),
                  }
                : undefined

        return [stats, legends, calculatorProps, installationStats]
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
            <AnalyticsPageTitle>Extensions</AnalyticsPageTitle>

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
                {calculatorProps && <TimeSavedCalculatorGroup {...calculatorProps} />}
                <div className={styles.suggestionBox}>
                    <H4 className="my-3">Suggestions</H4>
                    <div className={classNames(styles.border, 'mb-3')} />
                    {installationStats && (
                        <ul className="mb-3 pl-3">
                            <Text as="li">
                                {installationStats.vscode}% of users have installed the{' '}
                                <AnchorLink to="/help/integration/editor" target="_blank">
                                    VS Code extension
                                </AnchorLink>
                                . Promote installation to increase the value.
                            </Text>
                            <Text as="li">
                                {installationStats.jetbrains}% of users have installed the{' '}
                                <AnchorLink to="/help/integration/editor" target="_blank">
                                    JetBrains plugin
                                </AnchorLink>
                                . Promote installation to increase the value.
                            </Text>
                            <Text as="li">
                                {installationStats.browser}% of users have installed the{' '}
                                <AnchorLink to="/help/integration/browser_extension" target="_blank">
                                    browser extension
                                </AnchorLink>
                                . Promote installation to increase the value.
                            </Text>
                        </ul>
                    )}
                </div>
            </Card>
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table and are updated every 24 hours.
            </Text>
        </>
    )
}
