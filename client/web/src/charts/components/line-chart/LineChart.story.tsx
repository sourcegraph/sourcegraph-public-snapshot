import React from 'react'

import { Meta, Story } from '@storybook/react'
import { ParentSize } from '@visx/responsive'

import { WebStory } from '../../../components/WebStory'
import { Series } from '../../types'

import { LineChart, LegendList, LegendItem, getLineColor } from '.'

export const StoryConfig: Meta = {
    title: 'web/charts/line',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default StoryConfig

export const LineChartsVitrina: Story = () => (
    <div className="d-flex flex-wrap">
        <PlainChart />
        <PlainStackedChart />
        <WithLegendExample />
        <WithHugeData />
        <WithZeroOneData />
        <WithDataSteps />
        <WithDataMissingValues />
        <StackedWithDataMissingValues />
    </div>
)

interface StandardDatum {
    a: number | null
    aLink: string
    b: number | null
    bLink: string
    c: number | null
    cLink: string
    x: number | null
}

const PlainChart = () => {
    const DATA: StandardDatum[] = [
        {
            x: 1588965700286 - 4 * 24 * 60 * 60 * 1000,
            a: 4000,
            aLink: 'https://google.com/search',
            b: 15000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
        {
            x: 1588965700286 - 3 * 24 * 60 * 60 * 1000,
            a: 4000,
            aLink: 'https://google.com/search',
            b: 26000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
        {
            x: 1588965700286 - 2 * 24 * 60 * 60 * 1000,
            a: 5600,
            aLink: 'https://google.com/search',
            b: 20000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
        {
            x: 1588965700286 - 24 * 60 * 60 * 1000,
            a: 9800,
            aLink: 'https://google.com/search',
            b: 19000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
        {
            x: 1588965700286,
            a: 6000,
            aLink: 'https://google.com/search',
            b: 17000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
    ]

    const SERIES: Series<StandardDatum>[] = [
        {
            dataKey: 'a',
            name: 'A metric',
            color: 'var(--blue)',
            getLinkURL: datum => datum.aLink,
        },
        {
            dataKey: 'b',
            name: 'B metric',
            color: 'var(--warning)',
            getLinkURL: datum => datum.bLink,
        },
        {
            dataKey: 'c',
            name: 'C metric',
            color: 'var(--green)',
            getLinkURL: datum => datum.cLink,
        },
    ]

    return <LineChart width={400} height={400} data={DATA} series={SERIES} xAxisKey="x" />
}

const PlainStackedChart = () => {
    const DATA: StandardDatum[] = [
        {
            x: 1588965700286 - 4 * 24 * 60 * 60 * 1000,
            a: 4000,
            aLink: 'https://google.com/search',
            b: 15000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
        {
            x: 1588965700286 - 3 * 24 * 60 * 60 * 1000,
            a: 4000,
            aLink: 'https://google.com/search',
            b: 26000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
        {
            x: 1588965700286 - 2 * 24 * 60 * 60 * 1000,
            a: 5600,
            aLink: 'https://google.com/search',
            b: 20000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
        {
            x: 1588965700286 - 24 * 60 * 60 * 1000,
            a: 9800,
            aLink: 'https://google.com/search',
            b: 19000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
        {
            x: 1588965700286,
            a: 6000,
            aLink: 'https://google.com/search',
            b: 17000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
    ]

    const SERIES: Series<StandardDatum>[] = [
        {
            dataKey: 'a',
            name: 'A metric',
            color: 'var(--blue)',
            getLinkURL: datum => datum.aLink,
        },
        {
            dataKey: 'b',
            name: 'B metric',
            color: 'var(--warning)',
            getLinkURL: datum => datum.bLink,
        },
        {
            dataKey: 'c',
            name: 'C metric',
            color: 'var(--green)',
            getLinkURL: datum => datum.cLink,
        },
    ]

    return <LineChart width={400} height={400} data={DATA} series={SERIES} xAxisKey="x" stacked={true} />
}

const WithLegendExample = () => {
    const DATA: StandardDatum[] = [
        {
            x: 1588965700286 - 4 * 24 * 60 * 60 * 1000,
            a: 4000,
            aLink: 'https://google.com/search',
            b: 15000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
        {
            x: 1588965700286 - 3 * 24 * 60 * 60 * 1000,
            a: 4000,
            aLink: 'https://google.com/search',
            b: 26000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
        {
            x: 1588965700286 - 2 * 24 * 60 * 60 * 1000,
            a: 5600,
            aLink: 'https://google.com/search',
            b: 20000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
        {
            x: 1588965700286 - 24 * 60 * 60 * 1000,
            a: 9800,
            aLink: 'https://google.com/search',
            b: 19000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
        {
            x: 1588965700286,
            a: 6000,
            aLink: 'https://google.com/search',
            b: 17000,
            bLink: 'https://yandex.com/search',
            c: 5000,
            cLink: 'https://twitter.com/search',
        },
    ]

    const SERIES: Series<StandardDatum>[] = [
        {
            dataKey: 'a',
            name: 'A metric',
            color: 'var(--blue)',
            getLinkURL: datum => datum.aLink,
        },
        {
            dataKey: 'b',
            name: 'B metric',
            color: 'var(--warning)',
            getLinkURL: datum => datum.bLink,
        },
        {
            dataKey: 'c',
            name: 'C metric',
            color: 'var(--green)',
            getLinkURL: datum => datum.cLink,
        },
    ]

    return (
        <div className="d-flex flex-column" style={{ width: 400, height: 400 }}>
            <ParentSize className="flex-1">
                {({ width, height }) => (
                    <LineChart<StandardDatum> width={width} height={height} data={DATA} series={SERIES} xAxisKey="x" />
                )}
            </ParentSize>
            <LegendList>
                {SERIES.map(line => (
                    <LegendItem key={line.dataKey.toString()} color={getLineColor(line)} name={line.name} />
                ))}
            </LegendList>
        </div>
    )
}

interface HugeDataDatum {
    series0: number | null
    series1: number | null
    dateTime: number
}

const WithHugeData = () => {
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

    const SERIES: Series<HugeDataDatum>[] = [
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

const WithZeroOneData = () => {
    const DATA: ZeroOneDatum[] = [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 0 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 1 },
    ]

    const SERIES: Series<ZeroOneDatum>[] = [
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

const WithDataSteps = () => {
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

    const SERIES: Series<StepDatum>[] = [
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
    c: number | null
    x: number
}

const WithDataMissingValues = () => {
    const DATA_WITH_STEP: DatumWithMissingData[] = [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: null, b: null, c: null },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: null, b: null, c: null },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 94, b: 200, c: 200 },
        { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, a: 134, b: null, c: 134 },
        { x: 1588965700286 - 1.3 * 24 * 60 * 60 * 1000, a: null, b: 150, c: null },
        { x: 1588965700286 - 24 * 60 * 60 * 1000, a: 134, b: 190, c: 134 },
        { x: 1588965700286, a: 123, b: 170, c: 123 },
    ]

    const SERIES: Series<DatumWithMissingData>[] = [
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

const StackedWithDataMissingValues = () => {
    const DATA_WITH_STEP: DatumWithMissingData[] = [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: null, b: null, c: null },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: null, b: null, c: null },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 94, b: null, c: null },
        { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, a: 134, b: null, c: 200 },
        { x: 1588965700286 - 1.3 * 24 * 60 * 60 * 1000, a: null, b: 150, c: 150 },
        { x: 1588965700286 - 24 * 60 * 60 * 1000, a: 134, b: 190, c: 190 },
        { x: 1588965700286, a: 123, b: 170, c: 170 },
    ]

    const SERIES: Series<DatumWithMissingData>[] = [
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
        {
            dataKey: 'c',
            name: 'C metric',
            color: 'var(--purple)',
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
                        stacked={true}
                    />
                )}
            </ParentSize>
        </div>
    )
}
