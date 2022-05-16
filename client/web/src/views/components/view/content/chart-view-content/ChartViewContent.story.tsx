import { storiesOf } from '@storybook/react'
import isChromatic from 'chromatic/isChromatic'
import { createMemoryHistory } from 'history'
import { ChartContent } from 'sourcegraph'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Typography } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../../components/WebStory'
import { LINE_CHART_CONTENT_MOCK, LINE_CHART_WITH_MANY_LINES } from '../../../../mocks/charts-content'

import { LineChartLayoutOrientation, LineChartSettingsContext } from './charts/line'
import { ChartViewContent } from './ChartViewContent'

import styles from './ChartViewContent.story.module.scss'

const history = createMemoryHistory()

const commonProps = {
    history,
    animate: !isChromatic(),
    viewID: '1',
    telemetryService: NOOP_TELEMETRY_SERVICE,
    className: styles.chart,
    locked: false,
}

const { add } = storiesOf('web/ChartViewContent', module).addDecorator(story => (
    <WebStory>{() => <div className={styles.charts}>{story()}</div>}</WebStory>
))

const DATA_WITH_STEP = [
    { dateTime: 1604188800000, series0: 3725 },
    {
        dateTime: 1606780800000,
        series0: 3725,
    },
    { dateTime: 1609459200000, series0: 3725 },
    {
        dateTime: 1612137600000,
        series0: 3725,
    },
    { dateTime: 1614556800000, series0: 3725 },
    {
        dateTime: 1617235200000,
        series0: 3725,
    },
    { dateTime: 1619827200000, series0: 3728 },
    {
        dateTime: 1622505600000,
        series0: 3827,
    },
    { dateTime: 1625097600000, series0: 3827 },
    {
        dateTime: 1627776000000,
        series0: 3827,
    },
    { dateTime: 1630458631000, series0: 3053 },
    {
        dateTime: 1633452311000,
        series0: 3053,
    },
    { dateTime: 1634952495000, series0: 3053 },
]

add('Line chart', () => (
    <>
        <ChartViewContent {...commonProps} content={LINE_CHART_CONTENT_MOCK} />
        <ChartViewContent
            {...commonProps}
            content={{
                chart: 'line',
                data: DATA_WITH_STEP,
                series: [
                    {
                        dataKey: 'series0',
                        name: 'Series 0',
                        stroke: 'var(--blue)',
                    },
                ],
                xAxis: {
                    dataKey: 'dateTime',
                    scale: 'time',
                    type: 'number',
                },
            }}
        />
    </>
))

add('Line chart with missing data', () => (
    <ChartViewContent
        {...commonProps}
        content={{
            chart: 'line',
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
                    stroke: 'var(--blue)',
                    linkURLs: {
                        [1588965700286 - 4 * 24 * 60 * 60 * 1000]: '#A:1st_data_point',
                        [1588965700286 - 3 * 24 * 60 * 60 * 1000]: '#A:2st_data_point',
                        [1588965700286 - 3 * 24 * 60 * 60 * 1000]: '#A:3rd_data_point',
                        [1588965700286 - 2 * 24 * 60 * 60 * 1000]: '#A:4th_data_point',
                        [1588965700286 - 1 * 24 * 60 * 60 * 1000]: '#A:5th_data_point',
                    },
                },
                {
                    dataKey: 'b',
                    name: 'B metric',
                    stroke: 'var(--warning)',
                },
            ],
            xAxis: {
                dataKey: 'x',
                scale: 'time',
                type: 'number',
            },
        }}
    />
))

add('Line chart with 0 to 1 data', () => (
    <ChartViewContent
        {...commonProps}
        content={{
            chart: 'line',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 0 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 1 },
            ],
            series: [
                {
                    dataKey: 'a',
                    name: 'A metric',
                    stroke: 'var(--red)',
                    linkURLs: ['#A:1st_data_point', 'https://example.com'],
                },
            ],
            xAxis: {
                dataKey: 'x',
                scale: 'time',
                type: 'number',
            },
        }}
    />
))

const DATA_WITH_HUGE_DATA: ChartContent = {
    chart: 'line',
    data: [
        { dateTime: 1606780800000, series0: 8394074, series1: 1001777 },
        {
            dateTime: 1609459200000,
            series0: 839476900,
            series1: 100180700,
        },
        { dateTime: 1612137600000, series0: 8395504, series1: 1001844 },
        {
            dateTime: 1614556800000,
            series0: 839684900,
            series1: 1001966,
        },
        { dateTime: 1617235200000, series0: 8397911, series1: 1002005 },
        {
            dateTime: 1619827200000,
            series0: 839922700,
            series1: 100202500,
        },
        { dateTime: 1622505600000, series0: 8400349, series1: 1002137 },
        {
            dateTime: 1625097600000,
            series0: 840148500,
            series1: 100218000,
        },
        { dateTime: 1627776000000, series0: 8402574, series1: 1002280 },
        {
            dateTime: 1630454400000,
            series0: 840362900,
            series1: 100237600,
        },
        { dateTime: 1633046400000, series0: 8374023, series1: null },
        {
            dateTime: 1635724800000,
            series0: 837455000,
            series1: null,
        },
    ],
    series: [
        { name: 'Fix', dataKey: 'series0', stroke: 'var(--oc-indigo-7)' },
        {
            name: 'Revert',
            dataKey: 'series1',
            stroke: 'var(--oc-orange-7)',
        },
    ],
    xAxis: { dataKey: 'dateTime', scale: 'time', type: 'number' },
}

add('Line chart with different data', () => <ChartViewContent {...commonProps} content={DATA_WITH_HUGE_DATA} />)

add('Bar chart', () => (
    <ChartViewContent
        {...commonProps}
        content={{
            chart: 'bar',
            data: [
                { name: 'A', value: 183 },
                { name: 'B', value: 145 },
                { name: 'C', value: 94 },
                { name: 'D', value: 134 },
                { name: 'E', value: 123 },
            ],
            series: [
                {
                    dataKey: 'value',
                    name: 'A metric',
                    fill: 'var(--oc-teal-7)',
                    linkURLs: [
                        '#1st_data_point',
                        '#2nd_data_point',
                        '#3rd_data_point',
                        '#4th_data_point',
                        '#5th_data_point',
                    ],
                },
            ],
            xAxis: {
                dataKey: 'name',
                type: 'category',
            },
        }}
    />
))

add('Pie chart', () => (
    <ChartViewContent
        {...commonProps}
        content={{
            chart: 'pie',
            pies: [
                {
                    dataKey: 'value',
                    nameKey: 'name',
                    fillKey: 'fill',
                    linkURLKey: 'linkURL',
                    data: [
                        {
                            name: 'Covered',
                            value: 0.3,
                            fill: 'var(--success)',
                            linkURL: '#Covered',
                        },
                        {
                            name: 'Not covered',
                            value: 0.7,
                            fill: 'var(--danger)',
                            linkURL: '#Not_covered',
                        },
                    ],
                },
            ],
        }}
    />
))

add('Line chart with horizontal layout', () => (
    <>
        <article>
            <Typography.H3>Middle width chart 2 lines</Typography.H3>
            <p>Legend block should be below the chart</p>
            <ChartViewContent {...commonProps} content={LINE_CHART_CONTENT_MOCK} />
        </article>

        <article>
            <Typography.H3>Big width chart 2 lines</Typography.H3>
            <p>
                Legend block should be below the chart even if we have enough space, but we have small number of series
            </p>
            <ChartViewContent {...commonProps} className={styles.chartLg} content={LINE_CHART_CONTENT_MOCK} />
        </article>

        <article>
            <Typography.H3>Middle width chart with many lines</Typography.H3>
            <p>Legend is placed below cause we don't have enough X space to put it aside</p>
            <ChartViewContent {...commonProps} content={LINE_CHART_WITH_MANY_LINES} />
        </article>

        <article>
            <Typography.H3>Big width chart with many lines</Typography.H3>
            <p>Legend is placed aside because we have enought X space</p>
            <ChartViewContent {...commonProps} className={styles.chartLg} content={LINE_CHART_WITH_MANY_LINES} />
        </article>

        <article>
            <Typography.H3>Middle width chart with many lines with explicit chart layout</Typography.H3>
            <LineChartSettingsContext.Provider
                value={{ zeroYAxisMin: false, layout: LineChartLayoutOrientation.Horizontal }}
            >
                <ChartViewContent {...commonProps} content={LINE_CHART_WITH_MANY_LINES} />
            </LineChartSettingsContext.Provider>
        </article>
    </>
))
