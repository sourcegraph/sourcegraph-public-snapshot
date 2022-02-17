import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import React from 'react'
import { LineChartContent, LineChartContent as LineChartContentType, LineChartSeries } from 'sourcegraph'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Link } from '@sourcegraph/wildcard'

import * as View from '../../../../../../views'
import { LegendBlock, LegendItem } from '../../../../../../views'
import {
    getLineStroke,
    LineChart,
} from '../../../../../../views/components/view/content/chart-view-content/charts/line/components/LineChartContent'
import { CodeInsightTrackType, useCodeInsightViewPings } from '../../../../pings'
import { encodeCaptureInsightURL } from '../../../insights/creation/capture-group'
import { DATA_SERIES_COLORS, encodeSearchInsightUrl } from '../../../insights/creation/search-insight'
import { CodeInsightsQueryBlock } from '../code-insights-query-block/CodeInsightsQueryBlock'

import styles from './CodeInsightsExamples.module.scss'

export interface CodeInsightsExamplesProps extends TelemetryProps, React.HTMLAttributes<HTMLElement> {}

export const CodeInsightsExamples: React.FunctionComponent<CodeInsightsExamplesProps> = props => {
    const { telemetryService, ...otherProps } = props
    return (
        <section {...otherProps}>
            <h2>Example insights</h2>
            <p className="text-muted">
                Here are a few example insights to show you what the tool can do.{' '}
                <Link to="/help/code_insights/references/common_use_cases" rel="noopener noreferrer" target="_blank">
                    Explore more use cases.
                </Link>
            </p>

            <div className={styles.section}>
                <CodeInsightSearchExample telemetryService={telemetryService} className={styles.card} />
                <CodeInsightCaptureExample telemetryService={telemetryService} className={styles.card} />
            </div>
        </section>
    )
}

interface ExampleCardProps extends TelemetryProps {
    className?: string
}

interface SeriesWithQuery extends LineChartSeries<any> {
    query: string
    name: string
}

type Content = Omit<LineChartContentType<any, string>, 'chart' | 'series'> & { series: SeriesWithQuery[] }

const SEARCH_INSIGHT_EXAMPLES_DATA: Content = {
    data: [
        { x: new Date('May 7, 2021'), a: 88, b: 410 },
        { x: new Date('June 7, 2021'), a: 95, b: 410 },
        { x: new Date('July 7, 2021'), a: 110, b: 315 },
        { x: new Date('August 7, 2021'), a: 160, b: 180 },
        { x: new Date('September 7, 2021'), a: 310, b: 90 },
        { x: new Date('October 7, 2021'), a: 520, b: 45 },
        { x: new Date('November 7, 2021'), a: 700, b: 10 },
    ],
    series: [
        {
            dataKey: 'a',
            name: 'CSS Modules',
            stroke: DATA_SERIES_COLORS.GREEN,
            query: 'select:file lang:scss file:module.scss patterntype:regexp archived:no fork:no',
        },
        {
            dataKey: 'b',
            name: 'Global CSS',
            stroke: DATA_SERIES_COLORS.RED,
            query: 'select:file lang:scss -file:module.scss patterntype:regexp archived:no fork:no',
        },
    ],
    xAxis: {
        dataKey: 'x',
        scale: 'time',
        type: 'number',
    },
}

const SEARCH_INSIGHT_CREATION_UI_URL_PARAMETERS = encodeSearchInsightUrl({
    title: 'Migration to CSS modules',
    repositories: 'repo:github.com/awesomeOrg/examplerepo',
    series: SEARCH_INSIGHT_EXAMPLES_DATA.series,
})

const CodeInsightSearchExample: React.FunctionComponent<ExampleCardProps> = props => {
    const { telemetryService, className } = props
    const { trackMouseEnter, trackMouseLeave } = useCodeInsightViewPings({
        telemetryService,
        insightType: CodeInsightTrackType.InProductLandingPageInsight,
    })

    const handleTemplateLinkClick = (): void => {
        telemetryService.log('InsightsGetStartedBigTemplateClick')
    }

    return (
        <View.Root
            title="Migration to CSS modules"
            subtitle={
                <CodeInsightsQueryBlock
                    as={SyntaxHighlightedSearchQuery}
                    query="repo:github.com/awesomeOrg/examplerepo"
                    className="mt-1"
                />
            }
            className={classNames(className)}
            actions={
                <Button
                    as={Link}
                    variant="link"
                    size="sm"
                    className={styles.actionLink}
                    to={`/insights/create/search?${SEARCH_INSIGHT_CREATION_UI_URL_PARAMETERS}`}
                    onClick={handleTemplateLinkClick}
                >
                    Use as template
                </Button>
            }
            onMouseEnter={trackMouseEnter}
            onMouseLeave={trackMouseLeave}
        >
            <div className={styles.chart}>
                <ParentSize>
                    {({ width, height }) => (
                        <LineChart {...SEARCH_INSIGHT_EXAMPLES_DATA} width={width} height={height} />
                    )}
                </ParentSize>
            </div>

            <LegendBlock className={styles.legend}>
                {SEARCH_INSIGHT_EXAMPLES_DATA.series.map(line => (
                    <LegendItem key={line.dataKey.toString()} color={getLineStroke<any>(line)}>
                        <span className={classNames(styles.legendMigrationItem, 'flex-shrink-0 mr-2')}>
                            {line.name}
                        </span>
                        <CodeInsightsQueryBlock as={SyntaxHighlightedSearchQuery} query={line.query} />
                    </LegendItem>
                ))}
            </LegendBlock>
        </View.Root>
    )
}

const CAPTURE_INSIGHT_EXAMPLES_DATA: LineChartContent<any, string> = {
    chart: 'line' as const,
    data: [
        { x: new Date('May 7, 2021'), a: 100, b: 160, c: 90, d: 75, e: 85, f: 20, g: 150 },
        { x: new Date('June 7, 2021'), a: 90, b: 155, c: 95, d: 85, e: 80, f: 25, g: 155 },
        { x: new Date('July 7, 2021'), a: 85, b: 150, c: 110, d: 90, e: 60, f: 40, g: 165 },
        { x: new Date('August 7, 2021'), a: 85, b: 150, c: 125, d: 80, e: 50, f: 50, g: 165 },
        { x: new Date('September 7, 2021'), a: 70, b: 155, c: 125, d: 75, e: 45, f: 55, g: 160 },
        { x: new Date('October 7, 2021'), a: 50, b: 150, c: 145, d: 70, e: 35, f: 60, g: 155 },
        { x: new Date('November 7, 2021'), a: 35, b: 160, c: 175, d: 75, e: 45, f: 65, g: 145 },
    ],
    series: [
        {
            dataKey: 'a',
            name: '3.1',
            stroke: DATA_SERIES_COLORS.INDIGO,
        },
        {
            dataKey: 'b',
            name: '3.5',
            stroke: DATA_SERIES_COLORS.RED,
        },
        {
            dataKey: 'c',
            name: '3.15',
            stroke: DATA_SERIES_COLORS.GREEN,
        },
        {
            dataKey: 'd',
            name: '3.8',
            stroke: DATA_SERIES_COLORS.GRAPE,
        },
        {
            dataKey: 'e',
            name: '3.9',
            stroke: DATA_SERIES_COLORS.ORANGE,
        },
        {
            dataKey: 'f',
            name: '3.9.2',
            stroke: DATA_SERIES_COLORS.TEAL,
        },
        {
            dataKey: 'g',
            name: '3.14',
            stroke: DATA_SERIES_COLORS.PINK,
        },
    ],
    xAxis: {
        dataKey: 'x',
        scale: 'time' as const,
        type: 'number',
    },
}

const CAPTURE_GROUP_INSIGHT_CREATION_UI_URL_PARAMETERS = encodeCaptureInsightURL({
    title: 'Alpine versions over all repos',
    allRepos: true,
    groupSearchQuery: 'patterntype:regexp FROM\\s+alpine:([\\d\\.]+) file:Dockerfile',
})

const CodeInsightCaptureExample: React.FunctionComponent<ExampleCardProps> = props => {
    const { telemetryService, className } = props
    const { trackMouseEnter, trackMouseLeave } = useCodeInsightViewPings({
        telemetryService,
        insightType: CodeInsightTrackType.InProductLandingPageInsight,
    })

    const handleTemplateLinkClick = (): void => {
        telemetryService.log('GetStartedBigTemplateClick')
    }

    return (
        <View.Root
            title="Alpine versions over all repos"
            subtitle={
                <CodeInsightsQueryBlock as={SyntaxHighlightedSearchQuery} query="All repositories" className="mt-1" />
            }
            actions={
                <Button
                    as={Link}
                    variant="link"
                    size="sm"
                    className={styles.actionLink}
                    to={`/insights/create/capture-group?${CAPTURE_GROUP_INSIGHT_CREATION_UI_URL_PARAMETERS}`}
                    onClick={handleTemplateLinkClick}
                >
                    Use as template
                </Button>
            }
            className={className}
            onMouseEnter={trackMouseEnter}
            onMouseLeave={trackMouseLeave}
        >
            <div className={styles.captureGroup}>
                <div className={styles.chart}>
                    <ParentSize className={styles.chartContent}>
                        {({ width, height }) => (
                            <LineChart {...CAPTURE_INSIGHT_EXAMPLES_DATA} width={width} height={height} />
                        )}
                    </ParentSize>
                </div>

                <LegendBlock className={styles.legendHorizontal}>
                    {CAPTURE_INSIGHT_EXAMPLES_DATA.series.map(line => (
                        <LegendItem key={line.dataKey.toString()} color={getLineStroke<any>(line)}>
                            {line.name}
                        </LegendItem>
                    ))}
                </LegendBlock>
            </div>
            <CodeInsightsQueryBlock
                as={SyntaxHighlightedSearchQuery}
                query="patterntype:regexp FROM\s+alpine:([\d\.]+) file:Dockerfile"
                className="mt-2"
            />
        </View.Root>
    )
}
