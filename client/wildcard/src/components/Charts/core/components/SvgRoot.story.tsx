import { Meta, Story } from '@storybook/react'
import { AxisScale } from '@visx/axis/lib/types'
import { ParentSize } from '@visx/responsive'
import { scaleBand, scaleLinear, scaleTime } from '@visx/scale'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { formatDateTick } from './axis/tick-formatters'
import { SvgRoot, SvgAxisLeft, SvgAxisBottom, SvgContent } from './SvgRoot'

const StoryConfig: Meta = {
    title: 'wildcard/Charts/Core',
    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],
    argTypes: {
        useMaxValuesForYScale: {
            type: 'boolean',
            defaultValue: false,
        },
    },
    parameters: { chromatic: { disableSnapshots: false } },
}

export default StoryConfig

interface TemplateProps {
    xScale: AxisScale
    yScale: AxisScale
    pixelsPerXTick?: number
    formatXLabel?: (value: any) => string
    color?: string
}

const SimpleChartTemplate: Story<TemplateProps> = args => (
    <ParentSize style={{ width: 400, height: 400 }} debounceTime={0} className="flex-shrink-0">
        {parent => (
            <SvgRoot width={parent.width} height={parent.height} xScale={args.xScale} yScale={args.yScale}>
                <SvgAxisLeft />
                <SvgAxisBottom tickFormat={args.formatXLabel} pixelsPerTick={args.pixelsPerXTick} maxRotateAngle={90} />

                <SvgContent>
                    {({ content }) => <rect fill={args.color} width={content.width} height={content.height} />}
                </SvgContent>
            </SvgRoot>
        )}
    </ParentSize>
)

export const SmartAxisDemo: Story = args => (
    <section style={{ display: 'flex', flexWrap: 'wrap', gap: 20 }}>
        <SimpleChartTemplate
            xScale={scaleTime<number>({
                domain: [new Date(2022, 8, 22), new Date(2022, 10, 22)],
                nice: true,
                clamp: true,
            })}
            yScale={scaleLinear({
                domain: [0, args.useMaxValuesForYScale ? 1000000000000000000000000000000000000 : 10000],
                nice: true,
                clamp: true,
            })}
            formatXLabel={formatDateTick}
            pixelsPerXTick={30}
            color="darkslateblue"
        />

        <SimpleChartTemplate
            xScale={scaleBand<string>({
                domain: ['hello', 'worlddddddddddd', 'this', 'is', 'rotation', 'speaking'],
                padding: 0.2,
            })}
            yScale={scaleLinear({
                domain: [0, args.useMaxValuesForYScale ? 1000000000000000000000000000000000000 : 10000],
                nice: true,
                clamp: true,
            })}
            color="pink"
        />
    </section>
)
