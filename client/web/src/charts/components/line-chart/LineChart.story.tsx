import { useState } from 'react'

import { Meta, Story } from '@storybook/react'
import { ParentSize } from '@visx/responsive'

import { WebStory } from '../../../components/WebStory'
import { Series } from '../../types'

import { LineChart, LegendList, LegendItem, getLineColor } from '.'

const StoryConfig: Meta = {
    title: 'web/charts/line',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default StoryConfig

export const LineChartsVitrina: Story = () => (
    <div className="d-flex flex-wrap" style={{ gap: 20 }}>
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
    value: number | null
    x: number
    link?: string
}

const getXValue = (datum: StandardDatum) => new Date(datum.x)
const getYValue = (datum: StandardDatum) => datum.value
const getLinkURL = (datum: StandardDatum) => datum.link

const STANDARD_SERIES: Series<StandardDatum>[] = [
    {
        id: 'series_001',
        data: [
            {
                x: 1588965700286 - 4 * 24 * 60 * 60 * 1000,
                value: 4000,
                link: 'https://google.com/search',
            },
            {
                x: 1588965700286 - 3 * 24 * 60 * 60 * 1000,
                value: 4000,
                link: 'https://google.com/search',
            },
            {
                x: 1588965700286 - 2 * 24 * 60 * 60 * 1000,
                value: 5600,
                link: 'https://google.com/search',
            },
            {
                x: 1588965700286 - 24 * 60 * 60 * 1000,
                value: 9800,
                link: 'https://google.com/search',
            },
            {
                x: 1588965700286,
                value: 6000,
            },
        ],
        name: 'A metric',
        color: 'var(--blue)',
        getXValue,
        getYValue,
        getLinkURL,
    },
    {
        id: 'series_003',
        data: [
            {
                x: 1588965700286 - 4 * 24 * 60 * 60 * 1000,
                value: 5000,
                link: 'https://twitter.com/search',
            },
            {
                x: 1588965700286 - 3 * 24 * 60 * 60 * 1000,
                value: 5000,
                link: 'https://twitter.com/search',
            },
            {
                x: 1588965700286 - 2 * 24 * 60 * 60 * 1000,
                value: 5000,
                link: 'https://twitter.com/search',
            },
            {
                x: 1588965700286 - 24 * 60 * 60 * 1000,
                value: 5000,
                link: 'https://twitter.com/search',
            },
            {
                x: 1588965700286,
                value: 5000,
                link: 'https://twitter.com/search',
            },
        ],
        name: 'C metric',
        color: 'var(--green)',
        getXValue,
        getYValue,
        getLinkURL,
    },
    {
        id: 'series_002',
        data: [
            {
                x: 1588965700286 - 4 * 24 * 60 * 60 * 1000,
                value: 15000,
                link: 'https://yandex.com/search',
            },
            {
                x: 1588965700286 - 3 * 24 * 60 * 60 * 1000,
                value: 26000,
                link: 'https://yandex.com/search',
            },
            {
                x: 1588965700286 - 2 * 24 * 60 * 60 * 1000,
                value: 20000,
                link: 'https://yandex.com/search',
            },
            {
                x: 1588965700286 - 24 * 60 * 60 * 1000,
                value: 19000,
                link: 'https://yandex.com/search',
            },
            {
                x: 1588965700286,
                value: 17000,
                link: 'https://yandex.com/search',
            },
        ],
        name: 'B metric',
        color: 'var(--warning)',
        getXValue,
        getYValue,
        getLinkURL,
    },
]

const PlainChart = () => (
    <div style={{ width: 400, height: 400 }}>
        <ParentSize className="flex-1">
            {({ width, height }) => <LineChart width={width} height={height} series={STANDARD_SERIES} />}
        </ParentSize>
    </div>
)

const PlainStackedChart = () => {
    const [active, setActive] = useState(false)

    return (
        <section>
            <button className="d-block" onClick={() => setActive(!active)}>
                Toggle zero Y axis state
            </button>
            <LineChart width={400} height={400} series={STANDARD_SERIES} stacked={true} zeroYAxisMin={active} />
        </section>
    )
}

const WithLegendExample = () => (
    <div className="d-flex flex-column" style={{ width: 400, height: 400 }}>
        <ParentSize className="flex-1">
            {({ width, height }) => <LineChart<StandardDatum> width={width} height={height} series={STANDARD_SERIES} />}
        </ParentSize>
        <LegendList>
            {STANDARD_SERIES.map(line => (
                <LegendItem key={line.id} color={getLineColor(line)} name={line.name} />
            ))}
        </LegendList>
    </div>
)

const WithHugeData = () => {
    const SERIES: Series<StandardDatum>[] = [
        {
            id: 'series_001',
            data: [
                { x: 1606780800000, value: 8394074 },
                { x: 1609459200000, value: 839476900 },
                { x: 1612137600000, value: 8395504 },
                { x: 1614556800000, value: 839684900 },
                { x: 1617235200000, value: 8397911 },
                { x: 1619827200000, value: 839922700 },
                { x: 1622505600000, value: 8400349 },
                { x: 1625097600000, value: 840148500 },
                { x: 1627776000000, value: 8402574 },
                { x: 1630454400000, value: 840362900 },
                { x: 1633046400000, value: 8374023 },
                { x: 1635724800000, value: 837455000 },
            ],
            name: 'Fix',
            color: 'var(--oc-indigo-7)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_002',
            data: [
                { x: 1606780800000, value: 1001777 },
                { x: 1609459200000, value: 100180700 },
                { x: 1612137600000, value: 1001844 },
                { x: 1614556800000, value: 1001966 },
                { x: 1617235200000, value: 1002005 },
                { x: 1619827200000, value: 100202500 },
                { x: 1622505600000, value: 1002137 },
                { x: 1625097600000, value: 100218000 },
                { x: 1627776000000, value: 1002280 },
                { x: 1630454400000, value: 100237600 },
                { x: 1633046400000, value: null },
                { x: 1635724800000, value: null },
            ],
            color: 'var(--oc-orange-7)',
            name: 'Revert',
            getXValue,
            getYValue,
        },
    ]

    return (
        <div style={{ width: 400, height: 400 }}>
            <ParentSize>
                {({ width, height }) => <LineChart<StandardDatum> width={width} height={height} series={SERIES} />}
            </ParentSize>
        </div>
    )
}

const WithZeroOneData = () => {
    const SERIES: Series<StandardDatum>[] = [
        {
            id: 'series_001',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 0 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 1 },
            ],
            name: 'A metric',
            color: 'var(--blue)',
            getXValue,
            getYValue,
        },
    ]

    return (
        <div style={{ width: 400, height: 400 }}>
            <ParentSize>
                {({ width, height }) => <LineChart<StandardDatum> width={width} height={height} series={SERIES} />}
            </ParentSize>
        </div>
    )
}

const WithDataSteps = () => {
    const SERIES: Series<StandardDatum>[] = [
        {
            id: 'series_001',
            data: [
                { x: 1604188800000, value: 3725 },
                { x: 1606780800000, value: 3725 },
                { x: 1609459200000, value: 3725 },
                { x: 1612137600000, value: 3725 },
                { x: 1614556800000, value: 3725 },
                { x: 1617235200000, value: 3725 },
                { x: 1619827200000, value: 3728 },
                { x: 1622505600000, value: 3827 },
                { x: 1625097600000, value: 3827 },
                { x: 1627776000000, value: 3827 },
                { x: 1630458631000, value: 3053 },
                { x: 1633452311000, value: 3053 },
                { x: 1634952495000, value: 3053 },
            ],
            name: 'A metric',
            color: 'var(--blue)',
            getXValue,
            getYValue,
        },
    ]

    return (
        <div style={{ width: 400, height: 400 }}>
            <ParentSize>
                {({ width, height }) => <LineChart<StandardDatum> width={width} height={height} series={SERIES} />}
            </ParentSize>
        </div>
    )
}

const WithDataMissingValues = () => {
    const SERIES: Series<StandardDatum>[] = [
        {
            id: 'series_001',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 94 },
                { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, value: 134 },
                { x: 1588965700286 - 1.3 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 24 * 60 * 60 * 1000, value: 134 },
                { x: 1588965700286, value: 123 },
            ],
            name: 'A metric',
            color: 'var(--blue)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_002',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 200 },
                { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 1.3 * 24 * 60 * 60 * 1000, value: 150 },
                { x: 1588965700286 - 24 * 60 * 60 * 1000, value: 190 },
                { x: 1588965700286, value: 170 },
            ],
            name: 'B metric',
            color: 'var(--warning)',
            getXValue,
            getYValue,
        },
    ]

    return (
        <div style={{ width: 400, height: 400 }}>
            <ParentSize>
                {({ width, height }) => (
                    <LineChart<StandardDatum> width={width} height={height} series={SERIES} zeroYAxisMin={true} />
                )}
            </ParentSize>
        </div>
    )
}

const StackedWithDataMissingValues = () => {
    const SERIES: Series<StandardDatum>[] = [
        {
            id: 'series_001',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 94 },
                { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, value: 134 },
                { x: 1588965700286 - 1.4 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 1.3 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 24 * 60 * 60 * 1000, value: 134 },
                { x: 1588965700286, value: 123 },
                { x: 1588965700286 + 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 + 1.3 * 24 * 60 * 60 * 1000, value: null },
            ],
            name: 'A metric',
            color: 'var(--blue)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_002',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 1.4 * 24 * 60 * 60 * 1000, value: 150 },
                { x: 1588965700286 - 1.3 * 24 * 60 * 60 * 1000, value: 150 },
                { x: 1588965700286 - 24 * 60 * 60 * 1000, value: 190 },
                { x: 1588965700286, value: 170 },
                { x: 1588965700286 + 24 * 60 * 60 * 1000, value: 200 },
                { x: 1588965700286 + 1.3 * 24 * 60 * 60 * 1000, value: 180 },
            ],
            name: 'C metric',
            color: 'var(--purple)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_003',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, value: 200 },
                { x: 1588965700286 - 1.4 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 1.3 * 24 * 60 * 60 * 1000, value: 150 },
                { x: 1588965700286 - 24 * 60 * 60 * 1000, value: 190 },
                { x: 1588965700286, value: 170 },
                { x: 1588965700286 + 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 + 1.3 * 24 * 60 * 60 * 1000, value: null },
            ],
            name: 'B metric',
            color: 'var(--warning)',
            getXValue,
            getYValue,
        },
    ]

    return (
        <div style={{ width: 400, height: 400 }}>
            <ParentSize>
                {({ width, height }) => (
                    <LineChart<StandardDatum> width={width} height={height} series={SERIES} stacked={true} />
                )}
            </ParentSize>
        </div>
    )
}
