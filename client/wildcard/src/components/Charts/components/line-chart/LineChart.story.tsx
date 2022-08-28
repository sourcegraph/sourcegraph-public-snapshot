import { useState } from 'react'

import { Meta, Story } from '@storybook/react'
import { ParentSize } from '@visx/responsive'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Badge } from '../../../Badge'
import { Button } from '../../../Button'
import { H2, Text, Code } from '../../../Typography'
import { Series } from '../../types'

import { LineChart, LegendList, LegendItem, getLineColor } from './index'

const StoryConfig: Meta = {
    title: 'wildcard/Charts',
    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],
    parameters: {
        chromatic: { disableSnapshots: false, enableDarkMode: true },
    },
}

export default StoryConfig

export const LineChartsVitrina: Story = () => (
    <main className="d-flex flex-wrap" style={{ rowGap: 40, columnGap: 20 }}>
        <PlainChartExample />
        <PlainStackedChartExample />
        <WithLegendExample />
        <WithHugeDataExample />
        <WithZeroOneDataExample />
        <StackedWithDataMissingValues />
    </main>
)

interface StandardDatum {
    value: number
    x: Date | number
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
                x: new Date(2020, 5, 5),
                value: 4000,
                link: 'https://google.com/search',
            },
            {
                x: new Date(2020, 5, 6),
                value: 4000,
                link: 'https://google.com/search',
            },
            {
                x: new Date(2020, 5, 7),
                value: 5600,
                link: 'https://google.com/search',
            },
            {
                x: new Date(2020, 5, 8),
                value: 9800,
                link: 'https://google.com/search',
            },
            {
                x: new Date(2020, 5, 9),
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
                x: new Date(2020, 5, 5),
                value: 5000,
                link: 'https://twitter.com/search',
            },
            {
                x: new Date(2020, 5, 6),
                value: 5000,
                link: 'https://twitter.com/search',
            },
            {
                x: new Date(2020, 5, 7),
                value: 5000,
                link: 'https://twitter.com/search',
            },
            {
                x: new Date(2020, 5, 8),
                value: 5000,
                link: 'https://twitter.com/search',
            },
            {
                x: new Date(2020, 5, 9),
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
                x: new Date(2020, 5, 5),
                value: 15000,
                link: 'https://yandex.com/search',
            },
            {
                x: new Date(2020, 5, 6),
                value: 26000,
                link: 'https://yandex.com/search',
            },
            {
                x: new Date(2020, 5, 7),
                value: 20000,
                link: 'https://yandex.com/search',
            },
            {
                x: new Date(2020, 5, 8),
                value: 19000,
                link: 'https://yandex.com/search',
            },
            {
                x: new Date(2020, 5, 9),
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

const PlainChartExample = () => {
    const [active, setActive] = useState(false)

    return (
        <section style={{ flexBasis: 0 }}>
            <H2>Plain chart</H2>

            <Text>
                A standard example of the line chart. The static dataset has a fixed size (400 x 400). The
                active/focused line is pulled in front of other non-active lines. (synthetic z-index for SVG elements).
            </Text>

            <Button variant="primary" size="sm" className="mb-2" onClick={() => setActive(!active)}>
                Start Y axis with zero
            </Button>

            <LineChart width={400} height={400} zeroYAxisMin={active} series={STANDARD_SERIES} />
        </section>
    )
}

const PlainStackedChartExample = () => {
    const [active, setActive] = useState(false)

    return (
        <section style={{ flexBasis: 0 }}>
            <H2>Plain stacked chart</H2>

            <Text>
                <Badge variant="merged">Experimental</Badge> Stacked line chart. Each series value is calculated based
                on the previous series in the series array.
            </Text>

            <Button variant="primary" size="sm" className="mb-2" onClick={() => setActive(!active)}>
                Start Y axis with zero
            </Button>

            <LineChart stacked={true} width={400} height={400} series={STANDARD_SERIES} zeroYAxisMin={active} />
        </section>
    )
}

const WithLegendExample = () => (
    <section style={{ flexBasis: 0 }}>
        <H2>Line chart with legend</H2>

        <Text>
            The following chart layout (chart and legend blocks) has a fixed size of 400x400. Chart size is calculated
            (Chart = ParentSize - LegendSize).
        </Text>

        <div className="d-flex flex-column" style={{ width: 400, height: 400 }}>
            <ParentSize className="flex-1">
                {({ width, height }) => (
                    <LineChart<StandardDatum> width={width} height={height} series={STANDARD_SERIES} />
                )}
            </ParentSize>
            <LegendList className="mt-2">
                {STANDARD_SERIES.map(line => (
                    <LegendItem key={line.id} color={getLineColor(line)} name={line.name} />
                ))}
            </LegendList>
        </div>
    </section>
)

const WithHugeDataExample = () => {
    const SERIES: Series<StandardDatum>[] = [
        {
            id: 'series_001',
            data: [
                { x: new Date(2022, 1), value: 95_000 },
                { x: new Date(2022, 2), value: 125_000 },
                { x: new Date(2022, 3), value: 195_000 },
                { x: new Date(2022, 4), value: 235_000 },
                { x: new Date(2022, 5), value: 325_000 },
                { x: new Date(2022, 6), value: 400_000 },
                { x: new Date(2022, 7), value: 520_000 },
                { x: new Date(2022, 8), value: 720_000 },
                { x: new Date(2022, 9), value: 780_000 },
                { x: new Date(2022, 10), value: 800_000 },
                { x: new Date(2022, 11), value: 815_000 },
                { x: new Date(2022, 12), value: 840_000 },
            ],
            name: 'Fix',
            color: 'var(--oc-indigo-7)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_002',
            data: [
                { x: new Date(2022, 1), value: 92_000 },
                { x: new Date(2022, 2), value: 100_000 },
                { x: new Date(2022, 3), value: 106_000 },
                { x: new Date(2022, 4), value: 120_000 },
                { x: new Date(2022, 5), value: 130_000 },
                { x: new Date(2022, 6), value: 136_000 },
            ],
            color: 'var(--oc-orange-7)',
            name: 'Revert',
            getXValue,
            getYValue,
        },
    ]

    return (
        <section style={{ flexBasis: 0 }}>
            <H2>With unaligned by (x and y axes) data series</H2>

            <Text>
                It's a possible situation when some series has fewer points than others on the chart. In this case, we
                just show only existing points in the tooltip.
            </Text>

            <LineChart<StandardDatum> width={400} height={400} series={SERIES} />
        </section>
    )
}

const WithZeroOneDataExample = () => {
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
        <section style={{ flexBasis: 0 }}>
            <H2>Short datasets</H2>

            <Text>
                <Badge variant="warning">Should be improved</Badge> Currently, line charts and axis components may not
                handle short datasets properly. This is something that we want to improve in the future.
            </Text>

            <LineChart<StandardDatum> width={400} height={400} series={SERIES} />
        </section>
    )
}

const SERIES: Series<StandardDatum>[] = [
    {
        id: 'series_001',
        data: [
            { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 94 },
            { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, value: 134 },
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
            { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, value: 200 },
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

const StackedWithDataMissingValues = () => (
    <section style={{ flexBasis: 0 }}>
        <H2>Unaligned stacked datasets</H2>

        <Text>
            <Badge variant="merged">Experimental</Badge> In case some datasets are unaligned, the stacked value will
            take the closest points of previous data series, interpolate values and apply a synthetically calculated
            value as a basis for the current series point.
            <br />
            <Code>
                ---- X ---- <br />
                -----|----- <br />
                --S--D--E-- <br />X = D (interpolation between S and E) + X (value)
            </Code>
        </Text>

        <LineChart<StandardDatum> stacked={true} width={400} height={400} series={SERIES} />
    </section>
)
