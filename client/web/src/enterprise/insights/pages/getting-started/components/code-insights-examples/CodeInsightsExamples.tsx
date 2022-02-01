import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'
import { LineChartContent, LineChartContent as LineChartContentType, LineChartSeries } from 'sourcegraph'

import { Button } from '@sourcegraph/wildcard'

import * as View from '../../../../../../views'
import { LegendBlock, LegendItem } from '../../../../../../views'
import {
    getLineStroke,
    LineChart,
} from '../../../../../../views/components/view/content/chart-view-content/charts/line/components/LineChartContent'
import { encodeCaptureInsightURL } from '../../../insights/creation/capture-group'
import { DATA_SERIES_COLORS, encodeSearchInsightUrl } from '../../../insights/creation/search-insight'

import styles from './CodeInsightsExamples.module.scss'

export interface CodeInsightsExamples extends React.HTMLAttributes<HTMLElement> {}

export const CodeInsightsExamples: React.FunctionComponent<CodeInsightsExamples> = props => (
    <section {...props}>
        <h2>Example insights</h2>
        <p className="text-muted">
            We've created a few common simple insights to show you what the tool can do.{' '}
            <a href="/help/code_insights/references/common_use_cases" rel="noopener noreferrer" target="_blank">
                Explore more use cases.
            </a>
        </p>

        <div className={styles.section}>
            <CodeInsightSearchExample className={styles.card} />
            <CodeInsightCaptureExample className={styles.card} />
        </div>
    </section>
)

interface ExampleCardProps {
    className?: string
}

interface SeriesWithQuery extends LineChartSeries<any> {
    query: string
    name: string
}

type Content = Omit<LineChartContentType<any, string>, 'chart' | 'series'> & { series: SeriesWithQuery[] }

const SEARCH_INSIGHT_EXAMPLES_DATA: Content = {
    data: [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: null, b: null },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: null, b: null },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 94, b: 200 },
        { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, a: 134, b: null },
        { x: 1588965700286 - 1.3 * 24 * 60 * 60 * 1000, a: null, b: 150 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, a: 134, b: 190 },
        { x: 1588965700286, a: 123, b: 170 },
    ],
    series: [
        {
            dataKey: 'a',
            name: 'A metric',
            stroke: DATA_SERIES_COLORS.BLUE,
            query: 'file:README archived:no fork:no',
        },
        {
            dataKey: 'b',
            name: 'B metric',
            stroke: DATA_SERIES_COLORS.ORANGE,
            query: '-file:README archived:no fork:no',
        },
    ],
    xAxis: {
        dataKey: 'x',
        scale: 'time',
        type: 'number',
    },
}

const SEARCH_INSIGHT_CREATION_UI_URL_PARAMETERS = encodeSearchInsightUrl({
    title: 'Repos with READMEs / without READMEs',
    allRepos: true,
    series: SEARCH_INSIGHT_EXAMPLES_DATA.series,
})

const CodeInsightSearchExample: React.FunctionComponent<ExampleCardProps> = props => {
    const { className } = props

    return (
        <View.Root
            title="Repos with READMEs / without READMEs"
            subtitle={<InlineCodeBlock className="mt-1">All repositories</InlineCodeBlock>}
            className={classNames(className)}
            actions={
                <Button
                    as={Link}
                    variant="link"
                    size="sm"
                    className={styles.actionLink}
                    to={`/insights/create/search?${SEARCH_INSIGHT_CREATION_UI_URL_PARAMETERS}`}
                >
                    Use as template
                </Button>
            }
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
                        <span className="flex-shrink-0 mr-2">{line.name}</span>
                        <InlineCodeBlock>{line.query}</InlineCodeBlock>
                    </LegendItem>
                ))}
            </LegendBlock>
        </View.Root>
    )
}

const CAPTURE_INSIGHT_EXAMPLES_DATA: LineChartContent<any, string> = {
    chart: 'line' as const,
    data: [
        { x: 1588965700286 - 6 * 24 * 60 * 60 * 1000, a: 20, b: 200 },
        { x: 1588965700286 - 5 * 24 * 60 * 60 * 1000, a: 40, b: 177 },
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 110, b: 150 },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: 105, b: 165 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 160, b: 100 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, a: 184, b: 85 },
        { x: 1588965700286, a: 200, b: 50 },
    ],
    series: [
        {
            dataKey: 'a',
            name: 'Go 1.11',
            stroke: 'var(--oc-indigo-7)',
        },
        {
            dataKey: 'b',
            name: 'Go 1.12',
            stroke: 'var(--oc-orange-7)',
        },
    ],
    xAxis: {
        dataKey: 'x',
        scale: 'time' as const,
        type: 'number',
    },
}

const CAPTURE_GROUP_INSIGHT_CREATION_UI_URL_PARAMETERS = encodeCaptureInsightURL({
    title: 'Node.js versions (present or most popular)',
    repositories: 'github.com/awesomeOrg/examplerepo',
    groupSearchQuery: 'nvm install ([0-9]+\\.[0-9]+) archived:no fork:no',
})

const CodeInsightCaptureExample: React.FunctionComponent<ExampleCardProps> = props => {
    const { className } = props

    return (
        <View.Root
            title="Node.js versions (present or most popular)"
            subtitle={<InlineCodeBlock className="mt-1">repo:github.com/awesomeOrg/examplerepo</InlineCodeBlock>}
            actions={
                <Button
                    as={Link}
                    variant="link"
                    size="sm"
                    className={styles.actionLink}
                    to={`/insights/create/capture-group?${CAPTURE_GROUP_INSIGHT_CREATION_UI_URL_PARAMETERS}`}
                >
                    Use as template
                </Button>
            }
            className={classNames(className)}
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
            <InlineCodeBlock className="mt-2">nvm install ([0-9]+\.[0-9]+) archived:no fork:no</InlineCodeBlock>
        </View.Root>
    )
}

const InlineCodeBlock: React.FunctionComponent<React.HTMLAttributes<HTMLElement>> = props => (
    <code className={classNames(styles.code, props.className)}>{props.children}</code>
)
