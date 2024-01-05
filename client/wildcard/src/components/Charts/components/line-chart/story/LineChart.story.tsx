import { useState } from 'react'

import type { Meta, StoryFn } from '@storybook/react'
import { ParentSize } from '@visx/responsive'
import { ResizableBox } from 'react-resizable'

import { BrandedStory } from '../../../../../stories/BrandedStory'
import { Badge } from '../../../../Badge'
import { Button } from '../../../../Button'
import { H2, Text, Code } from '../../../../Typography'
import type { Series } from '../../../types'
import { LineChart, LegendList, LegendItem, getLineColor } from '../index'

import {
    FLAT_SERIES,
    STANDARD_SERIES,
    SERIES_WITH_HUGE_DATA,
    UNALIGNED_SERIES,
    type StandardDatum,
    FLAT_XY_SERIES,
} from './mocks'

const StoryConfig: Meta = {
    title: 'wildcard/Charts',
    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],
    parameters: {
        chromatic: { disableSnapshots: false, enableDarkMode: true },
    },
}

export default StoryConfig

export const LineChartsDemo: StoryFn = () => (
    <main
        style={{
            display: 'flex',
            flexWrap: 'wrap',
            rowGap: 40,
            columnGap: 20,
            paddingBottom: 40,
        }}
    >
        <PlainChartExample />
        <FlatChartExample />
        <PlainStackedChartExample />
        <ResponsiveChartExample />
        <WithLegendExample />
        <WithHugeDataExample />
        <WithZeroOneDataExample />
        <StackedWithDataMissingValues />
    </main>
)

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
                Start Y axis at zero
            </Button>

            <LineChart width={400} height={400} zeroYAxisMin={active} series={FLAT_SERIES} />
        </section>
    )
}

const FlatChartExample = () => (
    <section style={{ flexBasis: 0 }}>
        <H2>Flat chart</H2>

        <Text>
            A standard example of the line chart but with fully flat datasets, try to navigate with arrows keyboard
            navigation.
        </Text>

        <LineChart width={400} height={400} series={FLAT_XY_SERIES} />
    </section>
)

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
                Start Y axis at zero
            </Button>

            <LineChart stacked={true} width={400} height={400} series={STANDARD_SERIES} zeroYAxisMin={active} />
        </section>
    )
}

const ResponsiveChartExample = () => (
    <section style={{ flexBasis: 0 }}>
        <H2>Responsive chart</H2>

        <Text style={{ maxWidth: 400, minWidth: 400 }}>
            SVG charts always require width and height values. In order to track parent size you can use ParentSize
            helper. Try to resize the following chart (drag any corner of the chart block).
            <br />
            <br />
            Note: Resize logic comes from react-resize package and not from this chart package.
        </Text>

        <ResizableBox width={400} height={400} axis="both" minConstraints={[200, 200]} className="p-3">
            <ParentSize debounceTime={0}>
                {parent => <LineChart width={parent.width} height={parent.height} series={STANDARD_SERIES} />}
            </ParentSize>
        </ResizableBox>
    </section>
)

const WithLegendExample = () => (
    <section style={{ flexBasis: 0 }}>
        <H2>Line chart with legend</H2>

        <Text>
            The following chart layout (chart and legend blocks) has a fixed size of 400x400. Chart size is calculated
            (Chart = ParentSize - LegendSize).
        </Text>

        <div className="d-flex flex-column" style={{ width: 400, height: 400 }}>
            <ParentSize className="flex-1">
                {({ width, height }) => <LineChart width={width} height={height} series={STANDARD_SERIES} />}
            </ParentSize>
            <LegendList className="mt-2">
                {STANDARD_SERIES.map(line => (
                    <LegendItem key={line.id} color={getLineColor(line)} name={line.name} />
                ))}
            </LegendList>
        </div>
    </section>
)

const WithHugeDataExample = () => (
    <section style={{ flexBasis: 0 }}>
        <H2>With unaligned by (x and y axes) data series</H2>

        <Text>
            It's a possible situation when some series has fewer points than others on the chart. In this case, we just
            show only existing points in the tooltip.
        </Text>

        <LineChart width={400} height={400} series={SERIES_WITH_HUGE_DATA} />
    </section>
)

const WithZeroOneDataExample = () => {
    const SERIES: Series<StandardDatum>[] = [
        {
            id: 'series_001',
            data: [
                { x: new Date(2022, 6, 1), value: 0 },
                { x: new Date(2022, 6, 3), value: 5 },
            ],
            name: 'A metric',
            color: 'var(--blue)',
            getXValue: datum => new Date(datum.x),
            getYValue: datum => datum.value,
        },
    ]

    return (
        <section style={{ flexBasis: 0 }}>
            <H2>Short datasets</H2>

            <Text>
                <Badge variant="warning">Should be improved</Badge> Currently, line charts and axis components may not
                handle short datasets properly. This is something that we want to improve in the future.
            </Text>

            <LineChart width={400} height={400} series={SERIES} />
        </section>
    )
}

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

        <LineChart stacked={true} width={400} height={400} series={UNALIGNED_SERIES} />
    </section>
)
