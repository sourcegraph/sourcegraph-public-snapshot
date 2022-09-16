import React, { useMemo, useEffect, useState } from 'react'

import classNames from 'classnames'
import { groupBy, sortBy, startCase, sumBy } from 'lodash'
import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import {
    Card,
    H2,
    Text,
    LoadingSpinner,
    AnchorLink,
    H4,
    LineChart,
    Series,
    BarChart,
    LegendList,
    LegendItem,
    Link,
} from '@sourcegraph/wildcard'

import { CodeIntelStatisticsResult, CodeIntelStatisticsVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { ChartContainer } from '../components/ChartContainer'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { TimeSavedCalculatorGroup } from '../components/TimeSavedCalculatorGroup'
import { ToggleSelect } from '../components/ToggleSelect'
import { ValueLegendList, ValueLegendListProps } from '../components/ValueLegendList'
import { useChartFilters } from '../useChartFilters'
import { formatNumber, StandardDatum } from '../utils'

import { CODEINTEL_STATISTICS } from './queries'

import styles from './index.module.scss'

export const AnalyticsCodeIntelPage: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    const { dateRange, aggregation, grouping } = useChartFilters({ name: 'CodeIntel' })
    const { data, error, loading } = useQuery<CodeIntelStatisticsResult, CodeIntelStatisticsVariables>(
        CODEINTEL_STATISTICS,
        {
            variables: {
                dateRange: dateRange.value,
                grouping: grouping.value,
            },
        }
    )
    useEffect(() => {
        eventLogger.logPageView('AdminAnalyticsCodeIntel')
    }, [])

    type Kind = 'inApp' | 'codeHost' | 'crossRepo' | 'precise'

    const [kindToMinPerItem, setKindToMinPerItem] = useState<Record<Kind, number>>({
        inApp: 0.5,
        codeHost: 1.5,
        crossRepo: 3,
        precise: 1,
    })

    const [stats, legends, calculatorProps] = useMemo(() => {
        if (!data) {
            return []
        }
        const {
            referenceClicks,
            definitionClicks,
            inAppEvents,
            codeHostEvents,
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
                    aggregation.selected === 'count'
                        ? '"Find references" clicked'
                        : 'Users who clicked "Find references"',
                color: 'var(--cyan)',
                data: referenceClicks.nodes.map(
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
                id: 'definitions',
                name:
                    aggregation.selected === 'count'
                        ? '"Go to definition" clicked'
                        : 'Users who clicked "Go to definition"',
                color: 'var(--orange)',
                data: definitionClicks.nodes.map(
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
                value:
                    referenceClicks.summary[aggregation.selected === 'count' ? 'totalCount' : 'totalRegisteredUsers'],
                description: aggregation.selected === 'count' ? 'References' : 'Users using references',
                color: 'var(--cyan)',
                tooltip:
                    aggregation.selected === 'count'
                        ? "The number of times users clicked 'References' in code navigation hovers to view usages of an item."
                        : "The number of users who clicked 'References'. in code navigation hovers to view usages of an item.",
            },
            {
                value:
                    definitionClicks.summary[aggregation.selected === 'count' ? 'totalCount' : 'totalRegisteredUsers'],
                description: aggregation.selected === 'count' ? 'Definitions' : 'Users using definitions',
                color: 'var(--orange)',
                tooltip:
                    aggregation.selected === 'count'
                        ? "The number of times users clicked 'Definitions' in code navigation hovers to view the definition of an item."
                        : "The number of users who clicked 'Definitions' in code navigation hovers to view the definition of an item.",
            },
            {
                value: Math.floor((crossRepoEvents.summary.totalCount * totalEvents) / totalHoverEvents || 0),
                description: 'Cross repo events',
                position: 'right',
                color: 'var(--body-color)',
                tooltip:
                    'Cross repository code navigation identifies symbols in code throughout your Sourcegraph instance, in a single click, without locating and downloading a repository.',
            },
        ]

        const calculatorProps: React.ComponentProps<typeof TimeSavedCalculatorGroup> = {
            page: 'CodeIntel',
            label: 'Navigation events',
            dateRange: dateRange.value,
            color: 'var(--purple)',
            description:
                'Code navigation helps users quickly understand a codebase, identify dependencies, reuse code, and perform more efficient and accurate code reviews.<br/><br/>We’ve broken this calculation down into use cases and types of code navigation to be able to independently value product capabilities.',
            value: totalEvents,
            items: [
                {
                    label: 'In app code navigation',
                    minPerItem: kindToMinPerItem.inApp,
                    onMinPerItemChange: minPerItem => setKindToMinPerItem(old => ({ ...old, inApp: minPerItem })),
                    value: inAppEvents.summary.totalCount,
                    description:
                        'In app code navigation supports developers finding the impact of a change or code to reuse by listing references and finding definitions.',
                },
                {
                    label: 'Code navigation on code hosts <br/> via the browser extension',
                    minPerItem: kindToMinPerItem.codeHost,
                    onMinPerItemChange: minPerItem => setKindToMinPerItem(old => ({ ...old, codeHost: minPerItem })),
                    value: codeHostEvents.summary.totalCount,
                    description:
                        'Navigation events on the code host typically occur during PR reviews, where the ability to quickly understand code is key to efficient reviews.',
                },
                {
                    label: 'Cross repository <br/> code navigation events',
                    minPerItem: kindToMinPerItem.crossRepo,
                    onMinPerItemChange: minPerItem => setKindToMinPerItem(old => ({ ...old, crossRepo: minPerItem })),
                    value: Math.floor((crossRepoEvents.summary.totalCount * totalEvents) / totalHoverEvents || 0),
                    description:
                        'Cross repository code navigation identifies the correct symbol in code throughout your entire code base in a single click, without locating and downloading a repository.',
                },
                {
                    label: 'Precise code navigation*',
                    minPerItem: kindToMinPerItem.precise,
                    onMinPerItemChange: minPerItem => setKindToMinPerItem(old => ({ ...old, precise: minPerItem })),
                    value: Math.floor((preciseEvents.summary.totalCount * totalEvents) / totalHoverEvents || 0),
                    description:
                        'Compiler-accurate code navigation takes users to the correct result as defined by SCIP, and does so cross repository. The reduction in false positives produced by other search engines represents significant additional time savings.',
                },
            ],
        }

        return [stats, legends, calculatorProps]
    }, [
        data,
        dateRange.value,
        aggregation.selected,
        kindToMinPerItem.codeHost,
        kindToMinPerItem.crossRepo,
        kindToMinPerItem.inApp,
        kindToMinPerItem.precise,
    ])

    if (error) {
        throw error
    }

    if (loading) {
        return <LoadingSpinner />
    }

    const repos = data?.site.analytics.repos
    const groupingLabel = startCase(grouping.value.toLowerCase())

    const preciseFraction = data
        ? data.site.analytics.codeIntel.preciseEvents.summary.totalCount /
          (data.site.analytics.codeIntel.preciseEvents.summary.totalCount +
              data.site.analytics.codeIntel.searchBasedEvents.summary.totalCount)
        : undefined

    interface TopRepo {
        repoName: string
        events: number
        hoursSaved: number
        preciseEnabled: boolean
        preciseNavigation: JSX.Element
    }

    const langToIndexerUrl: Record<string, string> = {
        python: 'https://github.com/sourcegraph/scip-python',
        typescript: 'https://github.com/sourcegraph/scip-typescript',
        java: 'https://github.com/sourcegraph/scip-java',
        ruby: 'https://github.com/sourcegraph/scip-ruby',
        go: 'https://github.com/sourcegraph/lsif-go',
        rust: 'https://github.com/rust-analyzer/rust-analyzer',
        scala: 'https://github.com/sourcegraph/lsif-java',
        cpp: 'https://github.com/sourcegraph/lsif-clang',
        csharp: 'https://github.com/tcz717/LsifDotnet',
        dart: 'https://github.com/sourcegraph/lsif-dart',
        haskell: 'https://github.com/mpickering/hie-lsif',
        kotlin: 'https://github.com/sourcegraph/lsif-java',
    }

    const topRepos: TopRepo[] | undefined = (() => {
        const allRows = data?.site.analytics.codeIntelTopRepositories
        if (!allRows) {
            return undefined
        }

        const unsortedRepos = Object.entries(groupBy(allRows, row => row.name)).map(([name, rows]) => ({
            repoName: name,
            events: sumBy(rows, row => row.events),
            hoursSaved: sumBy(
                rows,
                row => (((kindToMinPerItem[row.kind as Kind] as number | undefined) ?? 0) * row.events) / 60
            ),
            preciseEnabled: rows[0]?.hasPrecise ?? false,
            preciseNavigation: ((): JSX.Element => {
                interface Item {
                    brand: 'precise' | 'configurable'
                    element: React.ReactNode
                }

                const items: Item[] = Object.entries(groupBy(rows, row => row.language))
                    .map(([lang, rows]): Item | undefined => {
                        const searchBased = sumBy(
                            rows.filter(row => row.precision === 'search-based'),
                            row => row.events
                        )
                        const precise = sumBy(
                            rows.filter(row => row.precision === 'precise'),
                            row => row.events
                        )
                        const total = searchBased + precise

                        if (precise > 0) {
                            return {
                                brand: 'precise',
                                element: (
                                    <div key={lang} className={styles.preciseItem}>
                                        <strong>{Math.round((precise / total) * 100)}%</strong> Precise coverage for{' '}
                                        <strong>{lang}</strong>
                                    </div>
                                ),
                            }
                        }
                        if (lang in langToIndexerUrl) {
                            return {
                                brand: 'configurable',
                                element: <Link to={langToIndexerUrl[lang]}>{lang}</Link>,
                            }
                        }
                        return undefined
                    })
                    .filter((item): item is Item => item !== undefined)

                return (
                    <>
                        {items.filter(item => item.brand === 'precise').map(item => item.element)}
                        {items.some(item => item.brand === 'configurable') && (
                            <div className={styles.preciseItem}>
                                Configure precise navigation for{' '}
                                {items
                                    .filter(item => item.brand === 'configurable')
                                    .map(item => item.element)
                                    .reduce((acc, item) => [acc, ', ', item])}
                            </div>
                        )}
                    </>
                )
            })(),
        }))

        return sortBy(unsortedRepos, repo => -repo.events)
    })()

    return (
        <>
            <AnalyticsPageTitle>Code navigation</AnalyticsPageTitle>

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
                    <ul className="mb-3 pl-3">
                        <Text as="li">
                            Promote installation of the{' '}
                            <AnchorLink to="/help/integration/browser_extension" target="_blank">
                                browser extension
                            </AnchorLink>{' '}
                            to add code intelligence to your code hosts.
                        </Text>
                        {repos && (
                            <Text as="li">
                                <b>{repos.preciseCodeIntelCount}</b> of your <b>{repos.count}</b> repositories have
                                precise code navigation.{' '}
                                <AnchorLink
                                    to="/help/code_navigation/explanations/precise_code_intelligence"
                                    target="_blank"
                                >
                                    Learn how to improve precise code navigation coverage.
                                </AnchorLink>
                            </Text>
                        )}
                    </ul>
                </div>
                <div>
                    <H4 className="my-3">Events by language</H4>
                    {data && (
                        <div className={styles.events}>
                            <ChartContainer className={styles.chart} labelX="Languages" labelY="Events">
                                {width => (
                                    <>
                                        <LegendList>
                                            <LegendItem color={color('precise')} name="precise" />
                                            <LegendItem color={color('search-based')} name="search-based" />
                                        </LegendList>
                                        <BarChart
                                            stacked={true}
                                            width={width}
                                            height={300}
                                            data={data.site.analytics.codeIntelByLanguage}
                                            getDatumColor={value => color(value.precision)}
                                            getDatumValue={value => value.count}
                                            getDatumName={value => value.language}
                                            getDatumHover={value => `${value.language} ${value.precision}`}
                                        />
                                    </>
                                )}
                            </ChartContainer>
                            <div className={styles.percentContainer}>
                                <div className={styles.percent}>
                                    {preciseFraction ? (100 * preciseFraction).toFixed(1) : '...'}%
                                </div>
                                <div>Precise code navigation</div>
                            </div>
                        </div>
                    )}
                    <H4 className="my-3">Top repositories</H4>
                    {topRepos && (
                        <div className={styles.repos}>
                            <div className="text-muted text-nowrap">{/* Repository */}</div>
                            <div className="text-center text-muted text-nowrap">Events</div>
                            <div className="text-center text-muted text-nowrap">Hours saved</div>
                            <div className="text-center text-muted text-nowrap">Precise enabled</div>
                            <div className="text-muted text-nowrap">Precise navigation</div>
                            {topRepos.map((repo, index) => (
                                <React.Fragment key={index}>
                                    <Text className="text-muted">{repo.repoName}</Text>
                                    <Text weight="bold" className="text-center">
                                        {formatNumber(repo.events)}
                                    </Text>
                                    <Text weight="bold" className="text-center">
                                        {formatNumber(repo.hoursSaved)}
                                    </Text>
                                    <Text weight="bold" className="text-center">
                                        {repo.preciseEnabled ? 'Yes' : 'No'}
                                    </Text>
                                    <Text>{repo.preciseNavigation}</Text>
                                </React.Fragment>
                            ))}
                        </div>
                    )}
                </div>
            </Card>
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table and are updated every 24 hours.
                <br />* Calculated from precise code navigation events
            </Text>
        </>
    )
}

const color = (precision: string): string => {
    switch (precision) {
        case 'precise':
            return 'rgb(255, 184, 109)'
        case 'search-based':
            return 'rgb(155, 211, 255)'
        default:
            return 'gray'
    }
}
