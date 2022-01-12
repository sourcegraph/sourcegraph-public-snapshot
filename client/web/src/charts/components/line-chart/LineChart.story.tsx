import { Meta } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { LineChartSeries } from './types'

import { LineChart, LegendList, ParentSize } from '.'

export default {
    title: 'web/charts/line',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta

interface StandardDatum {
    a: number | null
    b: number | null
    c: number | null
    x: number | null
}

export const PlainChart = () => {
    const DATA: StandardDatum[] = [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, c: 5000, a: 4000, b: 15000 },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, c: 5000, a: 4000, b: 26000 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, c: 5000, a: 5600, b: 20000 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, c: 5000, a: 9800, b: 19000 },
        { x: 1588965700286, c: 5000, a: 6000, b: 17000 },
    ]

    const SERIES: LineChartSeries<StandardDatum>[] = [
        {
            dataKey: 'a',
            name: 'A metric',
            color: 'var(--blue)',
            linkURLs: [
                'https://google.com/search',
                'https://google.com/search',
                'https://google.com/search',
                'https://google.com/search',
                'https://google.com/search',
            ],
        },
        {
            dataKey: 'b',
            name: 'B metric',
            color: 'var(--warning)',
            linkURLs: [
                'https://yandex.com/search',
                'https://yandex.com/search',
                'https://yandex.com/search',
                'https://yandex.com/search',
                'https://yandex.com/search',
            ],
        },
        {
            dataKey: 'c',
            name: 'C metric',
            color: 'var(--green)',
            linkURLs: [
                'https://twitter.com/search',
                'https://twitter.com/search',
                'https://twitter.com/search',
                'https://twitter.com/search',
                'https://twitter.com/search',
            ],
        },
    ]

    return <LineChart width={400} height={400} data={DATA} series={SERIES} xAxisKey="x" />
}

export const WithLegendExample = () => {
    const DATA: StandardDatum[] = [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, c: 5000, a: 4000, b: 15000 },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, c: 5000, a: 4000, b: 26000 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, c: 5000, a: 5600, b: 20000 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, c: 5000, a: 9800, b: 19000 },
        { x: 1588965700286, c: 5000, a: 6000, b: 17000 },
    ]

    const SERIES: LineChartSeries<StandardDatum>[] = [
        {
            dataKey: 'a',
            name: 'A metric',
            color: 'var(--blue)',
            linkURLs: [
                'https://google.com/search',
                'https://google.com/search',
                'https://google.com/search',
                'https://google.com/search',
                'https://google.com/search',
            ],
        },
        {
            dataKey: 'b',
            name: 'B metric',
            color: 'var(--warning)',
            linkURLs: [
                'https://yandex.com/search',
                'https://yandex.com/search',
                'https://yandex.com/search',
                'https://yandex.com/search',
                'https://yandex.com/search',
            ],
        },
        {
            dataKey: 'c',
            name: 'C metric',
            color: 'var(--green)',
            linkURLs: [
                'https://twitter.com/search',
                'https://twitter.com/search',
                'https://twitter.com/search',
                'https://twitter.com/search',
                'https://twitter.com/search',
            ],
        },
    ]

    return (
        <div className="d-flex flex-column" style={{ width: 400, height: 400 }}>
            <ParentSize className="flex-1">
                {({ width, height }) => (
                    <LineChart<StandardDatum> width={width} height={height} data={DATA} series={SERIES} xAxisKey="x" />
                )}
            </ParentSize>
            <LegendList series={SERIES} />
        </div>
    )
}

interface HugeDataDatum {
    series0: number | null
    series1: number | null
    dateTime: number
}

export const WithHugeData = () => {
    const DATA: HugeDataDatum[] = [
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
    ]

    const SERIES: LineChartSeries<HugeDataDatum>[] = [
        { name: 'Fix', dataKey: 'series0', color: 'var(--oc-indigo-7)' },
        { name: 'Revert', dataKey: 'series1', color: 'var(--oc-orange-7)' },
    ]

    return (
        <div style={{ width: 400, height: 400 }}>
            <ParentSize>
                {({ width, height }) => (
                    <LineChart<HugeDataDatum>
                        width={width}
                        height={height}
                        data={DATA}
                        series={SERIES}
                        xAxisKey="dateTime"
                    />
                )}
            </ParentSize>
        </div>
    )
}

interface ZeroOneDatum {
    a: number
    x: number
}

export const WithZeroOneData = () => {
    const DATA: ZeroOneDatum[] = [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 0 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 1 },
    ]

    const SERIES: LineChartSeries<ZeroOneDatum>[] = [
        {
            dataKey: 'a',
            name: 'A metric',
            color: 'var(--blue)',
        },
    ]

    return (
        <div style={{ width: 400, height: 400 }}>
            <ParentSize>
                {({ width, height }) => (
                    <LineChart<ZeroOneDatum> width={width} height={height} data={DATA} series={SERIES} xAxisKey="x" />
                )}
            </ParentSize>
        </div>
    )
}

interface StepDatum {
    series0: number
    dateTime: number
}

export const WithDataSteps = () => {
    const DATA_WITH_STEP: StepDatum[] = [
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

    const SERIES: LineChartSeries<StepDatum>[] = [
        {
            dataKey: 'series0',
            name: 'A metric',
            color: 'var(--blue)',
        },
    ]

    return (
        <div style={{ width: 400, height: 400 }}>
            <ParentSize>
                {({ width, height }) => (
                    <LineChart<StepDatum>
                        width={width}
                        height={height}
                        data={DATA_WITH_STEP}
                        series={SERIES}
                        xAxisKey="dateTime"
                    />
                )}
            </ParentSize>
        </div>
    )
}

interface DatumWithMissingData {
    a: number | null
    b: number | null
    x: number
}

export const WithDataMissingValues = () => {
    const DATA_WITH_STEP: DatumWithMissingData[] = [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: null, b: null },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: null, b: null },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 94, b: 200 },
        { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, a: 134, b: null },
        { x: 1588965700286 - 1.3 * 24 * 60 * 60 * 1000, a: null, b: 150 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, a: 134, b: 190 },
        { x: 1588965700286, a: 123, b: 170 },
    ]

    const SERIES: LineChartSeries<DatumWithMissingData>[] = [
        {
            dataKey: 'a',
            name: 'A metric',
            color: 'var(--blue)',
        },
        {
            dataKey: 'b',
            name: 'B metric',
            color: 'var(--warning)',
        },
    ]

    return (
        <div style={{ width: 400, height: 400 }}>
            <ParentSize>
                {({ width, height }) => (
                    <LineChart<DatumWithMissingData>
                        width={width}
                        height={height}
                        data={DATA_WITH_STEP}
                        series={SERIES}
                        xAxisKey="x"
                    />
                )}
            </ParentSize>
        </div>
    )
}
