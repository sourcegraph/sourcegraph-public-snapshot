import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import React from 'react'
import { LineChartContent, LineChartContent as LineChartContentType, LineChartSeries } from 'sourcegraph'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { Button, Link } from '@sourcegraph/wildcard'

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
            <Link to="/help/code_insights/references/common_use_cases" rel="noopener noreferrer" target="_blank">
                Explore more use cases.
            </Link>
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
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 88, b: 410 },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: 95, b: 410 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 110, b: 315 },
        { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, a: 160, b: 180 },
        { x: 1588965700286 - 1.3 * 24 * 60 * 60 * 1000, a: 310, b: 90 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, a: 520, b: 45 },
        { x: 1588965700286, a: 700, b: 10 },
    ],
    series: [
        {
            dataKey: 'a',
            name: 'CSS Modules',
            stroke: DATA_SERIES_COLORS.GREEN,
            query: 'type:file lang:scss file:module.scss patterntype:regexp archived:no fork:no',
        },
        {
            dataKey: 'b',
            name: 'Global CSS',
            stroke: DATA_SERIES_COLORS.RED,
            query: 'type:file lang:scss -file:module.scss patterntype:regexp archived:no fork:no',
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
    const { className } = props

    return (
        <View.Root
            title="Migration to CSS modules"
            subtitle={<InlineCodeBlock query="repo:github.com/awesomeOrg/examplerepo" className="mt-1" />}
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
                        <span className={classNames(styles.legendMigrationItem, 'flex-shrink-0 mr-2')}>
                            {line.name}
                        </span>
                        <InlineCodeBlock query={line.query} />
                    </LegendItem>
                ))}
            </LegendBlock>
        </View.Root>
    )
}

const CAPTURE_INSIGHT_EXAMPLES_DATA: LineChartContent<any, string> = {
    chart: 'line' as const,
    data: [
        { x: 1588965700286 - 6 * 24 * 60 * 60 * 1000, a: 200, b: 160, c: 150, d: 75, e: 45, f: 20 },
        { x: 1588965700286 - 5 * 24 * 60 * 60 * 1000, a: 200, b: 160, c: 150, d: 75, e: 60, f: 20 },
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 200, b: 160, c: 150, d: 75, e: 45, f: 20 },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: 200, b: 160, c: 150, d: 75, e: 45, f: 20 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 200, b: 160, c: 150, d: 75, e: 45, f: 20 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, a: 200, b: 160, c: 150, d: 75, e: 45, f: 20 },
        { x: 1588965700286, a: 200, b: 160, c: 150, d: 75, e: 45, f: 20 },
    ],
    series: [
        {
            dataKey: 'a',
            name: '17.3.1',
            stroke: DATA_SERIES_COLORS.ORANGE,
        },
        {
            dataKey: 'b',
            name: '17.3.0',
            stroke: DATA_SERIES_COLORS.BLUE,
        },
        {
            dataKey: 'c',
            name: '17.2.0',
            stroke: DATA_SERIES_COLORS.RED,
        },
        {
            dataKey: 'd',
            name: '17.1.1',
            stroke: DATA_SERIES_COLORS.GREEN,
        },
        {
            dataKey: 'e',
            name: '17.1.0',
            stroke: DATA_SERIES_COLORS.CYAN,
        },
        {
            dataKey: 'e',
            name: '17.0.1',
            stroke: DATA_SERIES_COLORS.GRAPE,
        },
    ],
    xAxis: {
        dataKey: 'x',
        scale: 'time' as const,
        type: 'number',
    },
}

const CAPTURE_GROUP_INSIGHT_CREATION_UI_URL_PARAMETERS = encodeCaptureInsightURL({
    title: 'Terraform versions (present or most popular)',
    allRepos: true,
    groupSearchQuery: 'app.terraform.io/(.*)\\n version =(.*)([0-9].[0-9].[0-9]) lang:Terraform archived:no fork:no',
})

const CodeInsightCaptureExample: React.FunctionComponent<ExampleCardProps> = props => {
    const { className } = props

    return (
        <View.Root
            title="Terraform versions (present or most popular)"
            subtitle={<InlineCodeBlock query="All repositories" className="mt-1" />}
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
            <InlineCodeBlock
                query="app.terraform.io/(.*)\n version =(.*)([0-9].[0-9].[0-9]) lang:Terraform archived:no fork:no"
                className="mt-2"
            />
        </View.Root>
    )
}

interface InlineCodeBlockProps extends React.HTMLAttributes<HTMLElement> {
    query: string
}

const InlineCodeBlock: React.FunctionComponent<InlineCodeBlockProps> = props => (
    <SyntaxHighlightedSearchQuery query={props.query ?? ''} className={classNames(styles.code, props.className)} />
)
