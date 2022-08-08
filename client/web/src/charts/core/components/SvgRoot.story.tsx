import { boolean } from '@storybook/addon-knobs'
import { Meta, Story } from '@storybook/react'
import { AxisScale } from '@visx/axis/lib/types'
import { ParentSize } from '@visx/responsive'
import { scaleBand, scaleLinear, scaleTime } from '@visx/scale'

import { WebStory } from '../../../components/WebStory'

import { formatDateTick } from './axis/tick-formatters'
import { SvgRoot, SvgAxisLeft, SvgAxisBottom, SvgContent } from './SvgRoot'

const StoryConfig: Meta = {
    title: 'web/charts/core/axis',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
    parameters: { chromatic: { disableSnapshots: false } }
}

export default StoryConfig

interface TemplateProps {
    xScale: AxisScale
    pixelsPerXTick?: number
    formatXLabel?: (value: any) => string
    yScale: AxisScale
    color?: string
}

const SimpleChartTemplate: Story<TemplateProps> = args => (
    <ParentSize style={{ width: 400, height: 400 }} debounceTime={0} className="flex-shrink-0">
        {parent => (
            <SvgRoot width={parent.width} height={parent.height} xScale={args.xScale} yScale={args.yScale}>
                <SvgAxisLeft />
                <SvgAxisBottom tickFormat={args.formatXLabel} pixelsPerTick={args.pixelsPerXTick} />

                <SvgContent>
                    {({ content }) => <rect fill={args.color} width={content.width} height={content.height} />}
                </SvgContent>
            </SvgRoot>
        )}
    </ParentSize>
)

export const MainAxisDemo: Story = () => (
    <section style={{ display: 'flex', flexWrap: 'wrap', gap: 20 }}>
        <SimpleChartTemplate
            xScale={scaleTime<number>({
                domain: [new Date(2022, 8, 22), new Date(2022, 10, 22)],
                nice: true,
                clamp: true,
            })}
            yScale={scaleLinear({
                domain: [0, boolean('useMaxValuesForYScale', false) ? 1000000000000000000000000000000000000 : 10000],
                nice: true,
                clamp: true,
            })}
            formatXLabel={formatDateTick}
            pixelsPerXTick={20}
            color="darkslateblue"
        />

        <SimpleChartTemplate
            xScale={scaleBand<string>({
                domain: ['hello', 'worlddddddddddd', 'this', 'is', 'rotation', 'speaking'],
                padding: 0.2,
            })}
            yScale={scaleLinear({
                domain: [0, boolean('useMaxValuesForYScale', false) ? 1000000000000000000000000000000000000 : 10000],
                nice: true,
                clamp: true,
            })}
            color="pink"
        />
    </section>
)
